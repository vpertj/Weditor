<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="close">
        <div class="modal-container" @click.stop>
          <div class="modal-header">
            <div class="header-content">
              <div class="header-icon">
                <svg viewBox="0 0 24 24" width="28" height="28">
                  <path fill="currentColor" d="M13,9H11V7H13M13,17H11V11H13M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z"/>
                </svg>
              </div>
              <div class="header-text">
                <h2>{{ $t('about.title') }}</h2>
                <p>{{ $t('about.subtitle') }}</p>
              </div>
            </div>
            <button class="close-btn" @click="close">
              <svg viewBox="0 0 24 24" width="24" height="24">
                <path fill="currentColor" d="M19,6.41L17.59,5L12,10.59L6.41,5L5,6.41L10.59,12L5,17.59L6.41,19L12,13.41L17.59,19L19,17.59L13.41,12L19,6.41Z"/>
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <div class="about-sections">
              <!-- 软件信息 -->
              <section class="about-section">
                <div class="section-header">
                  <div class="section-icon">
                    <svg viewBox="0 0 24 24" width="20" height="20">
                      <path fill="currentColor" d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20A8,8 0 0,0 20,12A8,8 0 0,0 12,4M11,16.5L6.5,12L7.91,10.59L11,13.67L16.59,8.09L18,9.5L11,16.5Z"/>
                    </svg>
                  </div>
                  <h3>{{ $t('about.softwareInfo') }}</h3>
                </div>
                <div class="info-list">
                  <div class="info-item">
                    <div class="info-label">{{ $t('about.version') }}</div>
                    <div class="info-value">{{ versionText }}</div>
                  </div>

                  <div class="info-item">
                    <div class="info-label">{{ $t('about.author') }}</div>
                    <div class="info-value">vpertj</div>
                  </div>

                  <div class="info-item">
                    <div class="info-label">{{ $t('about.buildDate') }}</div>
                    <div class="info-value">{{ buildDateText }}</div>
                  </div>

                  <div class="info-item">
                    <div class="info-label">{{ $t('about.license') }}</div>
                    <div class="info-value">MIT License</div>
                  </div>
                </div>
              </section>

              <!-- 功能介绍 -->
              <section class="about-section">
                <div class="section-header">
                  <div class="section-icon">
                    <svg viewBox="0 0 24 24" width="20" height="20">
                      <path fill="currentColor" d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
                    </svg>
                  </div>
                  <h3>{{ $t('about.features') }}</h3>
                </div>
                <div class="features-list">
                  <div class="feature-item">
                    <div class="feature-icon">
                      <svg viewBox="0 0 24 24" width="16" height="16">
                        <path fill="currentColor" d="M3,3H11V5H3V3M13,3H21V5H13V3M3,7H11V9H3V7M13,7H21V9H13V7M3,11H11V13H3V11M13,11H21V13H13V11M3,15H11V17H3V15M13,15H21V17H13V15M3,19H11V21H3V19M13,19H21V21H13V19Z"/>
                      </svg>
                    </div>
                    <div class="feature-text">
                      <h4>{{ $t('about.feature1.title') }}</h4>
                      <p>{{ $t('about.feature1.description') }}</p>
                    </div>
                  </div>

                  <div class="feature-item">
                    <div class="feature-icon">
                      <svg viewBox="0 0 24 24" width="16" height="16">
                        <path fill="currentColor" d="M14.6,16.6L19.2,12L14.6,7.4L16,6L22,12L16,18L14.6,16.6M9.4,16.6L4.8,12L9.4,7.4L8,6L2,12L8,18L9.4,16.6Z"/>
                      </svg>
                    </div>
                    <div class="feature-text">
                      <h4>{{ $t('about.feature2.title') }}</h4>
                      <p>{{ $t('about.feature2.description') }}</p>
                    </div>
                  </div>

                  <div class="feature-item">
                    <div class="feature-icon">
                      <svg viewBox="0 0 24 24" width="16" height="16">
                        <path fill="currentColor" d="M3 5h18v2H3V5m0 4h18v2H3V9m0 4h10v2H3v-2z"/>
                      </svg>
                    </div>
                    <div class="feature-text">
                      <h4>{{ $t('about.feature3.title') }}</h4>
                      <p>{{ $t('about.feature3.description') }}</p>
                    </div>
                  </div>

                  <div class="feature-item">
                    <div class="feature-icon">
                      <svg viewBox="0 0 24 24" width="16" height="16">
                        <path fill="currentColor" d="M12,15.5A3.5,3.5 0 0,1 8.5,12A3.5,3.5 0 0,1 12,8.5A3.5,3.5 0 0,1 15.5,12A3.5,3.5 0 0,1 12,15.5M19.43,12.97C19.47,12.65 19.5,12.33 19.5,12C19.5,11.67 19.47,11.34 19.43,11L21.54,9.37C21.73,9.22 21.78,8.95 21.66,8.73L19.66,5.27C19.54,5.05 19.27,4.96 19.05,5.05L16.56,6.05C16.04,5.66 15.5,5.32 14.87,5.07L14.5,2.42C14.46,2.18 14.25,2 14,2H10C9.75,2 9.54,2.18 9.5,2.42L9.13,5.07C8.5,5.32 7.96,5.66 7.44,6.05L4.95,5.05C4.73,4.96 4.46,5.05 4.34,5.27L2.34,8.73C2.22,8.95 2.27,9.22 2.46,9.37L4.57,11C4.53,11.34 4.5,11.67 4.5,12C4.5,12.33 4.53,12.65 4.57,12.97L2.46,14.63C2.27,14.78 2.22,15.05 2.34,15.27L4.34,18.73C4.46,18.95 4.73,19.03 4.95,18.95L7.44,17.94C7.96,18.34 8.5,18.68 9.13,18.93L9.5,21.58C9.54,21.82 9.75,22 10,22H14C14.25,22 14.46,21.82 14.5,21.58L14.87,18.93C15.5,18.68 16.04,18.34 16.56,17.94L19.05,18.95C19.27,19.03 19.54,18.95 19.66,18.73L21.66,15.27C21.78,15.05 21.73,14.78 21.54,14.63L19.43,12.97Z"/>
                      </svg>
                    </div>
                    <div class="feature-text">
                      <h4>{{ $t('about.feature4.title') }}</h4>
                      <p>{{ $t('about.feature4.description') }}</p>
                    </div>
                  </div>
                </div>
              </section>


            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script>
