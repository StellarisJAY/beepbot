<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { marked } from 'marked'
import type { MarkedExtension } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js'

// 配置 marked 使用 highlight.js
const highlightExtension: MarkedExtension = markedHighlight({
  highlight: function (code: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch {
        // ignore
      }
    }
    return hljs.highlightAuto(code).value
  },
})

marked.use(highlightExtension)
marked.setOptions({
  breaks: true,
  gfm: true,
})

const props = defineProps<{
  content: string
}>()

const renderedContent = computed(() => {
  if (!props.content) return ''
  return marked.parse(props.content) as string
})

const containerRef = ref<HTMLElement | null>(null)

// 代码块复制功能
onMounted(() => {
  addCopyButtons()
})

watch(renderedContent, () => {
  setTimeout(addCopyButtons, 0)
})

function addCopyButtons() {
  if (!containerRef.value) return
  const codeBlocks = containerRef.value.querySelectorAll('pre')
  codeBlocks.forEach((pre) => {
    if (pre.querySelector('.copy-button')) return
    const button = document.createElement('button')
    button.className = 'copy-button'
    button.textContent = '复制'
    button.onclick = async () => {
      const code = pre.querySelector('code')?.textContent || ''
      await navigator.clipboard.writeText(code)
      button.textContent = '已复制'
      setTimeout(() => {
        button.textContent = '复制'
      }, 2000)
    }
    pre.style.position = 'relative'
    pre.appendChild(button)
  })
}
</script>

<template>
  <div ref="containerRef" class="markdown-content" v-html="renderedContent"></div>
</template>

<style scoped>
.markdown-content {
  line-height: 1.6;
  word-wrap: break-word;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  margin-top: 1em;
  margin-bottom: 0.5em;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-content :deep(h1) {
  font-size: 1.5em;
  border-bottom: 1px solid var(--border-color, #eaecef);
  padding-bottom: 0.3em;
}

.markdown-content :deep(h2) {
  font-size: 1.25em;
  border-bottom: 1px solid var(--border-color, #eaecef);
  padding-bottom: 0.3em;
}

.markdown-content :deep(h3) {
  font-size: 1.1em;
}

.markdown-content :deep(p) {
  margin: 0.5em 0;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  padding-left: 2em;
  margin: 0.5em 0;
}

.markdown-content :deep(li) {
  margin: 0.25em 0;
}

.markdown-content :deep(code) {
  padding: 0.2em 0.4em;
  margin: 0;
  font-size: 85%;
  background-color: var(--code-bg, rgba(27, 31, 35, 0.05));
  border-radius: 3px;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.markdown-content :deep(pre) {
  padding: 1em;
  overflow: auto;
  font-size: 85%;
  line-height: 1.45;
  background-color: var(--pre-bg, #f6f8fa);
  border-radius: 6px;
  margin: 0.5em 0;
}

.markdown-content :deep(pre code) {
  padding: 0;
  margin: 0;
  font-size: 100%;
  background-color: transparent;
  border-radius: 0;
}

.markdown-content :deep(blockquote) {
  padding: 0 1em;
  color: var(--blockquote-color, #6a737d);
  border-left: 0.25em solid var(--blockquote-border, #dfe2e5);
  margin: 0.5em 0;
}

.markdown-content :deep(table) {
  border-spacing: 0;
  border-collapse: collapse;
  margin: 0.5em 0;
  width: 100%;
}

.markdown-content :deep(table th),
.markdown-content :deep(table td) {
  padding: 6px 13px;
  border: 1px solid var(--table-border, #dfe2e5);
}

.markdown-content :deep(table th) {
  font-weight: 600;
  background-color: var(--table-header-bg, #f6f8fa);
}

.markdown-content :deep(table tr:nth-child(2n)) {
  background-color: var(--table-stripe-bg, #f6f8fa);
}

.markdown-content :deep(img) {
  max-width: 100%;
  box-sizing: content-box;
  background-color: #fff;
}

.markdown-content :deep(hr) {
  height: 0.25em;
  padding: 0;
  margin: 1.5em 0;
  background-color: var(--hr-color, #e1e4e8);
  border: 0;
}

.markdown-content :deep(a) {
  color: var(--link-color, #0366d6);
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

/* 复制按钮样式 */
.markdown-content :deep(.copy-button) {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px 8px;
  font-size: 12px;
  color: var(--text-color-secondary, #666);
  background-color: var(--component-bg, #fff);
  border: 1px solid var(--border-color, #d9d9d9);
  border-radius: 4px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.markdown-content :deep(pre:hover .copy-button) {
  opacity: 1;
}

.markdown-content :deep(.copy-button:hover) {
  color: var(--primary-color, #1890ff);
  border-color: var(--primary-color, #1890ff);
}

/* 深色模式适配 */
:global(.dark) .markdown-content :deep(code) {
  background-color: var(--code-bg, rgba(255, 255, 255, 0.1));
}

:global(.dark) .markdown-content :deep(pre) {
  background-color: var(--pre-bg, #1e1e1e);
}

:global(.dark) .markdown-content :deep(blockquote) {
  color: var(--blockquote-color, #8b949e);
  border-left-color: var(--blockquote-border, #30363d);
}

:global(.dark) .markdown-content :deep(table th),
:global(.dark) .markdown-content :deep(table td) {
  border-color: var(--table-border, #30363d);
}

:global(.dark) .markdown-content :deep(table th) {
  background-color: var(--table-header-bg, #21262d);
}

:global(.dark) .markdown-content :deep(table tr:nth-child(2n)) {
  background-color: var(--table-stripe-bg, #161b22);
}

:global(.dark) .markdown-content :deep(hr) {
  background-color: var(--hr-color, #21262d);
}

:global(.dark) .markdown-content :deep(a) {
  color: var(--link-color, #58a6ff);
}
</style>