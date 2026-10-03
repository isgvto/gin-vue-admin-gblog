import assert from 'node:assert/strict'
import { test } from 'node:test'
import { captureEditorSnapshot, applySnapshot, undoSnapshot, textareaOffsetToSource, sourceOffsetToTextarea } from '../../src/components/blog/editorSnapshot.js'
import { createPinia, setActivePinia } from 'pinia'
import { useAiStore } from '../../src/pinia/modules/ai.js'

const original = '前文\r\n\r\n待修改😀\r\n\r\n后文\r\n'
const state = { active: true, editorId: 'editor-a', documentId: 'article-a', revision: 3, content: original }
const start = original.indexOf('待修改')
const end = original.indexOf('\r\n', start)

for (const [from, to] of [[0, 2], [start, end], [original.indexOf('后文'), original.length], [0, original.length]]) {
  test(`固定选区替换及撤销 ${from}:${to}`, () => {
    const snapshot = captureEditorSnapshot(state, { start: from, end: to })
    const result = applySnapshot(state, snapshot, 'AI')
    assert.equal(result.ok, true)
    assert.equal(result.content, original.slice(0, from) + 'AI' + original.slice(to))
    const restored = undoSnapshot({ ...state, content: result.content }, result.undo)
    assert.equal(restored.content, original)
  })
}

test('无选区或无编辑器时不可将状态误解为整篇替换', () => {
  assert.equal(captureEditorSnapshot(state, { start: 0, end: 0 }), null)
  assert.equal(captureEditorSnapshot({ ...state, active: false }, { start, end }), null)
  assert.equal(applySnapshot(state, null, 'AI').ok, false)
})

test('正文、文章、实例或版本改变后拒绝应用，离开文章后也拒绝', () => {
  const snapshot = captureEditorSnapshot(state, { start, end })
  for (const change of [{ content: original + '人工编辑' }, { revision: 4 }, { documentId: 'article-b' }, { editorId: 'editor-b' }, { active: false }]) {
    assert.equal(applySnapshot({ ...state, ...change }, snapshot, 'AI').ok, false)
  }
})

test('撤销不覆盖后来输入的内容，不跨文章恢复', () => {
  const snapshot = captureEditorSnapshot(state, { start, end })
  const result = applySnapshot(state, snapshot, 'AI')
  assert.equal(undoSnapshot({ ...state, content: result.content + '新输入' }, result.undo).ok, false)
  assert.equal(undoSnapshot({ ...state, content: result.content, documentId: 'article-b' }, result.undo).ok, false)
})

test('CRLF、中文和 emoji 的 DOM 偏移正确映射回原文', () => {
  const displayed = original.replace(/\r\n?/g, '\n')
  const domStart = displayed.indexOf('待修改')
  const domEnd = displayed.indexOf('\n', domStart)
  assert.equal(textareaOffsetToSource(original, domStart), start)
  assert.equal(textareaOffsetToSource(original, domEnd), end)
  assert.equal(sourceOffsetToTextarea(original, end), domEnd)
})

test('未处理修改保留原文，单处即时采用失败不改变对比或正文', async () => {
  setActivePinia(createPinia())
  const store = useAiStore()
  let current = { ...state }
  const handle = { getEditorState: () => current, applySelectionSnapshot: (saved, text) => {
    const result = applySnapshot(current, saved, text)
    if (result.ok) current = { ...current, content: result.content, revision: current.revision + 1 }
    return result
  } }
  store.registerContext('editor', handle)
  const snapshot = captureEditorSnapshot(state, { start, end })
  assert.equal(store.openEditorDiff(snapshot, 'AI').ok, true)
  assert.equal(store.diffPreviewText, original)
  assert.equal((await store.setDiffChoice({ index: 0, takeRevised: true }, { getEditorState: () => current })).ok, false)
  assert.equal(store.diff.active, true)
  assert.equal(store.diff.blocks[0].reviewed, false)
  assert.equal((await store.setDiffChoice({ index: 0, takeRevised: true })).ok, true)
  assert.equal(current.content, original.slice(0, start) + 'AI' + original.slice(end))
  assert.equal(store.diff.active, false)
})

