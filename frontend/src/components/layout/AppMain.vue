<template>
  <main class="app-main" :class="`view-mode-${viewMode}`">
    <!-- 左侧：Markdown编辑器 -->
    <div v-if="viewMode === 'both' || viewMode === 'editor'" class="editor-panel">
      <div class="panel-header">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20">
            <path fill="currentColor" d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
          </svg>
          {{ $t('main.editor.title') }}
        </h3>
        <div class="panel-actions">
          <button @click="$emit('clear-content')" class="btn-small" :title="$t('main.editor.clear')">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M19,4H15.5L14.5,3H9.5L8.5,4H5V6H19M6,19A2,2 0 0,0 8,21H16A2,2 0 0,0 18,19V7H6V19Z"/>
            </svg>
          </button>
          <button @click="$emit('load-sample')" class="btn-small" :title="$t('main.editor.loadSample')">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
            </svg>
          </button>
          <!-- 新增：导入 .md（更新图标） -->
          <button @click="$emit('import-markdown')" class="btn-small" :title="$t('main.editor.importMd')">
            <svg viewBox="0 0 1024 1024" width="20" height="20" xmlns="http://www.w3.org/2000/svg">
              <path d="M628.1 357.9h107.4c2.8 0 5.7 0.7 8.1 2.4 18.4 13.1 18.2 34.3 6.4 46.3L532.4 627.5c-11.1 11.2-29.2 11.4-40.4 0.3l-0.3-0.3-217.6-220.9c-11.1-11.3-10.9-29.4 0.4-40.4 5.3-5.2 12.5-8.2 19.9-8.2H396V142.3c0-7.9 6.4-14.3 14.3-14.3h203.5c7.9 0 14.3 6.4 14.3 14.3v215.6zM283.4 652.4v87.5c0 7.9 6.4 14.3 14.3 14.3h428.8c7.9 0 14.3-6.4 14.3-14.3v-87.5c0-7.9 6.4-14.3 14.3-14.3H898c7.9 0 14.3 6.4 14.3 14.3V914c0 7.9-6.4 14.3-14.3 14.3H126.2c-7.9 0-14.3-6.4-14.3-14.3V652.4c0-7.9 6.4-14.3 14.3-14.3h142.9c7.9 0 14.3 6.4 14.3 14.3z" fill="currentColor"></path>
            </svg>
          </button>
        </div>
      </div>
      <MarkdownEditor
        :model-value="markdownContent"
        :sync-scroll-enabled="syncScrollEnabled"
        @update:model-value="$emit('update:markdown-content', $event)"
        class="editor-content"
      />

      <!-- 空状态引导卡片 -->
      <div v-if="!markdownContent && showEmptyGuide" class="empty-guide" @click="showEmptyGuide = false">
        <div class="empty-guide-card" @click.stop>
          <h3>{{ $t('main.emptyGuide.title') }}</h3>
          <p class="guide-subtitle">{{ $t('main.emptyGuide.subtitle') }}</p>

          <div class="guide-section">
            <h4>{{ $t('main.emptyGuide.shortcutsTitle') }}</h4>
            <ul class="guide-shortcuts">
              <li><kbd>⌘/Ctrl N</kbd><span>{{ $t('main.emptyGuide.shortcutNew') }}</span></li>
              <li><kbd>⌘/Ctrl O</kbd><span>{{ $t('main.emptyGuide.shortcutOpen') }}</span></li>
              <li><kbd>⌘/Ctrl S</kbd><span>{{ $t('main.emptyGuide.shortcutSave') }}</span></li>
              <li><kbd>⌘/Ctrl 1-3</kbd><span>{{ $t('main.emptyGuide.shortcutView') }}</span></li>
            </ul>
          </div>

          <div class="guide-section">
            <h4>{{ $t('main.emptyGuide.syntaxTitle') }}</h4>
            <div class="guide-syntax">
              <code>{{ $t('main.emptyGuide.syntaxHeading') }}</code>
              <code>{{ $t('main.emptyGuide.syntaxBold') }}</code>
              <code>{{ $t('main.emptyGuide.syntaxList') }}</code>
              <code>{{ $t('main.emptyGuide.syntaxCode') }}</code>
            </div>
          </div>

          <div class="guide-actions">
            <button class="btn-guide-primary" @click="$emit('load-sample')">{{ $t('main.emptyGuide.loadSample') }}</button>
            <button class="btn-guide-ghost" @click="showEmptyGuide = false">{{ $t('main.emptyGuide.close') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧：预览/WYSIWYG 面板 -->
    <div v-if="viewMode === 'both' || viewMode === 'preview' || viewMode === 'wysiwyg'" class="preview-panel">
      <div class="panel-header">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20">
            <path fill="currentColor" d="M12,9A3,3 0 0,0 9,12A3,3 0 0,0 12,15A3,3 0 0,0 15,12A3,3 0 0,0 12,9M12,17A5,5 0 0,1 7,12A5,5 0 0,1 12,7A5,5 0 0,1 17,12A5,5 0 0,1 12,17M12,4.5C7,4.5 2.73,7.61 1,12C2.73,16.39 7,19.5 12,19.5C17,19.5 21.27,16.39 23,12C21.27,7.61 17,4.5 12,4.5Z"/>
          </svg>
          {{ viewMode === 'wysiwyg' ? $t('main.preview.wysiwyg') : $t('main.preview.preview') }}
        </h3>
        <span v-if="viewMode === 'wysiwyg' && !wysiwygReady" class="wysiwyg-warming-badge">排版编辑器初始化中…</span>
        <div class="panel-actions" v-if="viewMode !== 'wysiwyg'">
          <!-- 同步滚动开关（移自 Footer） -->
          <label class="toggle-label" :title="$t('main.preview.syncScroll')">
            <span class="toggle-text">{{ $t('main.preview.syncScroll') }}</span>
            <input
              type="checkbox"
              :checked="syncScrollEnabled"
              @change="$emit('toggle-sync-scroll')"
              class="toggle-checkbox"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>

      <!-- 快速预览：非 wysiwyg 模式，或 wysiwyg 尚未就绪时显示（保证内容 ms 级可见） -->
      <PreviewPane
        v-show="viewMode !== 'wysiwyg' || !wysiwygReady"
        :markdown="markdownContent"
        :sync-scroll-enabled="syncScrollEnabled"
        @html-generated="(val) => viewMode !== 'wysiwyg' && $emit('html-generated', val)"
        class="preview-content"
      />
      <!-- 所见即所得（Milkdown）：后台预热，就绪后无缝切换 -->
      <WysiwygPane
        v-if="wysiwygEverMounted"
        v-show="viewMode === 'wysiwyg' && wysiwygReady"
        :model-value="markdownContent"
        :active="viewMode === 'wysiwyg'"
        @update:model-value="(val) => $emit('update:markdown-content', val)"
        @ready="wysiwygReady = true"
        class="preview-content"
      />
    </div>
  </main>
</template>

<script setup>
import PreviewPane from '../PreviewPane.vue'
import { ref, watch, defineAsyncComponent } from 'vue'
// MarkdownEditor 依赖 CodeMirror（体积大）：仅「分栏/源码」模式需要，按需加载。
// WYSIWYG 默认打开方式的用户启动时完全不加载 CodeMirror，冷启动更快
const MarkdownEditor = defineAsyncComponent(() => import('../MarkdownEditor.vue'))
// WysiwygPane 依赖 milkdown/prosemirror（体积大）：仅在切换到所见即所得模式时按需加载，
// 避免拖慢首屏与双击打开速度
const WysiwygPane = defineAsyncComponent(() => import('../WysiwygPane.vue'))

const props = defineProps({
  markdownContent: {
    type: String,
    required: true
  },
  syncScrollEnabled: {
    type: Boolean,
    default: true
  },
  viewMode: {
    type: String,
    default: 'both'
  },
  // 文件打开计数：新文件打开时 WYSIWYG 先回退到快速预览，等待 Milkdown 同步
  fileOpenTick: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits([
  'update:markdown-content',
  'clear-content',
  'load-sample',
  'html-generated',
  'import-markdown',
  'toggle-sync-scroll'
])

// 空状态引导卡片：首次进入且内容为空时显示；一旦开始输入内容即永久隐藏（一次性引导）
const showEmptyGuide = ref(true)
watch(() => props.markdownContent, (val) => {
  if (val) showEmptyGuide.value = false
})

// ==========================================
// WYSIWYG 混合模式：进入所见即所得时先用快速预览立刻显示内容，
// Milkdown 在后台预热（就绪后 emit ready），再无缝切换到所见即所得。
// ==========================================
const wysiwygReady = ref(false)
// 初始即可能处于所见即所得模式（启动时按默认打开方式进入），需预置挂载状态
const wysiwygEverMounted = ref(props.viewMode === 'wysiwyg')

watch(() => props.viewMode, (mode) => {
  if (mode === 'wysiwyg') {
    wysiwygEverMounted.value = true
    // 每次进入都重新走「预览 → 就绪 → 切换」流程，避免显示过期内容
    wysiwygReady.value = false
  }
})

// 新文件打开：WYSIWYG 模式先回退到快速预览，Milkdown 同步完成后由 ready 事件切回
watch(() => props.fileOpenTick, () => {
  if (props.viewMode === 'wysiwyg') {
    wysiwygReady.value = false
  }
})
</script>

<style scoped>
.app-main {
  flex: 1;
  display: flex;
  gap: var(--spacing-3xl);
  padding: var(--spacing-4xl);
  overflow: hidden;
  background: #f5f5f5;
  position: relative;
}

.editor-panel,
.preview-panel {
  flex: 1;
  min-width: 400px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.3s ease;
  position: relative;
}

.editor-panel:hover,
.preview-panel:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-color: #d0d0d0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #f8f8f8;
  border-bottom: 1px solid #e0e0e0;
  position: relative;
  overflow: visible; /* 允许下拉菜单溢出显示 */
  /* 无原生标题栏（TitleBarHiddenInset），面板头部作为窗口拖拽区域 */
  -webkit-app-region: drag;
}

/* 面板头部内的交互元素不参与拖拽，保持可点击 */
.panel-header .panel-actions,
.panel-header .toggle-label,
.panel-header .btn-small {
  -webkit-app-region: no-drag;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--theme-text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.panel-header h3 svg {
  opacity: 0.7;
}

/* 排版编辑器初始化徽标 */
.wysiwyg-warming-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--theme-text-secondary);
  background: var(--theme-bg-tertiary);
  border: 1px solid var(--theme-border-light);
  border-radius: 999px;
  padding: 3px 10px;
  margin-left: 8px;
  -webkit-app-region: no-drag;
}

.wysiwyg-warming-badge::before {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: 50%;
  border: 2px solid var(--theme-primary-25);
  border-top-color: var(--theme-primary);
  animation: wysiwyg-spin 0.8s linear infinite;
}

@keyframes wysiwyg-spin {
  to { transform: rotate(360deg); }
}

.panel-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 轻量版同步滚动开关样式，贴合面板头部高度 */
.toggle-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  transition: all 0.2s ease;
  padding: 4px 8px;
  border-radius: 12px;
  background: transparent;
}
.toggle-label:hover {
  color: var(--theme-primary);
  background: rgba(0,0,0,0.04);
}
.toggle-checkbox { display: none; }
.toggle-slider {
  position: relative;
  width: 32px;
  height: 18px;
  background: #e2e8f0;
  border-radius: 9px;
  transition: all 0.3s ease;
  box-shadow: inset 0 1px 2px rgba(0,0,0,0.1);
}
.toggle-slider::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 14px;
  height: 14px;
  background: #ffffff;
  border-radius: 50%;
  transition: all 0.3s ease;
  box-shadow: 0 1px 2px rgba(0,0,0,0.2);
}
.toggle-checkbox:checked + .toggle-slider { background: var(--theme-primary); }
.toggle-checkbox:checked + .toggle-slider::before { transform: translateX(14px); }

