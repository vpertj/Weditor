# WEditor UI 可用性优化 — 设计

日期：2026-08-11
状态：已批准

## 背景

对 WEditor 的 UI/UX 分析发现三处影响可用性的短板：左侧竖栏按钮只有图标无文字（可发现性差）、视图切换按钮语义不直观、空内容时无使用引导。用户确认实施三项优化。

## 目标

1. **竖栏收窄 + 文字标签**：左侧竖栏 80px → 64px；按钮（48×48 纯图标）改为图标+下方 10px 小字标签（56×56）；解除 `.header-btn span { display:none }`；删除 App.vue scoped 中与 app-header.css 重复的 `.app-header` 定义。
2. **视图切换加文字**：three 视图按钮（both/editor/wysiwyg）下方加短标签，新增 i18n key：zh `分栏/源码/排版`，en `Split/Source/Visual`。
3. **空状态引导**：编辑器内容为空且引导未关闭时，在编辑器面板覆盖引导卡片（欢迎语、快捷键提示、基础语法速览、「加载示例内容」按钮、关闭按钮）；输入内容后自动消失。

## 改动文件

| 文件 | 改动 |
|------|------|
| `frontend/src/App.vue` | 删除 scoped 中重复的 `.app-header` 块（:391-402） |
| `frontend/src/styles/components/layout/app-header.css` | `.app-header` width 80→64px；`.header-btn`/`.view-toggle-btn`/`.dropdown-trigger` 48→56px 并支持图标+下方标签；`.header-btn span` 取消 display:none 并设 10px 小字样式 |
| `frontend/src/components/layout/AppHeader.vue` | view-toggle-btn 内加 `<span class="view-toggle-label">`（i18n 短标签） |
| `frontend/src/components/layout/AppMain.vue` | editor-panel 内加空状态引导卡片（v-if="!markdownContent && showEmptyGuide"），本地 ref + 「加载示例」emit |
| `frontend/src/locales/zh-CN.json` / `en.json` | 新增 `header.view.bothShort/editorShort/wysiwygShort` 与 `main.emptyGuide.*` 文案（中英同步） |

## 明确不做

- 不改变竖栏整体存在方式（不收进菜单）
- 不改编辑器/预览/主题/后端逻辑
- 不新增依赖

## 验收标准

1. 竖栏宽度 64px，功能按钮显示文字标签
2. 视图按钮显示短标签（分栏/源码/排版）
3. 空内容时显示引导卡片，输入后消失，可关闭
4. `npm run build` 通过；`go test ./...` 通过
5. zh/en 文案同步
