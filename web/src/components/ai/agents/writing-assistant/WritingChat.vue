<template>
  <section class="writing-chat" aria-label="写作对话">
    <header class="chat-heading">
      <div><strong>一起把文章写好</strong><span>理清思路 · 修改表达 · 审阅文章</span></div>
      <el-button link size="small" :disabled="busy || blocked" @click="newConversation">新对话</el-button>
    </header>
    <div class="session-controls">
      <el-select v-model="activeId" size="small" aria-label="选择写作会话" :disabled="busy || blocked" @change="switchSession">
        <el-option v-for="session in sessions" :key="session.id" :value="session.id" :label="session.title" :aria-label="session.title">
          <div class="session-option"><span>{{ session.title }}</span><el-button link type="danger" size="small" :aria-label="`删除会话：${session.title}`" :disabled="busy || blocked || storageConflict" @mousedown.prevent.stop @keydown.stop @click.stop="deleteConversation(session.id)">删除</el-button></div>
        </el-option>
      </el-select>
      <el-button link size="small" :disabled="busy || blocked" @click="clearConversation">清空</el-button>
    </div>
    <div v-if="storageNotice" class="chat-notice" role="status">{{ storageNotice }}<el-button v-if="storageConflict" link size="small" :disabled="busy || blocked" @click="reload">载入本地记录</el-button></div>
    <div ref="transcript" class="chat-transcript" role="log" aria-label="聊天记录" tabindex="0" @scroll="onTranscriptScroll">
      <div v-if="!messages.length" class="chat-empty"><span class="empty-mark">✎</span><h3>从你想表达的内容开始</h3><p>告诉我你的想法，或选中一段正文。可以先讨论，再决定是否修改。</p></div>
      <article v-for="message in messages" :key="message.id" class="chat-message" :class="message.role" :aria-label="message.role === 'user' ? '你的消息' : '助手回复'">
        <div class="message-heading"><span>{{ message.role === 'user' ? '你' : '写作助手' }}</span><small>{{ message.scope || '' }}</small></div>
        <p v-if="message.role === 'user'" class="user-text">{{ message.content }}</p>
        <div v-else-if="message.content" class="message-content" v-mermaid="message.content" v-html="renderSafeMarkdown(message.content)" />
        <p v-else-if="message.status === 'generating'" class="chat-thinking">{{ progress || '正在思考…' }}</p>
        <p v-if="message.notice" class="message-note">{{ message.notice }}</p>
        <p v-if="message.error" class="message-error" role="status">{{ message.error }}</p>
        <template v-if="message.role === 'assistant' && message.status === 'complete'">
          <div v-if="message.result.kind === 'title'" class="chat-titles"><div v-for="title in message.result.titles" :key="title"><span>{{ title }}</span><el-button size="small" :disabled="unavailable(message)" @click="adoptTitle(message, title)">采用标题</el-button></div></div>
          <div v-if="message.result.kind === 'tags'" class="chat-tags">
            <el-checkbox v-if="message.result.category" v-model="message.includeCategory">分类：{{ message.result.category }}</el-checkbox>
            <el-checkbox-group v-model="message.selectedTags"><el-checkbox v-for="tag in message.result.tags" :key="tag" :value="tag">{{ tag }}</el-checkbox></el-checkbox-group>
            <small v-if="message.result.newTags.length">可选新标签，采用并保存文章时创建：</small>
            <el-checkbox-group v-model="message.selectedNewTags"><el-checkbox v-for="tag in message.result.newTags" :key="tag" :value="tag">{{ tag }}</el-checkbox></el-checkbox-group>
          </div>
          <div v-if="message.result.kind === 'review'" class="review-issues">
            <section v-for="(issue, index) in message.result.issues" :key="index" class="review-issue">
              <blockquote>{{ issue.quote }}</blockquote><p>{{ issue.reason }}</p><small>{{ issue.suggestion }}</small>
              <div class="message-actions"><el-button size="small" :disabled="busy || blocked || !issueTarget(message, issue)" @click="locateIssue(message, issue)">定位原文</el-button><el-button size="small" :disabled="busy || blocked || !issueTarget(message, issue)" @click="discussIssue(message, issue)">讨论这一处</el-button><el-button size="small" :disabled="busy || blocked || !issueTarget(message, issue)" @click="discussIssue(message, issue, true)">按建议修改</el-button></div>
              <small v-if="!issueTarget(message, issue)" class="message-note">原文无法唯一定位或正文已变化，请重新审阅。</small>
            </section>
          </div>
          <div class="message-actions">
            <el-button v-if="message.result.kind === 'edit'" type="primary" size="small" :disabled="unavailable(message)" @click="compare(message)">查看修改对比</el-button>
            <el-button v-if="['insert', 'outline'].includes(message.result.kind)" type="primary" size="small" :disabled="unavailable(message)" @click="insert(message)">插入原光标位置</el-button>
            <el-button v-if="['insert', 'outline'].includes(message.result.kind)" size="small" :disabled="unavailable(message)" @click="append(message)">追加到文末</el-button>
            <el-button v-if="message.result.kind === 'outline'" size="small" :disabled="busy || blocked" @click="emit('outline', message.content)">按大纲逐节写作</el-button>
            <el-button v-if="message.result.kind === 'summary'" type="primary" size="small" :disabled="unavailable(message)" @click="adoptSummary(message)">采用摘要</el-button>
            <el-button v-if="message.result.kind === 'tags'" type="primary" size="small" :disabled="unavailable(message)" @click="adoptTags(message)">采用分类与标签</el-button>
            <el-button size="small" :disabled="busy || blocked" @click="revise(message)">继续调整</el-button>
            <el-button link size="small" @click="copy(message.content)">复制</el-button>
          </div>
          <small v-if="message.adopted" class="adopted-note">已采用；继续调整将引用实际保留的正文。文章仍需保存。</small><small v-else-if="message.resolved" class="message-note">已保留原文；继续调整将引用原文。</small>
          <small v-else-if="message.target && !currentTarget(message)" class="message-note">正文已变化，旧结果仍可查看和复制。请重新生成后再应用。</small>
        </template>
        <template v-if="['error', 'stopped'].includes(message.status)"><el-button size="small" :disabled="busy || blocked" @click="retry(message)">重新尝试</el-button><el-button v-if="message.content" link size="small" @click="copy(message.content)">复制部分内容</el-button></template>
      </article>
    </div>
    <el-button v-if="newReply" class="new-reply" size="small" @click="scrollBottom(true)">查看最新回复 ↓</el-button>
    <footer class="chat-composer">
      <div v-if="replyTo" class="followup-context" role="status"><span>{{ followupLabel }}</span><el-button link size="small" :disabled="busy || blocked" @click="cancelFollowup">取消引用</el-button></div>
      <div class="scope-row"><el-select v-model="scope" size="small" aria-label="对话参考范围" :disabled="busy || blocked || !!replyTo"><el-option label="自动：选区或继续上文" value="auto" /><el-option label="参考整篇文章" value="article" /><el-option label="当前选区" value="selection" :disabled="!selectionText" /></el-select><span>{{ scopeLabel }}</span></div>
      <div class="chat-shortcuts"><button v-for="item in shortcuts" :key="item.label" type="button" :disabled="busy || blocked" @click="send(combineShortcut(item.text, draft))">{{ item.label }}</button></div>
      <el-input v-model="draft" type="textarea" :rows="3" :maxlength="4000" aria-label="写作要求" placeholder="说说你想怎么写，例如：这段再简洁一点，保留例子。" :disabled="busy || blocked" @keydown="onKeydown" />
      <div class="composer-actions"><small>记录仅存于此浏览器 · Enter 发送，Shift+Enter 换行</small><el-button v-if="busy" size="small" @click="stop">停止生成</el-button><el-button v-else type="primary" size="small" :disabled="!draft.trim() || blocked || storageConflict" @click="send()">发送</el-button></div>
      <p v-if="blocked" class="message-note">请先在正文完成当前修改对比，再继续对话。</p>
    </footer>
  </section>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAiStore } from '@/pinia/modules/ai'
