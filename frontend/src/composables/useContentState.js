/**
 * @file src/composables/useContentState.js
 * @description 内容状态管理 Composable
 *
 * 专门管理与内容相关的状态和操作，包括：
 * - Markdown 内容管理
 * - HTML 内容管理
 * - 内容统计（字符数、是否有内容等）
 * - 示例内容加载
 */

import { ref, computed } from 'vue'

// 初始内容（用于初始化和示例加载）
const INITIAL_CONTENT = `# Markdown 格式完整测试

## 📋 所有格式枚举测试

这是一个包含所有 Markdown 格式的完整测试文档，用于检查社交平台兼容性。

---

## 1️⃣ 标题层级测试

# 一级标题 H1
## 二级标题 H2
### 三级标题 H3
#### 四级标题 H4
##### 五级标题 H5
###### 六级标题 H6

---

## 2️⃣ 文本格式测试

### 基础文本格式
- **粗体文本** 和 __另一种粗体语法__
- *斜体文本* 和 _另一种斜体语法_
- ***粗斜体文本*** 和 ___另一种粗斜体___
- ~~删除线文本~~
- \`行内代码\` 示例
- 普通文本和 **混合** *格式* ~~测试~~

### 特殊字符和转义
- 反引号: \\\`code\\\`
- 星号: \\*text\\*
- 下划线: \\_text\\_
- 波浪号: \\~\\~text\\~\\~

---

## 3️⃣ 列表测试

### 无序列表
- 第一项
- 第二项
  - 嵌套项 1
  - 嵌套项 2
    - 深层嵌套项
    - 另一个深层项
  - 嵌套项 3
- 第三项
- 包含 **粗体** 和 *斜体* 的项
- 包含 \`代码\` 的项

### 有序列表
1. 第一项
2. 第二项
   1. 嵌套有序项 1
   2. 嵌套有序项 2
      1. 深层嵌套
      2. 另一个深层嵌套
   3. 嵌套有序项 3
3. 第三项
4. 包含 **格式** 的项
5. 包含 \`代码\` 的项

### 任务列表
- [x] 已完成任务
- [x] 另一个已完成任务
- [ ] 未完成任务
- [ ] 包含 **粗体** 的任务
- [x] 包含 \`代码\` 的已完成任务
- [ ] 包含 [链接](https://github.com) 的任务

---

## 4️⃣ 引用块测试

### 简单引用
> 这是一个简单的引用块。

### 多行引用
> 这是多行引用的第一行。
>
> 这是第二行，中间有空行。
>
> 这是第三行。

### 嵌套引用
> 这是外层引用。
>
> > 这是嵌套引用。
> >
> > > 这是更深层的嵌套引用。
>
> 回到外层引用。

### 引用中的格式
> 引用中可以包含 **粗体**、*斜体* 和 \`代码\`。
>
> 也可以包含 [链接](https://github.com)。

---

## 5️⃣ 代码测试

### 行内代码
这是 \`行内代码\` 示例，包含 \`console.log('Hello')\` 这样的代码片段。

### 代码块（无语言标识）
\`\`\`
这是没有语言标识的代码块
可以包含任意文本
保持原有格式和缩进
\`\`\`

### JavaScript 代码块
\`\`\`javascript
// JavaScript 代码示例
function greet(name) {
  console.log(\`Hello, \${name}!\`);
  return \`Welcome, \${name}\`;
}

const user = 'World';
greet(user);
\`\`\`


### Mermaid 复杂图示例
\`\`\`mermaid
flowchart TB
  %% 子图与泳道
  subgraph Cluster_A [订单流转域]
    direction TB
    A[用户下单] --> B{库存检查}
    B -- 库存充足 --> C[预占库存]
    B -- 库存不足 --> D[下发缺货通知]
    C --> E{支付结果}
    E -- 成功 --> F[生成订单]
    E -- 失败 --> G[释放库存]
  end

  subgraph Cluster_B [履约域]
    direction LR
    F --> H[拣货打包]
    H --> I{发货方式}
    I -- 自营 --> J[自营配送]
    I -- 第三方 --> K[三方物流]
  end

  %% 跨域编排
  J --> L[签收]
  K --> L
  L --> M{是否售后}
  M -- 否 --> N[订单完成]
  M -- 是 --> O[售后流程]

  %% 样式与连线定义
  classDef critical fill:#ffe6e6,stroke:#ff4d4f,stroke-width:1px,color:#a8071a
  classDef ok fill:#e6fffb,stroke:#36cfc9,stroke-width:1px,color:#006d75
  classDef neutral fill:#f6ffed,stroke:#73d13d,stroke-width:1px,color:#135200

  class B,O critical
  class C,F,H ok
  class A,N neutral

  linkStyle 0 stroke:#1890ff,stroke-width:2px
  linkStyle 3 stroke:#fa8c16,stroke-dasharray: 5 3

  %% 子图外说明
  R[/监控面板/]:::ok --> B
\`\`\`

---

## 6️⃣ 表格测试

### 简单表格
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 数据1 | 数据2 | 数据3 |
| 数据4 | 数据5 | 数据6 |

### 对齐表格
| 左对齐 | 居中对齐 | 右对齐 |
|:-------|:-------:|-------:|
| 左 | 中 | 右 |
| 数据较长的内容 | 居中内容 | 右侧内容 |

### 包含格式的表格
| 功能 | 语法 | 示例 |
|------|------|------|
| **粗体** | \`**text**\` | **示例文本** |
| *斜体* | \`*text*\` | *示例文本* |
| [链接](https://github.com) | \`[text](url)\` | [GitHub](https://github.com) |

---

## 7️⃣ 分割线测试

使用三个或更多连字符：

---

使用三个或更多星号：

***

使用三个或更多下划线：

___

---

## 📝 测试总结

以上包含了所有常用的 Markdown 格式：

✅ **已测试格式**：
- 标题（H1-H6）
- 文本格式（粗体、斜体、删除线、行内代码）
- 列表（有序、无序、任务列表、嵌套）
- 引用块（简单、多行、嵌套）
- 代码块（多种语言）
- 表格（简单、对齐、包含格式）
- 分割线

🎯 **测试目标**：检查所有格式在社交平台中的显示效果和兼容性。`

