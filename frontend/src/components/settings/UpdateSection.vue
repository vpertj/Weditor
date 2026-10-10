<template>
  <section class="settings-section">
    <div class="section-header">
      <div class="section-icon">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M11,7V13L16.2,16.1L17,14.9L12.5,12.2V7H11Z" />
        </svg>
      </div>
      <h3>{{ $t('settings.update.title') }}</h3>
    </div>

    <!-- 当前版本 -->
    <div class="card-base setting-row">
      <div class="setting-info">
        <div class="title-secondary">{{ $t('settings.update.currentVersion') }}</div>
        <div class="description-small">
          {{ currentVersion }}
          <span v-if="buildInfo.commit && buildInfo.commit !== 'dev'"> · {{ buildInfo.commit }}</span>
          <span v-if="isDevBuild"> · {{ $t('settings.update.localBuild') }}</span>
        </div>
      </div>
      <button class="action-btn" :disabled="isBusy" @click="onCheckNow">
        {{ state === 'checking' ? $t('settings.update.checking') : $t('settings.update.checkNow') }}
      </button>
    </div>

    <!-- 自动检查 -->
    <div class="card-base setting-row">
      <div class="setting-info">
        <div class="title-secondary">{{ $t('settings.update.autoCheck.title') }}</div>
        <div class="description-small">{{ $t('settings.update.autoCheck.description') }}</div>
      </div>
      <button
        class="toggle-switch"
        :class="{ active: settings.autoCheck }"
        :aria-pressed="settings.autoCheck"
        @click="saveSettings({ autoCheck: !settings.autoCheck })"
      >
        <span class="toggle-knob"></span>
      </button>
    </div>

    <!-- 更新通道 -->
    <div class="card-base setting-row channel-row">
      <div class="setting-info">
        <div class="title-secondary">{{ $t('settings.update.channel.title') }}</div>
        <div class="description-small">{{ $t('settings.update.channel.description') }}</div>
      </div>
      <div class="channel-options" role="radiogroup" :aria-label="$t('settings.update.channel.title')">
        <button
          v-for="option in channels"
          :key="option.value"
          class="channel-option"
          :class="{ active: settings.channel === option.value }"
          role="radio"
          :aria-checked="settings.channel === option.value"
          @click="saveSettings({ channel: option.value })"
        >
          {{ $t(option.labelKey) }}
        </button>
      </div>
    </div>

    <!-- 检查结果 / 上次检查时间 -->
    <div class="card-base result-row" v-if="resultText">
      <div class="description-small">{{ resultText }}</div>
      <button v-if="hasUpdate" class="link-btn" @click="showDialog">
        {{ $t('settings.update.viewDetail') }}
      </button>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUpdater } from '../../composables/index.js'

const { t } = useI18n()

const {
  buildInfo,
  settings,
  info,
  state,
  errorMessage,
  dialogVisible,
  hasUpdate,
  isBusy,
  checkForUpdate,
  saveSettings,
} = useUpdater()

const channels = [
  { value: 'stable', labelKey: 'settings.update.channel.stable' },
  { value: 'dev', labelKey: 'settings.update.channel.dev' },
]

const currentVersion = computed(() => `v${(buildInfo.value.version || '').replace(/^v/, '')}`)

/** 本地 wails dev / wails build 没有注入构建时间 */
const isDevBuild = computed(() => !buildInfo.value.buildTime)

const resultText = computed(() => {
  if (state.value === 'checking') return t('settings.update.checking')
  if (errorMessage.value) return errorMessage.value
  if (hasUpdate.value) return t('settings.update.foundNew', { version: info.value?.latestVersion })
  if (info.value?.reason) return info.value.reason
  if (settings.value.lastCheckAt) {
    return t('settings.update.lastCheckAt', { time: formatTime(settings.value.lastCheckAt) })
  }
  return ''
})

const formatTime = (raw) => {
  const date = new Date(raw)
  return Number.isNaN(date.getTime()) ? raw : date.toLocaleString()
}

const onCheckNow = () => {
  checkForUpdate(true)
}

const showDialog = () => {
  dialogVisible.value = true
}
</script>

<style scoped>
@import '../../styles/components/settings/section.css';

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 20px;
  cursor: default;
  margin-bottom: 12px;
}

.setting-info {
  min-width: 0;
}

/* 立即检查按钮 */
.action-btn {
  flex-shrink: 0;
  padding: 8px 18px;
  border-radius: 8px;
  border: 1px solid var(--theme-border-light);
  background: transparent;
  color: var(--theme-text-primary, var(--text-primary));
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.action-btn:hover:not(:disabled) {
  background: var(--bg-hover);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 更新通道 */
.channel-options {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.channel-option {
  padding: 7px 16px;
  border-radius: 8px;
  border: 1px solid var(--theme-border-light);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.channel-option.active {
  border-color: var(--theme-primary);
  color: var(--theme-primary);
  background: rgba(59, 130, 246, 0.08);
}

/* 结果行 */
.result-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 20px;
}

.link-btn {
  flex-shrink: 0;
  border: none;
  background: none;
  padding: 0;
  color: var(--theme-primary);
  font-size: 13px;
  cursor: pointer;
}

.link-btn:hover {
  text-decoration: underline;
}

/* Toggle 开关 */
.toggle-switch {
  position: relative;
  flex-shrink: 0;
  width: 52px;
  height: 28px;
  border-radius: 14px;
  border: none;
  background: var(--theme-border-light);
  cursor: pointer;
  transition: background 0.3s ease, box-shadow 0.3s ease;
  padding: 0;
  outline: none;
}

.toggle-switch:hover {
  box-shadow: 0 0 0 4px rgba(0, 0, 0, 0.06);
}

.toggle-switch.active {
  background: var(--theme-primary);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.35);
}

.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.toggle-switch.active .toggle-knob {
  transform: translateX(24px);
}

@media (max-width: 1000px) {
  .setting-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
