<template>
  <div id="app">
    <div class="app-container">
      <!-- 左侧应用头部 -->
      <AppHeader
        :show-settings-panel="showSettingsPanel"
        :view-mode="viewMode"
        :copy-format-options="copyFormatOptions"
        :selected-copy-format="selectedCopyFormat"
        :has-content="hasContent"
        logo-src="./logo.svg"
        logo-alt="WEditor"
        @open-github="openGithub"
        @toggle-settings="toggleSettingsPanel"
        @set-view-mode="setViewMode"
        @show-guide="showGuide"
        @show-about="showAbout"
        @copy-format-select="handleCopyFormatSelect"
        @update:selected-copy-format="selectedCopyFormat.value = $event"
      />

      <!-- 右侧主内容区域 -->
      <div class="app-content">
        <!-- 隐藏文件输入：用于导入 .md -->
        <input ref="fileInputRef" type="file" accept=".md,text/markdown,.txt" style="display:none" @change="handleFileChosen" />

        <!-- 主内容区域 -->
        <AppMain
          :markdown-content="markdownContent"
          :sync-scroll-enabled="syncScrollEnabled"
          :view-mode="viewMode"
          @update:markdown-content="updateMarkdownContent"
          @clear-content="clearContent"
          @load-sample="loadSample"
          @html-generated="updateHtmlContent"
          @import-markdown="triggerImportMd"
          @toggle-sync-scroll="toggleSyncScroll"
        />

        <!-- 应用底部 -->
        <AppFooter
          :character-count="characterCount"
          :line-count="lineCount"
          :word-count="wordCount"
          :estimated-read-time="estimatedReadTime"
          :sync-scroll-enabled="syncScrollEnabled"
          @toggle-sync-scroll="toggleSyncScroll"
        />
      </div>
    </div>

    <!-- 全局回到顶部悬浮按钮 -->
    <BackToTopFloat />

    <!-- 草稿恢复横幅 -->
    <div v-if="showDraftBanner" class="draft-banner">
      <div class="draft-banner-content">
        <span class="draft-banner-icon">💾</span>
        <span class="draft-banner-text">
          检测到未保存的草稿
          <small v-if="draftTime">({{ draftTime }})</small>
        </span>
        <div class="draft-banner-actions">
          <button class="draft-btn recover" @click="handleRecoverDraft">恢复草稿</button>
          <button class="draft-btn discard" @click="handleDiscardDraft">丢弃</button>
        </div>
      </div>
    </div>

    <!-- 自动保存状态 -->
    <div v-if="lastSavedAt && !isSaving" class="autosave-indicator">
      <span>{{ isMac ? '自动保存' : '自动保存' }} {{ lastSavedAt }}</span>
    </div>

    <!-- 通知组件 -->
    <div v-if="notifications.length > 0" class="notification-container">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        :class="['notification', notification.type, { 'slide-out': notification.isRemoving }]"
      >
        {{ notification.message }}
      </div>
    </div>

    <!-- 设置面板 -->
    <SettingsPanel
      :visible="showSettingsPanel"
      @close="closeSettingsPanel"
      @show-notification="showNotification"
    />

    <!-- Markdown 语法指南 -->
    <MarkdownGuide
      :show="showMarkdownGuide"
      @close="closeGuide"
    />

    <!-- 关于弹窗 -->
    <AboutDialog
      :show="showAboutDialog"
      @close="closeAbout"
    />

  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAppState, useWails, usePlatform, useAutoSave, useWindowDrag } from './composables/index.js'
import { useGlobalThemeManager } from './composables/index.js'
import { nextTick } from 'vue'
import AppHeader from './components/layout/AppHeader.vue'
import AppMain from './components/layout/AppMain.vue'
import AppFooter from './components/layout/AppFooter.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import MarkdownGuide from './components/MarkdownGuide.vue'
import AboutDialog from './components/AboutDialog.vue'
import BackToTopFloat from './components/BackToTopFloat.vue'

// ==========================================
// 平台检测
// ==========================================
const { isMac, isWindows, isLinux, platform } = usePlatform()

// ==========================================
// 应用状态管理
// ==========================================
const {
  // 状态
  markdownContent,
  htmlContent,
  showSettingsPanel,
  showMarkdownGuide,
  showAboutDialog,
  syncScrollEnabled,
  viewMode,
  notifications,
  selectedCopyFormat,
  copyFormatOptions,

  // 计算属性
  hasContent,
  characterCount,
  lineCount,
  wordCount,
  estimatedReadTime,

  // 方法
  updateMarkdownContent,
  clearContent,
  loadSample,
  updateHtmlContent,
  toggleSettingsPanel,
  closeSettingsPanel,
  showGuide,
  closeGuide,
  showAbout,
  closeAbout,
  toggleSyncScroll,
  setViewMode,
  showNotification,
  handleCopyFormatSelect,
  openGithub,

  // 默认打开方式
  defaultOpenMode
} = useAppState()

