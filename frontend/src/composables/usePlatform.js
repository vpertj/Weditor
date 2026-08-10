/**
 * @file src/composables/usePlatform.js
 * @description 平台检测 Composable
 *
 * 提供跨平台能力检测和快捷键显示适配：
 * - macOS: 显示 ⌘ 键
 * - Windows/Linux: 显示 Ctrl 键
 */
import { ref } from 'vue'

const platform = ref('unknown')
const isMac = ref(false)
const isWindows = ref(false)
const isLinux = ref(false)
const modKey = ref('Ctrl')
const modSymbol = ref('Ctrl')

function detect() {
  const ua = navigator.userAgent || navigator.platform || ''

  if (ua.indexOf('Mac') !== -1 && ua.indexOf('Mobile') === -1) {
    isMac.value = true
    isWindows.value = false
    isLinux.value = false
    platform.value = 'macOS'
    modKey.value = 'Cmd'
    modSymbol.value = '⌘'
  } else if (ua.indexOf('Win') !== -1) {
    isMac.value = false
    isWindows.value = true
    isLinux.value = false
    platform.value = 'Windows'
    modKey.value = 'Ctrl'
    modSymbol.value = 'Ctrl'
  } else if (ua.indexOf('Linux') !== -1 || ua.indexOf('X11') !== -1) {
    isMac.value = false
    isWindows.value = false
    isLinux.value = true
    platform.value = 'Linux'
    modKey.value = 'Ctrl'
    modSymbol.value = 'Ctrl'
  } else {
    // 在 Wails 环境下可以从 Go 后端获取平台信息
    platform.value = 'unknown'
    modKey.value = 'Ctrl'
    modSymbol.value = 'Ctrl'
  }
}

/**
 * 将快捷键文本适配到当前平台
 * 例如: "Ctrl+B" → "⌘B" (macOS) 或 "Ctrl+B" (Windows)
 * @param {string} shortcut - 原始快捷键文本
 * @returns {string} 平台适配后的快捷键文本
 */
function formatShortcut(shortcut) {
  if (!shortcut) return ''
  if (isMac.value) {
    return shortcut
      .replace(/Ctrl\+/gi, '⌘')
      .replace(/Shift\+/gi, '⇧')
      .replace(/Alt\+/gi, '⌥')
      .replace(/Cmd\+/gi, '⌘')
      .replace(/Enter/gi, '↵')
      .replace(/Escape/gi, '⎋')
      .replace(/Delete/gi, '⌫')
      .replace(/Tab/gi, '⇥')
  }
  return shortcut
}

/**
 * 生成平台适配的快捷键显示
 * @param {string} key - 按键名称
 * @param {Object} modifiers - 修饰键 { shift?: boolean, alt?: boolean }
 * @returns {string} 格式化后的快捷键
 */
function shortcut(modifiers = {}, key) {
  const parts = []
  if (isMac.value) {
    if (modifiers.shift) parts.push('⇧')
    if (modifiers.alt) parts.push('⌥')
    parts.push('⌘')
  } else {
    parts.push('Ctrl')
    if (modifiers.shift) parts.push('+Shift')
    if (modifiers.alt) parts.push('+Alt')
  }
  parts.push(key.toUpperCase())
  return isMac.value ? parts.join('') : parts.join('+')
}

// 初始化检测
detect()

export function usePlatform() {
  return {
    platform,
    isMac,
    isWindows,
    isLinux,
    modKey,
    modSymbol,
    formatShortcut,
    shortcut,
    detect,
  }
}
