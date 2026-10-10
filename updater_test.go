package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPickDevRelease(t *testing.T) {
	releases := []ghRelease{
		{TagName: "v1.0.0", Prerelease: false, PublishedAt: "2026-08-10T00:00:00Z"},
		{TagName: "v1.0.0-dev-old", Prerelease: true, PublishedAt: "2026-08-01T00:00:00Z"},
		{TagName: "v1.0.0-dev-new", Prerelease: true, PublishedAt: "2026-08-12T10:00:00Z"},
		{TagName: "v1.0.0-dev-draft", Prerelease: true, Draft: true, PublishedAt: "2026-08-13T00:00:00Z"},
	}
	got, ok := pickDevRelease(releases)
	if !ok {
		t.Fatal("应挑出 prerelease")
	}
	if got.TagName != "v1.0.0-dev-new" {
		t.Errorf("挑出的是 %q，期望最新发布时间的 v1.0.0-dev-new", got.TagName)
	}

	if _, ok := pickDevRelease([]ghRelease{{TagName: "v1.0.0", PublishedAt: "x"}}); ok {
		t.Error("没有 prerelease 时不应返回结果")
	}
}

func TestSelectAsset(t *testing.T) {
	exact := runtime.GOOS + "-" + runtime.GOARCH

	full := UpdateManifest{Assets: map[string]UpdateAsset{
		exact:              {Name: "exact", URL: "https://example.com/exact"},
		"darwin-universal": {Name: "universal", URL: "https://example.com/universal"},
	}}
	asset, key, ok := selectAsset(full)
	if !ok || key != exact || asset.Name != "exact" {
		t.Errorf("应优先精确匹配当前架构，got ok=%v key=%q name=%q", ok, key, asset.Name)
	}

	// 只有 universal 包时，macOS 应能回退使用；其他平台不应误用
	universalOnly := UpdateManifest{Assets: map[string]UpdateAsset{
		"darwin-universal": {Name: "universal", URL: "https://example.com/universal"},
	}}
	_, key2, ok2 := selectAsset(universalOnly)
	if runtime.GOOS == "darwin" {
		if !ok2 || key2 != "darwin-universal" {
			t.Errorf("macOS 应回退到 darwin-universal，got ok=%v key=%q", ok2, key2)
		}
	} else if ok2 {
		t.Errorf("%s 不应匹配 darwin-universal 包", runtime.GOOS)
	}

	// URL 为空的条目视为缺失
	if _, _, ok3 := selectAsset(UpdateManifest{Assets: map[string]UpdateAsset{exact: {Name: "no-url"}}}); ok3 {
		t.Error("URL 为空的资产不应被选中")
	}
}

func TestPublishedAfter(t *testing.T) {
	cases := []struct {
		published, built string
		want             bool
	}{
		{"2026-08-12T10:00:00Z", "2026-08-12T09:00:00Z", true},
		{"2026-08-12T09:00:00Z", "2026-08-12T10:00:00Z", false},
		{"2026-08-12T10:00:00Z", "2026-08-12T10:00:00Z", false},
		// 无法解析时一律判定为"不更新"，避免误触发
		{"", "2026-08-12T10:00:00Z", false},
		{"2026-08-12T10:00:00Z", "", false},
	}
	for _, c := range cases {
		if got := publishedAfter(c.published, c.built); got != c.want {
			t.Errorf("publishedAfter(%q, %q) = %v, 期望 %v", c.published, c.built, got, c.want)
		}
	}
}

