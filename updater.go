package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	updateRepoOwner = "vpertj"
	updateRepoName  = "Weditor"

	// 开发通道：prerelease 无法通过 releases/latest 拿到，必须查 API 定位最新版本
	devReleasesAPI = "https://api.github.com/repos/" + updateRepoOwner + "/" + updateRepoName + "/releases?per_page=10"

	updateManifestAssetName = "latest.json"

	updateChannelStable = "stable"
	updateChannelDev    = "dev"

	// 清单/API 这类小请求的短超时；安装包体量大，另有独立的长超时客户端
	metaHTTPTimeout = 15 * time.Second
	// 清单再大也不会超过这个量级，防止异常响应把内存吃满
	maxMetaBytes = 1 << 20
)

// ==========================================
// 对外数据结构
// ==========================================

// UpdateAsset 清单里单个平台的安装包
type UpdateAsset struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
	Sha256 string `json:"sha256"`
}

// UpdateManifest 由 CI 在发布时生成，随 Release 一起上传为 latest.json
type UpdateManifest struct {
	Version     string                 `json:"version"`
	Channel     string                 `json:"channel"`
	PublishedAt string                 `json:"publishedAt"`
	Notes       string                 `json:"notes"`
	Assets      map[string]UpdateAsset `json:"assets"`
}

// UpdateInfo 一次检查的结果，直接交给前端展示
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Channel        string `json:"channel"`
	PublishedAt    string `json:"publishedAt"`
	Notes          string `json:"notes"`
	AssetName      string `json:"assetName"`
	AssetSize      int64  `json:"assetSize"`
	AssetURL       string `json:"assetUrl"`
	Sha256         string `json:"sha256"`
	Reason         string `json:"reason"` // 不提供更新的原因（已最新/已跳过/平台不支持…）
	Error          string `json:"error"`  // 检查本身失败的原因
}

// UpdateSettings 持久化到 update.json 的更新偏好
type UpdateSettings struct {
	Channel        string `json:"channel"`
	AutoCheck      bool   `json:"autoCheck"`
	SkippedVersion string `json:"skippedVersion"`
	LastCheckAt    string `json:"lastCheckAt"`
}

// UpdateResult 下载结果
type UpdateResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FilePath string `json:"filePath"`
}

// UpdateProgress 通过 update-progress 事件推送给前端
type UpdateProgress struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
	Speed      int64   `json:"speed"` // 字节/秒
}

// stagedPackage 已下载并校验通过、等待安装的安装包
type stagedPackage struct {
	Path    string
	Version string
	Sha256  string
}

// 更新流程状态。Wails 应用只有一个 App 实例，且下载/安装必须全局串行，
// 因此用包级状态而不是塞进 App 结构体。
var (
	updateMu     sync.Mutex
	stagedUpdate *stagedPackage
)

// ==========================================
// 设置持久化
// ==========================================

func defaultUpdateSettings() UpdateSettings {
	// 默认稳定通道 + 自动检查：普通用户不该被 dev 构建打扰
	return UpdateSettings{Channel: updateChannelStable, AutoCheck: true}
}

func normalizeUpdateSettings(s UpdateSettings) UpdateSettings {
	if s.Channel != updateChannelDev {
		s.Channel = updateChannelStable
	}
	s.SkippedVersion = strings.TrimSpace(s.SkippedVersion)
	return s
}

func (a *App) updateSettingsPath() string {
	return filepath.Join(a.appDataDir, "update.json")
}

// GetUpdateSettings 读取更新设置，文件不存在或损坏时返回默认值
func (a *App) GetUpdateSettings() UpdateSettings {
	s := defaultUpdateSettings()
	if data, err := os.ReadFile(a.updateSettingsPath()); err == nil {
		// 反序列化只覆盖 JSON 中出现的字段，未出现的保留默认值
		json.Unmarshal(data, &s)
	}
	return normalizeUpdateSettings(s)
}

// SetUpdateSettings 保存更新设置并返回规范化后的结果
func (a *App) SetUpdateSettings(s UpdateSettings) UpdateSettings {
	s = normalizeUpdateSettings(s)
	if data, err := json.MarshalIndent(s, "", "  "); err == nil {
		os.WriteFile(a.updateSettingsPath(), data, 0644)
	}
	return s
}

// SkipVersion 记录用户选择跳过的版本，之后不再提示该版本
func (a *App) SkipVersion(version string) UpdateSettings {
	s := a.GetUpdateSettings()
	s.SkippedVersion = version
	return a.SetUpdateSettings(s)
}

// ==========================================
// HTTP 辅助
// ==========================================

type etagEntry struct {
	etag string
	body []byte
}

// 进程内的 ETag 缓存：用户在设置里反复点「检查更新」时，
// GitHub 返回 304 不再消耗 API 配额。
var metaCache = struct {
	sync.Mutex
	entries map[string]etagEntry
}{entries: map[string]etagEntry{}}

