package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	APP_NAME    = "WEditor"
	APP_VERSION = "1.0.0"
)

// RecentFile 最近打开的文件信息
type RecentFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	OpenedAt  string `json:"openedAt"`
}

// App struct
type App struct {
	ctx         context.Context
	appDataDir  string
	recentFiles []RecentFile
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.appDataDir = a.getAppDataDir()
	a.loadRecentFiles()
}

// ==========================================
// 文件对话框操作
// ==========================================

// OpenFileDialog 打开文件选择对话框，返回 (文件路径, 文件内容, 错误)
func (a *App) OpenFileDialog() (string, string, error) {
	filePath, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "打开 Markdown 文件",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Markdown 文件 (*.md)", Pattern: "*.md;*.markdown;*.mdown;*.mkd"},
			{DisplayName: "文本文件 (*.txt)", Pattern: "*.txt"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if filePath == "" {
		return "", "", nil // 用户取消了选择
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 添加到最近文件列表
	fileName := filepath.Base(filePath)
	a.addRecentFile(filePath, fileName)

	return filePath, string(content), nil
}

// SaveFileDialog 打开保存文件对话框，返回选择的路径
func (a *App) SaveFileDialog(defaultName string) (string, error) {
	if defaultName == "" {
		defaultName = "untitled.md"
	}
	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "保存 Markdown 文件",
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Markdown 文件 (*.md)", Pattern: "*.md"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("保存文件对话框失败: %w", err)
	}
	return filePath, nil
}

// ==========================================
// 文件读写操作
// ==========================================

// SaveFile 将内容写入指定路径
func (a *App) SaveFile(filePath string, content string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 更新最近文件
	fileName := filepath.Base(filePath)
	a.addRecentFile(filePath, fileName)

	return nil
}

// ReadFile 读取文件内容
func (a *App) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	fileName := filepath.Base(filePath)
	a.addRecentFile(filePath, fileName)

	return string(content), nil
}

// ==========================================
// 最近文件管理
// ==========================================

// GetRecentFiles 获取最近打开的文件列表
func (a *App) GetRecentFiles() []RecentFile {
	return a.recentFiles
}

// AddRecentFile 添加文件到最近列表
func (a *App) AddRecentFile(filePath string, name string) {
	a.addRecentFile(filePath, name)
}

// ClearRecentFiles 清空最近文件列表
func (a *App) ClearRecentFiles() {
	a.recentFiles = []RecentFile{}
	a.saveRecentFiles()
}

// ==========================================
// 导出功能
// ==========================================

// ExportHTMLResult 导出 HTML 的结果
type ExportHTMLResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"filePath"`
	Message  string `json:"message"`
}

// ExportHTML 导出 HTML 文件（接收原始 Markdown，后端生成基本 HTML 包装）
// 实际渲染由前端完成，这里仅做文件保存
func (a *App) ExportHTML(htmlContent string, defaultName string) ExportHTMLResult {
	if defaultName == "" {
		defaultName = "export.html"
	}
	if !strings.HasSuffix(defaultName, ".html") {
		defaultName = strings.TrimSuffix(defaultName, filepath.Ext(defaultName)) + ".html"
	}

	filePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "导出 HTML",
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "HTML 文件 (*.html)", Pattern: "*.html;*.htm"},
		},
	})
	if err != nil {
		return ExportHTMLResult{Success: false, Message: fmt.Sprintf("对话框失败: %v", err)}
	}
	if filePath == "" {
		return ExportHTMLResult{Success: false, Message: "用户取消了操作"}
	}

	// 添加完整的 HTML 包装
	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { max-width: 860px; margin: 40px auto; padding: 0 20px; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif; line-height: 1.8; color: #333; }
        img { max-width: 100%%; }
        pre { background: #f5f5f5; padding: 16px; border-radius: 8px; overflow-x: auto; }
        code { font-family: "JetBrains Mono", "Fira Code", "Menlo", monospace; font-size: 0.9em; }
        blockquote { border-left: 4px solid #4A90D9; margin: 0; padding: 4px 16px; color: #666; background: #f8f9fa; }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
        th { background: #f5f5f5; }
    </style>
</head>
<body>
%s
</body>
</html>`, defaultName, htmlContent)

	if err := os.WriteFile(filePath, []byte(fullHTML), 0644); err != nil {
		return ExportHTMLResult{Success: false, Message: fmt.Sprintf("写入文件失败: %v", err)}
	}

	return ExportHTMLResult{Success: true, FilePath: filePath}
}

// ==========================================
// 应用信息
// ==========================================

// GetAppVersion 返回应用版本
func (a *App) GetAppVersion() string {
	return APP_VERSION
}

// GetAppDataDir 返回应用数据目录
func (a *App) GetAppDataDir() string {
	return a.appDataDir
}

// GetPlatform 返回当前平台信息
func (a *App) GetPlatform() string {
	env := wailsRuntime.Environment(a.ctx)
	return fmt.Sprintf("%s/%s", env.Platform, env.Arch)
}

// ==========================================
// 通知/事件（前端通过 Events 系统通信）
// ==========================================


// ==========================================
// 内部辅助方法
// ==========================================

// getAppDataDir 获取应用数据目录
func (a *App) getAppDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	var configDir string
	switch wailsRuntime.Environment(a.ctx).Platform {
	case "darwin":
		configDir = filepath.Join(homeDir, "Library", "Application Support", APP_NAME)
	case "windows":
		configDir = filepath.Join(homeDir, "AppData", "Roaming", APP_NAME)
	default:
		configDir = filepath.Join(homeDir, ".config", APP_NAME)
	}

	os.MkdirAll(configDir, 0755)
	return configDir
}

// recentFilePath 返回最近文件记录的存储路径
func (a *App) recentFilePath() string {
	return filepath.Join(a.appDataDir, "recent.json")
}

// loadRecentFiles 从磁盘加载最近文件列表
func (a *App) loadRecentFiles() {
	data, err := os.ReadFile(a.recentFilePath())
	if err != nil {
		a.recentFiles = []RecentFile{}
		return
	}
	json.Unmarshal(data, &a.recentFiles)
}

// saveRecentFiles 保存最近文件列表到磁盘
func (a *App) saveRecentFiles() {
	data, err := json.MarshalIndent(a.recentFiles, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(a.recentFilePath(), data, 0644)
}

// addRecentFile 添加一个文件到最近列表（去重，最多保留 20 条）
func (a *App) addRecentFile(filePath string, name string) {
	if filePath == "" {
		return
	}

	// 去重：移除已存在的相同路径
	filtered := make([]RecentFile, 0, len(a.recentFiles))
	for _, f := range a.recentFiles {
		if f.Path != filePath {
			filtered = append(filtered, f)
		}
	}

	// 添加新记录到最前面
	newEntry := RecentFile{
		Path:     filePath,
		Name:     name,
		OpenedAt: time.Now().Format("2006-01-02 15:04"),
	}
	a.recentFiles = append([]RecentFile{newEntry}, filtered...)

	// 最多保留 20 条
	if len(a.recentFiles) > 20 {
		a.recentFiles = a.recentFiles[:20]
	}

	a.saveRecentFiles()
}

// Greet 保留的模板方法
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