func TestLooksLikeDevBuild(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(repo, "build", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if !looksLikeDevBuild(filepath.Join(binDir, "WEditor")) {
		t.Error("git 工作区内的构建应被识别为开发构建")
	}

	outside := t.TempDir()
	if looksLikeDevBuild(filepath.Join(outside, "WEditor")) {
		t.Error("非 git 工作区不应被识别为开发构建")
	}
}

func TestHTTPGetJSONETag(t *testing.T) {
	var full, notModified int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"v1"` {
			notModified++
			w.WriteHeader(http.StatusNotModified)
			return
		}
		full++
		w.Header().Set("Etag", `"v1"`)
		io.WriteString(w, `{"version":"v1.2.3","assets":{"darwin-universal":{"name":"a","url":"u"}}}`)
	}))
	defer srv.Close()

	var first UpdateManifest
	if err := httpGetJSON(srv.URL, &first); err != nil {
		t.Fatalf("首次请求失败: %v", err)
	}
	if first.Version != "v1.2.3" {
		t.Errorf("版本 = %q, 期望 v1.2.3", first.Version)
	}

	var second UpdateManifest
	if err := httpGetJSON(srv.URL, &second); err != nil {
		t.Fatalf("二次请求失败: %v", err)
	}
	if second.Version != "v1.2.3" {
		t.Error("304 时应回落到本地缓存的内容")
	}
	if full != 1 || notModified != 1 {
		t.Errorf("完整响应 %d 次 / 304 %d 次，期望各 1 次", full, notModified)
	}
}

func TestDownloadFileVerifiesSha256(t *testing.T) {
	payload := []byte("fake installer payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	sum := sha256.Sum256(payload)
	dir := t.TempDir()
	app := &App{appDataDir: dir}

	dest := filepath.Join(dir, "pkg.bin")
	if err := app.downloadFile(srv.URL, dest, int64(len(payload)), hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("下载应成功: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != string(payload) {
		t.Errorf("落盘内容不匹配: %v", err)
	}

	// 摘要不匹配必须报错，让调用方删除文件而不是继续安装
	if err := app.downloadFile(srv.URL, filepath.Join(dir, "bad.bin"), 0, "deadbeef"); err == nil {
		t.Error("sha256 不匹配时应返回错误")
	}
}

func TestManifestContract(t *testing.T) {
	// 与 CI 中 "Generate update manifest" 步骤产出的结构保持一致
	const raw = `{
  "version": "v1.4.2",
  "channel": "stable",
  "publishedAt": "2026-10-10T07:40:18Z",
  "notes": "Release v1.4.2",
  "assets": {
    "darwin-universal": {
      "name": "WEditor-macos.app.zip",
      "url": "https://github.com/vpertj/Weditor/releases/download/v1.4.2/WEditor-macos.app.zip",
      "size": 200000,
      "sha256": "ee15d14a1e0db3c71cd3e01ade5ff08bc1353e9eea0354b3a17dae19fcecefab"
    },
    "windows-amd64": {
      "name": "WEditor-windows.exe",
      "url": "https://github.com/vpertj/Weditor/releases/download/v1.4.2/WEditor-windows.exe",
      "size": 120000,
      "sha256": "869fdd33a0e48b28a25ed707ddae77b6e2edd1cff8f066b4c0e85234eaf48f5b"
    }
  }
}`

	var m UpdateManifest
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("解析 CI 产出的清单失败: %v", err)
	}
	if m.Version != "v1.4.2" || m.Channel != updateChannelStable {
		t.Errorf("清单字段解析错误: %+v", m)
	}
	if !newerThan(m.Version, "v1.0.0") {
		t.Error("v1.4.2 应被判定为比 v1.0.0 新")
	}
	if !publishedAfter(m.PublishedAt, "2026-10-01T00:00:00Z") {
		t.Error("publishedAt 应能被解析用于开发通道比较")
	}

	// 当前平台上必须能选中一个带校验和的资产
	asset, key, ok := selectAsset(m)
	if !ok {
		t.Fatalf("清单中没有 %s/%s 可用资产", runtime.GOOS, runtime.GOARCH)
	}
	if asset.URL == "" || asset.Sha256 == "" || asset.Name == "" {
		t.Errorf("资产字段不完整 (%s): %+v", key, asset)
	}
}

func TestUpdateSettingsRoundTrip(t *testing.T) {
	app := &App{appDataDir: t.TempDir()}

	def := app.GetUpdateSettings()
	if def.Channel != updateChannelStable || !def.AutoCheck {
		t.Errorf("默认值应为稳定通道 + 自动检查，实际 %+v", def)
	}

	saved := app.SetUpdateSettings(UpdateSettings{Channel: updateChannelDev, SkippedVersion: " v1.0.0 "})
	if saved.SkippedVersion != "v1.0.0" {
		t.Errorf("跳过版本应被 trim，实际 %q", saved.SkippedVersion)
	}
	if saved.AutoCheck {
		t.Error("显式关掉自动检查后不应被默认值覆盖")
	}

	got := app.GetUpdateSettings()
	if got.Channel != updateChannelDev || got.AutoCheck || got.SkippedVersion != "v1.0.0" {
		t.Errorf("读取结果与写入不一致: %+v", got)
	}

	// 非法通道回落到稳定通道
	if app.SetUpdateSettings(UpdateSettings{Channel: "nightly"}).Channel != updateChannelStable {
		t.Error("非法通道应回落到稳定通道")
	}
}

func TestSkipVersion(t *testing.T) {
	app := &App{appDataDir: t.TempDir()}
	if got := app.SkipVersion("v9.9.9"); got.SkippedVersion != "v9.9.9" {
		t.Errorf("跳过版本 = %q, 期望 v9.9.9", got.SkippedVersion)
	}
	if app.GetUpdateSettings().SkippedVersion != "v9.9.9" {
		t.Error("跳过版本应被持久化")
	}
}

// TestUpdateCheckAndDownloadEndToEnd 用本地假清单跑完整的"检查 → 下载 → 校验"链路。
// 不触碰真实 Release，也不需要联网。
func TestUpdateCheckAndDownloadEndToEnd(t *testing.T) {
	payload := []byte("WEditor fake installer payload v1.5.0")
	sum := sha256.Sum256(payload)
	sha := hex.EncodeToString(sum[:])

	var assetHits int
	assetSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assetHits++
		w.Write(payload)
	}))
	defer assetSrv.Close()

	manifest := UpdateManifest{
		Version:     "v1.5.0",
		Channel:     updateChannelStable,
		PublishedAt: "2026-10-10T07:40:18Z",
		Notes:       "Release v1.5.0",
		Assets: map[string]UpdateAsset{
			runtime.GOOS + "-" + runtime.GOARCH: {
				Name:   "WEditor-fake.bin",
				URL:    assetSrv.URL,
				Size:   int64(len(payload)),
				Sha256: sha,
			},
		},
	}
	manifestSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(manifest)
	}))
	defer manifestSrv.Close()

	// 只在本测试内把更新源指向本地服务
	originalURL := stableManifestURL
	stableManifestURL = manifestSrv.URL
	defer func() { stableManifestURL = originalURL }()

	originalVersion := APP_VERSION
	APP_VERSION = "v1.0.0"
	defer func() { APP_VERSION = originalVersion }()

	// 每个用例都用独立的 App 实例与数据目录，避免相互污染
	app := &App{appDataDir: t.TempDir()}

	info := app.CheckForUpdate(true)
	if info.Error != "" {
		t.Fatalf("检查更新失败: %s", info.Error)
	}
	if !info.Available {
		t.Fatalf("v1.0.0 应当检测到 v1.5.0，实际 reason=%q", info.Reason)
	}
	if info.LatestVersion != "v1.5.0" || info.Sha256 != sha || info.AssetName != "WEditor-fake.bin" {
		t.Errorf("检查结果字段不正确: %+v", info)
	}

	result := app.DownloadUpdate()
	if !result.Success {
		t.Fatalf("下载失败: %s", result.Message)
	}
	got, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("读取下载结果失败: %v", err)
	}
	if string(got) != string(payload) {
		t.Error("下载内容与源文件不一致")
	}
	if stagedUpdate == nil || stagedUpdate.Version != "v1.5.0" {
		t.Errorf("下载完成后应记录待安装版本，实际 %+v", stagedUpdate)
	}

	// 同一版本再次下载应复用已校验的包，不重复请求
	before := assetHits
	if second := app.DownloadUpdate(); !second.Success {
		t.Fatalf("重复下载应成功: %s", second.Message)
	}
	if assetHits != before {
		t.Errorf("同一版本不应重复下载，资产请求次数 %d → %d", before, assetHits)
	}

	// 已是最新版本时不应提示更新
	APP_VERSION = "v1.5.0"
	stagedUpdate = nil
	if latest := app.CheckForUpdate(true); latest.Available {
		t.Errorf("v1.5.0 不应再提示更新: %+v", latest)
	}

	// 跳过的版本不再提示
	APP_VERSION = "v1.0.0"
	app.SkipVersion("v1.5.0")
	if skipped := app.CheckForUpdate(true); skipped.Available || skipped.Reason == "" {
		t.Errorf("跳过的版本不应提示更新，实际 %+v", skipped)
	}
}

// TestDownloadUpdateRejectsTamperedPackage 校验和不匹配时必须拒绝安装
func TestDownloadUpdateRejectsTamperedPackage(t *testing.T) {
	assetSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "tampered")
	}))
	defer assetSrv.Close()

	manifestSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UpdateManifest{
			Version: "v2.0.0",
			Assets: map[string]UpdateAsset{
				runtime.GOOS + "-" + runtime.GOARCH: {
					Name:   "WEditor-fake.bin",
					URL:    assetSrv.URL,
					Sha256: "0000000000000000000000000000000000000000000000000000000000000000",
				},
			},
		})
	}))
	defer manifestSrv.Close()

	originalURL := stableManifestURL
	stableManifestURL = manifestSrv.URL
	defer func() { stableManifestURL = originalURL }()

	app := &App{appDataDir: t.TempDir()}
	result := app.DownloadUpdate()
	if result.Success {
		t.Fatal("校验和不匹配时必须拒绝该安装包")
	}
	if stagedUpdate != nil {
		t.Error("校验失败不应留下待安装状态")
	}
	// 失败后不能把损坏的文件留在更新目录里
	files, _ := os.ReadDir(app.updatesDir())
	for _, f := range files {
		t.Errorf("更新目录不应残留文件: %s", f.Name())
	}
}
