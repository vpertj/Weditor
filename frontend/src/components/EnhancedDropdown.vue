<template>
  <div class="enhanced-dropdown" :class="dropdownClasses">
    <!-- 触发器 -->
    <div
      ref="triggerRef"
      class="dropdown-trigger"
      @click="toggleDropdown"
    >
      <slot name="trigger"></slot>
    </div>

    <!-- 下拉内容 -->
    <Teleport to="body" :disabled="!teleport">
      <div
        v-if="isOpen"
        ref="contentRef"
        class="dropdown-content"
        :style="contentStyles"
        @click="handleContentClick"
      >
        <div
          v-for="(item, index) in items"
          :key="item.id || index"
          class="dropdown-item"
          :class="getItemClasses(item)"
          @click="selectItem(item)"
        >
          <!-- 图标 -->
          <div v-if="item.icon" class="item-icon">
            <svg-icon :name="item.icon" />
          </div>

          <!-- 文本 -->
          <div class="item-text">
            {{ item.title }}
          </div>

          <!-- 快捷键 -->
          <div v-if="item.shortcut" class="item-shortcut">
              {{ item.shortcut }}
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

// Props
const props = defineProps({
  items: {
    type: Array,
    default: () => []
  },
  position: {
    type: String,
    default: 'bottom-start',
    validator: (value) => [
      'top-start', 'top', 'top-end',
      'bottom-start', 'bottom', 'bottom-end',
      'left-start', 'left', 'left-end',
      'right-start', 'right', 'right-end'
    ].includes(value)
  },
  disabled: {
    type: Boolean,
    default: false
  },
  teleport: {
    type: Boolean,
    default: true
  },
  maxHeight: {
    type: String,
    default: '300px'
  },
  width: {
    type: String,
    default: 'auto'
})

// Emits
const emit = defineEmits(['select', 'open', 'close'])

// 状态
const isOpen = ref(false)
const triggerRef = ref(null)
const contentRef = ref(null)

// 计算属性
const dropdownClasses = computed(() => ({
  'dropdown-open': isOpen.value,
  'dropdown-disabled': props.disabled
})

const contentStyles = computed(() => {
  const styles = {
    '--dropdown-max-height': props.maxHeight,
    '--dropdown-width': props.width
  }
  
  return styles
})

// 方法
const toggleDropdown = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

const selectItem = (item) => {
  emit('select', item)
  closeDropdown()
}

const closeDropdown = () => {
  isOpen.value = false
  emit('close')
}

const getItemClasses = (item) => ({
  'item-disabled': item.disabled,
  'item-danger': item.danger,
  'item-divider': item.divider
})

const handleContentClick = (event) => {
  event.stopPropagation()
}

// 点击外部关闭
const handleClickOutside = (event) => {
  if (!isOpen.value) return
  
  const target = event.target
  const trigger = triggerRef.value
  const content = contentRef.value
  
  if (trigger && !trigger.contains(target) && content && !content.contains(target)) {
    closeDropdown()
  }
}

// 键盘导航
const handleKeydown = (event) => {
  if (!isOpen.value) return
  
  switch (event.key) {
    case 'Escape':
      closeDropdown()
      break
    case 'ArrowDown':
      event.preventDefault()
      // TODO: 实现键盘导航
      break
    case 'ArrowUp':
      event.preventDefault()
      // TODO: 实现键盘导航
      break
    case 'Enter':
      event.preventDefault()
      // TODO: 实现键盘导航
      break
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
}

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})

// 监听状态变化
watch(isOpen, (newValue) => {
  if (newValue) {
    emit('open')
  }
})
</script>

<style scoped>
.enhanced-dropdown {
  position: relative;
  display: inline-block;
}

.dropdown-trigger {
  cursor: pointer;
  user-select: none;
}

.dropdown-content {
  position: fixed;
  z-index: 1000;
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  padding: var(--space-1) 0;
  max-height: var(--dropdown-max-height, 300px);
  width: var(--dropdown-width, auto);
  min-width: 160px;
  overflow-y: auto;
  animation: dropdownFadeIn 0.2s var(--ease-out);
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
}

.dropdown-item:hover {
  background: var(--bg-hover);
}

.dropdown-item:active {
  background: var(--bg-active);
}

.item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  color: var(--text-secondary);
}

.item-text {
  flex: 1;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.item-shortcut {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

/* 禁用状态 */
.item-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.item-disabled:hover {
  background: transparent;
}

/* 危险操作 */
.item-danger {
  color: var(--color-error);
}

.item-danger:hover {
  background: var(--error-50);
}

/* 分隔线 */
.item-divider {
  height: 1px;
  background: var(--border-primary);
  margin: var(--space-1) 0;
  padding: 0;
  cursor: default;
}

.item-divider:hover {
  background: var(--border-primary);
}

/* 动画 */
@keyframes dropdownFadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 响应式 */
@media (max-width: 768px) {
  .dropdown-content {
    min-width: 140px;
  }
}
</style>