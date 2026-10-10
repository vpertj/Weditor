<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="modal-overlay" @click="onOverlayClick">
        <div class="modal-container" @click.stop>
          <div class="modal-header">
            <div class="header-content">
              <div class="header-icon" :class="`state-${state}`">
                <svg viewBox="0 0 24 24" width="28" height="28">
                  <path fill="currentColor" :d="headerIconPath" />
                </svg>
              </div>
              <div class="header-text">
                <h2>{{ headerTitle }}</h2>
                <p v-if="info?.latestVersion">{{ currentVersion }} → {{ info.latestVersion }}</p>
                <p v-else>{{ currentVersion }}</p>
              </div>
            </div>
            <button class="close-btn" :disabled="isBusy" @click="close">
              <svg viewBox="0 0 24 24" width="24" height="24">
                <path fill="currentColor" d="M19,6.41L17.59,5L12,10.59L6.41,5L5,6.41L10.59,12L5,17.59L6.41,19L12,13.41L17.59,19L19,17.59L13.41,12L19,6.41Z"/>
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <!-- 检查中 -->
            <div v-if="state === 'checking'" class="status-row">
              <span class="spinner"></span>
              <span>{{ t('update.checking') }}</span>
            </div>

            <!-- 已是最新 / 已跳过 -->
            <div v-else-if="state === 'idle'" class="status-row success">
              <svg viewBox="0 0 24 24" width="18" height="18">
                <path fill="currentColor" d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"/>
              </svg>
              <span>{{ info?.reason || t('update.upToDate') }}</span>
            </div>

            <!-- 出错 / 无法自动更新 -->
            <div v-else-if="state === 'error' || (info && !info.available)" class="status-block error">
              <div class="status-row">
                <svg viewBox="0 0 24 24" width="18" height="18">
                  <path fill="currentColor" d="M13,13H11V7H13M13,17H11V15H13M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z"/>
                </svg>
                <span>{{ errorMessage || info?.reason || t('update.unknownError') }}</span>
              </div>
              <p class="hint">{{ t('update.manualHint') }}</p>
            </div>

            <!-- 有新版本 -->
            <template v-else-if="info">
              <p class="hint">{{ t('update.restartHint') }}</p>

              <div v-if="info.publishedAt" class="meta-row">
                <span>{{ t('update.publishedAt') }}</span>
                <span>{{ formatDate(info.publishedAt) }}</span>
              </div>
              <div v-if="assetSizeText" class="meta-row">
                <span>{{ t('update.packageSize') }}</span>
                <span>{{ assetSizeText }}</span>
              </div>

              <div v-if="info.notes" class="notes">
                <div class="notes-title">{{ t('update.notes') }}</div>
                <p>{{ info.notes }}</p>
              </div>

              <!-- 下载进度 -->
              <div v-if="state === 'downloading' || state === 'ready'" class="progress-block">
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: `${progress.percent || 0}%` }"></div>
                </div>
                <div class="progress-text">
                  <span>{{ Math.round(progress.percent || 0) }}%</span>
                  <span>{{ formatBytes(progress.downloaded) }} / {{ formatBytes(progress.total) }}</span>
                  <span v-if="state === 'downloading' && progress.speed">{{ formatBytes(progress.speed) }}/s</span>
                </div>
              </div>
            </template>
          </div>

          <div class="modal-footer">
            <!-- 有可用更新 -->
            <template v-if="info?.available">
              <button class="btn btn-ghost" :disabled="isBusy" @click="onSkip">
                {{ t('update.skipVersion') }}
              </button>
              <button class="btn btn-ghost" :disabled="isBusy" @click="close">
                {{ t('update.later') }}
              </button>
              <button
                class="btn btn-primary"
                :disabled="isBusy"
                @click="onPrimary"
              >
                {{ primaryLabel }}
              </button>
            </template>

            <!-- 无可用更新或出错 -->
            <template v-else>
              <button class="btn btn-ghost" @click="onReveal">
                {{ t('update.revealFile') }}
              </button>
              <button class="btn btn-ghost" @click="onOpenPage">
                {{ t('update.openReleasePage') }}
              </button>
              <button class="btn btn-primary" @click="onCheckAgain">
                {{ t('update.checkAgain') }}
              </button>
            </template>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUpdater } from '../composables/useUpdater.js'

const { t } = useI18n()