// 默认内容开关（模块级单例，localStorage 持久化，默认启用）
const DEFAULT_CONTENT_KEY = 'weditor.defaultContentEnabled'

const getDefaultContentEnabled = () => {
  try {
    const stored = localStorage.getItem(DEFAULT_CONTENT_KEY)
    return stored === null ? true : stored === 'true'
  } catch {
    return true
  }
}

const enableDefaultContent = ref(getDefaultContentEnabled())

const setDefaultContentEnabled = (enabled) => {
  enableDefaultContent.value = enabled
  try {
    localStorage.setItem(DEFAULT_CONTENT_KEY, String(enabled))
  } catch (error) {
    console.warn('保存默认内容设置失败:', error)
  }
}

/**
 * 内容状态管理 Composable
 * @param {Object} options - 配置选项
 * @param {Function} options.onNotify - 通知回调函数
 * @returns {Object} 内容状态和方法
 */
export function useContentState(options = {}) {
  const { onNotify } = options

  // 状态（默认内容开关控制初始是否填充示例文档）
  const markdownContent = ref(getDefaultContentEnabled() ? INITIAL_CONTENT : '')
  const htmlContent = ref('')

  // 计算属性
  const hasContent = computed(() => markdownContent.value.trim().length > 0)
  const isHtmlReady = computed(() => htmlContent.value.length > 0)
  const characterCount = computed(() => markdownContent.value.length)

  // 新增统计信息
  const lineCount = computed(() => {
    if (!markdownContent.value) return 0
    return markdownContent.value.split('\n').length
  })

  const wordCount = computed(() => {
    if (!markdownContent.value) return 0
    // 中英文混合计数：中文字符按字计算，英文按单词计算
    const text = markdownContent.value.trim()
    if (!text) return 0

    // 移除代码块内容（避免影响统计）
    const withoutCodeBlocks = text.replace(/```[\s\S]*?```/g, '')

    // 统计中文字符
    const chineseChars = (withoutCodeBlocks.match(/[\u4e00-\u9fff]/g) || []).length

    // 统计英文单词（移除中文字符后按空格分割）
    const englishText = withoutCodeBlocks.replace(/[\u4e00-\u9fff]/g, ' ')
    const englishWords = englishText.split(/\s+/).filter(word => word.length > 0).length

    return chineseChars + englishWords
  })

  const estimatedReadTime = computed(() => {
    const wordsPerMinute = 200 // 中文阅读速度约200字/分钟
    return Math.max(1, Math.ceil(wordCount.value / wordsPerMinute))
  })

  // 方法
  const updateMarkdownContent = (content) => {
    markdownContent.value = content
  }

  const updateHtmlContent = (html) => {
    htmlContent.value = html
  }

  const clearContent = () => {
    markdownContent.value = ''
    htmlContent.value = ''
    onNotify?.('内容已清空', 'info')
  }

  const loadSample = () => {
    if (confirm('确定要加载示例内容吗？这将覆盖当前内容。')) {
      markdownContent.value = INITIAL_CONTENT
      onNotify?.('示例内容已加载', 'success')
    }
  }

  // 切换默认内容开关
  const toggleDefaultContent = () => {
    const next = !enableDefaultContent.value
    setDefaultContentEnabled(next)
    onNotify?.(next ? '默认内容已启用：启动时将显示默认内容' : '默认内容已禁用：启动时将显示空白文档', 'success')
  }

  return {
    // 状态
    markdownContent,
    htmlContent,
    enableDefaultContent,

    // 计算属性
    hasContent,
    isHtmlReady,
    characterCount,
    lineCount,
    wordCount,
    estimatedReadTime,

    // 方法
    updateMarkdownContent,
    updateHtmlContent,
    clearContent,
    loadSample,
    setDefaultContentEnabled,
    toggleDefaultContent
  }
}
