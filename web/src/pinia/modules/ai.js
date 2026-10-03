import { defineStore } from 'pinia'
import { computed, markRaw, nextTick, reactive, ref, shallowRef } from 'vue'
import { diffMarkdownBlocks, applyDiffBlocks } from '../../components/ai/agents/writing-assistant/diff.js'
import { snapshotError, replaceSnapshotRange } from '../../components/blog/editorSnapshot.js'

/**
 * AI 平台全局状态：
 * - contexts: 各页面注册的可共享上下文（如编辑器句柄、表单回填回调）
 * - diff: 润色/改写的编辑器内 diff 状态（MarkdownEditor 开启 enable-ai-diff 后展示）
 * - agent 注册表渲染 AiDock 时从这里判断能力是否可用
 */
export const useAiStore = defineStore('ai', () => {
  const contexts = reactive({})
  const selectionAction = shallowRef(null)
  const writingBusy = ref(false)
  const requestSelectionAction = (action, snapshot) => {
    if (!['polish', 'rewrite'].includes(action)) return { ok: false, message: '不支持的选区操作' }
    const message = snapshotError(snapshot, contexts.editor?.getEditorState?.())
    if (message) return { ok: false, message }
    if (writingBusy.value || selectionAction.value || diff.active) return { ok: false, message: '请先完成当前 AI 任务或对比' }
    selectionAction.value = { action, snapshot }
    openDock('writing-assistant')
    return { ok: true }
  }

  // AiDock 抽屉开关
  const dockVisible = ref(false)
  // 当前激活的 agent id
  const activeAgentId = ref('writing-assistant')

  // 编辑器内 diff 状态
  const lastDiffResolution = shallowRef(null)
  const diff = reactive({ active: false, blocks: [], snapshot: null, applying: false, sessionId: 0 })

  const mergedDiffText = computed(() =>
    diff.active ? applyDiffBlocks(diff.blocks) : ''
  )

  const diffPreviewText = computed(() => diff.active && diff.snapshot
    ? replaceSnapshotRange(diff.snapshot, mergedDiffText.value)
    : '')

  const openEditorDiff = (snapshot, resultText) => {
    const message = snapshotError(snapshot, contexts.editor?.getEditorState?.())
    if (message) return { ok: false, message }
    if (diff.active) return { ok: false, message: '请先完成当前对比' }
    const blocks = diffMarkdownBlocks(snapshot.text, resultText)
    if (blocks.every(block => block.type === 'equal')) return { ok: false, message: 'AI 结果与原文一致，没有需要处理的修改' }
    // 待处理修改始终保留原文；点击采纳时才真正写入正文。
    diff.blocks = blocks.map(block => ({ ...block, takeRevised: false, reviewed: block.type === 'equal' }))
    diff.snapshot = snapshot
    diff.sessionId++
    diff.active = true
    return { ok: true }
  }

  const closeEditorDiff = () => {
    if (diff.active && diff.snapshot) lastDiffResolution.value = {
      sessionId: diff.sessionId, snapshot: { ...diff.snapshot },
      blocks: diff.blocks.map(block => ({ ...block }))
    }
    diff.sessionId++
    diff.active = false
    diff.blocks = []
    diff.snapshot = null
    diff.applying = false
  }

  // 每次决策即时写回固定范围，再绑定更新后的正文版本，避免前一处变长导致后一处偏移。
  const setDiffChoice = async ({ index, takeRevised }, handle = contexts.editor) => {
    if (!diff.active || diff.applying) return { ok: false, message: '请等待当前修改完成' }
    const message = snapshotError(diff.snapshot, handle?.getEditorState?.())
    if (message) return { ok: false, message }
    const blocks = diff.blocks.map((block, current) => !block.reviewed && (index === undefined || current === index)
      ? { ...block, reviewed: true, takeRevised: Boolean(takeRevised) } : { ...block })
    if (blocks.every((block, current) => block.reviewed === diff.blocks[current].reviewed)) return { ok: false, message: '这处修改已处理或不存在' }
    const snapshot = diff.snapshot
    const sessionId = diff.sessionId
    const finalText = applyDiffBlocks(blocks)
    diff.applying = true
    try {
      if (finalText !== snapshot.text) {
        const result = handle?.applySelectionSnapshot?.(snapshot, finalText)
        if (!result?.ok) return result || { ok: false, message: '当前编辑器无法应用修改' }
        await nextTick()
      }
      if (!diff.active || diff.sessionId !== sessionId) return { ok: false, message: '对比已结束' }
      const state = handle?.getEditorState?.()
      if (!state?.active || state.editorId !== snapshot.editorId || state.documentId !== snapshot.documentId || state.content !== replaceSnapshotRange(snapshot, finalText)) return { ok: false, message: '正文已变化，请结束对比后重新生成' }
      diff.blocks = blocks
      diff.snapshot = { ...snapshot, content: state.content, revision: state.revision, text: finalText, end: snapshot.start + finalText.length }
      if (blocks.every(block => block.reviewed)) closeEditorDiff()
      return { ok: true }
    } finally {
      if (diff.sessionId === sessionId) diff.applying = false
    }
  }

  const registerContext = (name, handle) => {
    if (name === 'editor' && contexts[name] !== handle) selectionAction.value = null
    contexts[name] = markRaw(handle)
  }

  const unregisterContext = (name, handle) => {
    if (handle && contexts[name] !== handle) return
    if (name === 'editor') selectionAction.value = null
    if (name === 'editor' && diff.snapshot?.editorId === contexts[name]?.getEditorState?.()?.editorId) {
      closeEditorDiff()
    }
    delete contexts[name]
  }

  const hasContext = (name) => computed(() => Boolean(contexts[name]))

  const openDock = (agentId) => {
    if (agentId) activeAgentId.value = agentId
    dockVisible.value = true
  }

  const closeDock = () => {
    dockVisible.value = false
  }

  return {
    contexts,
    selectionAction,
    writingBusy,
    requestSelectionAction,
    dockVisible,
    activeAgentId,
    diff,
    lastDiffResolution,
    mergedDiffText,
    diffPreviewText,
    openEditorDiff,
    closeEditorDiff,
    setDiffChoice,
    registerContext,
    unregisterContext,
    hasContext,
    openDock,
    closeDock
  }
})