// ==========================================
// 桌面集成 (Wails)
// ==========================================
const {
  openFile,
  saveFile,
  saveFileAs,
  newFile,
  exportHTML,
  loadStartupFile,
  openFilePath,
  setupMenuListeners,
  setupFileDropListener,
  currentFilePath,
  currentFileName
} = useWails()

// ==========================================
// 自动保存
// ==========================================
const {
  lastSavedAt,
  isSaving,
  isDraftAvailable,
  draftTime,
  recoverDraft,
  clearDraft
} = useAutoSave(markdownContent, {
  delay: 2000,
  enabled: true,
  saveToDisk: async (content) => {
    // 如果当前有已打开的文件路径，自动保存到磁盘
    if (currentFilePath.value) {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.SaveFile(currentFilePath.value, content)
      }
    }
  },
  filePath: currentFilePath
})

// ==========================================
// 草稿恢复
// ==========================================
const showDraftBanner = ref(false)

// 检测是否有草稿（在启动文件加载之后调用，避免双击打开文件时误弹“未保存草稿”提示）
function checkDraftBanner() {
  if (isDraftAvailable.value && !hasContent.value) {
    showDraftBanner.value = true
  }
}

function handleRecoverDraft() {
  const draft = recoverDraft()
  if (draft) {
    updateMarkdownContent(draft)
    showDraftBanner.value = false
    showNotification('已恢复上次编辑的草稿', 'success')
  }
}

function handleDiscardDraft() {
  clearDraft()
  showDraftBanner.value = false
}

// ==========================================
// 默认打开方式
// ==========================================
// 通过系统打开（Finder 双击 / 命令行 / 拖放 / 打开对话框）文件时，
// 切换到用户设置的默认视图模式（默认排版，可在设置中修改）
function applyDefaultOpenMode() {
  setViewMode(defaultOpenMode.value)
}

// ==========================================
// Finder 双击 / Apple Events 打开文件（尽早注册，避免冷启动竞态）
// ==========================================
// 尽早注册 app-file-open 监听，随后通知后端前端已就绪：
// 就绪前到达的文件事件由后端暂存，经 GetStartupFile 在启动时获取；
// 就绪后到达的实时推送，由本监听接收。
window.runtime?.EventsOn?.('app-file-open', async (filePath) => {
  if (!filePath) return
  try {
    const result = await openFilePath(filePath)
    if (result.success) {
      updateMarkdownContent(result.content)
      clearDraft()
      applyDefaultOpenMode()
      showNotification(`已打开: ${currentFileName.value}`, 'success')
    } else {
      showNotification(`无法读取文件: ${result.message}`, 'error')
    }
  } catch (e) {
    showNotification(`无法读取文件: ${e.message}`, 'error')
  }
})