import { useUserStore } from '@/pinia/modules/user'
import { streamAiChat } from '@/api/blog/ai'
import { renderSafeMarkdown } from '@/utils/safeMarkdown'
import { vMermaid } from '@/utils/mermaid'
import { chatStorageKey, createChatId, createSession, loadChat, saveChat, parseChatResult, partialChatContent, restoreChatTarget } from './chatSession.js'
import { combineShortcut, reviewTarget, appliedTarget, insertedTarget, historyResult, nearTranscriptBottom } from './writingFollowup.js'
const props = defineProps({ blocked: Boolean, preferences: Object })
const emit = defineEmits(['busy', 'outline'])
const store = useAiStore(), user = useUserStore(), context = computed(() => store.contexts.editor)
const sessions = ref([createSession()]), activeId = ref(sessions.value[0].id), draft = ref(''), scope = ref('auto')
const messages = computed(() => sessions.value.find(s => s.id === activeId.value)?.messages || [])
const busy = ref(false), transcript = ref(), selectionText = ref(''), progress = ref(''), replyTo = ref(null)
const pinnedTarget = ref(null), followBottom = ref(true), newReply = ref(false)
const storageNotice = ref(''), storageConflict = ref(false)
let storageKey = null, version = 0, handle = null, running = null, draftTimer = null, disposed = false, corrupt = false
const account = computed(() => user.userInfo?.uuid || user.userInfo?.ID || user.userInfo?.id || user.userInfo?.userName || '')
const identity = computed(() => chatStorageKey(account.value, context.value?.getEditorState?.()?.active ? context.value.getEditorState().documentId : null))
watch(busy, value => emit('busy', value), { flush: 'sync' })
const syncSelection = () => { selectionText.value = context.value?.getSelection?.()?.text || '' }
document.addEventListener('selectionchange', syncSelection)
const previousTarget = computed(() => [...messages.value].reverse().find(m => m.role === 'assistant' && m.target))
const chosenTarget = () => {
  const ctx = context.value, state = ctx?.getEditorState?.()
  if (!state?.active) return null
  if (replyTo.value) return restoreChatTarget(pinnedTarget.value || replyTo.value.continuationTarget || replyTo.value.target, state)
  const selection = scope.value !== 'article' ? ctx.captureSelection?.() : null
  if (!selection && scope.value === 'auto') {
    const previous = restoreChatTarget(previousTarget.value?.continuationTarget || previousTarget.value?.target, state)
    if (previous?.selection) return previous
  }
  return { ...state, selection, cursor: ctx.getSelection?.()?.start ?? state.content.length, title: ctx.getTitle?.() || '', description: ctx.getDescription?.() || '' }
}
const followupLabel = computed(() => { const index = messages.value.filter(m => m.role === 'assistant').indexOf(replyTo.value) + 1; return pinnedTarget.value ? `正在讨论第 ${index} 条审阅中的这段原文` : `继续第 ${index} 条回复${replyTo.value?.application ? ' · 以实际保留的正文为准' : ''}` })
const scopeLabel = computed(() => replyTo.value ? followupLabel.value : scope.value !== 'article' && selectionText.value ? `选区 ${[...selectionText.value].length} 字` : scope.value === 'auto' && restoreChatTarget(previousTarget.value?.continuationTarget || previousTarget.value?.target, context.value?.getEditorState?.())?.selection ? '沿用上一段选区' : '文章正文')
const shortcuts = [{label:'润色',text:'润色当前选区，尽量保留原句和我的文风。'}, {label:'改写',text:'改写当前选区，换一种更清楚的表达，保留原意。'}, {label:'续写',text:'从当前光标前的内容自然续写一到两段。'}, {label:'起标题',text:'根据文章给我五个不夸张的候选标题。'}, {label:'摘要',text:'根据文章生成80到150字的摘要。'}, {label:'推荐标签',text:'为文章推荐分类和标签。'}, {label:'审阅文章',text:'从读者角度指出文章最需要改进的两三个问题，先给建议，不修改正文。'}, {label:'生成大纲',text:'根据文章主题和我的想法生成Markdown大纲。'}]
function persist() {
  if (storageConflict.value) return
  try {
    const result = saveChat(localStorage, storageKey, { sessions: sessions.value, activeId: activeId.value, draft: draft.value }, version)
    if (!result.ok) {
      storageConflict.value = result.reason === 'conflict'
      storageNotice.value = ({ identity:'当前未识别登录用户或文章，记录暂存在页面中。', conflict:'另一标签页已更新记录，请先载入本地记录。', capacity:'这次内容过大，暂未保存，请复制重要结果。' })[result.reason]
    } else {
      version = result.data.version
      // 只同步淘汰结果，保留 Vue 消息对象与流式任务引用。
      const kept = new Set(result.data.sessions.map(s => s.id))
      sessions.value = sessions.value.filter(s => kept.has(s.id))
      for (const session of sessions.value) { const saved = result.data.sessions.find(s => s.id === session.id); session.messages = session.messages.filter(m => saved.messages.some(item => item.id === m.id)) }
      storageNotice.value = ''
    }
  } catch { storageNotice.value = '浏览器本地保存失败，当前对话仍可使用；请复制重要结果。' }
}
function reload() {
  if (busy.value || props.blocked) return
  clearTimeout(draftTimer); cancelFollowup()
  try {
    const data = loadChat(localStorage, storageKey)
    sessions.value = data.sessions; activeId.value = data.sessions.some(s => s.id === data.activeId) ? data.activeId : data.sessions[0].id
    draft.value = data.draft; version = data.version || 0; storageConflict.value = false; storageNotice.value = ''; corrupt = false
  } catch (error) {
    sessions.value = [createSession()]; activeId.value = sessions.value[0].id; draft.value = ''
    const unavailable = ['SecurityError', 'QuotaExceededError'].includes(error?.name)
    storageNotice.value = unavailable ? '浏览器不允许本地存储，对话暂存在当前页面，刷新后会丢失。' : '本地聊天记录无法读取，请清空后开始；现有记录未被覆盖。'
    storageConflict.value = !unavailable; corrupt = !unavailable; version = 0
  }
  syncSelection(); scrollBottom(true)
}
function onTranscriptScroll() { followBottom.value = nearTranscriptBottom(transcript.value); if (followBottom.value) newReply.value = false }
function scrollBottom(force = false) {
  if (force) { followBottom.value = true; newReply.value = false }
  if (!followBottom.value) { newReply.value = true; return }
  nextTick(() => { if (transcript.value && followBottom.value) transcript.value.scrollTop = transcript.value.scrollHeight })
}
function cancelFollowup() { replyTo.value = null; pinnedTarget.value = null }
function newConversation() { if (busy.value || props.blocked || storageConflict.value) return; cancelFollowup(); const session = createSession(); sessions.value.push(session); activeId.value = session.id; draft.value = ''; scope.value = 'auto'; persist() }
async function deleteConversation(id) {
  if (busy.value || props.blocked || storageConflict.value) return
  const key = storageKey, session = sessions.value.find(item => item.id === id)
  if (!session) return
  try { await ElMessageBox.confirm(`删除对话「${session.title}」及其聊天记录？文章正文会保留。`, '删除对话', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }) } catch { return }
  if (storageKey !== key || busy.value || props.blocked || storageConflict.value) return
  const index = sessions.value.findIndex(item => item.id === id)
  if (index < 0) return
  sessions.value.splice(index, 1)
  if (!sessions.value.length) sessions.value.push(createSession())
  if (activeId.value === id) {
    activeId.value = sessions.value[Math.min(index, sessions.value.length - 1)].id
    cancelFollowup(); draft.value = ''; scope.value = 'auto'; scrollBottom(true)
  }
  persist()
}
async function clearConversation() {
  if (busy.value || props.blocked) return
  const key = storageKey, id = activeId.value
  try { await ElMessageBox.confirm('清空当前对话记录？文章正文和其他对话会保留。', '清空对话', {confirmButtonText:'清空',cancelButtonText:'保留',type:'warning'}) } catch { return }
  if (storageKey !== key || activeId.value !== id || busy.value || props.blocked) return
  if (corrupt) { try { localStorage.removeItem(storageKey); version = 0; corrupt = false; storageConflict.value = false } catch { return ElMessage.warning('浏览器不允许清理本地记录') } }
  if (storageConflict.value) { storageNotice.value = '请先载入另一标签页的记录再清空。'; return }
  const session = sessions.value.find(s => s.id === id); session.messages = []; session.title = '新对话'; cancelFollowup(); draft.value = ''; scope.value = 'auto'; persist()
}
function switchSession() { cancelFollowup(); draft.value = ''; scope.value = 'auto'; persist(); scrollBottom(true) }
function currentTarget(message) { return restoreChatTarget(message.target, context.value?.getEditorState?.()) }
function unavailable(message) { return busy.value || props.blocked || (message.adopted || message.resolved) || !currentTarget(message) }
function normalizeTags(result) {
  if (result.kind !== 'tags') return
  const catalog = context.value?.getTaxonomy?.() || {categories:[],tags:[]}
  const names = new Map(catalog.tags.map(t => [t.tagName.toLowerCase(), t.tagName]))
  const all = [...new Set([...result.tags, ...result.newTags].map(t => t.trim()))]
  result.tags = all.filter(t => names.has(t.toLowerCase())).map(t => names.get(t.toLowerCase())).slice(0, 3)
  result.newTags = all.filter(t => !names.has(t.toLowerCase())).slice(0, 2)
  result.category = catalog.categories.find(c => c.categoryName.toLowerCase() === result.category.trim().toLowerCase())?.categoryName || ''
}
async function send(text = draft.value, explicitTarget = null) {
  if (busy.value || props.blocked || storageConflict.value || typeof text !== 'string' || !text.trim()) return
  if (text.trim().length > 4000) return ElMessage.warning('合并后的写作要求超过4000字，请精简后发送；输入已保留')
  if (scope.value === 'selection' && !selectionText.value && !explicitTarget && !replyTo.value) return ElMessage.warning('请先选中正文')
  const sourceTarget = explicitTarget || chosenTarget()
  const target = sourceTarget ? { ...sourceTarget, title:context.value?.getTitle?.() || '', description:context.value?.getDescription?.() || '' } : null
  if (replyTo.value?.target && !target) return ElMessage.warning('正文已变化，请重新选择范围后生成')
  const request = { action:'conversation', instruction:text.trim(), title:context.value?.getTitle?.() || '', content:target?.content || '', selection:target?.selection?.text || '', cursorOffset:target?.cursor, ...props.preferences }
  const taxonomy = context.value?.getTaxonomy?.() || { categories:[], tags:[] }
  const catalog = { categories:taxonomy.categories.slice(0, 50).map(c => c.categoryName), tags:taxonomy.tags.slice(0, 100).map(t => t.tagName) }
  while (JSON.stringify(catalog).length > 4000 && (catalog.tags.length || catalog.categories.length)) { if (catalog.tags.length) catalog.tags.pop(); else catalog.categories.pop() }
  request.catalog = JSON.stringify(catalog)
  const previousMessages = replyTo.value ? messages.value.slice(0, messages.value.indexOf(replyTo.value) + 1) : messages.value
  request.history = previousMessages.filter(m => m.role === 'user' || m.status === 'complete').slice(-6).map(m => ({ role:m.role, content:[...(m.role === 'assistant' && m.result ? historyResult(m) : m.content)].slice(0, 8000).join('') }))
  cancelFollowup()
  const session = sessions.value.find(s => s.id === activeId.value)
  const userMessage = { id:createChatId(), role:'user', content:text.trim(), createdAt:Date.now(), scope:target?.selection ? `选区 ${[...target.selection.text].length} 字` : '文章正文' }
  const reply = reactive({ id:createChatId(), role:'assistant', content:'', status:'generating', createdAt:Date.now(), target, scope:userMessage.scope })
  session.messages.push(userMessage, reply)
  if (session.title === '新对话') session.title = text.trim().slice(0, 20)
  draft.value = ''; persist(); busy.value = true; running = reply; progress.value = ''; scrollBottom(true)
  let raw = '', complete = false
  const controller = streamAiChat(request, {
    onDelta(delta) { if (running !== reply) return; raw += delta; reply.content = partialChatContent(raw); scrollBottom() },
    onContext(info) { if (running === reply) reply.notice = info.notice || '' },
    onTool(tool) { if (running === reply) progress.value = tool.name === 'search_my_blogs' ? '正在参考你的历史文章…' : '正在阅读相关内容…' },
    onDone() { complete = true }
  })
  handle = controller
  const result = await controller.promise
  if (running !== reply) return
  try {
    if (result.aborted) { reply.status = 'stopped'; reply.error = '已停止，未完成的回复仅供复制。' }
    else if (result.errorMessage || !complete) { reply.status = 'error'; reply.error = result.errorMessage || '回复未完整结束，请重试。' }
    else {
      const parsed = parseChatResult(raw)
      if (parsed.kind === 'edit' && !target?.selection) throw new Error('没有有效选区，请选中要修改的段落后重试')
      normalizeTags(parsed)
      reply.result = parsed; reply.content = parsed.content; reply.status = 'complete'; reply.selectedTags = [...(parsed.tags || [])]; reply.selectedNewTags = []; reply.includeCategory = true
    }
  } catch (error) { reply.status = 'error'; reply.error = error.message }
  finally { busy.value = false; handle = null; running = null; persist(); scrollBottom() }
}
function stop() { if (!handle) return; handle.abort(); if (running) { running.status = 'stopped'; running.error = '已停止，未完成的回复仅供复制。' }; handle = null; running = null; busy.value = false; persist() }
function onKeydown(event) { if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) { event.preventDefault(); send() } }
function retry(message) { const index = messages.value.indexOf(message); const text = messages.value[index-1]?.content; if (text) send(text, currentTarget(message) || chosenTarget()) }
function revise(message) { const target = restoreChatTarget(message.continuationTarget || message.target, context.value?.getEditorState?.()); if (!target && message.target) return ElMessage.warning('正文已变化，请重新选择范围后提出修改要求'); replyTo.value = message; pinnedTarget.value = null; if (!draft.value.trim()) draft.value = '请在这一版基础上'; scrollBottom(true) }
function issueTarget(message, issue) { return reviewTarget(currentTarget(message), issue.quote) }
async function locateIssue(message, issue) {
  const target = issueTarget(message, issue)
  if (!target) return ElMessage.warning('原文无法唯一定位，请重新审阅')
  store.closeDock(); await nextTick()
  const result = await context.value?.selectSnapshot?.(target.selection)
  if (!result?.ok) ElMessage.warning(result?.message || '当前编辑器无法定位原文')
}
function discussIssue(message, issue, edit = false) {
  const target = issueTarget(message, issue)
  if (!target) return ElMessage.warning('原文无法唯一定位，请重新审阅')
  replyTo.value = message; pinnedTarget.value = target
  const instruction = `讨论这段原文：${issue.reason}\n建议方向：${issue.suggestion}\n${edit ? '请按照建议修改这段，保留原意与作者文风。' : '请先说明怎么改。'}`
  draft.value = combineShortcut(instruction, draft.value); scrollBottom(true)
}
async function finish(message, result, success) { if (!result?.ok && result !== true) return ElMessage.warning(result?.message || '当前页面无法应用，请复制使用'); message.adopted = true
  if (['insert', 'outline'].includes(message.result.kind)) {
    await nextTick()
    const state = context.value?.getEditorState?.()
    const target = result.range ? appliedTarget(message.target, state, result.range.start, result.range.end) : insertedTarget(message.target, state)
    if (target) { message.continuationTarget = target; message.application = { content: target.selection?.text || '', status: 'adopted' } }
  }
  persist(); ElMessage.success(success) }
