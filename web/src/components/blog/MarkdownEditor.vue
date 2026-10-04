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
      <span class="ai-diff-tip">{{ diffConflict || '点击采用立即替换，保留原文立即结束这一处；全部处理后自动退出对比' }}</span>
      <div class="ai-diff-actions">
        <el-button size="small" :disabled="aiStore.diff.applying" @click="cancelAiDiff">结束对比</el-button>
      </div>
    </div>

    <div class="markdown-body" :style="{ minHeight: editorHeight }">
      <div v-if="aiDiffActive" class="editor-pane diff-pane">
        <ParagraphDiff :blocks="aiStore.diff.blocks" :before-text="diffBeforeText" :after-text="diffAfterText" :disabled="aiStore.diff.applying || Boolean(diffConflict)" @change="resolveAiDiff" />
      </div>
      <div v-show="!aiDiffActive" class="editor-pane" :class="{ 'is-alone': !previewVisible, 'has-paragraph-numbers': enableAiDiff }">
        <div v-if="enableAiDiff" class="paragraph-gutter" :style="{ height: `${textareaViewportHeight}px` }" aria-label="正文段落编号">
          <button v-for="paragraph in gutterParagraphs" :key="paragraph.number" type="button"
            class="paragraph-number" :class="{ 'is-current': paragraph.number === currentParagraph }"
            :style="{ top: `${paragraph.top - textareaScrollTop}px` }"
            :aria-label="`选中第${paragraph.number}段`" :title="`第${paragraph.number}段，点击选中；可对 AI 说：润色第${paragraph.number}段`"
            :disabled="aiStore.writingBusy || aiDiffActive" @click="selectParagraph(paragraph)">{{ paragraph.number }}</button>
        </div>
        <textarea
          ref="textareaRef"
          v-model="value"
          class="markdown-textarea"
          :placeholder="placeholder"
          spellcheck="false"
          @scroll="textareaScrollTop = $event.target.scrollTop"
          @select="syncCaret" @click="syncCaret" @keyup="syncCaret"
          @keydown.tab.prevent="insertText('  ', '', '')"
        />
      </div>
      <div v-if="previewVisible || aiDiffActive" class="preview-pane">
        <div v-if="aiDiffActive" class="markdown-preview" v-mermaid="mergedPreviewHtml" v-html="mergedPreviewHtml" />
        <div v-else-if="value" class="markdown-preview" v-mermaid="previewHtml" v-html="previewHtml" />
        <div v-else class="preview-empty">Markdown 预览</div>
      </div>
    </div>

    <div class="markdown-footer">
      <span v-if="enableAiDiff" class="paragraph-help" title="空行不计数；标题、列表、引用、表格、完整代码块各按一段编号。编号仅在编辑时显示，不写入文章。">{{ paragraphs.length }} 段 · 可对 AI 说“润色第3段”</span>
      <span>{{ stats.characters }} 字符</span>
      <span>{{ stats.words }} 字</span>
      <span>{{ stats.lines }} 行</span>
    </div>
    <SelectionAiToolbar :textarea="textareaRef" :enabled="selectionAiEnabled" @action="requestSelectionAi" />
  </div>
</template>

