<template>
  <div class="markdown-editor" :class="{ 'is-fullscreen': fullscreen }">
    <div class="markdown-toolbar">
      <div class="toolbar-group">
        <el-tooltip
          v-for="tool in tools"
          :key="tool.key"
          :content="tool.label"
          placement="top"
        >
          <el-button
            class="tool-button"
            :icon="tool.icon"
            text
            :disabled="aiDiffActive"
            @click="insertMarkdown(tool)"
          />
        </el-tooltip>
      </div>
      <div class="toolbar-group">
        <el-tooltip v-if="aiUndoHistory.length" :content="undoHint" placement="top">
          <span>
            <el-button size="small" :disabled="aiDiffActive || !canUndoAi" @click="undoAiChange">撤销 AI 修改</el-button>
          </span>
        </el-tooltip>
        <el-tooltip content="复制内容" placement="top">
          <el-button class="tool-button" :icon="CopyDocument" text @click="copyContent" />
        </el-tooltip>
        <el-tooltip :content="previewVisible ? '隐藏预览' : '显示预览'" placement="top">
          <el-button class="tool-button" :icon="View" text @click="previewVisible = !previewVisible" />
        </el-tooltip>
        <el-tooltip :content="fullscreen ? '退出全屏' : '全屏编辑'" placement="top">
          <el-button class="tool-button" :icon="FullScreen" text @click="fullscreen = !fullscreen" />
        </el-tooltip>
      </div>
    </div>

    <div v-if="aiDiffActive" class="ai-diff-banner">
      <span class="ai-diff-tip">{{ diffConflict || 'AI 润色对比中：逐块选择采用或保留，右侧预览全文效果' }}</span>
      <div class="ai-diff-actions">
        <el-button type="primary" size="small" :disabled="Boolean(diffConflict)" @click="applyAiDiff">应用修改</el-button>
        <el-button size="small" @click="cancelAiDiff">取消</el-button>
      </div>
    </div>

    <div class="markdown-body" :style="{ minHeight: editorHeight }">
      <div v-if="aiDiffActive" class="editor-pane diff-pane">
        <ParagraphDiff :blocks="aiStore.diff.blocks" @change="aiStore.setDiffChoice" />
      </div>
      <div v-show="!aiDiffActive" class="editor-pane" :class="{ 'is-alone': !previewVisible }">
        <textarea
          ref="textareaRef"
          v-model="value"
          class="markdown-textarea"
          :placeholder="placeholder"
          spellcheck="false"
          @keydown.tab.prevent="insertText('  ', '', '')"
        />
      </div>
      <div v-if="previewVisible || aiDiffActive" class="preview-pane">
        <div v-if="aiDiffActive" class="markdown-preview" v-html="mergedPreviewHtml" />
        <div v-else-if="value" class="markdown-preview" v-html="previewHtml" />
        <div v-else class="preview-empty">Markdown 预览</div>
      </div>
    </div>

    <div class="markdown-footer">
      <span>{{ stats.characters }} 字符</span>
      <span>{{ stats.words }} 字</span>
      <span>{{ stats.lines }} 行</span>
    </div>
    <SelectionAiToolbar :textarea="textareaRef" :enabled="selectionAiEnabled" @action="requestSelectionAi" />
  </div>
</template>

