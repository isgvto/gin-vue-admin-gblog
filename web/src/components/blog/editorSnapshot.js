// textarea 会把 CRLF 显示为 LF，DOM 选区偏移需要映射回原始 Markdown。
export function textareaOffsetToSource(content, position) {
  let offset = 0
  for (let i = 0; i < position && offset < content.length; i++) {
    offset += content[offset] === '\r' && content[offset + 1] === '\n' ? 2 : 1
  }
  return offset
}

export function sourceOffsetToTextarea(content, position) {
  return content.slice(0, position).replace(/\r\n?/g, '\n').length
}

/** 快照只留在本地；固定文章、编辑器实例、正文版本及原选区。 */
export function captureEditorSnapshot(state, selection) {
  if (!state?.active || !selection || !Number.isInteger(selection.start) ||
      !Number.isInteger(selection.end) || selection.start < 0 ||
      selection.end <= selection.start || selection.end > state.content.length) return null
  return Object.freeze({
    editorId: state.editorId,
    documentId: state.documentId,
    revision: state.revision,
    content: state.content,
    start: selection.start,
    end: selection.end,
    text: state.content.slice(selection.start, selection.end)
  })
}

export function snapshotError(snapshot, state) {
  if (!snapshot || !state?.active || snapshot.editorId !== state.editorId ||
      snapshot.documentId !== state.documentId) return '文章已切换，请回到原文章重新生成'
  if (snapshot.revision !== state.revision || snapshot.content !== state.content) {
    return '正文已发生变化，请保留当前内容并重新生成'
  }
  if (!Number.isInteger(snapshot.start) || !Number.isInteger(snapshot.end) ||
      snapshot.start < 0 || snapshot.end <= snapshot.start || snapshot.end > state.content.length ||
      state.content.slice(snapshot.start, snapshot.end) !== snapshot.text) {
    return '原选区已失效，请重新选择文本并生成'
  }
  return ''
}

export function replaceSnapshotRange(snapshot, replacement) {
  return snapshot.content.slice(0, snapshot.start) + replacement + snapshot.content.slice(snapshot.end)
}

export function applySnapshot(state, snapshot, replacement) {
  const message = snapshotError(snapshot, state)
  if (message) return { ok: false, message }
  const content = replaceSnapshotRange(snapshot, replacement)
  return {
    ok: true,
    content,
    undo: {
      editorId: state.editorId,
      documentId: state.documentId,
      before: state.content,
      after: content,
      start: snapshot.start,
      end: snapshot.end
    }
  }
}

export function undoSnapshot(state, entry) {
  if (!entry || !state?.active || entry.editorId !== state.editorId || entry.documentId !== state.documentId) {
    return { ok: false, message: '当前文章没有可撤销的 AI 修改' }
  }
  if (state.content !== entry.after) {
    return { ok: false, message: '正文已继续编辑，无法直接撤销此前的 AI 修改' }
  }
  return { ok: true, content: entry.before, start: entry.start, end: entry.end }
}

// 章节采纳只追加到当前正文末尾，并复用 AI 撤销历史；不依赖光标或选区。
export function appendDocumentSnapshot(state, expected, text) {
  if (!state?.active || !expected || state.editorId !== expected.editorId || state.documentId !== expected.documentId) {
    return { ok: false, message: '文章已切换，请重新确认大纲' }
  }
  if (state.revision !== expected.revision || state.content !== expected.content) {
    return { ok: false, message: '正文刚刚发生变化，请重新点击采纳' }
  }
  if (typeof text !== 'string' || !text.trim()) return { ok: false, message: '章节内容不能为空' }
  const newline = state.content.includes('\r\n') ? '\r\n' : '\n'
  const separator = !state.content || /(?:\r?\n){2}$/.test(state.content) ? '' : /\r?\n$/.test(state.content) ? newline : newline + newline
  const content = state.content + separator + text.trim().replace(/\r\n?|\n/g, newline) + newline
  return { ok: true, content, undo: {
    editorId: state.editorId, documentId: state.documentId, before: state.content, after: content,
    start: state.content.length, end: state.content.length
  } }
}

export function insertDocumentSnapshot(state, expected, text) {
  if (!state?.active || !expected || state.editorId !== expected.editorId || state.documentId !== expected.documentId || state.content !== expected.content || state.revision !== expected.revision) return { ok: false, message: '正文或文章已变化，请重新选择位置并生成' }
  const pos = expected.cursor
  if (!Number.isInteger(pos) || pos < 0 || pos > state.content.length || typeof text !== 'string' || !text.trim()) return { ok: false, message: '插入位置或内容无效' }
  const newline = state.content.includes('\r\n') ? '\r\n' : '\n'
  const content = state.content.slice(0, pos) + text.trim().replace(/\r\n?|\n/g, newline) + newline + state.content.slice(pos)
  return { ok: true, content, undo: { editorId: state.editorId, documentId: state.documentId, before: state.content, after: content, start: pos, end: pos } }
}