<script setup>
  import { computed, getCurrentInstance, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
  import { renderSafeMarkdown } from '@/utils/safeMarkdown'
  import { vMermaid } from '@/utils/mermaid'
  import { textareaSelectionPoint } from './selectionToolbarPosition.js'
  import { appendDocumentSnapshot, insertDocumentSnapshot, applySnapshot, captureEditorSnapshot, snapshotError, sourceOffsetToTextarea, textareaOffsetToSource, undoSnapshot } from './editorSnapshot'
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
  import { indexParagraphs } from './paragraphIndex.js'
  import { measureParagraphGutter } from './paragraphGutter.js'

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
  const diffBeforeText = computed(() => aiDiffActive.value ? aiStore.diff.snapshot.content.slice(0, aiStore.diff.snapshot.start) : '')
  const diffAfterText = computed(() => aiDiffActive.value ? aiStore.diff.snapshot.content.slice(aiStore.diff.snapshot.end) : '')

  const resolveAiDiff = async (choice) => {
    const result = await aiStore.setDiffChoice(choice, { applySelectionSnapshot, getEditorState })
    if (!result.ok) ElMessage.warning(result.message)
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

  const paragraphs = computed(() => props.enableAiDiff ? indexParagraphs(value.value) : [])
  const gutterParagraphs = ref([]), textareaScrollTop = ref(0), textareaViewportHeight = ref(0), caretOffset = ref(0)
  const currentParagraph = computed(() => paragraphs.value.find(paragraph => caretOffset.value >= paragraph.start && caretOffset.value <= paragraph.end)?.number)
  const syncCaret = () => { caretOffset.value = getSelection().start }
  let gutterFrame = 0, gutterObserver
  const updateGutter = () => {
    cancelAnimationFrame(gutterFrame)
    gutterFrame = requestAnimationFrame(() => {
      gutterParagraphs.value = measureParagraphGutter(textareaRef.value, paragraphs.value)
      textareaScrollTop.value = textareaRef.value?.scrollTop || 0
      textareaViewportHeight.value = textareaRef.value?.clientHeight || 0
      syncCaret()
    })
  }
  watch([value, previewVisible, aiDiffActive, () => props.enableAiDiff], () => nextTick(updateGutter))
  onMounted(() => {
    if (textareaRef.value) {
      gutterObserver = new ResizeObserver(updateGutter)
      gutterObserver.observe(textareaRef.value)
    }
    updateGutter()
  })
  onActivated(() => nextTick(updateGutter))
  onBeforeUnmount(() => { cancelAnimationFrame(gutterFrame); gutterObserver?.disconnect() })
  async function selectParagraph(paragraph) {
    const snapshot = captureEditorSnapshot(getEditorState(), paragraph)
    const result = await selectSnapshot(snapshot)
    if (!result.ok) ElMessage.warning(result.message)
    syncCaret()
  }

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
  const selectSnapshot = async (snapshot) => {
    const state = getEditorState()
    const message = snapshotError(snapshot, state)
    if (message) return { ok: false, message }
    await restoreSelection(snapshot.start, snapshot.end)
    const el = textareaRef.value
    if (el) {
      const point = textareaSelectionPoint(el, true)
      if (point) el.scrollTop += point.y - el.getBoundingClientRect().top - el.clientHeight / 2
      el.scrollIntoView({ block: 'center' })
    }
    return { ok: true }
  }
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
    return { ok: true, range: { start: result.undo.start, end: result.undo.start + result.content.length - result.undo.before.length } }
  }
  const insertAiAt = (text, expected) => {
    if (aiStore.diff.active) return { ok: false, message: '请先完成当前对比' }
    const result = insertDocumentSnapshot(getEditorState(), expected, text)
    if (!result.ok) return result
    aiUndoHistory.value = [...aiUndoHistory.value.slice(-19), result.undo]
    value.value = result.content
    return { ok: true, range: { start: result.undo.start, end: result.undo.start + result.content.length - result.undo.before.length } }
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
    selectSnapshot,
    applySelectionSnapshot,
    appendChapter,
    insertAiAt,
    replaceSelection,
    insertAtCursor,
    getFullText
  })
</script>

<style scoped lang="scss">
.markdown-editor {
  --md-token-red: #cf222e;
  --md-token-purple: #6f42c1;
  --md-token-blue: #005cc5;
  --md-token-string: #032f62;
  --md-token-orange: #b65300;
  --md-token-comment: #6a737d;
  color: var(--admin-text, var(--el-text-color-primary));
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
  border: 1px solid var(--admin-border, var(--el-border-color));
  border-radius: 6px;
  background: var(--admin-surface, var(--el-bg-color));
}

:global(html.dark .markdown-editor) {
  --md-token-red: #ff9aa8;
  --md-token-purple: #d2b5ff;
  --md-token-blue: #9bc7ff;
  --md-token-string: #a5d9b5;
  --md-token-orange: #ffc58d;
  --md-token-comment: #a6b4c6;
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
  background: var(--el-fill-color-light);
}

.markdown-toolbar {
  border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter));
}

.ai-diff-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--el-color-warning-light-5);
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
  font-size: 13px;

  .ai-diff-actions {
    display: flex;
    gap: 8px;
  }
}

.diff-pane {
  overflow: auto;
  padding: 10px;
  background: var(--el-fill-color-lighter);
}

.markdown-footer {
  gap: 16px;
  flex-wrap: wrap;
  row-gap: 6px;
  justify-content: flex-end;
  border-top: 1px solid var(--admin-border, var(--el-border-color-lighter));
  color: var(--admin-muted, var(--el-text-color-secondary));
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
  background: var(--admin-surface, var(--el-bg-color));
}

.editor-pane,
.preview-pane {
  min-width: 0;
}

.editor-pane {
  position: relative;
  border-right: 1px solid var(--admin-border, var(--el-border-color-lighter));
}

