import { Lexer } from 'marked'
import { captureEditorSnapshot } from './editorSnapshot.js'

// Top-level Markdown blocks are atomic: never split a list, table or code fence.
export function indexParagraphs(content) {
  const source = String(content || '')
  const normalized = source.replace(/\r\n?/g, '\n')
  const offsets = [0]
  for (let i = 0; i < source.length;) {
    i += source[i] === '\r' && source[i + 1] === '\n' ? 2 : 1
    offsets.push(i)
  }
  const blocks = []
  let cursor = 0
  const append = (from, to, kind) => {
    // Exclude blank separators, but retain indentation and Markdown hard breaks.
    const raw = normalized.slice(from, to)
    if (!raw.trim()) return
    const leading = raw.match(/^(?:[ \t]*\n)+/)?.[0].length || 0
    const trailing = raw.match(/(?:\n[ \t]*)+$/)?.[0].length || 0
    const start = from + leading, end = to - trailing
    if (end <= start) return
    blocks.push({ number: blocks.length + 1, kind, start: offsets[start], end: offsets[end], textareaStart: start })
  }
  for (const token of Lexer.lex(normalized, { gfm: true })) {
    const start = normalized.indexOf(token.raw, cursor)
    if (start < cursor) throw new Error('无法定位 Markdown 段落')
    // Link definitions may be omitted by Lexer; retain them as source blocks.
    append(cursor, start, 'definition')
    if (token.type !== 'space') append(start, start + token.raw.length, token.type)
    cursor = start + token.raw.length
  }
  append(cursor, normalized.length, 'definition')
  return blocks
}

function paragraphNumber(text) {
  if (/^[0-9０-９]+$/.test(text)) return Number(text.replace(/[０-９]/g, digit => String(digit.charCodeAt(0) - 0xff10)))
  const digits = { 零: 0, 〇: 0, 一: 1, 二: 2, 两: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9 }
  let result = 0, digit = 0
  for (const character of text) {
    if (character in digits) digit = digits[character]
    else if (character === '十' || character === '百' || character === '千') {
      result += (digit || 1) * ({ 十: 10, 百: 100, 千: 1000 }[character]); digit = 0
    } else return NaN
  }
  return result + digit
}

export function resolveParagraphReference(instruction, state) {
  const number = '[0-9０-９零〇一二两三四五六七八九十百千]+'
  const expression = new RegExp(`第\\s*(${number})\\s*(?:段\\s*)?(?:[至到—–~～-]\\s*(?:第\\s*)?(${number})\\s*)?段(?:落)?`, 'g')
  // Quoted examples and fenced code are material, not a request to locate text.
  const request = String(instruction).replace(/```[\s\S]*?```|`[^`]*`|“[^”]*”|「[^」]*」|"[^"\n]*"/g, '')
  const shorthand = new RegExp(`第\\s*${number}\\s*(?:[、，,和及与]\\s*(?:第\\s*)?${number}\\s*)+段`)
  if (shorthand.test(request)) return { matched: true, error: '本次请指定一个段落或连续范围，例如“修改第2至4段”；多个不连续段落请分别处理' }
  const references = [...request.matchAll(expression)]
  if (!references.length) return { matched: false }
  if (!state?.active) return { matched: true, error: '请在文章编辑页使用段落编号，或直接粘贴要处理的文字' }
  const ranges = references.map(match => ({ from: paragraphNumber(match[1]), to: paragraphNumber(match[2] || match[1]) }))
  const range = ranges[0]
  if (ranges.some(item => item.from !== range.from || item.to !== range.to)) {
    return { matched: true, error: '本次请指定一个段落或连续范围，例如“修改第2至4段”；多个不连续段落请分别处理' }
  }
  const paragraphs = indexParagraphs(state.content)
  if (!Number.isSafeInteger(range.from) || range.from < 1 || range.to < range.from || range.to > paragraphs.length) {
    return { matched: true, error: `段落编号无效，当前正文共${paragraphs.length}段；请查看左侧编号后重试` }
  }
  const selection = captureEditorSnapshot(state, { start: paragraphs[range.from - 1].start, end: paragraphs[range.to - 1].end })
  const label = range.from === range.to ? `第${range.from}段` : `第${range.from}至${range.to}段`
  return { matched: true, target: { ...state, selection, cursor: selection.start, paragraphLabel: label }, label }
}
