/**
 * @file src/composables/useUpdater.js
 * @description 自动更新：检查 / 下载 / 安装重启
 *
 * 状态放在模块级：设置面板与更新弹窗共享同一份状态，
 * 避免两处各查一次、进度不同步。
 */
import { ref, computed } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js'
import {
  CheckForUpdate,
  DownloadUpdate,
  ApplyUpdateAndRestart,
  GetUpdateSettings,
  SetUpdateSettings,
  SkipVersion,
  RevealUpdateFile,
  OpenUpdateDownloadPage,
  GetBuildInfo,
} from '../../wailsjs/go/main/App.js'

// ==========================================
// 模块级共享状态
// ==========================================

const buildInfo = ref({ version: '', commit: '', buildTime: '', platform: '' })
const settings = ref({ channel: 'stable', autoCheck: true, skippedVersion: '', lastCheckAt: '' })
const info = ref(null)
const progress = ref({ downloaded: 0, total: 0, percent: 0, speed: 0 })
/** idle | checking | available | downloading | ready | installing | error */
const state = ref('idle')
const errorMessage = ref('')
const dialogVisible = ref(false)

let initialized = false
let progressBound = false
let startupTimer = null

const isWailsEnv = () =>
  typeof window !== 'undefined' && window.go && window.go.main && window.go.main.App

/** 启动后延迟多久做静默检查：不阻塞窗口出现，也避开启动瞬间的资源争用 */
const STARTUP_CHECK_DELAY = 5000

// ==========================================
// 内部
// ==========================================

const bindProgressEvent = () => {
  if (progressBound) return
  progressBound = true
  EventsOn('update-progress', (payload) => {
    progress.value = {
      downloaded: payload?.downloaded || 0,
      total: payload?.total || 0,
      percent: payload?.percent || 0,
      speed: payload?.speed || 0,
    }
  })
}

const loadSettings = async () => {
  try {
    settings.value = await GetUpdateSettings()
  } catch (e) {
    console.warn('读取更新设置失败:', e)
  }
}

const loadBuildInfo = async () => {
  try {
    buildInfo.value = await GetBuildInfo()
  } catch (e) {
    console.warn('读取构建信息失败:', e)
  }
}

const scheduleStartupCheck = () => {
  if (startupTimer) {
    clearTimeout(startupTimer)
    startupTimer = null
  }
  if (!settings.value.autoCheck) return
  startupTimer = setTimeout(() => {
    checkForUpdate(false)
  }, STARTUP_CHECK_DELAY)
}

// ==========================================
// 对外操作
// ==========================================

/**
 * 检查更新
 * @param {boolean} manual 用户主动触发时为 true：需要展示结果与错误；自动检查失败则静默
 */
const checkForUpdate = async (manual = true) => {
  if (!isWailsEnv()) return null
  if (state.value === 'checking' || state.value === 'downloading' || state.value === 'installing') {
    return info.value
  }

  state.value = 'checking'
  errorMessage.value = ''
  if (manual) dialogVisible.value = true

  try {
    const result = await CheckForUpdate(manual)
    info.value = result
    await loadSettings()

    if (result?.error) {
      state.value = 'error'
      errorMessage.value = result.error
      return result
    }
    if (result?.available) {
      state.value = 'available'
      // 只有真的发现新版本（或用户主动检查）才打扰用户
      dialogVisible.value = true
      return result
    }
    state.value = 'idle'
    return result
  } catch (e) {
    state.value = 'error'
    errorMessage.value = e?.message || String(e)
    return null
  }
}

/** 下载新版本安装包，进度通过 update-progress 事件更新 */
const startDownload = async () => {
  if (!isWailsEnv()) return false
  if (state.value === 'downloading' || state.value === 'installing') return false

  state.value = 'downloading'
  errorMessage.value = ''
  progress.value = { downloaded: 0, total: info.value?.assetSize || 0, percent: 0, speed: 0 }

  try {
    const result = await DownloadUpdate()
    if (result?.success) {
      state.value = 'ready'
      return true
    }
    state.value = 'error'
    errorMessage.value = result?.message || '下载失败'
    return false
  } catch (e) {
    state.value = 'error'
    errorMessage.value = e?.message || String(e)
    return false
  }
}

/** 安装已下载的更新并重启；成功后后端会在约 0.8 秒后退出进程 */
const applyAndRestart = async () => {
  if (!isWailsEnv()) return false
  state.value = 'installing'
  errorMessage.value = ''
  try {
    await ApplyUpdateAndRestart()
    return true
  } catch (e) {
    state.value = 'error'
    errorMessage.value = e?.message || String(e)
    return false
  }
}

/** 跳过后不再提示当前最新版本 */
const skipCurrentVersion = async () => {
  if (!isWailsEnv()) return
  const version = info.value?.latestVersion
  dialogVisible.value = false
  if (!version) return
  try {
    settings.value = await SkipVersion(version)
  } catch (e) {
    console.warn('记录跳过版本失败:', e)
  }
}

/** 保存更新偏好（自动检查开关、更新通道） */
const saveSettings = async (patch) => {
  if (!isWailsEnv()) return
  try {
    settings.value = await SetUpdateSettings({ ...settings.value, ...patch })
    // 关掉自动检查时取消已排期的启动检查
    if (!settings.value.autoCheck) scheduleStartupCheck()
  } catch (e) {
    console.error('保存更新设置失败:', e)
  }
}

/** 打开 Release 页面（无法自动更新时的降级路径） */
const openDownloadPage = () => {
  if (!isWailsEnv()) return
  OpenUpdateDownloadPage()
}

/** 在文件管理器中定位已下载的安装包 */
const revealUpdateFile = async () => {
  if (!isWailsEnv()) return ''
  try {
    return await RevealUpdateFile()
  } catch (e) {
    console.warn('定位安装包失败:', e)
    return ''
  }
}

/** 应用启动后调用一次：绑定事件、加载设置、排期静默检查 */
const initUpdater = async () => {
  if (initialized || !isWailsEnv()) return
  initialized = true
  bindProgressEvent()
  await loadSettings()
  await loadBuildInfo()
  scheduleStartupCheck()
}

const disposeUpdater = () => {
  if (startupTimer) {
    clearTimeout(startupTimer)
    startupTimer = null
  }
  if (progressBound) {
    EventsOff('update-progress')
    progressBound = false
  }
  initialized = false
}

// ==========================================
// 导出
// ==========================================

export function useUpdater() {
  return {
    // 状态
    buildInfo,
    settings,
    info,
    progress,
    state,
    errorMessage,
    dialogVisible,
    hasUpdate: computed(() => !!info.value?.available),
    isBusy: computed(() => ['checking', 'downloading', 'installing'].includes(state.value)),

    // 操作
    initUpdater,
    disposeUpdater,
    checkForUpdate,
    startDownload,
    applyAndRestart,
    skipCurrentVersion,
    saveSettings,
    openDownloadPage,
    revealUpdateFile,
  }
}