const {
  buildInfo,
  info,
  progress,
  state,
  errorMessage,
  dialogVisible,
  isBusy,
  checkForUpdate,
  startDownload,
  applyAndRestart,
  skipCurrentVersion,
  openDownloadPage,
  revealUpdateFile,
} = useUpdater()

const visible = dialogVisible

// 有未保存内容时，主按钮改为先引导用户保存，避免更新重启丢内容
const primaryLabel = computed(() => {
  if (state.value === 'downloading') return `${Math.round(progress.value.percent || 0)}%`
  if (state.value === 'ready') return t('update.installNow')
  if (state.value === 'installing') return t('update.installing')
  return t('update.downloadAndInstall')
})

const headerTitle = computed(() => {
  switch (state.value) {
    case 'checking':
      return t('update.checking')
    case 'downloading':
      return t('update.downloading')
    case 'ready':
      return t('update.readyToInstall')
    case 'installing':
      return t('update.installing')
    case 'error':
      return t('update.checkFailed')
    default:
      return info.value?.available ? t('update.newVersionFound') : t('update.noUpdate')
  }
})

const headerIconPath = computed(() => {
  if (state.value === 'error') {
    return 'M13,13H11V7H13M13,17H11V15H13M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z'
  }
  if (state.value === 'idle') {
    return 'M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z'
  }
  return 'M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M11,7V13L16.2,16.1L17,14.9L12.5,12.2V7H11Z'
})

const currentVersion = computed(() => `v${(buildInfo.value.version || '').replace(/^v/, '')}`)

const assetSizeText = computed(() => (info.value?.assetSize ? formatBytes(info.value.assetSize) : ''))

const formatBytes = (bytes) => {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }
  return `${value.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

const formatDate = (raw) => {
  const date = new Date(raw)
  return Number.isNaN(date.getTime()) ? raw : date.toLocaleString()
}

// 下载/安装过程中不允许关掉弹窗，避免用户以为已经取消
const close = () => {
  if (isBusy.value) return
  dialogVisible.value = false
}

const onOverlayClick = () => close()

const onPrimary = async () => {
  if (state.value === 'available') {
    const ok = await startDownload()
    if (ok) await applyAndRestart()
    return
  }
  if (state.value === 'downloading') return
  if (state.value === 'ready') {
    await applyAndRestart()
  }
}

const onSkip = () => skipCurrentVersion()

const onCheckAgain = async () => {
  await checkForUpdate(true)
}

const onOpenPage = () => openDownloadPage()

const onReveal = async () => {
  const path = await revealUpdateFile()
  if (!path) onOpenPage()
}
</script>

<style scoped>
@import '../styles/components/notifications.css';

.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.modal-container {
  background-color: var(--bg-primary);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  width: min(520px, 90vw);
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 20px 22px 16px;
  border-bottom: 1px solid var(--border-light);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 14px;
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background-color: var(--bg-secondary);
  color: var(--theme-primary-light);
}

.header-icon.state-error {
  color: #e57373;
}

.header-icon.state-idle {
  color: #66bb6a;
}

.header-text h2 {
  margin: 0;
  font-size: 16px;
  color: var(--text-primary);
}

.header-text p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 2px;
  border-radius: 6px;
}

.close-btn:hover:not(:disabled) {
  background-color: var(--bg-hover);
}

.close-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.modal-body {
  padding: 18px 22px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-primary);
}

.status-row.success {
  color: #66bb6a;
}

.status-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-block .status-row {
  color: var(--text-primary);
}

.status-block.error .status-row svg {
  color: #e57373;
}

.hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.meta-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary);
}

.notes {
  background-color: var(--bg-secondary);
  border-radius: 8px;
  padding: 12px 14px;
}

.notes-title {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.notes p {
  margin: 0;
  font-size: 13px;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.progress-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-bar {
  height: 6px;
  border-radius: 3px;
  background-color: var(--bg-secondary);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--theme-primary-light);
  transition: width 0.2s ease;
}

.progress-text {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 22px 18px;
  border-top: 1px solid var(--border-light);
}

.btn {
  border: none;
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.15s ease, opacity 0.15s ease;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background-color: var(--theme-primary-light);
  color: #fff;
}

.btn-ghost {
  background-color: transparent;
  color: var(--text-secondary);
}

.btn-ghost:hover:not(:disabled) {
  background-color: var(--bg-hover);
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-light);
  border-top-color: var(--theme-primary-light);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
