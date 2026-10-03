const PREFIX = 'gblog:writing-chat:v1:'
export const MAX_MESSAGES = 40
const MAX_CHARS = 500000
// randomUUID 仅在安全上下文可用，兼容通过内网 HTTP 访问的后台。
export function createChatId(random = globalThis.crypto) {
  if (random?.randomUUID) return random.randomUUID()
  if (random?.getRandomValues) return [...random.getRandomValues(new Uint32Array(4))].map(n => n.toString(16)).join('-')
  return Date.now().toString(36) + '-' + Math.random().toString(36).slice(2) + '-' + Math.random().toString(36).slice(2)
}
export const createSession = () => ({ id: createChatId(), title: '新对话', messages: [], createdAt: Date.now() })

export function chatStorageKey(user, document) {
  return user && document ? PREFIX + encodeURIComponent(String(user)) + ':' + encodeURIComponent(String(document)) : null
}

export function loadChat(storage, key) {
  if (!key) return { sessions: [createSession()], activeId: '', draft: '', version: 0 }
  const raw = storage.getItem(key)
  if (!raw) return { sessions: [createSession()], activeId: '', draft: '', version: 0 }
  const data = JSON.parse(raw)
  if (!Array.isArray(data.sessions) || !data.sessions.length || data.sessions.some(s => !s.id || !Array.isArray(s.messages))) throw new Error('聊天记录格式无效')
  for (const session of data.sessions) {
    session.messages = session.messages.filter(m => ['user', 'assistant'].includes(m.role) && typeof m.content === 'string').slice(-MAX_MESSAGES)
    for (const message of session.messages) {
      if (message.status === 'generating') message.status = 'stopped'
      if (message.role === 'assistant' && message.status === 'complete') {
        try { parseChatResult(JSON.stringify(message.result)) } catch { message.status = 'error'; message.error = '本地回复格式无效，请重新生成'; message.result = null }
      }
    }
  }
  return { ...data, sessions: data.sessions.slice(-5), draft: typeof data.draft === 'string' ? data.draft.slice(0, 4000) : '' }
}

// 一条完整回复才写入；不持久化模型调用句柄、令牌或组件引用。
export function saveChat(storage, key, data, expectedVersion) {
  if (!key) return { ok: false, reason: 'identity' }
  const current = storage.getItem(key)
  if ((current && JSON.parse(current).version !== expectedVersion) || (!current && expectedVersion)) return { ok: false, reason: 'conflict' }
  const snapshot = JSON.parse(JSON.stringify(data))
  snapshot.sessions = snapshot.sessions.slice(-5)
  for (const session of snapshot.sessions) session.messages = session.messages.slice(-MAX_MESSAGES)
  snapshot.version = createChatId()
  snapshot.updatedAt = Date.now()
  let encoded = JSON.stringify(snapshot)
  // 淘汰旧记录，不截断单条修改结果或其安全快照。
  while (encoded.length > MAX_CHARS) {
    const old = snapshot.sessions.find(s => s.id !== snapshot.activeId)
    if (old) snapshot.sessions.splice(snapshot.sessions.indexOf(old), 1)
    else if (snapshot.sessions[0].messages.length > 2) snapshot.sessions[0].messages.splice(0, 2)
    else return { ok: false, reason: 'capacity' }
    encoded = JSON.stringify(snapshot)
  }
  // 整个功能最多保存约 2 MB 文本字符；淘汰当前浏览器最旧的其他文章会话。
  const keys = []
  for (let i = 0; i < storage.length; i++) {
    const candidate = storage.key(i)
    if (candidate?.startsWith(PREFIX) && candidate !== key) keys.push(candidate)
  }
  const updated = key => { try { const data = JSON.parse(storage.getItem(key)); return data.updatedAt || Number(data.version) || 0 } catch { return 0 } }
  keys.sort((a, b) => updated(a) - updated(b))
  let total = encoded.length + keys.reduce((n, k) => n + (storage.getItem(k)?.length || 0), 0)
  while (total > 1000000 && keys.length) { const old = keys.shift(); total -= storage.getItem(old)?.length || 0; storage.removeItem(old) }
  storage.setItem(key, encoded)
  return { ok: true, data: snapshot }
}

export function parseChatResult(raw) {
  const text = String(raw).trim().replace(/^```(?:json)?\s*\n/, '').replace(/\n```$/, '')
  const result = JSON.parse(text)
  if (!result || typeof result.content !== 'string' || !result.content.trim() || result.content.length > 32000 || !['advice', 'edit', 'insert', 'outline', 'title', 'summary', 'tags', 'review'].includes(result.kind)) throw new Error('回复格式不完整，请重试；文章未被修改')
  if (result.kind === 'review' && (!Array.isArray(result.issues) || !result.issues.length || result.issues.length > 3 || result.issues.some(issue => !issue || ['quote', 'reason', 'suggestion'].some(key => typeof issue[key] !== 'string' || !issue[key].trim() || issue[key].length > 2000)))) throw new Error('审阅建议格式无效，请重试')
  if (result.kind === 'title' && (!Array.isArray(result.titles) || !result.titles.length || result.titles.length > 5 || result.titles.some(t => typeof t !== 'string' || !t.trim() || [...t].length > 120 || /[\r\n<>]/.test(t)))) throw new Error('候选标题格式无效，请重试')
  if (result.kind === 'tags') {
    if (typeof result.category !== 'string' || !Array.isArray(result.tags) || !Array.isArray(result.newTags) || result.tags.length > 3 || result.newTags.length > 2 || [...result.tags, ...result.newTags].some(t => typeof t !== 'string' || !t.trim() || [...t].length > 32 || /[<>\p{Cc}]/u.test(t))) throw new Error('标签建议格式无效，请重试')
  }
  return result
}

export function partialChatContent(raw) {
  const match = raw.match(/"content"\s*:\s*"((?:\\.|[^"\\])*)/)
  if (!match) return ''
  try { return JSON.parse('"' + match[1] + '"') } catch { return '' }
}

// 重载后编辑器实例会改变，只在文章和正文完全一致时重新绑定快照。
export function restoreChatTarget(target, state) {
  if (!target || !state?.active || target.documentId !== state.documentId || target.content !== state.content) return null
  const selection = target.selection
  if (selection && (!Number.isInteger(selection.start) || !Number.isInteger(selection.end) || selection.end > state.content.length || state.content.slice(selection.start, selection.end) !== selection.text || selection.start < 0 || selection.end <= selection.start)) return null
  return { ...target, editorId: state.editorId, revision: state.revision, selection: selection ? { ...selection, editorId: state.editorId, revision: state.revision } : null }
}