function compare(message) {
  const target = currentTarget(message)
  if (!target?.selection || unavailable(message)) return ElMessage.warning('原选区已变化，请重新生成')
  const result = store.openEditorDiff(target.selection, message.content)
  if (result.ok) { store.closeDock(); diffMessage = { message, sessionId: store.diff.sessionId } } else ElMessage.warning(result.message)
}
let diffMessage = null
watch(() => store.lastDiffResolution, resolution => {
  if (!diffMessage || resolution?.sessionId !== diffMessage.sessionId) return
  const { message } = diffMessage; diffMessage = null
  const state = context.value?.getEditorState?.(), snapshot = resolution.snapshot
  if (state?.content !== snapshot.content || state?.documentId !== snapshot.documentId) return
  const target = appliedTarget(message.target, state, snapshot.start, snapshot.end)
  if (!target) return
  message.continuationTarget = target; message.resolved = true
  message.adopted = state.content !== message.target.content
  message.application = { content: snapshot.text, status: message.adopted ? 'adopted_with_original_choices' : 'kept_original' }
  persist()
}, { flush: 'sync' })
function insert(message) { const target = currentTarget(message); if (!target || unavailable(message)) return; finish(message, context.value?.insertAiAt?.(message.content, target), '已插入原光标位置，可撤销') }
function append(message) { if (unavailable(message)) return; finish(message, context.value?.appendChapter?.(message.content, context.value.getEditorState()), '已追加到文末，可撤销') }
function adoptTitle(message, title) { if (unavailable(message)) return; finish(message, context.value?.fillTitle?.(title, message.target.title), '已采用标题') }
function adoptSummary(message) {
  if (unavailable(message)) return
  if ((context.value?.getDescription?.() || '') !== message.target.description) return ElMessage.warning('摘要已被编辑，请保留当前摘要或重新生成')
  finish(message, context.value?.fillDescription?.(message.content), '已采用摘要')
}
function adoptTags(message) { if (unavailable(message)) return; finish(message, context.value?.applySuggestion?.({ category:message.includeCategory ? message.result.category : '', tags:message.selectedTags, newTags:message.selectedNewTags }), '已采用分类与标签') }
async function copy(text) { try { await navigator.clipboard.writeText(text); ElMessage.success('已复制') } catch { ElMessage.warning('复制失败，请手动选择文本复制') } }
watch(draft, () => { clearTimeout(draftTimer); draftTimer = setTimeout(() => { if (!busy.value) persist() }, 400) })
watch([identity, () => context.value, () => context.value?.getEditorState?.()?.active], () => { clearTimeout(draftTimer); stop(); storageKey = identity.value; storageConflict.value = false; diffMessage = null; cancelFollowup(); scope.value = 'auto'; reload() }, { immediate:true, flush:'sync' })
function externalUpdate(event) { if (event.key === storageKey) { storageConflict.value = true; storageNotice.value = '另一标签页已更新记录，当前内容不会覆盖它。请载入本地记录。' } }
window.addEventListener('storage', externalUpdate)
function flush() { if (!disposed) { stop(); persist() } }
window.addEventListener('pagehide', flush)
watch(() => store.selectionAction, async request => {
  if (!request || busy.value || props.blocked) return
  store.selectionAction = null
  syncSelection(); scope.value = 'selection'
  await send(request.action === 'polish' ? '润色当前选区，尽量保留原句与作者文风。' : '改写当前选区，换一种表达，保留原意。', { ...request.snapshot, selection:request.snapshot, cursor:request.snapshot.start, title:context.value?.getTitle?.() || '', description:context.value?.getDescription?.() || '' })
}, { immediate:true })
onBeforeUnmount(() => { flush(); disposed = true; clearTimeout(draftTimer); document.removeEventListener('selectionchange', syncSelection); window.removeEventListener('storage', externalUpdate); window.removeEventListener('pagehide', flush); emit('busy', false) })
</script>

