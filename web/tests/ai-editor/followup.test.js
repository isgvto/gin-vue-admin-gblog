import test from 'node:test'
import assert from 'node:assert/strict'
import { combineShortcut, reviewTarget, appliedTarget, insertedTarget, historyResult, nearTranscriptBottom } from '../../src/components/ai/agents/writing-assistant/writingFollowup.js'
import { parseChatResult, restoreChatTarget } from '../../src/components/ai/agents/writing-assistant/chatSession.js'
const state = { editorId: 'editor', documentId: 1, revision: 2, active: true, content: '开头\r\n\r\n保留😀原文\r\n\r\n结尾' }
test('快捷操作合并未发送要求，空输入保持原快捷指令', () => {
  assert.equal(combineShortcut('润色', '  保留专业术语  '), '润色\n补充要求：保留专业术语')
  assert.equal(combineShortcut('润色', '  '), '润色')
})
test('审阅引用只绑定唯一原文，重复、缺失或空引用不能定位', () => {
  const target = reviewTarget(state, '保留😀原文')
  assert.equal(state.content.slice(target.selection.start, target.selection.end), target.selection.text)
  assert.equal(reviewTarget({ ...state, content: '重复重复' }, '重复'), null)
  assert.equal(reviewTarget(state, '不存在'), null)
  assert.equal(reviewTarget(state, ''), null)
})
test('采用后的追问引用实际保留的正文，刷新绑定；手工修改后拒绝复用', () => {
  const target = appliedTarget({ ...state, content: '旧文' }, state, 6, 12)
  assert.equal(target.selection.text, '保留😀原文')
  assert.equal(restoreChatTarget(target, { ...state, editorId: 'new' }).selection.editorId, 'new')
  assert.equal(restoreChatTarget(target, { ...state, content: state.content + '人工补充' }), null)
  assert.equal(appliedTarget(state, { ...state, documentId: 2 }, 0, 3), null)
  const result = JSON.parse(historyResult({ result: {kind:'edit', content:'被拒绝的新内容'}, application:{content:'保留原文',status:'kept_original'} }))
  assert.equal(result.content, '保留原文')
})
test('插入后的追问引用实际新增范围，替换正文时不猜测范围', () => {
  const before = { ...state, content:'前文\n\n后文' }
  const after = { ...state, content:'前文\n\n新😀内容\n\n后文' }
  assert.equal(insertedTarget(before, after).selection.text, '新😀内容\n\n')
  assert.equal(insertedTarget(before, { ...state, content:'不同正文' }), null)
})
test('结构化审阅拒绝空引用、错误字段和超过三个问题', () => {
  const result = { kind:'review', content:'先调整这处表达', issues:[{ quote:'原文', reason:'不清楚', suggestion:'写明主体' }] }
  assert.deepEqual(parseChatResult(JSON.stringify(result)), result)
  assert.throws(() => parseChatResult(JSON.stringify({...result, issues:[{...result.issues[0],quote:''}]})))
  assert.throws(() => parseChatResult(JSON.stringify({...result, issues:Array(4).fill(result.issues[0])})))
})
test('阅读历史时暂停跟随，接近底部时恢复跟随', () => {
  assert.equal(nearTranscriptBottom({scrollHeight:1000,clientHeight:300,scrollTop:100}), false)
  assert.equal(nearTranscriptBottom({scrollHeight:1000,clientHeight:300,scrollTop:680}), true)
})
