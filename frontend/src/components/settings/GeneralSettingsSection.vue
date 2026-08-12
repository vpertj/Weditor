<template>
  <section class="settings-section">
    <div class="section-header">
      <div class="section-icon">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M19.14,12.94C19.18,12.64 19.2,12.33 19.2,12C19.2,11.68 19.18,11.36 19.13,11.06L21.16,9.48C21.34,9.34 21.39,9.07 21.28,8.87L19.36,5.55C19.24,5.33 18.99,5.26 18.77,5.33L16.38,6.29C15.88,5.91 15.32,5.59 14.71,5.35L14.5,2.81C14.47,2.58 14.27,2.4 14.04,2.4H10.02C9.78,2.4 9.59,2.57 9.56,2.81L9.36,5.35C8.75,5.59 8.19,5.91 7.69,6.29L5.3,5.33C5.08,5.26 4.83,5.33 4.71,5.55L2.79,8.87C2.68,9.07 2.73,9.34 2.91,9.48L4.94,11.06C4.89,11.36 4.87,11.67 4.87,12C4.87,12.33 4.89,12.64 4.94,12.94L2.91,14.52C2.73,14.66 2.68,14.93 2.79,15.13L4.71,18.45C4.83,18.67 5.08,18.74 5.3,18.67L7.69,17.71C8.19,18.09 8.75,18.41 9.36,18.65L9.56,21.19C9.59,21.42 9.78,21.6 10.02,21.6H14.04C14.27,21.6 14.47,21.42 14.5,21.19L14.71,18.65C15.32,18.41 15.88,18.09 16.38,17.71L18.77,18.67C18.99,18.74 19.24,18.67 19.36,18.45L21.28,15.13C21.39,14.93 21.34,14.66 21.16,14.52L19.13,12.94M12,15.6A3.6,3.6 0 1,1 12,8.4A3.6,3.6 0 0,1 12,15.6M12,10.2A1.8,1.8 0 1,0 12,13.8A1.8,1.8 0 0,0 12,10.2Z" />
        </svg>
      </div>
      <h3>{{ $t('settings.general.title') }}</h3>
    </div>

    <!-- 默认打开方式 -->
    <div class="card-base setting-row open-mode-row">
      <div class="setting-info">
        <div class="title-secondary">{{ $t('settings.general.openMode.title') }}</div>
        <div class="description-small">{{ $t('settings.general.openMode.description') }}</div>
      </div>

      <div class="open-mode-options" role="radiogroup" :aria-label="$t('settings.general.openMode.title')">
        <button
          v-for="mode in openModes"
          :key="mode.value"
          class="open-mode-option"
          :class="{ active: defaultOpenMode === mode.value }"
          role="radio"
          :aria-checked="defaultOpenMode === mode.value"
          @click="$emit('select-default-open-mode', mode.value)"
        >
          <span class="open-mode-icon">
            <svg v-if="mode.value === 'editor'" viewBox="0 0 24 24" width="16" height="16">
              <path fill="currentColor" d="M14.6,16.6L19.2,12L14.6,7.4L16,6L22,12L16,18L14.6,16.6M9.4,16.6L4.8,12L9.4,7.4L8,6L2,12L8,18L9.4,16.6Z"/>
            </svg>
            <svg v-else-if="mode.value === 'both'" viewBox="0 0 24 24" width="16" height="16">
              <path fill="currentColor" d="M3,3H11V5H3V3M13,3H21V5H13V3M3,7H11V9H3V7M13,7H21V9H13V7M3,11H11V13H3V11M13,11H21V13H13V11M3,15H11V17H3V15M13,15H21V17H13V15M3,19H11V21H3V19Z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="16" height="16">
              <path fill="currentColor" d="M3 5h18v2H3V5m0 4h18v2H3V9m0 4h10v2H3v-2z"/>
            </svg>
          </span>
          <span class="open-mode-label">{{ $t(mode.labelKey) }}</span>
          <span v-if="defaultOpenMode === mode.value" class="open-mode-check">
            <svg width="14" height="14" viewBox="0 0 20 20" fill="none">
              <path d="M4 10l4 4 8-9" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
        </button>
      </div>
    </div>

    <!-- 默认内容开关 -->
    <div class="card-base setting-row">
      <div class="setting-info">
        <div class="title-secondary">{{ $t('settings.general.defaultContent.title') }}</div>
        <div class="description-small">{{ $t('settings.general.defaultContent.description') }}</div>
      </div>
      <button
        class="toggle-switch"
        :class="{ active: enableDefaultContent }"
        :aria-pressed="enableDefaultContent"
        :title="enableDefaultContent
          ? $t('settings.general.defaultContent.disableTip')
          : $t('settings.general.defaultContent.enableTip')"
        @click="$emit('toggle-default-content')"
      >
        <span class="toggle-knob"></span>
      </button>
    </div>
  </section>
</template>

<script setup>
defineProps({
  enableDefaultContent: {
    type: Boolean,
    default: true
  },
  defaultOpenMode: {
    type: String,
    default: 'wysiwyg'
  },
  openModes: {
    type: Array,
    default: () => []
  }
})

defineEmits(['toggle-default-content', 'select-default-open-mode'])
</script>

<style scoped>
@import '../../styles/components/settings/section.css';

/* 设置行：左信息 + 右开关/选项 */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 20px;
  cursor: default;
  margin-bottom: 16px;
}

.setting-row:hover {
  transform: none;
}

.setting-info {
  flex: 1;
  min-width: 0;
}

/* 默认打开方式选项 */
.open-mode-options {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.open-mode-option {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 1.5px solid var(--theme-border-light);
  border-radius: 12px;
  background: var(--theme-bg-primary);
  color: var(--theme-text-secondary);
  cursor: pointer;
  transition: all 0.25s ease;
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  outline: none;
}

.open-mode-option:hover {
  border-color: var(--theme-primary);
  color: var(--theme-text-primary);
  transform: translateY(-1px);
}

.open-mode-option.active {
  border-color: var(--theme-primary);
  color: var(--theme-primary);
  background: var(--theme-primary-15, rgba(59, 130, 246, 0.1));
  box-shadow: 0 0 0 1px var(--theme-primary);
}

.open-mode-icon {
  display: flex;
  align-items: center;
  opacity: 0.85;
}

.open-mode-check {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--theme-primary);
  margin-left: 2px;
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

/* 响应式 */
@media (max-width: 1000px) {
  .open-mode-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .open-mode-options {
    width: 100%;
    flex-wrap: wrap;
  }

  .open-mode-option {
    flex: 1;
    justify-content: center;
    min-width: 100px;
  }
}
</style>