// ==========================================
// 菜单监听器
// ==========================================
nextTick(async () => {
  console.log(`🖥️ 运行平台: ${platform.value}, Wails 菜单监听器初始化`)

  setupMenuListeners({
    // 新建文件
    onNewFile: () => {
      console.log('📄 菜单: 新建文件')
      if (newFile()) {
        clearContent()
        showNotification('已创建新文件', 'success')
      }
    },

    // 打开文件
    onOpenFile: async () => {
      console.log('📁 菜单: 打开文件')
      const result = await openFile()
      if (!result) {
        console.log('用户取消了操作')
        return
      }
      if (result.success) {
        updateMarkdownContent(result.content)
        clearDraft()
        applyDefaultOpenMode()
        showNotification(`已打开: ${currentFileName.value}`, 'success')
      } else {
        showNotification(`打开失败: ${result.message}`, 'error')
      }
    },

    // 保存文件
    onSaveFile: async () => {
      console.log('💾 菜单: 保存文件')
      const result = await saveFile(markdownContent.value)
      if (result.success) {
        clearDraft()
        showNotification(`已保存: ${currentFileName.value}`, 'success')
      } else {
        showNotification(`保存失败: ${result.message}`, 'error')
      }
    },

    // 另存为
    onSaveAsFile: async () => {
      console.log('💾 菜单: 另存为')
      const result = await saveFileAs(markdownContent.value)
      if (result.success) {
        clearDraft()
        showNotification(`已保存: ${currentFileName.value}`, 'success')
      } else if (result.message !== '用户取消了操作') {
        showNotification(`保存失败: ${result.message}`, 'error')
      }
    },

    // 导入文件（通过文件选择器）
    onImportFile: () => {
      triggerImportMd()
    },

    // 视图切换
    onViewBoth: () => setViewMode('both'),
    onViewEditor: () => setViewMode('editor'),
    onViewPreview: () => setViewMode('preview'),
    onToggleSyncScroll: () => toggleSyncScroll(),

    // 面板
    onToggleSettings: () => toggleSettingsPanel(),
    onShowGuide: () => showGuide(),
    onShowAbout: () => showAbout(),

    // 导出 HTML
    onExportHTML: async () => {
      console.log('🌐 菜单: 导出 HTML')
      if (!htmlContent.value) {
        showNotification('预览尚未生成，请等待渲染完成', 'warning')
        return
      }
      const result = await exportHTML(htmlContent.value, '')
      if (result && result.success) {
        showNotification(`已导出: ${result.filePath.split('/').pop()}`, 'success')
      } else if (result && result.message !== '用户取消了操作') {
        showNotification(`导出失败: ${result.message}`, 'error')
      }
    }
  })

  // 文件拖放支持
  setupFileDropListener(async (filePath) => {
    try {
      const result = await openFilePath(filePath)
      if (result.success) {
        updateMarkdownContent(result.content)
        clearDraft()
        applyDefaultOpenMode()
        showNotification(`已拖放打开: ${currentFileName.value}`, 'success')
      } else {
        showNotification(`无法读取文件: ${result.message}`, 'error')
      }
    } catch (e) {
      showNotification(`无法读取文件: ${e.message}`, 'error')
    }
  })

  // 通知后端前端已就绪：就绪后到达的文件打开事件走实时推送；
  // 就绪前到达的已由后端暂存，经下面的 GetStartupFile 获取
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.MarkFrontendReady) {
    try { await window.go.main.App.MarkFrontendReady() } catch (e) { console.warn('MarkFrontendReady 调用失败:', e) }
  }

  // 加载通过 Finder 双击 / 命令行打开的文件（启动期后端暂存的）
  const startupResult = await loadStartupFile()
  if (startupResult && startupResult.success) {
    updateMarkdownContent(startupResult.content)
    clearDraft()
    applyDefaultOpenMode()
    showNotification(`已打开: ${currentFileName.value}`, 'success')
  } else if (startupResult && !startupResult.success) {
    showNotification(`打开失败: ${startupResult.message}`, 'error')
  } else {
    // 无启动文件时才检测草稿，避免与双击打开文件冲突
    checkDraftBanner()
  }
})

// ==========================================
// 主题管理器
// ==========================================
useGlobalThemeManager()

// ==========================================
// 窗口拖拽兜底（CSS -webkit-app-region 不可靠时的 JS 方案）
// ==========================================
useWindowDrag()

// ==========================================
// 导入 .md 文件（通过浏览器文件选择器）
// ==========================================
const fileInputRef = ref(null)
const triggerImportMd = () => fileInputRef.value && fileInputRef.value.click()

const handleFileChosen = async (e) => {
  const file = e.target.files && e.target.files[0]
  e.target.value = ''
  if (!file) return
  if (!/\.(md|markdown)$/i.test(file.name) && !/text\/markdown|text\/plain/.test(file.type)) {
    showNotification('仅支持导入 .md 文件', 'warning')
    return
  }
  try {
    const text = await file.text()
    updateMarkdownContent(text)
    clearDraft()
    showNotification(`已导入：${file.name}`, 'success')
  } catch (err) {
    showNotification(`导入失败：${err.message}`, 'error')
  }
}
</script>

<style scoped>
/* 导入原来的样式 */
@import './styles/components/layout/app-layout.css';
@import './styles/components/notifications.css';

/* 水平布局 */
.app-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* 右侧主内容区域 */
.app-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 通知容器样式调整 */
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* 草稿恢复横幅 */
.draft-banner {
  position: fixed;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1001;
  background: var(--theme-primary, #4A90D9);
  color: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  animation: draft-slide-in 0.3s ease-out;
  max-width: 500px;
  /* 横幅悬浮于顶部，自身及其按钮不参与窗口拖拽 */
  -webkit-app-region: no-drag;
}

.draft-banner-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
}

.draft-banner-icon {
  font-size: 18px;
}

.draft-banner-text {
  flex: 1;
  font-size: 14px;
}

.draft-banner-text small {
  opacity: 0.8;
  font-size: 12px;
}

.draft-banner-actions {
  display: flex;
  gap: 8px;
}

.draft-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.draft-btn.recover {
  background: white;
  color: var(--theme-primary, #4A90D9);
}

.draft-btn.recover:hover {
  background: #f0f0f0;
}

.draft-btn.discard {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.draft-btn.discard:hover {
  background: rgba(255, 255, 255, 0.35);
}

@keyframes draft-slide-in {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

/* 自动保存状态指示器 */
.autosave-indicator {
  position: fixed;
  bottom: 10px;
  right: 10px;
  z-index: 1001;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.6;
  pointer-events: none;
}
</style>
