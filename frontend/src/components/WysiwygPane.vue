<template>
  <div class="wysiwyg-pane">
    <div class="preview-toolbar">
      <div class="preview-mode-selector">
        <label class="mode-label">所见即所得</label>
      </div>
      <div class="viewport-info"></div>
    </div>

    <div class="preview-container" :class="[`theme-system-${currentLayoutId}`]">
      <div
        ref="editorRoot"
        class="preview-rendered markdown-body modern-markdown wysiwyg-rendered"
        :class="getPreviewClasses()"
      ></div>
    </div>
  </div>
</template>

<script>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useGlobalThemeManager } from '../composables/index.js'

// Milkdown core & plugins
import { Editor, defaultValueCtx, rootCtx } from '@milkdown/core'
import { commonmark } from '@milkdown/preset-commonmark'
import { gfm } from '@milkdown/preset-gfm'
import { clipboard } from '@milkdown/plugin-clipboard'
import { history } from '@milkdown/plugin-history'
import { listener, listenerCtx } from '@milkdown/plugin-listener'
import { replaceAll, getMarkdown } from '@milkdown/utils'
import { prism } from '@milkdown/plugin-prism'
import '../plugins/prism-setup.js'

export default {
  name: 'WysiwygPane',
  props: {
    modelValue: {
      type: String,
      default: ''
    },
    // 是否处于激活（所见即所得）状态：激活时同步最新内容并发出 ready
    active: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:modelValue', 'ready'],
  setup(props, { emit }) {
    const editorRoot = ref(null)
    const editorRef = ref(null)
    const isApplyingExternal = ref(false)
    const hasFocus = ref(false)
    const lastEmittedMarkdown = ref(props.modelValue || '')
    const lastEmittedNormalized = ref('')
    let emitTimer = null

    const toPlainMarkdown = (md = '') => {
      return (md || '')
        // Convert HTML <br> tags into newlines
        .replace(/<br\s*\/?>/gi, '\n')
        // In case of <br ...> with attributes (rare)
        .replace(/<br[^>]*>/gi, '\n')
        // Convert non-breaking spaces to normal spaces (named & numeric)
        .replace(/\u00a0|&nbsp;|&#xA0;|&#160;/gi, ' ')
        // Convert numeric space entities used by remark/stringify to preserve spaces
        .replace(/&#x20;|&#32;/gi, ' ')
    }

    const normalizeMarkdown = (md = '') => (md || '')
      .replace(/\r\n/g, '\n')
      .replace(/\r/g, '\n')

    // Don't use BR placeholders - let Milkdown handle blank lines natively
    const markdownToEditorInput = (md = '') => {
      return toPlainMarkdown(md)
    }

    // init normalized cache
    lastEmittedNormalized.value = normalizeMarkdown(lastEmittedMarkdown.value)

    const themeManager = useGlobalThemeManager()
    const { currentThemeSystemId: currentLayoutId, currentColorTheme, currentCodeStyle, currentFontSettings, initialize } = themeManager

    const getPreviewClasses = () => {
      return {
        [`theme-system-${currentLayoutId.value}`]: true,
        [`code-style-${currentCodeStyle.value?.id || 'default'}`]: true
      }
    }

    const createEditor = async () => {
      if (!editorRoot.value) return

      const editor = await Editor.make()
        .config((ctx) => {
          ctx.set(rootCtx, editorRoot.value)
          ctx.set(defaultValueCtx, markdownToEditorInput(props.modelValue || ''))
        })
        .use(commonmark)
        .use(gfm)
        .use(prism)
        .use(history)
        .use(clipboard)
        .use(listener)
        // Mermaid NodeView plugin
        .use((await import('../plugins/mermaid-nodeview.js')).mermaidNodeViewPlugin)
        .config((ctx) => {
          const l = ctx.get(listenerCtx)
          l.focus(() => { hasFocus.value = true })
          l.blur(() => { hasFocus.value = false })
          l.markdownUpdated((_, markdown) => {
            if (isApplyingExternal.value) return
            const plain = toPlainMarkdown(markdown)
            const normalized = normalizeMarkdown(plain)
            if (normalized === lastEmittedNormalized.value) return
            if (emitTimer) clearTimeout(emitTimer)
            emitTimer = setTimeout(() => {
              lastEmittedMarkdown.value = plain
              lastEmittedNormalized.value = normalized
              emit('update:modelValue', plain)
            }, 60)
          })
        })
        .create()

      editorRef.value = editor
    }

    // 外部内容同步串行化：避免多次拖入/输入期间并发执行 replaceAll，
    // 导致 ProseMirror 状态错乱（乱码）；同步期间的新内容按序排队，
    // 过期的同步结果不再发 ready（由最新一次负责）
    let syncChain = Promise.resolve()
    let syncVersion = 0

    const doApplyModelValue = async (val) => {
      const editor = editorRef.value
      if (!editor) return true
      const incoming = normalizeMarkdown(val || '')
      if (incoming === lastEmittedNormalized.value) return true // 已同步
      const current = await editor.action(getMarkdown())
      const currentNorm = normalizeMarkdown(toPlainMarkdown(current || ''))
      const incomingPlain = markdownToEditorInput(val || '')
      const incomingNorm = normalizeMarkdown(incomingPlain)
      if (currentNorm !== incomingNorm) {
        isApplyingExternal.value = true
        try {
          await editor.action(replaceAll(incomingPlain))
          lastEmittedMarkdown.value = incomingPlain
          lastEmittedNormalized.value = incomingNorm
        } finally {
          setTimeout(() => { isApplyingExternal.value = false }, 30)
        }
      }
      return true
    }

    // 返回 Promise<boolean>：本次同步是否为最新内容（是则应由它发 ready）
    const applyModelValue = (val) => {
      const version = ++syncVersion
      syncChain = syncChain
        .catch(() => {}) // 前一次失败不阻塞后续
        .then(async () => {
          await doApplyModelValue(val)
          return version === syncVersion
        })
      return syncChain
    }

    // 外部内容变化：同步到编辑器；激活状态下同步完成后通知 ready
    watch(
      () => props.modelValue,
      async (val, oldVal) => {
        if (val === oldVal) return
        const synced = await applyModelValue(val)
        if (props.active && synced) emit('ready')
      }
    )

    // 进入激活状态：确保内容同步后发出 ready（编辑器未创建时由 onMounted 负责）
    watch(() => props.active, async (act) => {
      if (!act) return
      if (editorRef.value) {
        await applyModelValue(props.modelValue)
        emit('ready')
      }
    })

    // React to theme changes; actual variables are applied globally
    watch([currentColorTheme, currentCodeStyle, currentFontSettings, currentLayoutId], () => {}, { deep: true })

    onMounted(async () => {
      initialize()
      await createEditor()
      setupInlineCodeObserver()
      // 编辑器已创建并持有最新内容，通知父组件可以切换到所见即所得
      emit('ready')
    })

    onBeforeUnmount(() => {
      const editor = editorRef.value
      if (editor && typeof editor.destroy === 'function') {
        editor.destroy()
      }
      editorRef.value = null
      if (inlineCodeObserver) {
        inlineCodeObserver.disconnect()
        inlineCodeObserver = null
      }
      // 清理可能未触发的节流定时器，避免内存泄漏
      if (emitTimer) {
        clearTimeout(emitTimer)
        emitTimer = null
      }
    })



    // WYSIWYG 表格兼容处理：使“语法”列中的转义反引号整体以 code 呈现
    let inlineCodeObserver = null

    const normalizeWysiwygTables = () => {
      const root = editorRoot.value
      if (!root) return
      const cells = root.querySelectorAll('.wysiwyg-rendered table td, .wysiwyg-rendered table th')
      cells.forEach((cell) => {
        // 若单元格已是“仅一个 <code>...</code> 且没有其他文本”的情况，直接跳过
        if (cell.children.length === 1 && cell.children[0].tagName?.toLowerCase() === 'code') {
          const onlyCode = (cell.textContent || '').trim() === (cell.children[0].textContent || '').trim()
          if (onlyCode) return
        }

        const plain = (cell.textContent || '').trim()
        const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

        // 模式1：标准形式 \`text\`
        const m1 = plain.match(/^\\`([^`]+)\\`$/)
        if (m1) {
          cell.innerHTML = `<code>\`${esc(m1[1])}\`</code>`
          return
        }

        // 模式2：Milkdown 断裂形式 "\\" + text + "``"（光标在末尾时尤为常见）
        // e.g. "\\code``" 或 "\\ code``" -> <code>`code`</code>
        const m2 = plain.match(/^\\\s*([^`]+)``$/)
        if (m2) {
          cell.innerHTML = `<code>\`${esc(m2[1].trim())}\`</code>`
          return
        }

        // 模式3：包含 \` 且至少两个反引号 —— 将整格包裹为 code 胶囊（用于更宽泛的兼容）
        const backticks = (plain.match(/`/g) || []).length
        if (/\\`/.test(plain) && backticks >= 2) {
          cell.innerHTML = `<code>${esc(plain)}</code>`
          return
        }
      })
    }

    const setupInlineCodeObserver = () => {
      const root = editorRoot.value
      if (!root || inlineCodeObserver) return
      inlineCodeObserver = new MutationObserver(() => {
        requestAnimationFrame(normalizeWysiwygTables)
      })
      inlineCodeObserver.observe(root, { childList: true, characterData: true, subtree: true })
      // 初始执行一次
      normalizeWysiwygTables()
    }


    return {
      editorRoot,
      currentLayoutId,
      getPreviewClasses
    }
  }
}
</script>

<style scoped>
/* Fill available height like PreviewPane */
.wysiwyg-pane,
.preview-container,
.wysiwyg-rendered {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Content area look & scroll */
.wysiwyg-rendered {
  padding: 16px;
  overflow: auto;
}

/* Mermaid NodeView styles */
.wysiwyg-rendered :deep(.md-mermaid) {
  position: relative;
  margin: 0.75rem 0;
  border-radius: 12px;
  border: 1px solid var(--theme-border-light, #e5e7eb);
  background: var(--theme-bg-primary, #fff);
  overflow: hidden;
}
.wysiwyg-rendered :deep(.md-mermaid.is-selected) { outline: 2px solid #60a5fa; outline-offset: -2px; }
.wysiwyg-rendered :deep(.md-mermaid__toolbar) {
  display: flex;
  justify-content: flex-end;
  gap: .5rem;
  padding: .25rem .5rem;
  border-bottom: 1px solid var(--theme-border-light, #e5e7eb);
  background: var(--theme-bg-secondary, #f9fafb);
}
.wysiwyg-rendered :deep(.md-mermaid__btn) {
  font: inherit;
  padding: 4px 10px;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #fff;
  cursor: pointer;
}
.wysiwyg-rendered :deep(.md-mermaid__btn:hover) { background: #f3f4f6; }
.wysiwyg-rendered :deep(.md-mermaid__svg) { padding: 12px; overflow: auto; }
.wysiwyg-rendered :deep(.md-mermaid__source) { margin: 0; padding: 8px 12px; overflow: auto; }
.wysiwyg-rendered :deep(.md-mermaid__error) { color: #b91c1c; white-space: pre-wrap; }
.wysiwyg-rendered :deep(.md-mermaid__empty) { color: #6b7280; padding: 12px; }
.wysiwyg-rendered :deep(.md-mermaid__source) { display: none; }
.wysiwyg-rendered :deep(.md-mermaid[data-editing="true"] .md-mermaid__svg) { display: none; }
.wysiwyg-rendered :deep(.md-mermaid[data-editing="true"] .md-mermaid__source) { display: block; }


/* HR 分隔线（WYSIWYG 对齐预览页当前样式） */
.wysiwyg-rendered :deep(hr) {
  border: none !important;
  height: 2px !important;
  background: linear-gradient(to right, transparent, var(--theme-primary), transparent) !important;
  margin: 2rem 0 !important;
}

/* ProseMirror baseline; rely on markdown-body for tag styles */
:deep(.ProseMirror) {
  outline: none;
  white-space: pre-wrap;
  min-height: 100%;
}

/* Tables/blocks consistency */
:deep(table) {
  width: 100%;
  border-collapse: collapse;
}

/* 表头与单元格样式（WYSIWYG 与预览页对齐） */
.wysiwyg-rendered :deep(th),
.wysiwyg-rendered :deep(td) {
  border: 1px solid var(--theme-border-light) !important;
  padding: 8px 12px !important;
  font-size: var(--markdown-font-size, 16px) !important;
  font-family: var(--markdown-font-family, var(--theme-font-family)) !important;
  line-height: var(--markdown-line-height, 1.6) !important;
  color: var(--theme-text-primary) !important;
}

.wysiwyg-rendered :deep(th) {
  background-color: var(--theme-bg-secondary) !important;
  font-weight: 600 !important;
}
/* 保障表头加粗的多重兜底（适配 Milkdown 结构差异） */
.wysiwyg-rendered :deep(thead th),
.wysiwyg-rendered :deep(tr:first-child > th),
.wysiwyg-rendered :deep(tr:first-child > td) {
  font-weight: 600 !important;
  background-color: var(--theme-bg-secondary) !important;
}
/* ProseMirror 在单元格内包裹 p，导致 th 的字重被覆盖，这里强制加粗 */
.wysiwyg-rendered :deep(th > p) { font-weight: 600 !important; margin: 0 !important; }

/* 强调文本样式 - 与预览页面保持一致 */
.wysiwyg-rendered :deep(strong),
.wysiwyg-rendered :deep(b) {
  font-weight: 600;
  color: var(--theme-primary) !important;  /* 使用主题色 */
}

.wysiwyg-rendered :deep(em),
.wysiwyg-rendered :deep(i) {
  font-style: italic;
  color: var(--theme-text-secondary) !important;
}

/* 行内代码样式 - 对齐预览页的颜色主题（使用 inline-code 变量） */
.wysiwyg-rendered :deep(code:not(pre code)) {
  background-color: var(--theme-inline-code-bg) !important;
  color: var(--theme-inline-code-text) !important;
  padding: 0.2rem 0.4rem !important;
  border-radius: var(--radius-sm) !important;
  font-family: var(--theme-code-font-family) !important;
  font-size: 14px !important;
  border: 1px solid var(--theme-inline-code-border) !important;
}

/* 代码块样式 */
.wysiwyg-rendered :deep(pre) {
  background: var(--code-bg, var(--theme-code-bg)) !important;
  color: var(--code-color, var(--theme-code-text)) !important;
  padding: var(--code-padding, 16px) !important;
  border-radius: var(--code-border-radius, var(--radius-md)) !important;
  overflow-x: auto !important;
  margin: 1.5rem 0 !important;
  border: var(--code-border, 1px solid var(--theme-code-border)) !important;
  position: relative !important;
}

/* Mac-style header bar & traffic lights (pure CSS, no DOM changes) */
.wysiwyg-rendered.code-style-mac :deep(pre)::before {
  content: '';
  display: block;
  height: 28px;
  margin: calc(var(--code-padding, 16px) * -1) calc(var(--code-padding, 16px) * -1) 12px calc(var(--code-padding, 16px) * -1); /* cancel pre padding dynamically */
  border-radius: var(--code-border-radius, var(--radius-md)) var(--code-border-radius, var(--radius-md)) 0 0;
  background: var(--code-bg, var(--theme-code-bg));
  border-bottom: none;
}

.wysiwyg-rendered.code-style-mac :deep(pre)::after {
  content: '';
  position: absolute;
  top: 12px;
  left: 20px;
  width: 8px;   /* smaller: half size */
  height: 8px;  /* smaller: half size */
  border-radius: 50%;
  background: #ff5f56; /* red */
  /* increase spacing: double the offsets */
  box-shadow: 14px 0 0 #ffbd2e, 28px 0 0 #27c93f; /* yellow, green */
  opacity: 0.95;
}

/* GitHub style header */
.wysiwyg-rendered.code-style-github :deep(pre)::before {
  content: '';
  display: block;
  height: 32px;
  margin: calc(var(--code-padding, 16px) * -1) calc(var(--code-padding, 16px) * -1) 12px calc(var(--code-padding, 16px) * -1);
  border-radius: var(--code-border-radius, var(--radius-md)) var(--code-border-radius, var(--radius-md)) 0 0;
  background: #f1f3f4;
  border-bottom: 1px solid #d0d7de;
}
.wysiwyg-rendered.code-style-github :deep(pre)::after {
  content: '📄 代码';
  position: absolute;
  top: 8px;
  left: 16px;
  font-size: 12px;
  color: #656d76;
}

/* VS Code style header */
.wysiwyg-rendered.code-style-vscode :deep(pre)::before {
  content: '';
  display: block;
  height: 32px;
  margin: calc(var(--code-padding, 16px) * -1) calc(var(--code-padding, 16px) * -1) 12px calc(var(--code-padding, 16px) * -1);
  border-radius: var(--code-border-radius, var(--radius-md)) var(--code-border-radius, var(--radius-md)) 0 0;
  background: linear-gradient(135deg, #2d2d30 0%, #3c3c3c 100%);
  border-bottom: 1px solid #3c3c3c;
}
.wysiwyg-rendered.code-style-vscode :deep(pre)::after {
  content: '⚡ 代码片段';
  position: absolute;
  top: 8px;
  left: 16px;
  font-size: 13px;
  color: #cccccc;
}

/* Terminal style header */
.wysiwyg-rendered.code-style-terminal :deep(pre)::before {
  content: '';
  display: block;
  height: 28px;
  margin: calc(var(--code-padding, 16px) * -1) calc(var(--code-padding, 16px) * -1) 12px calc(var(--code-padding, 16px) * -1);
  border-radius: var(--code-border-radius, var(--radius-md)) var(--code-border-radius, var(--radius-md)) 0 0;
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
}
.wysiwyg-rendered.code-style-terminal :deep(pre)::after {
  content: '$ terminal';
  position: absolute;
  top: 8px;
  left: 16px;
  font-size: 12px;
  color: #00ff00;
  font-family: 'Courier New', monospace;
}

.wysiwyg-rendered :deep(pre code) {
  background: none !important;
  padding: 0 !important;
  border-radius: 0 !important;
  border: none !important;
  font-size: 14px !important;
  /* Do not set explicit color here; allow Prism token colors to take effect */
  font-family: var(--code-font-family, var(--theme-code-font-family)) !important;
}

/* 链接样式 - 与预览页面保持一致 */
.wysiwyg-rendered :deep(a) {
  color: var(--theme-primary) !important;
  text-decoration: none !important;
  border-bottom: 1px solid transparent !important;
  transition: border-bottom-color 0.2s ease !important;
}

.wysiwyg-rendered :deep(a:hover) {
  border-bottom-color: var(--theme-primary) !important;
}

/* 删除线样式 */
.wysiwyg-rendered :deep(del),
.wysiwyg-rendered :deep(s),
.wysiwyg-rendered :deep(strike) {
  text-decoration: line-through !important;
  color: var(--theme-text-tertiary) !important;
}

/* 列表基础样式 - LI-based Indentation */
.wysiwyg-rendered :deep(ul),
.wysiwyg-rendered :deep(ol) {
  margin: 1rem 0 !important;
  padding-left: 0.5rem !important; /* Indentation is handled by LI elements */
  list-style: none !important;
}

.wysiwyg-rendered :deep(li) {
  margin: 0.5rem 0 !important;
  padding-left: 2rem !important; /* Default indentation for regular lists */
  position: relative !important;
}

.wysiwyg-rendered :deep(li > p) {
    display: inline !important;
    margin: 0 !important;
}

/* 基础项目符号（所有非任务列表） */
.wysiwyg-rendered :deep(li:not([data-item-type='task']))::before {
  position: absolute !important;
  top: 0;
  left: 0; /* Position at the start of the LI's padding */
  width: 2rem; /* Occupy the full padding width */
  height: 1.7em; /* Match line-height */
  line-height: 1.7em; /* Align with text */
  font-weight: 600 !important;
  color: var(--theme-primary) !important;
  text-align: center;
}

/* 无序列表 - 排除任务列表 */
.wysiwyg-rendered :deep(ul > li:not([data-item-type='task']))::before {
  content: '●' !important;
  font-size: 0.9em !important;
}

.wysiwyg-rendered :deep(ul > li:not([data-item-type='task']) > ul > li)::before {
  content: '○' !important;
  font-size: 0.8em !important;
}

.wysiwyg-rendered :deep(ul > li:not([data-item-type='task']) > ul > li > ul > li)::before {
  content: '▪' !important;
  font-size: 0.9em !important;
}

/* 有序列表 */
.wysiwyg-rendered :deep(ol) {
  counter-reset: list-counter !important;
}

.wysiwyg-rendered :deep(ol > li:not([data-item-type='task'])) {
  counter-increment: list-counter !important;
}

.wysiwyg-rendered :deep(ol > li:not([data-item-type='task']))::before {
  content: counter(list-counter) '.' !important;
}

.wysiwyg-rendered :deep(ol > li > ol) {
  counter-reset: list-counter-2 !important;
}

.wysiwyg-rendered :deep(ol > li > ol > li:not([data-item-type='task'])) {
  counter-increment: list-counter-2 !important;
}

.wysiwyg-rendered :deep(ol > li > ol > li:not([data-item-type='task']))::before {
  content: counter(list-counter-2, lower-alpha) '.' !important;
  opacity: 0.8 !important;
}

.wysiwyg-rendered :deep(ol > li > ol > li > ol) {
  counter-reset: list-counter-3 !important;
}

.wysiwyg-rendered :deep(ol > li > ol > li > ol > li:not([data-item-type='task'])) {
  counter-increment: list-counter-3 !important;
}

.wysiwyg-rendered :deep(ol > li > ol > li > ol > li:not([data-item-type='task']))::before {
  content: '(' counter(list-counter-3) ')' !important;
  opacity: 0.6 !important;
}

/* 任务列表 - LI-based Indentation */
.wysiwyg-rendered :deep(li[data-item-type="task"]) {
  padding-left: 0.5rem !important; /* Specific, smaller indentation */
  display: flex !important;
  align-items: center !important;
  margin: 0.25em 0 !important;
}

.wysiwyg-rendered :deep(li[data-item-type="task"])::before {
  content: '' !important;
  display: inline-block !important;
  width: 0.9em !important;
  height: 0.9em !important;
  border: 1px solid var(--theme-border-medium) !important;
  border-radius: var(--radius-sm) !important;
  background-color: var(--theme-bg-primary) !important;
  margin-right: 0.5em;
  flex-shrink: 0;
  position: static !important;
}

.wysiwyg-rendered :deep(li[data-item-type="task"][data-checked="true"])::before {
  background-color: var(--theme-primary) !important;
  border-color: var(--theme-primary) !important;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3e%3cpath fill-rule='evenodd' clip-rule='evenodd' d='M12 5L7 10L4 7L3 8L7 12L13 6L12 5Z' fill='white'/%3e%3c/svg%3e") !important;
  background-repeat: no-repeat !important;
  background-position: center !important;
}

.wysiwyg-rendered :deep(li[data-item-type="task"][data-checked="true"] > p) {
  text-decoration: line-through !important;
  color: var(--theme-text-secondary) !important;
  opacity: 0.7 !important;
}

/* 引用块样式 - 对齐预览页面（带斜向柔和渐变背景 + 左侧竖线圆弧效果） */
.wysiwyg-rendered :deep(blockquote) {
  margin: 1.5rem 0 !important;
  padding: 1rem 1.5rem !important;
  border-left: 4px solid var(--theme-primary) !important;
  /* 与预览页 ThemeProcessor 输出保持一致的条纹渐变 */
  background: linear-gradient(
    135deg,
    rgba(var(--theme-primary-rgb), 0.08) 0%,
    rgba(var(--theme-primary-rgb), 0.04) 50%,
    rgba(var(--theme-primary-rgb), 0.08) 100%
  ) !important;
  color: var(--theme-text-secondary) !important;
  font-style: italic !important;
  /* 关键：为整个容器添加圆角，令左侧竖线呈“括号”圆弧端点 */
  border-radius: var(--radius-md) !important;
  /* 轻微投影，接近预览页 legacy 配置的 0 4px 16px */
  box-shadow: 0 4px 16px rgba(var(--theme-primary-rgb), 0.10) !important;
  position: relative;
}

.wysiwyg-rendered :deep(blockquote blockquote) {
  margin-top: 1rem !important;
  margin-bottom: 1rem !important;
  padding-left: 1.5rem !important;
  border-left-color: var(--theme-primary) !important;
  background: linear-gradient(
    135deg,
    rgba(var(--theme-primary-rgb), 0.08) 0%,
    rgba(var(--theme-primary-rgb), 0.04) 50%,
    rgba(var(--theme-primary-rgb), 0.08) 100%
  ) !important;
  border-radius: var(--radius-md) !important;
  box-shadow: 0 2px 8px rgba(var(--theme-primary-rgb), 0.08) !important;
}

.wysiwyg-rendered :deep(blockquote p) {
  margin: 0.5rem 0 !important;
}

.wysiwyg-rendered :deep(blockquote p:first-child) {
  margin-top: 0 !important;
}

.wysiwyg-rendered :deep(blockquote p:last-child) {
  margin-bottom: 0 !important;
}

/* 标题样式补充 */
.wysiwyg-rendered :deep(h1),
.wysiwyg-rendered :deep(h2),
.wysiwyg-rendered :deep(h3),
.wysiwyg-rendered :deep(h4),
.wysiwyg-rendered :deep(h5),
.wysiwyg-rendered :deep(h6) {
  color: var(--theme-text-primary) !important;
  font-family: var(--markdown-font-family, var(--theme-font-family)) !important;
  line-height: var(--markdown-line-height, 1.6) !important;
  margin-top: 1.5rem !important;
  margin-bottom: 1rem !important;
}

.wysiwyg-rendered :deep(h1) {
  font-size: calc(var(--markdown-font-size, 16px) * 2) !important;
  font-weight: 700 !important;
}

.wysiwyg-rendered :deep(h2) {
  font-size: calc(var(--markdown-font-size, 16px) * 1.5) !important;
  font-weight: 600 !important;
}

.wysiwyg-rendered :deep(h3) {
  font-size: calc(var(--markdown-font-size, 16px) * 1.3) !important;
  font-weight: 600 !important;
}

.wysiwyg-rendered :deep(h4) {
  font-size: calc(var(--markdown-font-size, 16px) * 1.1) !important;
  font-weight: 600 !important;
}

.wysiwyg-rendered :deep(h5) {
  font-size: calc(var(--markdown-font-size, 16px) * 1) !important;
  font-weight: 600 !important;
}

.wysiwyg-rendered :deep(h6) {
  font-size: calc(var(--markdown-font-size, 16px) * 0.9) !important;
  font-weight: 600 !important;
}
</style>