.has-paragraph-numbers .markdown-textarea { padding-left: 64px; }
.paragraph-gutter {
  position: absolute;
  top: 0; bottom: 0; left: 0;
  width: 44px;
  overflow: hidden;
  border-right: 1px solid var(--admin-border, var(--el-border-color-lighter));
  background: var(--el-fill-color-light);
  z-index: 1;
}
.paragraph-number {
  position: absolute;
  left: 0;
  width: 43px;
  height: 24.5px;
  padding: 0 7px 0 2px;
  border: 0;
  background: transparent;
  color: var(--admin-muted, var(--el-text-color-secondary));
  font-family: ui-monospace, monospace;
  font-size: 12px;
  line-height: 24.5px;
  text-align: right;
  cursor: pointer;
}
.paragraph-number:hover, .paragraph-number.is-current {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
.paragraph-number:focus-visible { outline: 1px solid var(--el-color-primary); outline-offset: -2px; }
.paragraph-help { margin-right: auto; cursor: help; }

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
  color: var(--admin-text, var(--el-text-color-primary));
  background: var(--admin-surface, var(--el-bg-color));
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 14px;
  line-height: 1.75;
}

/* 光标落点提示当前块，提升“正在编辑哪一段”的反馈 */
.markdown-textarea::placeholder {
  color: var(--el-text-color-placeholder);
}

.markdown-textarea:focus {
  border: 0;
  box-shadow: none;
}

.preview-pane {
  overflow: auto;
  background: var(--admin-surface, var(--el-bg-color));
}

.preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--el-text-color-placeholder);
}

.markdown-preview {
  padding: 18px 22px;
  color: var(--admin-text, var(--el-text-color-primary));
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
    color: var(--admin-text, var(--el-text-color-primary));
    font-weight: 600;
    line-height: 1.3;
    text-align: left;
  }

  :deep(h1) {
    margin-top: 0.6em;
    padding-bottom: 0.3em;
    border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter));
    font-size: 1.9em;
  }

  :deep(h2) {
    padding-bottom: 0.3em;
    border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter));
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
    background: var(--el-border-color);
  }

  :deep(a) {
    color: var(--el-color-primary);
    text-decoration: none;
  }

  :deep(a:hover) {
    color: var(--el-color-primary-light-3);
    text-decoration: underline;
  }

  :deep(blockquote) {
    margin: 1em 0;
    padding: 0.8em 1.25em;
    border-left: 4px solid var(--el-color-primary);
    border-radius: 6px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-secondary);
  }

  :deep(blockquote) :deep(p:last-child),
  :deep(blockquote) :deep(ul:last-child),
  :deep(blockquote) :deep(ol:last-child) {
    margin-bottom: 0;
  }

  :deep(pre) {
    overflow: auto;
    padding: 14px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-light);
  }

  :deep(code) {
    margin: 0 2px;
    padding: 0.2em 0.4em;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 4px;
    background: var(--el-fill-color-light);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
    font-size: 0.88em;
    color: var(--md-token-red);
  }

  :deep(pre code) {
    margin: 0;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--admin-text, var(--el-text-color-primary));
  }

  :deep(.hljs) {
    color: var(--admin-text, var(--el-text-color-primary));
    background: var(--el-fill-color-light);
  }

  :deep(.hljs-doctag),
  :deep(.hljs-keyword),
  :deep(.hljs-meta .hljs-keyword),
  :deep(.hljs-template-tag),
  :deep(.hljs-template-variable),
  :deep(.hljs-type),
  :deep(.hljs-variable.language_) {
    color: var(--md-token-red);
  }

  :deep(.hljs-title),
  :deep(.hljs-title.class_),
  :deep(.hljs-title.class_.inherited__),
  :deep(.hljs-title.function_) {
    color: var(--md-token-purple);
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
    color: var(--md-token-blue);
  }

  :deep(.hljs-regexp),
  :deep(.hljs-string),
  :deep(.hljs-meta .hljs-string) {
    color: var(--md-token-string);
  }

  :deep(.hljs-built_in),
  :deep(.hljs-symbol) {
    color: var(--md-token-orange);
  }

  :deep(.hljs-code),
  :deep(.hljs-comment),
  :deep(.hljs-formula) {
    color: var(--md-token-comment);
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
    border: 1px solid var(--admin-border, var(--el-border-color));
  }

  :deep(th) {
    background: var(--el-fill-color-light);
    font-weight: 600;
  }

  :deep(tbody tr:nth-child(even)) {
    background: var(--el-fill-color-lighter);
  }
}

@media (max-width: 960px) {
  .markdown-body {
    grid-template-columns: 1fr;
  }

  .editor-pane {
    border-right: 0;
    border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter));
  }

  .preview-pane {
    min-height: 320px;
  }
}
</style>
