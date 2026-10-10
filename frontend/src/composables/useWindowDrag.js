/**
 * @file src/composables/useWindowDrag.js
 * @description 窗口拖拽 + 顶部栏双击最大化
 *
 * WKWebView 对 -webkit-app-region: drag 的支持并不可靠（Wails 无原生处理，
 * 完全依赖 WebKit 行为）。此 composable 通过 JS 监听拖拽区域 mousedown/mousemove，
 * 调用 Wails runtime WindowSetPosition 实现窗口移动：
 * - 若 WebKit 支持 CSS drag，事件被原生拦截，JS 不会触发（CSS 生效）
 * - 若 WebKit 不支持 CSS drag，事件到达 JS，由本方案兜底拖动（JS 生效）
 *
 * 双击顶部栏切换窗口最大化（macOS 上映射为 [window zoom:]：铺满屏幕可用区域、
 * 保留菜单栏与 Dock）。双击判定自己实现而不用 dblclick 事件：本处理里会调用
 * preventDefault，且 WKWebView 下 dblclick 的派发行为不确定，按时间/位置判定更可靠。
 * 若 CSS drag 生效导致事件到不了 JS，则由 AppKit 的原生标题栏双击行为兜底。
 */
import { onMounted, onBeforeUnmount } from 'vue'

// 可拖拽区域（需与样式中的 -webkit-app-region: drag 保持一致）
const DRAG_SELECTOR = '.window-drag-strip, .panel-header, .app-header'
// 拖拽区域内不可触发拖拽的交互元素
const NO_DRAG_SELECTOR =
  'button, a, input, textarea, select, label, .panel-actions, .toggle-label, .btn, .btn-small, .dropdown-trigger, .header-btn, .view-toggle-btn, .draft-banner'
// 双击最大化的区域：只认顶部整宽拖拽条，避免在左侧竖栏空白处误触发
const DOUBLE_CLICK_SELECTOR = '.window-drag-strip'

// 双击判定：同一位置在此时间内再次按下算双击
const DOUBLE_CLICK_MS = 400
// "同一位置"的容差（逻辑像素），避免手抖导致双击失效
const DOUBLE_CLICK_TOLERANCE = 4

/**
 * 切换窗口最大化：让内容铺满屏幕可用区域。
 * macOS 上 Wails 的 ToggleMaximise 即 [mainWindow zoom:]，与标题栏缩放行为一致。
 */
function toggleMaximise() {
  const rt = window.runtime
  if (!rt || typeof rt.WindowToggleMaximise !== 'function') return
  rt.WindowToggleMaximise()
}

export function useWindowDrag() {
  let dragging = false
  let startScreenX = 0
  let startScreenY = 0
  let winX = 0
  let winY = 0
  let winPosReady = false
  // 上一次在双击区域内的按下时间与位置
  let lastDownAt = 0
  let lastDownX = 0
  let lastDownY = 0

  function onMouseDown(e) {
    // 仅左键
    if (e.button !== 0) return

    const target = e.target
    if (!(target instanceof Element)) return
    // 必须位于拖拽区域内
    if (!target.closest(DRAG_SELECTOR)) return
    // 交互元素上不触发拖拽
    if (target.closest(NO_DRAG_SELECTOR)) return

    const rt = window.runtime
    if (!rt) return

    // 双击判定先于拖拽：双击时只最大化，不移动窗口
    const inDoubleClickZone = !!target.closest(DOUBLE_CLICK_SELECTOR)
    const now = Date.now()
    const isDoubleClick =
      inDoubleClickZone &&
      now - lastDownAt < DOUBLE_CLICK_MS &&
      Math.abs(e.screenX - lastDownX) <= DOUBLE_CLICK_TOLERANCE &&
      Math.abs(e.screenY - lastDownY) <= DOUBLE_CLICK_TOLERANCE

    if (inDoubleClickZone) {
      lastDownAt = now
      lastDownX = e.screenX
      lastDownY = e.screenY
    } else {
      // 区域外的按下不参与双击判定，避免"别处点一下 + 顶部栏点一下"被判成双击
      lastDownAt = 0
    }

    if (isDoubleClick) {
      // 置零，避免三连击被当成第二次双击
      lastDownAt = 0
      dragging = false
      toggleMaximise()
      return
    }

    if (typeof rt.WindowSetPosition !== 'function') return

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
