/**
 * @file src/composables/useWindowDrag.js
 * @description 窗口拖拽兜底方案
 *
 * WKWebView 对 -webkit-app-region: drag 的支持并不可靠（Wails 无原生处理，
 * 完全依赖 WebKit 行为）。此 composable 通过 JS 监听拖拽区域 mousedown/mousemove，
 * 调用 Wails runtime WindowSetPosition 实现窗口移动：
 * - 若 WebKit 支持 CSS drag，事件被原生拦截，JS 不会触发（CSS 生效）
 * - 若 WebKit 不支持 CSS drag，事件到达 JS，由本方案兜底拖动（JS 生效）
 */
import { onMounted, onBeforeUnmount } from 'vue'

// 可拖拽区域（需与样式中的 -webkit-app-region: drag 保持一致）
const DRAG_SELECTOR = '.window-drag-strip, .panel-header, .app-header'
// 拖拽区域内不可触发拖拽的交互元素
const NO_DRAG_SELECTOR =
  'button, a, input, textarea, select, label, .panel-actions, .toggle-label, .btn, .btn-small, .dropdown-trigger, .header-btn, .view-toggle-btn, .draft-banner'

export function useWindowDrag() {
  let dragging = false
  let startScreenX = 0
  let startScreenY = 0
  let winX = 0
  let winY = 0
  let winPosReady = false

  function onMouseDown(e) {
    // 仅左键
    if (e.button !== 0) return
    const rt = window.runtime
    if (!rt || typeof rt.WindowSetPosition !== 'function') return

    const target = e.target
    if (!(target instanceof Element)) return
    // 必须位于拖拽区域内
    if (!target.closest(DRAG_SELECTOR)) return
    // 交互元素上不触发拖拽
    if (target.closest(NO_DRAG_SELECTOR)) return

    e.preventDefault()
    dragging = true
    winPosReady = false
    startScreenX = e.screenX
    startScreenY = e.screenY

    // 异步获取窗口当前位置（Retina 下与 e.screenX/Y 同为逻辑坐标）
    if (typeof rt.WindowGetPosition === 'function') {
      Promise.resolve(rt.WindowGetPosition())
        .then((pos) => {
          if (pos && typeof pos.x === 'number' && typeof pos.y === 'number') {
            winX = pos.x
            winY = pos.y
          }
          winPosReady = true
        })
        .catch(() => {
          winPosReady = true
        })
    } else {
      winPosReady = true
    }
  }

  function onMouseMove(e) {
    if (!dragging || !winPosReady) return
    const dx = e.screenX - startScreenX
    const dy = e.screenY - startScreenY
    if (dx === 0 && dy === 0) return
    window.runtime.WindowSetPosition(winX + dx, winY + dy)
  }

  function stopDrag() {
    dragging = false
  }

  onMounted(() => {
    document.addEventListener('mousedown', onMouseDown, true)
    document.addEventListener('mousemove', onMouseMove, true)
    document.addEventListener('mouseup', stopDrag, true)
    // 拖拽中窗口失焦（如鼠标移出窗口）时停止
    window.addEventListener('blur', stopDrag)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('mousedown', onMouseDown, true)
    document.removeEventListener('mousemove', onMouseMove, true)
    document.removeEventListener('mouseup', stopDrag, true)
    window.removeEventListener('blur', stopDrag)
  })
}