import { useUpdater } from '../composables/useUpdater.js'

export default {
  name: 'AboutDialog',
  props: {
    show: {
      type: Boolean,
      default: false
    }
  },
  emits: ['close'],
  setup() {
    // 版本与构建时间都取自被注入的构建信息，不再写死
    const { buildInfo } = useUpdater()
    return { buildInfo }
  },
  computed: {
    versionText() {
      const version = (this.buildInfo?.version || '').replace(/^v/, '')
      return version ? `v${version}` : '-'
    },
    buildDateText() {
      const raw = this.buildInfo?.buildTime
      if (!raw) return '-'
      const date = new Date(raw)
      return Number.isNaN(date.getTime()) ? raw : date.toLocaleDateString()
    }
  },
  methods: {
    close() {
      this.$emit('close')
    }
  }
}
</script>

<style scoped>
@import '../styles/components/notifications.css';

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background-color: var(--bg-primary);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
  border-bottom: 1px solid var(--border-light);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background-color: var(--theme-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-text h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

.header-text p {
  margin: 4px 0 0 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: background-color 0.2s;
  color: var(--text-secondary);
}

.close-btn:hover {
  background-color: var(--bg-hover);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px 24px;
}

.about-sections {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.about-section {
  padding: 16px 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.section-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background-color: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.section-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.info-list,
.features-list,
.tech-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item,
.feature-item,
.tech-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.info-label,
.feature-icon,
.tech-icon {
  min-width: 80px;
  font-weight: 500;
  color: var(--text-secondary);
}

.info-value,
.feature-text,
.tech-text {
  flex: 1;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
}

.feature-text h4,
.tech-text h4 {
  margin: 0 0 4px 0;
  font-size: 1rem;
  font-weight: 600;
}

.feature-text p,
.tech-text p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

/* 过渡动画 */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>