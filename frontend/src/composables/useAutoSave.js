/**
 * @file src/composables/useAutoSave.js
 * @description 自动保存 Composable
 *
 * 功能：
 * - 监测内容变化，在指定延迟后自动保存到 localStorage（草稿）
 * - 在 Wails 桌面模式下，如果当前有打开的文件，同时保存到磁盘
 * - 提供草稿恢复功能
 * - 显示保存状态指示
 */
import { ref, watch, onBeforeUnmount } from 'vue'

const DRAFT_KEY = 'weditor-autosave-draft'
const DRAFT_TIME_KEY = 'weditor-autosave-draft-time'
const DEFAULT_DELAY = 2000 // ms

export function useAutoSave(contentRef, options = {}) {
  const {
    delay = DEFAULT_DELAY,
    enabled = true,
    saveToDisk = null,  // 可选: 异步保存到磁盘的函数 (content) => Promise
    filePath = null,    // 可选: reactrive ref for current file path
  } = options

  const lastSavedAt = ref(null)
  const isSaving = ref(false)
  const isDraftAvailable = ref(false)
  const draftTime = ref(null)

  let saveTimer = null

  // 检查是否有草稿
  function checkDraft() {
    try {
      const draft = localStorage.getItem(DRAFT_KEY)
      const time = localStorage.getItem(DRAFT_TIME_KEY)
      if (draft !== null && draft.trim() !== '') {
        isDraftAvailable.value = true
        draftTime.value = time
        return draft
      }
    } catch (e) {
      console.warn('读取草稿失败:', e)
    }
    isDraftAvailable.value = false
    return null
  }

  // 恢复草稿
  function recoverDraft() {
    try {
      const draft = localStorage.getItem(DRAFT_KEY)
      if (draft !== null && draft.trim() !== '') {
        clearDraft()
        return draft
      }
    } catch (e) {
      console.warn('恢复草稿失败:', e)
    }
    return null
  }

  // 清除草稿
  function clearDraft() {
    try {
      localStorage.removeItem(DRAFT_KEY)
      localStorage.removeItem(DRAFT_TIME_KEY)
      isDraftAvailable.value = false
      draftTime.value = null
    } catch (e) {
      // 忽略
    }
  }

  // 保存草稿
  async function saveDraft(content) {
    if (!enabled) return
    if (!content || content.trim() === '') return

    try {
      localStorage.setItem(DRAFT_KEY, content)
      const now = new Date().toLocaleString()
      localStorage.setItem(DRAFT_TIME_KEY, now)
      lastSavedAt.value = now

      // 如果提供了磁盘保存函数且当前有文件路径，同时保存到磁盘
      if (saveToDisk && filePath && filePath.value) {
        isSaving.value = true
        try {
          await saveToDisk(content)
          // 内容已持久化到磁盘，本地草稿冗余，清除避免下次启动误弹“未保存草稿”提示
          clearDraft()
        } catch (e) {
          console.warn('自动保存到磁盘失败:', e)
        } finally {
          isSaving.value = false
        }
      }
    } catch (e) {
      console.warn('自动保存草稿失败:', e)
    }
  }

  // 启用自动保存监听
  if (enabled && contentRef) {
    watch(
      () => contentRef.value,
      (newContent) => {
        if (saveTimer) clearTimeout(saveTimer)

        saveTimer = setTimeout(() => {
          saveDraft(newContent)
        }, delay)
      },
      { immediate: false }
    )
  }

  // 组件卸载时清理
  onBeforeUnmount(() => {
    if (saveTimer) {
      clearTimeout(saveTimer)
      // 最后一次保存
      if (contentRef && contentRef.value) {
        saveDraft(contentRef.value)
      }
    }
  })

  // 初始化：检查草稿
  checkDraft()

  return {
    lastSavedAt,
    isSaving,
    isDraftAvailable,
    draftTime,

    checkDraft,
    recoverDraft,
    clearDraft,
    saveDraft,
  }
}