var metaClient = &http.Client{Timeout: metaHTTPTimeout}

func userAgent() string {
	return fmt.Sprintf("%s/%s (+https://github.com/%s/%s)", APP_NAME, APP_VERSION, updateRepoOwner, updateRepoName)
}

// httpGetJSON 带 ETag 缓存的 JSON GET
func httpGetJSON(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent())
	if strings.Contains(url, "api.github.com") {
		// 只对 API 声明响应类型；release 资产下载走 CDN，带上反而多一层不确定性
		req.Header.Set("Accept", "application/vnd.github+json")
	}

	metaCache.Lock()
	cached, hasCache := metaCache.entries[url]
	metaCache.Unlock()
	if hasCache && cached.etag != "" {
		req.Header.Set("If-None-Match", cached.etag)
	}

	resp, err := metaClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body []byte
	switch resp.StatusCode {
	case http.StatusNotModified:
		if !hasCache {
			return errors.New("服务端返回 304 但没有本地缓存")
		}
		body = cached.body
	case http.StatusOK:
		body, err = io.ReadAll(io.LimitReader(resp.Body, maxMetaBytes))
		if err != nil {
			return err
		}
		if etag := resp.Header.Get("Etag"); etag != "" {
			metaCache.Lock()
			metaCache.entries[url] = etagEntry{etag: etag, body: body}
			metaCache.Unlock()
		}
	default:
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("响应解析失败: %w", err)
	}
	return nil
}

// 更新源地址。用变量而不是常量，是为了让集成测试能把它们指向本地假清单，
// 从而在不发布真实 Release 的前提下验证完整的检查+下载链路。
var (
	// 稳定通道：GitHub 的静态重定向地址，不消耗 API 配额，
	// 且自动跳过 prerelease（CI 在分支上发布的 dev 版本都是 prerelease）。
	stableManifestURL = "https://github.com/" + updateRepoOwner + "/" + updateRepoName + "/releases/latest/download/latest.json"
)

// ==========================================
// 清单获取
// ==========================================

