import assert from 'node:assert/strict'
import { test } from 'node:test'
import { chapterMarkdown, chapterOwnerError, chapterPayload, parseChapterOutline } from '../../src/components/ai/agents/writing-assistant/chapterWorkflow.js'
import { appendDocumentSnapshot, undoSnapshot } from '../../src/components/blog/editorSnapshot.js'

test('大纲按同级章节拆分，保留子标题与要点，排除文章标题', () => {
  const parsed = parseChapterOutline('# 全文标题\r\n\r\n## 背景\r\n- 要点\r\n### 现状\r\n说明\r\n## 实践\r\n- 案例')
  assert.equal(parsed.error, '')
  assert.deepEqual(parsed.sections, [{ title: '背景', brief: '- 要点\n### 现状\n说明' }, { title: '实践', brief: '- 案例' }])
  assert.equal(parseChapterOutline('## 唯一章节\n### 子节').sections.length, 1)
})

test('同级列表可作大纲，代码围栏内的标题不拆成章节', () => {
  assert.deepEqual(parseChapterOutline('1. 背景\n  - 原因\n2. 实践').sections.map(section => section.title), ['背景', '实践'])
  assert.equal(parseChapterOutline('## 背景\n```md\n## 假标题\n```\n## 实践').sections.length, 2)
  assert.ok(parseChapterOutline('未提供结构的正文').error)
  assert.ok(parseChapterOutline(Array.from({ length: 21 }, (_, i) => `## 第${i}章`).join('\n')).error)
  assert.ok(parseChapterOutline('## 章节\n' + '字'.repeat(601)).error)
})

const state = { active: true, editorId: 'e1', documentId: 'a1', revision: 1, content: '已有正文\r\n人工编辑' }
const context = { getEditorState: () => state, getFullText: () => state.content, getTitle: () => '文章' }
test('章节请求携带已确认大纲、最新正文与修改稿，不混入选区或快捷历史', () => {
  const chapters = [{ title: '背景', brief: '要点', draft: '手工修改稿', instruction: ' 补充案例 ' }]
  const payload = chapterPayload(chapters, 0, context, 'formal', 'shorter')
  assert.equal(payload.chapterDraft, '手工修改稿')
  assert.equal(payload.instruction, '补充案例')
  assert.equal(payload.content, state.content)
  assert.deepEqual(payload.outline, [{ title: '背景', brief: '要点' }])
  assert.equal(payload.selection, '')
  assert.deepEqual(payload.history, [])
  chapters[0].title = '后来修改'
  assert.equal(payload.outline[0].title, '背景')
})

test('章节采纳保留人工正文与CRLF并支持撤销，也可采纳到空文章', () => {
  const markdown = chapterMarkdown({ title: '背景', draft: '## 背景\n\n内容😀' })
  assert.equal(markdown, '## 背景\n\n内容😀')
  const result = appendDocumentSnapshot(state, state, markdown)
  assert.equal(result.content, state.content + '\r\n\r\n## 背景\r\n\r\n内容😀\r\n')
  assert.equal(undoSnapshot({ ...state, content: result.content }, result.undo).content, state.content)
  const empty = { ...state, content: '' }
  assert.equal(appendDocumentSnapshot(empty, empty, markdown).content, markdown + '\n')
  assert.equal(chapterMarkdown({ title: '背景', draft: '## 背景' }), '')
})

test('跨文章、失活、正文版本变化不能写回；相同文章继续手工编辑不丢失流程身份', () => {
  const owner = { context, editorId: state.editorId, documentId: state.documentId }
  assert.equal(chapterOwnerError(owner, context), '')
  assert.ok(chapterOwnerError(owner, { ...context }))
  for (const change of [{ documentId: 'a2' }, { editorId: 'e2' }, { active: false }, { content: '新的编辑' }, { revision: 2 }]) {
    assert.equal(appendDocumentSnapshot({ ...state, ...change }, state, '章节').ok, false)
  }
  assert.equal(appendDocumentSnapshot(state, state, ' ').ok, false)
})
