/**
 * @file src/composables/useOpenMode.js
 * @description 默认打开方式管理 Composable
 *
 * 通过双击 / 命令行 / Apple Events / 拖放等方式打开文件时，
 * 使用用户设置的默认视图模式（源码 / 分栏 / 排版）。
 *
 * 模块级单例 + localStorage 持久化，默认「排版」（所见即所得）。
 */

import { ref } from 'vue'

const OPEN_MODE_KEY = 'weditor.defaultOpenMode'

// 三种打开方式：源码(editor) / 分栏(both) / 排版(wysiwyg)
export const OPEN_MODES = [
  { value: 'editor', labelKey: 'settings.general.openMode.source' },
  { value: 'both', labelKey: 'settings.general.openMode.split' },
  { value: 'wysiwyg', labelKey: 'settings.general.openMode.typeset' }
]

export const DEFAULT_OPEN_MODE = 'wysiwyg'

const VALID_MODES = OPEN_MODES.map((m) => m.value)

const getStoredOpenMode = () => {
  try {
    const stored = localStorage.getItem(OPEN_MODE_KEY)
    if (stored && VALID_MODES.includes(stored)) return stored
  } catch (error) {
    console.warn('读取默认打开方式失败:', error)
  }
  return DEFAULT_OPEN_MODE
}

const defaultOpenMode = ref(getStoredOpenMode())

const setDefaultOpenMode = (mode) => {
  if (!VALID_MODES.includes(mode)) mode = DEFAULT_OPEN_MODE
  defaultOpenMode.value = mode
  try {
    localStorage.setItem(OPEN_MODE_KEY, mode)
  } catch (error) {
    console.warn('保存默认打开方式失败:', error)
  }
}

/**
 * 默认打开方式 Composable（模块级单例，跨组件共享状态）
 * @returns {Object} 默认打开方式状态与方法
 */
export function useOpenMode() {
  return {
    defaultOpenMode,
    setDefaultOpenMode,
    OPEN_MODES,
    DEFAULT_OPEN_MODE
  }
}
