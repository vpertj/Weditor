# WEditor 窗口宽度与工具栏按钮顺序调整 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 WEditor 默认窗口宽度从 1200 调整为 1440，并将左侧竖排工具栏的"设置"按钮移到最底部。

**Architecture:** 两处独立静态改动，无运行时状态变化：1) Wails 窗口配置在 `main.go` 的 `options.App`；2) 左侧竖排工具栏按钮顺序由 `frontend/src/components/layout/AppHeader.vue` 模板中的 DOM 顺序决定（flex column，视觉顺序 = DOM 顺序）。

**Tech Stack:** Go (Wails v2)、Vue 3 (Composition API, `<script setup>`)、Vite。

## Global Constraints

- 窗口高度 `768`、最小尺寸 `MinWidth: 800 / MinHeight: 500` 保持不变（来自 spec）。
- `viewMode` 默认值 `both` 不变（左右分栏行为不变）。
- "设置"按钮的 `toggle-settings` 事件、齿轮图标、`active` 高亮、`showSettingsPanel` 绑定不变。
- 不新增依赖、不改其它按钮逻辑。
- 提交时只 `git add` 本计划涉及的文件（工作区存在其它未提交改动，禁止一并提交）。

---

### Task 1: 调整窗口默认宽度（main.go）

**Files:**
- Modify: `main.go:28`（`Width: 1200` → `Width: 1440`）

**Interfaces:**
- Consumes: 无（独立改动）
- Produces: Wails 启动窗口默认宽度 1440

- [ ] **Step 1: 修改窗口宽度**

编辑 `main.go` 第 28 行：

```go
Width:     1440,
Height:    768,
```

（`Height`、`MinWidth`、`MinHeight` 保持不变）

- [ ] **Step 2: 验证改动生效**

Run: `grep -n "Width:" main.go`
Expected: 输出 `Width:     1440,`（gofmt 对齐下为 `Width:     1440,`）

- [ ] **Step 3: 编译验证**

Run: `go test ./...`
Expected: `?   mdeditor   [no test files]`，exit 0

- [ ] **Step 4: 提交**

```bash
git add main.go
git commit -m "feat: 窗口默认宽度调整至 1440 以优化左右分栏显示"
```

---

### Task 2: 设置按钮移至左侧竖排工具栏最底部（AppHeader.vue）

**Files:**
- Modify: `frontend/src/components/layout/AppHeader.vue:46-55`（"设置"按钮 `<button>` 块，从 `header-right` 第 2 位移到末尾）

**Interfaces:**
- Consumes: 无（事件名 `toggle-settings`、prop `showSettingsPanel` 已存在，不变）
- Produces: 竖排工具栏顺序变为 `视图切换 → 语法指南 → 复制格式 → 语言 → 关于 → 设置`

- [ ] **Step 1: 移动"设置"按钮模板块**

在 `AppHeader.vue` 中，将第 46-55 行的"设置"按钮 `<button>...</button>` 整块（从 `<button class="header-btn" @click="$emit('toggle-settings')"...>` 到 `</button>`）剪切，粘贴到 `header-right` 的**最后**（"关于"按钮 `</button>` 之后、`<!-- <button class="header-btn" @click="$emit('open-github')"...` 注释之前）。块内容原样保留：

```html
      <button
        class="header-btn"
        @click="$emit('toggle-settings')"
        :class="{ 'active': showSettingsPanel }"
      >
        <svg viewBox="0 0 24 24" width="16" height="16">
          <path fill="currentColor" d="M12,15.5A3.5,3.5 0 0,1 8.5,12A3.5,3.5 0 0,1 12,8.5A3.5,3.5 0 0,1 15.5,12A3.5,3.5 0 0,1 12,15.5M19.43,12.97C19.47,12.65 19.5,12.33 19.5,12C19.5,11.67 19.47,11.34 19.43,11L21.54,9.37C21.73,9.22 21.78,8.95 21.66,8.73L19.66,5.27C19.54,5.05 19.27,4.96 19.05,5.05L16.56,6.05C16.04,5.66 15.5,5.32 14.87,5.07L14.5,2.42C14.46,2.18 14.25,2 14,2H10C9.75,2 9.54,2.18 9.5,2.42L9.13,5.07C8.5,5.32 7.96,5.66 7.44,6.05L4.95,5.05C4.73,4.96 4.46,5.05 4.34,5.27L2.34,8.73C2.22,8.95 2.27,9.22 2.46,9.37L4.57,11C4.53,11.34 4.5,11.67 4.5,12C4.5,12.33 4.53,12.65 4.57,12.97L2.46,14.63C2.27,14.78 2.22,15.05 2.34,15.27L4.34,18.73C4.46,18.95 4.73,19.03 4.95,18.95L7.44,17.94C7.96,18.34 8.5,18.68 9.13,18.93L9.5,21.58C9.54,21.82 9.75,22 10,22H14C14.25,22 14.46,21.82 14.5,21.58L14.87,18.93C15.5,18.68 16.04,18.34 16.56,17.94L19.05,18.95C19.27,19.03 19.54,18.95 19.66,18.73L21.66,15.27C21.78,15.05 21.73,14.78 21.54,14.63L19.43,12.97Z"/>
        </svg>
        <span>{{ $t('header.settings') }}</span>
      </button>
```

- [ ] **Step 2: 验证顺序**

Run: `grep -n "toggle-settings\|show-guide\|show-about\|copy-format-select" frontend/src/components/layout/AppHeader.vue`
Expected: 行号顺序为 `show-guide`（语法指南）< `copy-format-select`（复制格式）< `show-about`（关于）< `toggle-settings`（设置，最后）

- [ ] **Step 3: 前端构建验证**

Run: `cd frontend && npm run build`
Expected: `✓ built in ...`，exit 0

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/layout/AppHeader.vue
git commit -m "feat: 左侧工具栏设置按钮移至最底部"
```

---

## Self-Review

**1. Spec coverage:** spec 两条目标（Width 1440、设置置底）分别由 Task 1、Task 2 覆盖；"明确不做"项（高度/最小尺寸/viewMode/其它按钮/无新依赖）均已写入 Global Constraints。✓

**2. Placeholder scan:** 无 TBD/TODO/占位；Step 1 均给出完整代码/文本。✓

**3. Type consistency:** 事件名 `toggle-settings`、prop `showSettingsPanel`、`header.settings` i18n key 均与现有代码一致，未新造标识符。✓
