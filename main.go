package main

import (
	"context"
	"embed"
	goRuntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// 构建应用菜单
	appMenu := buildAppMenu(app)

	err := wails.Run(&options.App{
		Title:     "WEditor",
		Width:     1000,
		Height:    1200,
		MinWidth:  800,
		MinHeight: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// 启用文件拖放：原生层返回拖入文件的绝对路径，前端 OnFileDrop 打开文件。
		// 注意：不设置 DisableWebViewDrop（会破坏编辑器内部文字拖拽）；
		// 前端已在捕获阶段拦截 DOM 文件 drop，防止 ProseMirror 把文件当内容插入导致乱码。
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnDomReady:       app.domReady,
		Bind: []interface{}{
			app,
		},
		// macOS 特定配置
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			// 处理 Finder 双击 / 已运行实例通过 Apple Events 打开的文件
			OnFileOpen: func(filePath string) {
				app.handleFileOpen(filePath)
			},
			About: &mac.AboutInfo{
				Title:   "WEditor",
				Message: "现代化 Markdown 编辑器\n版本 " + APP_VERSION,
				Icon:    nil,
			},
		},
		// 原生应用菜单
		Menu:    appMenu,
		// 启用拖放
		EnableDefaultContextMenu: true,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// shutdown 在应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	a.saveRecentFiles()
}

// domReady 在 DOM 就绪后调用
func (a *App) domReady(ctx context.Context) {
	runtime.LogInfo(ctx, "WEditor DOM ready")
}

// buildAppMenu 构建原生应用菜单
func buildAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// ===== 应用菜单 (macOS 专用) =====
	if goRuntime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
	}

	// ===== 文件菜单 =====
	fileMenu := appMenu.AddSubmenu("文件")
	fileMenu.AddText("新建", keys.CmdOrCtrl("N"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-new-file")
	})
	fileMenu.AddText("打开...", keys.CmdOrCtrl("O"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-open-file")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("保存", keys.CmdOrCtrl("S"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-save-file")
	})
	fileMenu.AddText("另存为...", keys.Combo("S", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-save-as-file")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("导入 Markdown...", keys.CmdOrCtrl("I"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-import-file")
	})

	// ===== 编辑菜单 =====
	editMenu := appMenu.AddSubmenu("编辑")
	editMenu.AddText("撤销", keys.CmdOrCtrl("Z"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-undo")
	})
	editMenu.AddText("重做", keys.Combo("Z", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-redo")
	})
	editMenu.AddSeparator()
	editMenu.AddText("剪切", keys.CmdOrCtrl("X"), nil)
	editMenu.AddText("复制", keys.CmdOrCtrl("C"), nil)
	editMenu.AddText("粘贴", keys.CmdOrCtrl("V"), nil)
	editMenu.AddText("全选", keys.CmdOrCtrl("A"), nil)

	// ===== 视图菜单 =====
	viewMenu := appMenu.AddSubmenu("视图")
	viewMenu.AddText("编辑+预览", keys.CmdOrCtrl("1"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-view-both")
	})
	viewMenu.AddText("仅编辑", keys.CmdOrCtrl("2"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-view-editor")
	})
	viewMenu.AddText("仅预览", keys.CmdOrCtrl("3"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-view-preview")
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("切换同步滚动", keys.Combo("T", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-toggle-sync-scroll")
	})

	// ===== 工具菜单 =====
	toolsMenu := appMenu.AddSubmenu("工具")
	toolsMenu.AddText("设置", keys.CmdOrCtrl(","), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-toggle-settings")
	})
	toolsMenu.AddText("Markdown 语法指南", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-show-guide")
	})
	toolsMenu.AddSeparator()
	toolsMenu.AddText("导出 HTML...", keys.Combo("E", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-export-html")
	})

	// ===== 窗口菜单 (macOS 专用) =====
	if goRuntime.GOOS == "darwin" {
		windowMenu := appMenu.AddSubmenu("窗口")
		windowMenu.AddText("最小化", keys.CmdOrCtrl("M"), func(cd *menu.CallbackData) {
			runtime.WindowMinimise(app.ctx)
		})
		windowMenu.AddText("缩放", nil, func(cd *menu.CallbackData) {
			runtime.WindowToggleMaximise(app.ctx)
		})
		windowMenu.AddSeparator()
		windowMenu.AddText("全部置前", nil, func(cd *menu.CallbackData) {
			// macOS 原生处理
		})
		appMenu.Append(menu.WindowMenu())
	}

	// ===== 帮助菜单 =====
	helpMenu := appMenu.AddSubmenu("帮助")
	helpMenu.AddText("检查更新...", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-check-update")
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("关于 WEditor", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu-show-about")
	})

	return appMenu
}