test('多处即时采用和保留按更新后的范围写回，结束对比保留已采用内容，支持逐步撤销', async () => {
  setActivePinia(createPinia())
  const store = useAiStore(), before = '前文\r\n\r\n甲\r\n\r\n乙😀\r\n\r\n丙\r\n\r\n后文'
  let current = { ...state, content: before }
  const undo = []
  const handle = { getEditorState: () => current, applySelectionSnapshot: (saved, text) => {
    const result = applySnapshot(current, saved, text)
    if (result.ok) { undo.push(result.undo); current = { ...current, content: result.content, revision: current.revision + 1 } }
    return result
  } }
  store.registerContext('editor', handle)
  const from = before.indexOf('甲'), to = before.indexOf('\r\n\r\n后文')
  store.openEditorDiff(captureEditorSnapshot(current, { start: from, end: to }), '很长的甲\r\n\r\n新版乙😀\r\n\r\n新版丙')
  assert.equal(store.diff.blocks.length, 3)
  assert.equal((await store.setDiffChoice({ index: 0, takeRevised: true })).ok, true)
  assert.equal(current.content, before.replace('甲', '很长的甲'))
  assert.equal(store.diffPreviewText, current.content)
  assert.equal((await store.setDiffChoice({ index: 1, takeRevised: false })).ok, true)
  assert.equal(undo.length, 1)
  assert.equal((await store.setDiffChoice({ index: 2, takeRevised: true })).ok, true)
  assert.equal(current.content, before.replace('甲', '很长的甲').replace('丙', '新版丙'))
  assert.equal(store.diff.active, false)
  for (const entry of undo.reverse()) { current = { ...current, content: undoSnapshot(current, entry).content } }
  assert.equal(current.content, before)
  store.openEditorDiff(captureEditorSnapshot(current, { start: from, end: to }), '新甲\r\n\r\n新乙\r\n\r\n新丙')
  await store.setDiffChoice({ index: 0, takeRevised: true })
  store.closeEditorDiff()
  assert.equal(current.content, before.replace('甲', '新甲'))
})

test('全部保留不写正文，冲突拒绝单处采纳，完全删除选区后自动退出', async () => {
  setActivePinia(createPinia())
  const store = useAiStore()
  let current = { ...state }, writes = 0
  store.registerContext('editor', { getEditorState: () => current, applySelectionSnapshot: (saved, text) => {
    const result = applySnapshot(current, saved, text)
    if (result.ok) { writes++; current = { ...current, content: result.content, revision: current.revision + 1 } }
    return result
  } })
  const snapshot = captureEditorSnapshot(current, { start, end })
  store.openEditorDiff(snapshot, 'AI')
  await store.setDiffChoice({ takeRevised: false })
  assert.equal(current.content, original)
  assert.equal(writes, 0)
  assert.equal(store.diff.active, false)
  store.openEditorDiff(snapshot, 'AI')
  current = { ...current, revision: current.revision + 1 }
  assert.equal((await store.setDiffChoice({ takeRevised: true })).ok, false)
  assert.equal(writes, 0)
  store.closeEditorDiff()
  store.openEditorDiff(captureEditorSnapshot(current, { start, end }), '')
  assert.equal((await store.setDiffChoice({ takeRevised: true })).ok, true)
  assert.equal(current.content, original.slice(0, start) + original.slice(end))
  assert.equal(store.diff.active, false)
})

test('旧页面注销不会清掉新编辑器，迟到的结果不能绑定新文章', () => {
  setActivePinia(createPinia())
  const store = useAiStore()
  const oldHandle = { getEditorState: () => state }
  const newHandle = { getEditorState: () => ({ ...state, editorId: 'editor-b' }) }
  store.registerContext('editor', oldHandle)
  store.registerContext('editor', newHandle)
  store.unregisterContext('editor', oldHandle)
  assert.equal(store.contexts.editor, newHandle)
  assert.equal(store.openEditorDiff(captureEditorSnapshot(state, { start, end }), 'AI').ok, false)
})