.btn-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--theme-border-light);
  border-radius: 6px;
  background: var(--theme-bg-primary);
  color: var(--theme-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.btn-small:hover {
  background: var(--theme-bg-secondary);
  color: var(--theme-text-primary);
  border-color: var(--theme-primary);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.btn-small:active {
  transform: translateY(0);
}

.editor-content,
.preview-content {
  flex: 1;
  overflow: hidden;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .app-main {
    flex-direction: column;
    gap: 20px;
    padding: 24px;
  }
  .editor-panel,
  .preview-panel {
    min-width: 100%;
    min-height: 400px;
  }
}

@media (max-width: 768px) {
  .panel-header {
    padding: 12px 16px;
  }
  
  .panel-header h3 {
    font-size: 14px;
  }
  
  .btn-small {
    width: 32px;
    height: 32px;
  }
  
  .btn-small svg {
    width: 16px;
    height: 16px;
  }
}



/* 滚动条样式 */
.editor-content::-webkit-scrollbar,
.preview-content::-webkit-scrollbar {
  width: 6px;
}

.editor-content::-webkit-scrollbar-track,
.preview-content::-webkit-scrollbar-track {
  background: var(--theme-bg-secondary);
}

.editor-content::-webkit-scrollbar-thumb,
.preview-content::-webkit-scrollbar-thumb {
  background: var(--theme-border-light);
  border-radius: 3px;
}

.editor-content::-webkit-scrollbar-thumb:hover,
.preview-content::-webkit-scrollbar-thumb:hover {
  background: var(--theme-primary);
}

/* 空状态引导卡片 */
.empty-guide {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(2px);
  z-index: 10;
  padding: 24px;
  overflow-y: auto;
}

.empty-guide-card {
  width: 100%;
  max-width: 380px;
  background: var(--theme-bg-primary, #ffffff);
  border-radius: 16px;
  border: 1px solid var(--theme-border-light, #e0e0e0);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  padding: 28px;
  text-align: center;
}

.empty-guide-card h3 {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 700;
  color: var(--theme-text-primary, #1e293b);
}

.guide-subtitle {
  margin: 0 0 20px;
  font-size: 13px;
  color: var(--theme-text-secondary, #64748b);
  line-height: 1.5;
}

.guide-section {
  text-align: left;
  margin-bottom: 18px;
}

.guide-section h4 {
  margin: 0 0 10px;
  font-size: 12px;
  font-weight: 600;
  color: var(--theme-text-secondary, #64748b);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.guide-shortcuts {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.guide-shortcuts li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--theme-text-primary, #1e293b);
}

.guide-shortcuts kbd {
  display: inline-block;
  min-width: 72px;
  padding: 3px 8px;
  font-family: inherit;
  font-size: 11px;
  text-align: center;
  color: var(--theme-text-secondary, #475569);
  background: var(--theme-bg-secondary, #f1f5f9);
  border: 1px solid var(--theme-border-light, #e2e8f0);
  border-bottom-width: 2px;
  border-radius: 6px;
}

.guide-syntax {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.guide-syntax code {
  padding: 4px 10px;
  font-size: 12px;
  color: var(--theme-primary, #3b82f6);
  background: var(--theme-primary-15, rgba(59, 130, 246, 0.1));
  border-radius: 8px;
}

.guide-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-top: 20px;
}

.btn-guide-primary {
  padding: 9px 20px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: var(--theme-primary, #3b82f6);
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-guide-primary:hover {
  background: var(--theme-primary-hover, #2563eb);
  transform: translateY(-1px);
}

.btn-guide-ghost {
  padding: 9px 16px;
  font-size: 13px;
  color: var(--theme-text-secondary, #64748b);
  background: transparent;
  border: 1px solid var(--theme-border-light, #e0e0e0);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-guide-ghost:hover {
  color: var(--theme-text-primary, #1e293b);
  border-color: var(--theme-text-secondary, #94a3b8);
}
</style>
