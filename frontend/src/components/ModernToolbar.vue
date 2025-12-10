<template>
  <div class="modern-toolbar" :class="toolbarClasses">
    <!-- 主工具栏区域 -->
    <div class="toolbar-main">
      <!-- 格式工具组 -->
      <div class="toolbar-group">
        <ToolbarButton
          v-for="item in formatTools"
          :key="item.id"
          :title="item.title"
          :icon="item.icon"
          :active="item.active"
          @click="handleToolClick(item)"
        />
      </div>

      <!-- 分隔线 -->
      <div class="toolbar-divider"></div>

      <!-- 插入工具组 -->
      <div class="toolbar-group">
        <ToolbarButton
          v-for="item in insertTools"
          :key="item.id"
          :title="item.title"
          :icon="item.icon"
          @click="handleToolClick(item)"
        />
      </div>

      <!-- 分隔线 -->
      <div class="toolbar-divider"></div>

      <!-- 布局工具组 -->
      <div class="toolbar-group">
        <ToolbarButton
          v-for="item in layoutTools"
          :key="item.id"
          :title="item.title"
          :icon="item.icon"
          @click="handleToolClick(item)"
        />
      </div>
    </div>

    <!-- 工具栏右侧区域 -->
    <div class="toolbar-right">
      <!-- 搜索按钮 -->
      <ToolbarButton
        title="搜索"
        icon="search"
        :active="searchActive"
        @click="toggleSearch"
      />

      <!-- 更多操作下拉菜单 -->
      <DropdownMenu
        :items="moreActions"
        position="bottom-end"
        @select="handleMoreAction"
      >
        <ToolbarButton
          title="更多操作"
          icon="more-horizontal"
        />
      </DropdownMenu>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import ToolbarButton from './ToolbarButton.vue'
import DropdownMenu from './DropdownMenu.vue'

// Props
const props = defineProps({
  compact: {
    type: Boolean,
    default: false
  },
  floating: {
    type: Boolean,
    default: false
  },
  theme: {
    type: String,
    default: 'light'
  }
})

// Emits
const emit = defineEmits(['tool-click', 'search-toggle'])

// 状态
const searchActive = ref(false)
const activeTools = ref(new Set())

// 工具栏类名计算
const toolbarClasses = computed(() => ({
  'toolbar-compact': props.compact,
  'toolbar-floating': props.floating,
  [`toolbar-${props.theme}`]: true
})

// 工具配置
const formatTools = ref([
  {
    id: 'bold',
    title: '粗体',
    icon: 'bold',
    action: 'toggleBold'
  },
  {
    id: 'italic',
    title: '斜体',
    icon: 'italic',
    action: 'toggleItalic'
  },
  {
    id: 'strikethrough',
    title: '删除线',
    icon: 'strikethrough',
    action: 'toggleStrikethrough'
  },
  {
    id: 'code',
    title: '行内代码',
    icon: 'code',
    action: 'toggleInlineCode'
  }
])

const insertTools = ref([
  {
    id: 'link',
    title: '插入链接',
    icon: 'link',
    action: 'insertLink'
  },
  {
    id: 'image',
    title: '插入图片',
    icon: 'image',
    action: 'insertImage'
  },
  {
    id: 'table',
    title: '插入表格',
    icon: 'table',
    action: 'insertTable'
  },
  {
    id: 'list',
    title: '插入列表',
    icon: 'list',
    action: 'insertList'
  }
])

const layoutTools = ref([
  {
    id: 'heading1',
    title: '一级标题',
    icon: 'heading-1',
    action: 'insertHeading1'
  },
  {
    id: 'heading2',
    title: '二级标题',
    icon: 'heading-2',
    action: 'insertHeading2'
  },
  {
    id: 'quote',
    title: '引用',
    icon: 'quote',
    action: 'insertQuote'
  }
])

const moreActions = ref([
  {
    id: 'export',
    title: '导出文档',
    icon: 'download',
    action: 'exportDocument'
  },
  {
    id: 'import',
    title: '导入文档',
    icon: 'upload',
    action: 'importDocument'
  },
  {
    id: 'clear',
    title: '清空内容',
    icon: 'trash-2',
    action: 'clearDocument'
  },
  {
    id: 'guide',
    title: '语法指南',
    icon: 'book-open',
    action: 'showGuide'
  }
])

// 事件处理
const handleToolClick = (tool) => {
  emit('tool-click', tool)
  
  // 切换工具激活状态
  if (activeTools.value.has(tool.id)) {
    activeTools.value.delete(tool.id)
  } else {
    activeTools.value.add(tool.id)
  }
}

const toggleSearch = () => {
  searchActive.value = !searchActive.value
  emit('search-toggle', searchActive.value)
}

const handleMoreAction = (action) => {
  emit('tool-click', action)
}

// 监听主题变化
watch(() => props.theme, (newTheme) => {
  console.log('工具栏主题已切换:', newTheme)
})
</script>

<style scoped>
.modern-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-primary);
  padding: var(--space-2) var(--space-4);
  transition: all var(--transition-normal);
  position: relative;
  z-index: 10;
}

.toolbar-main {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.toolbar-divider {
  width: 1px;
  height: 20px;
  background: var(--border-primary);
  margin: 0 var(--space-2);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

/* 紧凑模式 */
.toolbar-compact {
  padding: var(--space-1) var(--space-2);
}

.toolbar-compact .toolbar-group {
  gap: 0;
}

/* 浮动模式 */
.toolbar-floating {
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-primary);
}

/* 主题样式 */
.toolbar-light {
  background: var(--bg-primary);
}

.toolbar-dark {
  background: var(--bg-secondary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .modern-toolbar {
    padding: var(--space-2);
  }
  
  .toolbar-group {
    gap: 0;
  }
}

/* 动画效果 */
.modern-toolbar {
  animation: slideDown 0.3s var(--ease-out);
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>