<style scoped>
.writing-chat{display:flex;flex-direction:column;flex:1;min-height:0;gap:12px;color:var(--el-text-color-primary)}
.chat-heading{display:flex;justify-content:space-between;align-items:flex-start;gap:12px}.chat-heading strong{display:block;font-size:18px;letter-spacing:.02em}.chat-heading span{display:block;margin-top:5px;font-size:12px;color:var(--el-text-color-secondary)}
.session-controls{display:flex;gap:10px;align-items:center}.session-controls .el-select{flex:1}.chat-transcript{flex:1;min-height:150px;overflow-y:auto;overscroll-behavior:contain;padding:4px 8px 12px 0;scrollbar-gutter:stable}.chat-empty{padding:40px 20px;text-align:center}.empty-mark{font-size:36px;color:var(--el-color-primary)}.chat-empty h3{font-size:17px;font-weight:500}.chat-empty p{font-size:13px;line-height:1.8;color:var(--el-text-color-secondary)}
.chat-message{margin-bottom:20px;overflow-wrap:anywhere}.chat-message.user{padding:12px 14px;margin-left:24px;background:var(--el-fill-color-light);border-radius:12px 12px 2px 12px}.chat-message.assistant{padding:4px 0 12px;border-bottom:1px solid var(--el-border-color-lighter)}.message-heading{display:flex;justify-content:space-between;gap:8px;font-size:12px;font-weight:600;color:var(--el-text-color-secondary)}.message-heading small{font-weight:400}.user-text{white-space:pre-wrap;margin:8px 0 0;font-size:14px;line-height:1.7}.message-content{font-size:14px;line-height:1.85}.message-content :deep(pre){overflow:auto;background:var(--el-fill-color-light);padding:12px;border-radius:6px}.message-content :deep(img){max-width:100%}.message-content :deep(table){border-collapse:collapse;display:block;overflow:auto}.message-content :deep(td),.message-content :deep(th){border:1px solid var(--el-border-color);padding:5px}
.message-actions{display:flex;flex-wrap:wrap;gap:8px;margin-top:12px}.message-actions .el-button{margin-left:0}.chat-titles>div{display:flex;align-items:center;gap:12px;justify-content:space-between;padding:10px 0;border-bottom:1px solid var(--el-border-color-lighter)}.chat-titles span{font-size:14px;line-height:1.6}.chat-tags{display:flex;flex-direction:column;gap:6px}.chat-tags small,.message-note,.chat-thinking{font-size:12px;line-height:1.6;color:var(--el-text-color-secondary)}.message-note{display:block;margin-top:8px}.message-error,.chat-notice{font-size:12px;line-height:1.6;color:var(--el-color-warning)}.adopted-note{display:block;margin-top:8px;color:var(--el-color-success)}
.chat-composer{border-top:1px solid var(--el-border-color);padding-top:12px}.scope-row{display:flex;align-items:center;gap:8px;margin-bottom:10px}.scope-row .el-select{width:190px}.scope-row span{font-size:12px;color:var(--el-text-color-secondary)}.chat-shortcuts{display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px}.chat-shortcuts button{font:inherit;font-size:12px;padding:5px 9px;border:1px solid var(--el-border-color);border-radius:6px;background:var(--el-bg-color);color:var(--el-text-color-regular);cursor:pointer}.chat-shortcuts button:hover{border-color:var(--el-color-primary);color:var(--el-color-primary)}.chat-shortcuts button:disabled{opacity:.45;cursor:not-allowed}.composer-actions{display:flex;justify-content:space-between;align-items:center;gap:10px;margin-top:10px}.composer-actions small{font-size:11px;line-height:1.5;color:var(--el-text-color-secondary)}
.followup-context{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:8px 10px;margin-bottom:10px;border-left:3px solid var(--el-color-primary);background:var(--el-fill-color-light);font-size:12px}.new-reply{align-self:center}.review-issue{margin-top:12px;padding:12px;border:1px solid var(--el-border-color-lighter);border-radius:8px}.review-issue blockquote{margin:0;padding-left:10px;border-left:2px solid var(--el-border-color);white-space:pre-wrap;font-size:13px;color:var(--el-text-color-secondary)}.review-issue p{font-size:14px;line-height:1.6}.review-issue>small{font-size:12px;line-height:1.6}
.session-option{display:flex;align-items:center;justify-content:space-between;gap:16px;width:100%;height:100%}.session-option>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.session-option .el-button{flex-shrink:0}
</style>
