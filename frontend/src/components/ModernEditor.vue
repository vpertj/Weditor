<template>
  <div class="modern-editor" :class="editorClasses">
    <!-- 现代化工具栏 -->
    <ModernToolbar
      :compact="compactMode"
      :floating="floatingToolbar"
      :theme="currentTheme"
      @tool-click="handleToolAction"
      @search-toggle="toggleSearchMode"
    />

    <!-- 编辑器主体 -->
    <div class="editor-body">
      <!-- 搜索面板 -->
      <div v-if="searchActive" class="search-panel">
        <div class="search-header">
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            type="text"
            placeholder="搜索内容..."
            class="search-input"
            @keydown.escape="closeSearch"
          />
          <div class="search-controls">
            <ToolbarButton
              title="上一个"
              icon="chevron-up"
              @click="findPrevious"
          />
          <ToolbarButton
            title="下一个"
            icon="chevron-down"
            @click="findNext"
          />
          <ToolbarButton
            title="关闭"
            icon="x"
            @click="closeSearch"
          />
        </div>
      </div>

      <!-- 编辑器容器 -->
      <div class="editor-container">
        <div ref="editorElement" class="codemirror-wrapper"></div>
      </div>

      <!-- 状态栏 -->
      <div class="editor-statusbar">
        <div class="status-left">
          <span class="status-item">
            <svg-icon name="file-text" />
            {{ characterCount }} 字符
          </span>
          <span class="status-item">
            <svg-icon name="clock" />
            {{ estimatedReadTime }} 分钟
          </span>
        </div>
        <div class="status-right">
          <span class="status-item">
            <svg-icon name="type" />
            {{ wordCount }} 词
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useMarkdownEditor } from '../composables/editor/index.js'
import ModernToolbar from './ModernToolbar.vue'
import ToolbarButton from './ToolbarButton.vue'

// Props
const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  theme: {
    type: String,
    default: 'light'
  },
  compact: {
    type: Boolean,
    default: false
  },
  floating: {
    type: Boolean,
    default: false
  },
  syncScrollEnabled: {
    type: Boolean,
    default: true
  }
})

// Emits
const emit = defineEmits([
  'update:modelValue',
  'tool-action',
  'search-change'
])

// 状态
const searchActive = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const currentResultIndex = ref(-1)

// 编辑器实例
const editorElement = ref(null)
const searchInputRef = ref(null)

// 使用编辑器 composable
const editor = useMarkdownEditor({
  initialValue: props.modelValue,
  theme: props.theme,
  onContentChange: (content) => {
    emit('update:modelValue', content)
  },
  onScroll: (scrollPercentage) => {
    if (props.syncScrollEnabled) {
      handleEditorScroll(scrollPercentage)
    }
  }
})

// 计算属性
const editorClasses = computed(() => ({
  'editor-compact': props.compact,
  'editor-floating': props.floating,
  [`editor-${props.theme}`]: true
})

const currentTheme = computed(() => props.theme)
const compactMode = computed(() => props.compact)
const floatingToolbar = computed(() => props.floating)

// 统计信息
const characterCount = computed(() => editor.content.value.length)
const wordCount = computed(() => {
  const text = editor.content.value.trim()
  return text ? text.split(/\s+/).length : 0
})

const estimatedReadTime = computed(() => {
  const words = wordCount.value
  const readingSpeed = 200 // 词/分钟
  return Math.max(1, Math.ceil(words / readingSpeed))
})

// 事件处理
const handleToolAction = (tool) => {
  emit('tool-action', tool)
  
  // 执行对应的编辑器操作
  if (tool.action && typeof editor[tool.action] === 'function') {
    editor[tool.action]()
  }
}

const toggleSearchMode = (active) => {
  searchActive.value = active
  if (active) {
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
}

const closeSearch = () => {
  searchActive.value = false
  searchQuery.value = ''
  searchResults.value = []
  currentResultIndex.value = -1
}

const handleEditorScroll = (scrollPercentage) => {
  // 滚动同步逻辑
  const previewElement = document.querySelector('.preview-container .preview-rendered')
  if (previewElement) {
    const previewMaxScrollTop = Math.max(0, previewElement.scrollHeight - previewElement.clientHeight)
    const targetScrollTop = Math.round(previewMaxScrollTop * scrollPercentage)
    
    // 防止循环触发
    if (typeof window !== 'undefined' && !window.__scrollSyncLock) {
      window.__scrollSyncLock = true
      previewElement.scrollTop = targetScrollTop
      requestAnimationFrame(() => {
        window.__scrollSyncLock = false
      })
    }
  }
}

const findNext = () => {
  if (searchResults.value.length === 0) return
  
  currentResultIndex.value = (currentResultIndex.value + 1) % searchResults.value.length
  highlightCurrentResult()
}

const findPrevious = () => {
  if (searchResults.value.length === 0) return
  
  currentResultIndex.value = (currentResultIndex.value - 1 + searchResults.value.length) % searchResults.value.length
  highlightCurrentResult()
}

const highlightCurrentResult = () => {
  // TODO: 实现搜索结果高亮
}

// 监听搜索查询变化
watch(searchQuery, (newQuery) => {
  if (newQuery.trim()) {
    // TODO: 实现搜索功能
    searchResults.value = [{ start: 0, end: 10 }]
  currentResultIndex.value = 0
  emit('search-change', newQuery)
})

// 监听 props 变化
watch(() => props.modelValue, (newValue) => {
  if (newValue !== editor.content.value) {
    editor.updateContent(newValue)
  }
})

// 生命周期
onMounted(() => {
  // 初始化编辑器
  editor.initEditor(editorElement.value)
})

onUnmounted(() => {
  // 销毁编辑器
  editor.destroyEditor()
})
</script>

<style scoped>
.modern-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-normal);
}

.editor-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.search-panel {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-primary);
  padding: var(--space-3);
}

.search-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.search-input {
  flex: 1;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  transition: all var(--transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 0 2px var(--primary-100);
}

.search-controls {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.editor-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.codemirror-wrapper {
  height: 100%;
  width: 100%;
}

.editor-statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-primary);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.status-left,
.status-right {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.status-item {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

/* 紧凑模式 */
.editor-compact {
  border-radius: var(--radius-md);
}

.editor-compact .editor-statusbar {
  padding: var(--space-1) var(--space-2);
}

/* 浮动模式 */
.editor-floating {
  box-shadow: var(--shadow-xl);
  border: 1px solid var(--border-primary);
}

/* 主题样式 */
.editor-light {
  background: var(--bg-primary);
}

.editor-dark {
  background: var(--bg-secondary);
}

/* 搜索高亮样式 */
.search-highlight {
  background: var(--primary-200);
  border-radius: var(--radius-xs);
}

.search-highlight-current {
  background: var(--primary-500);
  color: var(--white);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .modern-editor {
    border-radius: 0;
  }
  
  .editor-statusbar {
    flex-direction: column;
    gap: var(--space-1);
  padding: var(--space-1) var(--space-2);
}
</style>