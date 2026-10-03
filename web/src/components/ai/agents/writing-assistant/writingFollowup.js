// 对话目标始终来自当前正文的精确范围，不对模糊匹配结果自动写入。
export function combineShortcut(action, draft) {
  return draft.trim() ? `${action}\n补充要求：${draft.trim()}` : action
}
export function reviewTarget(target, quote) {
  if (!target || !quote) return null
  const start = target.content.indexOf(quote)
  if (start < 0 || target.content.indexOf(quote, start + 1) >= 0) return null
  return { ...target, cursor: start, selection: { editorId: target.editorId, documentId: target.documentId, revision: target.revision, content: target.content, start, end: start + quote.length, text: quote } }
}
export function appliedTarget(target, state, start, end) {
  if (!state?.active || state.documentId !== target?.documentId || !Number.isInteger(start) || !Number.isInteger(end) || start < 0 || end < start || end > state.content.length) return null
  return { ...target, ...state, cursor: start, selection: end > start ? { ...state, start, end, text: state.content.slice(start, end) } : null }
}
export function insertedTarget(target, state) {
  if (!state?.active || state.documentId !== target?.documentId) return null
  const before = target.content, after = state.content
  let start = 0, suffix = 0
  while (start < before.length && start < after.length && before[start] === after[start]) start++
  while (suffix < before.length - start && suffix < after.length - start && before[before.length - 1 - suffix] === after[after.length - 1 - suffix]) suffix++
  if (before.slice(start, before.length - suffix)) return null
  return appliedTarget(target, state, start, after.length - suffix)
}
export function historyResult(message) {
  return JSON.stringify(message.application ? { ...message.result, content: message.application.content, application: message.application } : message.result)
}
export function nearTranscriptBottom(el) {
  return !el || el.scrollHeight - el.clientHeight - el.scrollTop < 48
}
