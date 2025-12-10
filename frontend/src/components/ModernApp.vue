<template>
  <div class="modern-app" :class="appClasses">
    <!-- 应用头部 -->
    <ModernHeader
      :title="appTitle"
      :theme="currentTheme"
      :show-settings="showSettingsPanel"
      @toggle-settings="toggleSettingsPanel"
      @toggle-theme="toggleTheme"
    />

    <!-- 主内容区域 -->
    <div class="app-main">
      <!-- 视图模式切换器 -->
      <div class="view-mode-switcher">
      <div class="mode-buttons">
        <button
          v-for="mode in viewModes"
          :key="mode.value"
          :class="['mode-button', { 'mode-active': viewMode === mode.value }]"
          @click="setViewMode(mode.value)"
        >
          <svg-icon :name="mode.icon" />
          <span>{{ mode.label }}</span>
        </button>
      </div>

      <!-- 编辑器与预览区域 -->
      <div class="editor-preview-container" :class="containerClasses">
        <!-- 编辑器区域 -->
        <div v-if="showEditor" class="editor-section">
          <ModernEditor
            v-model="markdownContent"
            :theme="currentTheme"
            :compact="isCompactMode"
            :sync-scroll-enabled="syncScrollEnabled"
            @tool-action="handleToolAction"
            @search-change="handleSearchChange"
          />
        </div>

        <!-- 预览区域 -->
        <div v-if="showPreview" class="preview-section">
          <ModernPreview
            :content="htmlContent"
            :theme="currentTheme"
            :sync-scroll-enabled="syncScrollEnabled"
          />
        </div>
      </div>

      <!-- 应用底部 -->
      <ModernFooter
        :stats="editorStats"
        :theme="currentTheme"
      />
    </div>

    <!-- 设置面板 -->
    <ModernSettingsPanel
      :visible="showSettingsPanel"
      :theme="currentTheme"
      @close="closeSettingsPanel"
        @theme-change="handleThemeChange"
        @font-change="handleFontChange"
      />

    <!-- 回到顶部按钮 -->
    <BackToTopFloat :theme="currentTheme" />

    <!-- 通知系统 -->
    <NotificationContainer
      :notifications="notifications"
      :theme="currentTheme"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useAppState } from '../composables/useAppState.js'
import ModernHeader from './ModernHeader.vue'
import ModernEditor from './ModernEditor.vue'
import ModernPreview from './ModernPreview.vue'
import ModernFooter from './ModernFooter.vue'
import ModernSettingsPanel from './ModernSettingsPanel.vue'
import BackToTopFloat from './BackToTopFloat.vue'
import NotificationContainer from './NotificationContainer.vue'

// 应用状态管理
const {
  // 状态
  markdownContent,
  showSettingsPanel,
  syncScrollEnabled,
  viewMode,
  notifications,
  htmlContent,

  // 计算属性
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
  toggleSyncScroll,
  setViewMode,
  showNotification
} = useAppState()

// 计算属性
const appClasses = computed(() => ({
  'app-dark': currentTheme.value === 'dark',
  'app-compact': isCompactMode.value
})

const currentTheme = ref('light')
const appTitle = ref('WEditor - 现代化 Markdown 编辑器'

// 视图模式配置
const viewModes = [
  { value: 'edit', label: '编辑', icon: 'edit' },
  { value: 'preview', label: '预览', icon: 'eye' },
  { value: 'split', label: '分屏', icon: 'layout' }
]

// 编辑器统计信息
const editorStats = computed(() => ({
  characters: characterCount.value,
  lines: lineCount.value,
  words: wordCount.value,
  readTime: estimatedReadTime.value
})

// 显示控制
const showEditor = computed(() => viewMode.value === 'edit' || viewMode.value === 'split'
})

const showPreview = computed(() => viewMode.value === 'preview' || viewMode.value === 'split'
})

const isCompactMode = computed(() => window.innerWidth < 1024)

const containerClasses = computed(() => ({
  'container-split': viewMode.value === 'split',
  'container-compact': isCompactMode.value
})

// 事件处理
const handleToolAction = (tool) => {
  console.log('工具操作:', tool)
  
  // 根据工具类型执行相应操作
  switch (tool.type) {
    case 'format':
      // 格式操作
      break
    case 'insert':
      // 插入操作
      break
    case 'action':
      // 动作操作
      break
  }
}

const handleSearchChange = (query) => {
  console.log('搜索查询:', query)
}

const toggleTheme = () => {
  currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  showNotification(`已切换到${currentTheme.value === 'light' ? '亮色' : '暗色'}主题`, 'info')
}

const handleThemeChange = (newTheme) => {
  currentTheme.value = newTheme
}

const handleFontChange = (fontSettings) => {
  console.log('字体设置变更:', fontSettings)
}

// 响应式处理
const handleResize = () => {
  // 窗口大小变化处理
}

// 生命周期
onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

// 监听主题变化
watch(currentTheme, (newTheme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
})
</script>

<style scoped>
.modern-app {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  transition: all var(--transition-normal);
}

.app-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.view-mode-switcher {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-3);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-primary);
}

.mode-buttons {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.mode-button {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.mode-button:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.mode-active {
  background: var(--primary-500);
  color: var(--white);
  border-color: var(--primary-500);
}

.editor-preview-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.editor-section,
.preview-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.container-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: var(--border-primary);
}

.container-split .editor-section,
.container-split .preview-section {
  flex: none;
}

/* 紧凑模式 */
.app-compact .view-mode-switcher {
  padding: var(--space-2);
}

.app-compact .mode-button {
  padding: var(--space-1) var(--space-2);
}

/* 暗色主题 */
.app-dark {
  background: var(--bg-secondary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .view-mode-switcher {
    padding: var(--space-2);
  }
  
  .mode-buttons {
    flex-direction: column;
    gap: var(--space-1);
  }
  
  .container-split {
    grid-template-columns: 1fr;
  grid-template-rows: 1fr 1fr;
  }
}
</style>