// ghRelease GitHub Releases API 的响应子集
type ghRelease struct {
	TagName     string `json:"tag_name"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// pickDevRelease 从发布列表里挑出发布时间最新的 prerelease。
//
// 用时间而不是版本号排序：dev 版本形如 v1.0.0-dev-<git sha>，sha 之间
// 没有大小关系（见 compareSemver）。GitHub 返回的都是 UTC 的 RFC3339，
// 同格式下字典序即时间序。
func pickDevRelease(releases []ghRelease) (ghRelease, bool) {
	var best ghRelease
	found := false
	for _, r := range releases {
		if r.Draft || !r.Prerelease {
			continue
		}
		if !found || r.PublishedAt > best.PublishedAt {
			best, found = r, true
		}
	}
	return best, found
}

// fetchDevManifest 通过 API 找到最新的 prerelease，再读它附带的 latest.json
func fetchDevManifest() (UpdateManifest, error) {
	var releases []ghRelease
	if err := httpGetJSON(devReleasesAPI, &releases); err != nil {
		return UpdateManifest{}, err
	}
	rel, ok := pickDevRelease(releases)
	if !ok {
		return UpdateManifest{}, errors.New("没有找到开发构建")
	}
	for _, asset := range rel.Assets {
		if asset.Name == updateManifestAssetName {
			var m UpdateManifest
			if err := httpGetJSON(asset.BrowserDownloadURL, &m); err != nil {
				return UpdateManifest{}, err
			}
			return m, nil
		}
	}
	return UpdateManifest{}, fmt.Errorf("开发构建 %s 缺少 %s", rel.TagName, updateManifestAssetName)
}

// fetchManifest 按通道获取更新清单
func fetchManifest(channel string) (UpdateManifest, error) {
	if channel == updateChannelDev {
		return fetchDevManifest()
	}
	var m UpdateManifest
	if err := httpGetJSON(stableManifestURL, &m); err != nil {
		return UpdateManifest{}, err
	}
	return m, nil
}

// ==========================================
// 平台资产匹配
// ==========================================

// assetKeyCandidates 返回当前平台在清单中可能使用的键，按优先级排列。
// CI 目前只产出 darwin-universal / windows-amd64 / linux-amd64，
// 保留架构粒度是为了将来发布分架构产物时不用改更新器。
func assetKeyCandidates() []string {
	keys := []string{runtime.GOOS + "-" + runtime.GOARCH}
	if runtime.GOOS == "darwin" {
		keys = append(keys, "darwin-universal")
	}
	return keys
}

// selectAsset 挑出当前平台对应的安装包
func selectAsset(m UpdateManifest) (UpdateAsset, string, bool) {
	for _, key := range assetKeyCandidates() {
		if asset, ok := m.Assets[key]; ok && asset.URL != "" {
			return asset, key, true
		}
	}
	return UpdateAsset{}, "", false
}

// ==========================================
// 检查更新
// ==========================================

// publishedAfter 判断发布时间是否晚于构建时间；任一无法解析时返回 false。
// 开发通道用它判定新旧，因为 dev 版本号之间无法定序。
func publishedAfter(publishedAt, buildTime string) bool {
	p, err1 := time.Parse(time.RFC3339, publishedAt)
	b, err2 := time.Parse(time.RFC3339, buildTime)
	if err1 != nil || err2 != nil {
		return false
	}
	return p.After(b)
}

// CheckForUpdate 检查更新。manual=true 表示用户主动触发，
// 此时前端需要展示失败原因；自动检查失败则静默。
func (a *App) CheckForUpdate(manual bool) UpdateInfo {
	return a.checkForUpdate(manual)
}

func (a *App) checkForUpdate(manual bool) UpdateInfo {
	settings := a.GetUpdateSettings()
	info := UpdateInfo{
		CurrentVersion: APP_VERSION,
		Channel:        settings.Channel,
	}

	manifest, err := fetchManifest(settings.Channel)
	if err != nil {
		info.Error = fmt.Sprintf("检查更新失败：%v", err)
		return info
	}

	settings.LastCheckAt = time.Now().Format(time.RFC3339)
	a.SetUpdateSettings(settings)

	info.LatestVersion = manifest.Version
	info.PublishedAt = manifest.PublishedAt
	info.Notes = manifest.Notes

	if manifest.Version == "" {
		info.Error = "更新清单缺少版本号"
		return info
	}

	if settings.Channel == updateChannelDev {
		if APP_BUILD_TIME == "" {
			info.Reason = "当前是本地构建（无构建时间），无法判断开发版本新旧"
			return info
		}
		if !publishedAfter(manifest.PublishedAt, APP_BUILD_TIME) {
			info.Reason = "已是最新版本"
			return info
		}
	} else if !newerThan(manifest.Version, APP_VERSION) {
		info.Reason = "已是最新版本"
		return info
	}

	if manifest.Version == settings.SkippedVersion {
		info.Reason = fmt.Sprintf("已跳过版本 %s", manifest.Version)
		return info
	}

	asset, _, ok := selectAsset(manifest)
	if !ok {
		info.Reason = fmt.Sprintf("新版本 %s 没有提供 %s/%s 的安装包", manifest.Version, runtime.GOOS, runtime.GOARCH)
		return info
	}

	info.Available = true
	info.AssetName = asset.Name
	info.AssetSize = asset.Size
	info.AssetURL = asset.URL
	info.Sha256 = asset.Sha256
	return info
}

// ==========================================
// 下载
// ==========================================

func (a *App) updatesDir() string {
	return filepath.Join(a.appDataDir, "updates")
}

// DownloadUpdate 下载当前平台的新版本安装包，进度通过 update-progress 事件推送
func (a *App) DownloadUpdate() UpdateResult {
	updateMu.Lock()
	defer updateMu.Unlock()

	info := a.checkForUpdate(true)
	if info.Error != "" {
		return UpdateResult{Message: info.Error}
	}
	if !info.Available {
		return UpdateResult{Message: orDefault(info.Reason, "当前已是最新版本")}
	}

	// 同一版本的包已经下好并校验通过，直接复用
	if stagedUpdate != nil && stagedUpdate.Version == info.LatestVersion && fileExists(stagedUpdate.Path) {
		return UpdateResult{Success: true, FilePath: stagedUpdate.Path, Message: "安装包已就绪"}
	}

	if err := os.MkdirAll(a.updatesDir(), 0755); err != nil {
		return UpdateResult{Message: fmt.Sprintf("创建更新目录失败：%v", err)}
	}

	dest := filepath.Join(a.updatesDir(), info.AssetName)
	part := dest + ".part"

	if err := a.downloadFile(info.AssetURL, part, info.AssetSize, info.Sha256); err != nil {
		os.Remove(part)
		return UpdateResult{Message: err.Error()}
	}
	if err := os.Rename(part, dest); err != nil {
		os.Remove(part)
		return UpdateResult{Message: fmt.Sprintf("保存安装包失败：%v", err)}
	}

	stagedUpdate = &stagedPackage{Path: dest, Version: info.LatestVersion, Sha256: info.Sha256}
	return UpdateResult{Success: true, FilePath: dest, Message: "下载完成"}
}

// downloadFile 流式下载并实时计算 sha256，未通过校验的文件不会落到最终路径
func (a *App) downloadFile(url, dest string, expectSize int64, expectSha string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent())

	// 安装包几十 MB，不能用清单请求那套 15s 超时
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败：%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败：HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("创建文件失败：%v", err)
	}
	defer f.Close()

	total := resp.ContentLength
	if total <= 0 {
		total = expectSize
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)
	buf := make([]byte, 64*1024)
	var downloaded int64
	started := time.Now()
	lastEmit := time.Time{}

	emit := func() {
		// 无 ctx（单元测试或未启动完成）时跳过事件推送
		if a.ctx == nil {
			return
		}
		progress := UpdateProgress{Downloaded: downloaded, Total: total}
		if total > 0 {
			progress.Percent = float64(downloaded) / float64(total) * 100
		}
		if elapsed := time.Since(started).Seconds(); elapsed > 0 {
			progress.Speed = int64(float64(downloaded) / elapsed)
		}
		wailsRuntime.EventsEmit(a.ctx, "update-progress", progress)
	}

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := writer.Write(buf[:n]); err != nil {
				return fmt.Errorf("写入文件失败：%v", err)
			}
			downloaded += int64(n)
			if time.Since(lastEmit) > 200*time.Millisecond {
				emit()
				lastEmit = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("下载中断：%v", readErr)
		}
	}
	emit()

	if err := f.Sync(); err != nil {
		return fmt.Errorf("写入文件失败：%v", err)
	}

	if expectSha != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(got, expectSha) {
			return errors.New("安装包校验失败（sha256 不匹配），已丢弃本次下载")
		}
	}
	return nil
}

// ==========================================
// 安装与重启
// ==========================================

// ApplyUpdateAndRestart 安装已下载的更新。成功后由 Go 侧延迟退出进程，
// 让前端的「正在重启」提示有机会渲染出来。
func (a *App) ApplyUpdateAndRestart() error {
	updateMu.Lock()
	defer updateMu.Unlock()

	if stagedUpdate == nil {
		return errors.New("没有已下载的更新包")
	}
	staged := *stagedUpdate

	// 落地前再校验一次：下载到安装之间文件可能被改动或截断
	if staged.Sha256 != "" {
		if err := verifyFileSha256(staged.Path, staged.Sha256); err != nil {
			return err
		}
	}

	if err := a.applyUpdatePlatform(staged.Path, staged.Version); err != nil {
		return err
	}

	stagedUpdate = nil
	time.AfterFunc(800*time.Millisecond, func() {
		if a.ctx != nil {
			wailsRuntime.Quit(a.ctx)
		}
	})
	return nil
}

// RevealUpdateFile 在系统文件管理器中定位已下载的安装包，返回其路径
func (a *App) RevealUpdateFile() string {
	if stagedUpdate == nil {
		return ""
	}
	return revealInFileManager(stagedUpdate.Path)
}

// OpenUpdateDownloadPage 打开 Release 页面，用于无法自动更新时的降级路径
func (a *App) OpenUpdateDownloadPage() {
	if a.ctx == nil {
		return
	}
	wailsRuntime.BrowserOpenURL(a.ctx, fmt.Sprintf("https://github.com/%s/%s/releases/latest", updateRepoOwner, updateRepoName))
}

// ==========================================
// 辅助函数
// ==========================================

// cleanupUpdatesDir 清理历史残留：未完成的下载片段立即删除，
// 备份与旧安装包保留 24 小时（留出问题排查和回滚的余地）后清理。
func (a *App) cleanupUpdatesDir() {
	dir := a.updatesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if strings.HasSuffix(entry.Name(), ".part") {
			os.RemoveAll(path)
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > 24*time.Hour {
			os.RemoveAll(path)
		}
	}
}

// looksLikeDevBuild 判断可执行文件是否来自 git 工作区内的开发构建。
// 自动替换会覆盖开发产物（build/bin 下的构建），这类构建只提示不自动安装。
func looksLikeDevBuild(exePath string) bool {
	dir := filepath.Dir(exePath)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return false
}

// ensureWritable 通过实际创建临时文件来验证目录可写。
// 不用权限位推断：macOS 的 ACL 与 POSIX 权限位可能不一致。
func ensureWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".weditor-write-test-")
	if err != nil {
		return err
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return nil
}

// verifyFileSha256 校验文件摘要
func verifyFileSha256(path, want string) error {
	got, err := sha256File(path)
	if err != nil {
		return fmt.Errorf("读取安装包失败：%v", err)
	}
	if !strings.EqualFold(got, want) {
		return errors.New("安装包校验失败（sha256 不匹配）")
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile 复制文件并保留可执行权限位
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// revealInFileManager 在系统文件管理器中选中文件，返回文件路径
func revealInFileManager(path string) string {
	if path == "" {
		return ""
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "windows":
		cmd = exec.Command("explorer", "/select,"+path)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	}
	if err := cmd.Start(); err != nil {
		return ""
	}
	return path
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