<script setup>
  import { computed, getCurrentInstance, nextTick, onActivated, onBeforeUnmount, onDeactivated, ref, watch } from 'vue'
  import { renderSafeMarkdown } from '@/utils/safeMarkdown'
  import { appendDocumentSnapshot, applySnapshot, captureEditorSnapshot, snapshotError, sourceOffsetToTextarea, textareaOffsetToSource, undoSnapshot } from './editorSnapshot'
  import {
    ChatLineSquare,
    CopyDocument,
    EditPen,
    Files,
    FullScreen,
    Link,
    List,
    MagicStick,
    Picture,
    Tickets,
    View
  } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import { useAiStore } from '@/pinia/modules/ai'
  import ParagraphDiff from '@/components/ai/agents/writing-assistant/ParagraphDiff.vue'
  import SelectionAiToolbar from './SelectionAiToolbar.vue'

  const props = defineProps({
    modelValue: {
      type: String,
      default: ''
    },
    placeholder: {
      type: String,
      default: '请输入 Markdown 内容'
    },
    height: {
      type: [Number, String],
      default: 520
    },
    // 开启后，AI 润色/改写完成的 diff 会直接铺在本编辑器中（仅正文编辑器开启）
    enableAiDiff: {
      type: Boolean,
      default: false
    },
    documentId: {
      type: String,
      default: ''
    }
  })

  const emit = defineEmits(['update:modelValue'])

  const aiStore = useAiStore()
  const editorId = `markdown-editor-${getCurrentInstance().uid}`
  const revision = ref(0)
  const active = ref(true)
  const aiUndoHistory = ref([])

  // 本编辑器实例是否处于 AI diff 模式
  const aiDiffActive = computed(() => props.enableAiDiff && aiStore.diff.active && aiStore.diff.snapshot?.editorId === editorId)
  const diffConflict = computed(() => aiDiffActive.value ? snapshotError(aiStore.diff.snapshot, getEditorState()) : '')
  const mergedPreviewHtml = computed(() => renderSafeMarkdown(diffConflict.value ? value.value : aiStore.diffPreviewText))

  const applyAiDiff = () => {
    const result = aiStore.applyEditorDiff({ applySelectionSnapshot })
    if (result.ok) {
      ElMessage.success('已应用 AI 修改')
    } else {
      ElMessage.warning(result.message)
    }
  }

  const cancelAiDiff = () => {
    aiStore.closeEditorDiff()
  }

  const textareaRef = ref()
  const previewVisible = ref(true)
  const fullscreen = ref(false)

  const value = computed({
    get: () => props.modelValue || '',
    set: (val) => emit('update:modelValue', val)
  })

  const editorHeight = computed(() => {
    if (typeof props.height === 'number') {
      return `${props.height}px`
    }
    return props.height
  })

  const previewHtml = computed(() => renderSafeMarkdown(value.value))

  watch(() => props.modelValue, () => revision.value++, { flush: 'sync' })
  watch(() => props.documentId, () => {
    revision.value++
    aiUndoHistory.value = []
    if (aiDiffActive.value) aiStore.closeEditorDiff()
  }, { flush: 'sync' })
  const deactivateEditor = () => {
    active.value = false
    revision.value++
    if (aiDiffActive.value) aiStore.closeEditorDiff()
  }
  onActivated(() => { active.value = true })
  onDeactivated(deactivateEditor)
  onBeforeUnmount(deactivateEditor)

  const getEditorState = () => ({
    editorId, documentId: props.documentId, revision: revision.value,
    content: value.value, active: active.value
  })
  const captureSelection = () => captureEditorSnapshot(getEditorState(), getSelection())
  const selectionAiEnabled = computed(() => props.enableAiDiff && active.value && !aiStore.diff.active &&
    !aiStore.writingBusy && !aiStore.selectionAction &&
    aiStore.contexts.editor?.getEditorState?.()?.editorId === editorId)
  const requestSelectionAi = (action) => {
    const result = aiStore.requestSelectionAction(action, captureSelection())
    if (!result.ok) ElMessage.warning(result.message)
    else fullscreen.value = false
  }
  const restoreSelection = (start, end) => nextTick(() => {
    textareaRef.value?.focus()
    textareaRef.value?.setSelectionRange(
      sourceOffsetToTextarea(value.value, start), sourceOffsetToTextarea(value.value, end)
    )
  })
  const applySelectionSnapshot = (snapshot, text) => {
    const result = applySnapshot(getEditorState(), snapshot, text)
    if (!result.ok) return result
    if (result.content !== value.value) {
      aiUndoHistory.value = [...aiUndoHistory.value.slice(-19), result.undo]
      value.value = result.content
    }
    restoreSelection(snapshot.start, snapshot.start + text.length)
    return { ok: true }
  }
  const undoResult = computed(() => undoSnapshot(getEditorState(), aiUndoHistory.value.at(-1)))
  const appendChapter = (text, expected) => {
    if (aiStore.diff.active) return { ok: false, message: '请先完成当前对比' }
    const result = appendDocumentSnapshot(getEditorState(), expected, text)
    if (!result.ok) return result
    aiUndoHistory.value = [...aiUndoHistory.value.slice(-19), result.undo]
    value.value = result.content
    return { ok: true }
  }
  const canUndoAi = computed(() => undoResult.value.ok)
  const undoHint = computed(() => undoResult.value.message || '恢复应用前的正文和选区')
  const undoAiChange = () => {
    if (aiDiffActive.value) return
    const result = undoResult.value
    if (!result.ok) return ElMessage.warning(result.message)
    aiUndoHistory.value.pop()
    value.value = result.content
    restoreSelection(result.start, result.end)
    ElMessage.success('已撤销 AI 修改')
  }

  const stats = computed(() => {
    const content = value.value || ''
    const words = content
      .replace(/```[\s\S]*?```/g, ' ')
      .replace(/[#>*_`~\-[\]()!|]/g, ' ')
      .match(/[\u4e00-\u9fa5]|[a-zA-Z0-9]+/g)

    return {
      characters: content.length,
      words: words ? words.length : 0,
      lines: content ? content.split(/\r?\n/).length : 0
    }
  })

  const tools = [
    { key: 'bold', label: '加粗', icon: EditPen, prefix: '**', suffix: '**', sample: '加粗文字' },
    { key: 'quote', label: '引用', icon: ChatLineSquare, block: '> 引用内容' },
    { key: 'list', label: '无序列表', icon: List, block: '- 列表项' },
    { key: 'code', label: '代码块', icon: Tickets, block: '```js\nconsole.log("hello")\n```' },
    { key: 'link', label: '链接', icon: Link, prefix: '[', suffix: '](https://)', sample: '链接文字' },
    { key: 'image', label: '图片', icon: Picture, prefix: '![', suffix: '](https://)', sample: '图片描述' },
    { key: 'table', label: '表格', icon: Files, block: '| 标题 | 内容 |\n| --- | --- |\n| 示例 | 文本 |' },
    { key: 'divider', label: '分割线', icon: MagicStick, block: '---' },
    { key: 'heading', label: '标题', icon: EditPen, block: '## 标题' }
  ]

  const focusTextarea = () => nextTick(() => textareaRef.value?.focus())

  const insertText = (prefix, suffix = '', sample = '') => {
    const textarea = textareaRef.value
    if (!textarea) return

    const { start, end } = getSelection()
    const selected = value.value.slice(start, end)
    const text = `${prefix}${selected || sample}${suffix}`
    value.value = `${value.value.slice(0, start)}${text}${value.value.slice(end)}`

    nextTick(() => {
      const cursorStart = start + prefix.length
      const cursorEnd = cursorStart + (selected || sample).length
      textarea.setSelectionRange(sourceOffsetToTextarea(value.value, cursorStart), sourceOffsetToTextarea(value.value, cursorEnd))
      textarea.focus()
    })
  }

  const insertBlock = (block) => {
    const textarea = textareaRef.value
    if (!textarea) return

    const { start, end } = getSelection()
    const before = value.value.slice(0, start)
    const after = value.value.slice(end)
    const selected = value.value.slice(start, end)
    const content = selected || block
    const leading = before && !before.endsWith('\n') ? '\n' : ''
    const trailing = after && !after.startsWith('\n') ? '\n' : ''
    const text = `${leading}${content}\n${trailing}`

    value.value = `${before}${text}${after}`

    nextTick(() => {
      const cursorStart = start + leading.length
      const cursorEnd = cursorStart + content.length
      textarea.setSelectionRange(sourceOffsetToTextarea(value.value, cursorStart), sourceOffsetToTextarea(value.value, cursorEnd))
      textarea.focus()
    })
  }

  const insertMarkdown = (tool) => {
    if (tool.block) {
      insertBlock(tool.block)
      return
    }
    insertText(tool.prefix, tool.suffix, tool.sample)
  }

  const copyContent = async () => {
    try {
      await navigator.clipboard.writeText(value.value)
      ElMessage.success('已复制')
    } catch (error) {
      ElMessage.error('复制失败')
    }
  }

  // ---- 供 AI 助手等外部场景调用的编辑器能力 ----
  const getSelection = () => {
    const textarea = textareaRef.value
    if (!textarea) return { text: '', start: 0, end: 0 }
    const start = textareaOffsetToSource(value.value, textarea.selectionStart)
    const end = textareaOffsetToSource(value.value, textarea.selectionEnd)
    return { text: value.value.slice(start, end), start, end }
  }

  const replaceSelection = (text) => {
    if (!textareaRef.value || aiDiffActive.value) return false
    const { start, end } = getSelection()
    value.value = `${value.value.slice(0, start)}${text}${value.value.slice(end)}`
    return true
  }

  const insertAtCursor = (text) => {
    const textarea = textareaRef.value
    if (!textarea || !active.value || aiDiffActive.value) return false
    const pos = textarea ? getSelection().start : value.value.length
    value.value = `${value.value.slice(0, pos)}${text}\n${value.value.slice(pos)}`
    return true
  }

  const getFullText = () => value.value

  defineExpose({
    focusTextarea,
    getSelection,
    getEditorState,
    captureSelection,
    applySelectionSnapshot,
    appendChapter,
    replaceSelection,
    insertAtCursor,
    getFullText
  })
</script>

<style scoped lang="scss">
.markdown-editor {
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #fff;
}

.markdown-editor.is-fullscreen {
  position: fixed;
  z-index: 3000;
  inset: 16px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 12px 36px rgb(0 0 0 / 16%);

  .markdown-body {
    flex: 1;
  }
}

.markdown-toolbar,
.markdown-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: #f7f8fa;
}

.markdown-toolbar {
  border-bottom: 1px solid #ebeef5;
}

.ai-diff-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border-bottom: 1px solid #f3d19e;
  background: #fdf6ec;
  color: #b88230;
  font-size: 13px;

  .ai-diff-actions {
    display: flex;
    gap: 8px;
  }
}

