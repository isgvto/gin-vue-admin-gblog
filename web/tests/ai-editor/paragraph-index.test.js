import assert from 'node:assert/strict'
import { test } from 'node:test'
import { indexParagraphs, resolveParagraphReference } from '../../src/components/blog/paragraphIndex.js'
import { applySnapshot } from '../../src/components/blog/editorSnapshot.js'

const state = content => ({ active: true, editorId: 'a', documentId: 'blog', revision: 1, content })
test('number Markdown blocks; fences, lists, tables and quotations remain atomic', () => {
  const content = '\n# 标题\n\n第一段\n连续行\n\n- 项目一\n- 项目二\n\n> 引用\n> 两行\n\n```js\nconst n=1\n\nconsole.log(n)\n```\n\n|A|B|\n|---|---|\n|1|2|\n\n结尾'
  const blocks = indexParagraphs(content)
  assert.deepEqual(blocks.map(p => p.kind), ['heading', 'paragraph', 'list', 'blockquote', 'code', 'table', 'paragraph'])
  assert.deepEqual(blocks.map(p => p.number), [1, 2, 3, 4, 5, 6, 7])
  assert.equal(content.slice(blocks[4].start, blocks[4].end), '```js\nconst n=1\n\nconsole.log(n)\n```')
})
test('preserve CRLF, repeated paragraphs, Unicode offsets and blank separators on replacement', () => {
  const content = '同样😀\r\n\r\n同样😀\r\n第二行  \r\n\r\n结尾\r\n'
  const resolved = resolveParagraphReference('润色第二段', state(content))
  assert.equal(resolved.target.selection.text, '同样😀\r\n第二行  ')
  assert.equal(resolved.target.selection.start, content.indexOf('同样😀', 2))
  assert.equal(applySnapshot(state(content), resolved.target.selection, '修改').content, '同样😀\r\n\r\n修改\r\n\r\n结尾\r\n')
})
test('support Chinese, full-width digits and contiguous paragraph ranges', () => {
  const content = Array.from({ length: 15 }, (_, i) => `段${i + 1}`).join('\n\n')
  for (const command of ['修改第2至4段', '修改第2段到第4段', '修改第二—四段', '修改第２～４段']) {
    const result = resolveParagraphReference(command, state(content))
    assert.equal(result.label, '第2至4段')
    assert.equal(result.target.selection.text, '段2\n\n段3\n\n段4')
  }
  assert.equal(resolveParagraphReference('解释第十二段', state(content)).target.selection.text, '段12')
})
test('invalid and ambiguous references never silently target other text', () => {
  const article = state('第一\n\n第二')
  for (const command of ['第0段', '第3段', '第2至1段', '修改第1段和第2段', '修改第1、2段', '修改第1和第2段', '修改第一、二、三段']) assert.ok(resolveParagraphReference(command, article).error)
  assert.ok(resolveParagraphReference('修改第1段', null).error)
  assert.equal(resolveParagraphReference('解释“第3段”是什么意思', article).matched, false)
  assert.equal(resolveParagraphReference('解释 `第3段` 这个词', article).matched, false)
  assert.equal(resolveParagraphReference('帮我润色', article).matched, false)
  assert.equal(resolveParagraphReference('第2段改短一点，第2段的例子保留', article).label, '第2段')
})
test('link definitions and indented code keep correct source positions', () => {
  const content = '[link]: https://example.com\n\n[链接][link]\n\n    a  \n    b\n\n结尾'
  const blocks = indexParagraphs(content)
  assert.equal(content.slice(blocks[1].start, blocks[1].end), '[链接][link]')
  assert.equal(content.slice(blocks[2].start, blocks[2].end), '    a  \n    b')
  assert.equal(indexParagraphs('   \r\n\r\n').length, 0)
})
