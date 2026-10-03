import test from 'node:test'
import assert from 'node:assert/strict'
import { chatStorageKey, createChatId, createSession, loadChat, saveChat, parseChatResult, partialChatContent, restoreChatTarget } from '../../src/components/ai/agents/writing-assistant/chatSession.js'
import { insertDocumentSnapshot, undoSnapshot } from '../../src/components/blog/editorSnapshot.js'
function storage() { const items = new Map(); return {get length(){return items.size},key:i=>[...items.keys()][i],getItem:k=>items.get(k)??null,setItem:(k,v)=>items.set(k,v),removeItem:k=>items.delete(k)} }
test('内网 HTTP 缺少 randomUUID 时仍能创建会话和消息 ID', () => { assert.equal(typeof createChatId(null), 'string'); assert.notEqual(createChatId(null), createChatId(null)) })
test('用户和文章隔离，消息、会话和未发送输入可恢复', () => {
  const local = storage(), key = chatStorageKey('user-a','article-a'), session = createSession()
  session.messages.push({role:'user',content:'再简洁一点'},{role:'assistant',content:'完整回复',status:'generating'})
  const result = saveChat(local,key,{sessions:[session],activeId:session.id,draft:'未发送'},0)
  assert.equal(result.ok,true)
  assert.equal(loadChat(local,key).sessions[0].messages[1].status,'stopped')
  assert.equal(loadChat(local,key).draft,'未发送')
  assert.equal(loadChat(local,chatStorageKey('user-b','article-a')).sessions[0].messages.length,0)
  assert.equal(loadChat(local,chatStorageKey('user-a','article-b')).sessions[0].messages.length,0)
  assert.equal(chatStorageKey('', 'article-a'),null)
})
test('另一个标签页的更新或清空不被旧状态覆盖', () => {
  const local=storage(),key=chatStorageKey('a','a'),session=createSession(),data={sessions:[session],activeId:session.id}
  const first=saveChat(local,key,data,0)
  const next=saveChat(local,key,data,first.data.version)
  assert.equal(saveChat(local,key,data,first.data.version).reason,'conflict')
  local.removeItem(key)
  assert.equal(saveChat(local,key,data,next.data.version).reason,'conflict')
})
test('超过限制淘汰旧消息，最新消息和快照完整保留', () => {
  const local=storage(),key=chatStorageKey('a','a'),session=createSession()
  session.messages=Array.from({length:60},(_,i)=>({id:i,role:i%2?'assistant':'user',content:String(i)}))
  const result=saveChat(local,key,{sessions:[session],activeId:session.id},0)
  assert.equal(result.data.sessions[0].messages.length,40)
  assert.equal(result.data.sessions[0].messages.at(-1).content,'59')
})
test('结构化输出区分可采纳结果，残缺JSON或无效标题标签拒绝采纳', () => {
  assert.equal(parseChatResult('{"content":"建议","kind":"advice"}').kind,'advice')
  assert.equal(partialChatContent('{"content":"第一行\\n第二行'),'第一行\n第二行')
  for (const raw of ['{"content":"残缺','{"content":"文本","kind":"overwrite"}','{"content":"标题","kind":"title","titles":["a\\nb"]}','{"content":"标签","kind":"tags","category":"","tags":["<script>"],"newTags":[]}']) assert.throws(()=>parseChatResult(raw))
})
test('刷新后只对正文完全相同的原文章重新绑定目标，固定位置插入并可撤销', () => {
  const target={documentId:'a',content:'前文😀后文',cursor:4,selection:null}
  const state={documentId:'a',editorId:'new',revision:2,content:target.content,active:true}
  const restored=restoreChatTarget(target,state)
  assert.equal(restored.editorId,'new')
  assert.equal(restoreChatTarget(target,{...state,content:'已修改'}),null)
  assert.equal(restoreChatTarget(target,{...state,documentId:'b'}),null)
  const applied=insertDocumentSnapshot(state,restored,'新内容')
  assert.equal(applied.content,'前文😀新内容\n后文')
  assert.equal(undoSnapshot({...state,content:applied.content},applied.undo).content,state.content)
  assert.equal(insertDocumentSnapshot({...state,revision:3},restored,'新内容').ok,false)
})