.diff-pane {
  overflow: auto;
  padding: 10px;
  background: #fafbfc;
}

.markdown-footer {
  gap: 16px;
  justify-content: flex-end;
  border-top: 1px solid #ebeef5;
  color: #909399;
  font-size: 12px;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
}

.tool-button {
  width: 30px;
  height: 30px;
  padding: 0;
}

.markdown-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  background: #fff;
}

.editor-pane,
.preview-pane {
  min-width: 0;
}

.editor-pane {
  border-right: 1px solid #ebeef5;
}

.editor-pane.is-alone {
  grid-column: 1 / -1;
  border-right: 0;
}

.markdown-textarea {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  min-height: inherit;
  padding: 18px;
  border: 0;
  outline: none;
  resize: none;
  color: #24292f;
  background: #fff;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 14px;
  line-height: 1.75;
}

/* 光标落点提示当前块，提升“正在编辑哪一段”的反馈 */
.markdown-textarea:focus {
  border: 0;
  box-shadow: none;
}

.preview-pane {
  overflow: auto;
  background: #fff;
}

.preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #c0c4cc;
}

.markdown-preview {
  padding: 18px 22px;
  color: #24292f;
  font-size: 16px;
  line-height: 1.7;
  text-align: justify;
  word-wrap: break-word;

  :deep(h1),
  :deep(h2),
  :deep(h3),
  :deep(h4),
  :deep(h5),
  :deep(h6) {
    margin: 1.4em 0 0.8em;
    color: #1f2328;
    font-weight: 600;
    line-height: 1.3;
    text-align: left;
  }

  :deep(h1) {
    margin-top: 0.6em;
    padding-bottom: 0.3em;
    border-bottom: 1px solid #ebeef5;
    font-size: 1.9em;
  }

  :deep(h2) {
    padding-bottom: 0.3em;
    border-bottom: 1px solid #ebeef5;
    font-size: 1.5em;
  }

  :deep(h3) {
    font-size: 1.25em;
  }

  :deep(h4) {
    font-size: 1.1em;
  }

  :deep(h5),
  :deep(h6) {
    font-size: 1em;
  }

  :deep(p) {
    margin: 0 0 1em;
  }

  :deep(ul),
  :deep(ol) {
    margin: 0 0 1em;
    padding-left: 1.8em;
    text-align: left;
  }

  :deep(ul) {
    list-style: disc;
  }

  :deep(ol) {
    list-style: decimal;
  }

  :deep(li) {
    margin: 0.3em 0;
  }

  :deep(li > ul) {
    margin: 0.2em 0 0;
    list-style: circle;
  }

  :deep(li > ol) {
    margin: 0.2em 0 0;
    list-style: lower-alpha;
  }

  :deep(.contains-task-list) {
    padding-left: 1.2em;
    list-style: none;
  }

  :deep(input[type='checkbox']) {
    margin-right: 6px;
  }

  :deep(hr) {
    height: 1px;
    margin: 2em 0;
    border: 0;
    background: #e7e9ee;
  }

  :deep(a) {
    color: #0969da;
    text-decoration: none;
  }

  :deep(a:hover) {
    color: #0550ae;
    text-decoration: underline;
  }

  :deep(blockquote) {
    margin: 1em 0;
    padding: 0.8em 1.25em;
    border-left: 4px solid #0969da;
    border-radius: 6px;
    background: #f6f8fa;
    color: #57606a;
  }

  :deep(blockquote) :deep(p:last-child),
  :deep(blockquote) :deep(ul:last-child),
  :deep(blockquote) :deep(ol:last-child) {
    margin-bottom: 0;
  }

  :deep(pre) {
    overflow: auto;
    padding: 14px;
    border: 1px solid #eceff2;
    border-radius: 6px;
    background: #f6f8fa;
  }

  :deep(code) {
    margin: 0 2px;
    padding: 0.2em 0.4em;
    border: 1px solid #e3e6ea;
    border-radius: 4px;
    background: #f6f8fa;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
    font-size: 0.88em;
    color: #cf222e;
  }

  :deep(pre code) {
    margin: 0;
    padding: 0;
    border: 0;
    background: transparent;
    color: #24292f;
  }

  :deep(.hljs) {
    color: #24292e;
    background: #f6f8fa;
  }

  :deep(.hljs-doctag),
  :deep(.hljs-keyword),
  :deep(.hljs-meta .hljs-keyword),
  :deep(.hljs-template-tag),
  :deep(.hljs-template-variable),
  :deep(.hljs-type),
  :deep(.hljs-variable.language_) {
    color: #d73a49;
  }

  :deep(.hljs-title),
  :deep(.hljs-title.class_),
  :deep(.hljs-title.class_.inherited__),
  :deep(.hljs-title.function_) {
    color: #6f42c1;
  }

  :deep(.hljs-attr),
  :deep(.hljs-attribute),
  :deep(.hljs-literal),
  :deep(.hljs-meta),
  :deep(.hljs-number),
  :deep(.hljs-operator),
  :deep(.hljs-selector-attr),
  :deep(.hljs-selector-class),
  :deep(.hljs-selector-id),
  :deep(.hljs-variable) {
    color: #005cc5;
  }

  :deep(.hljs-regexp),
  :deep(.hljs-string),
  :deep(.hljs-meta .hljs-string) {
    color: #032f62;
  }

  :deep(.hljs-built_in),
  :deep(.hljs-symbol) {
    color: #e36209;
  }

  :deep(.hljs-code),
  :deep(.hljs-comment),
  :deep(.hljs-formula) {
    color: #6a737d;
  }

  :deep(img) {
    display: block;
    max-width: 100%;
    margin: 0 auto;
    border-radius: 6px;
  }

  :deep(table) {
    width: 100%;
    margin: 0 0 1em;
    border-collapse: collapse;
  }

  :deep(th),
  :deep(td) {
    padding: 8px 12px;
    border: 1px solid #dcdfe6;
  }

  :deep(th) {
    background: #f6f8fa;
    font-weight: 600;
  }

  :deep(tbody tr:nth-child(even)) {
    background: #fafbfc;
  }
}

@media (max-width: 960px) {
  .markdown-body {
    grid-template-columns: 1fr;
  }

  .editor-pane {
    border-right: 0;
    border-bottom: 1px solid #ebeef5;
  }

  .preview-pane {
    min-height: 320px;
  }
}
</style>
