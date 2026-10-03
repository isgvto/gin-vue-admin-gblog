const count = text => [...text].length

export function outlineError(sections) {
  if (!sections.length) return '请使用 Markdown 章节标题（如 ## 第一章）或同级列表编写大纲'
  if (sections.length > 20) return '大纲最多支持20章，请合并部分章节'
  if (sections.some(section => !section.title.trim() || count(section.title) > 120 || /[\r\n<>]/.test(section.title))) return '章节标题须为不超过120字的单行文本'
  if (sections.some(section => count(section.brief) > 600)) return '每章要点不能超过600字，请精简大纲'
  if (sections.reduce((sum, section) => sum + count(section.title) + count(section.brief), 0) > 8000) return '大纲合计不能超过8000字'
  return ''
}

// 最浅层标题作为章节；单个一级文章标题不作为章节，子标题保留为本章要点。
export function parseChapterOutline(text) {
  const lines = String(text || '').replace(/\r\n?/g, '\n').split('\n')
  let fence = null
  const headings = []
  const bullets = []
  lines.forEach((line, index) => {
    const marker = line.match(/^\s*(`{3,}|~{3,})/)
    if (marker) {
      if (!fence) fence = marker[1]
      else if (marker[1][0] === fence[0] && marker[1].length >= fence.length) fence = null
      return
    }
    if (fence) return
    const heading = line.match(/^ {0,3}(#{1,6})\s+(.+?)\s*#*\s*$/)
    if (heading) headings.push({ index, level: heading[1].length, title: heading[2] })
    const bullet = line.match(/^(\s*)(?:[-*+]\s+|\d+[.)、]\s*)(.+)$/)
    if (bullet) bullets.push({ index, level: bullet[1].length, title: bullet[2] })
  })
  let candidates = headings
  if (headings.filter(item => item.level === 1).length === 1 && headings.some(item => item.level > 1)) {
    candidates = headings.filter(item => item.level !== 1)
  }
  if (!candidates.length) candidates = bullets
  const level = Math.min(...candidates.map(item => item.level))
  const chapters = candidates.filter(item => item.level === level)
  const sections = chapters.map((item, index) => ({
    title: item.title.replace(/^\*\*(.+)\*\*$/, '$1').trim(),
    brief: lines.slice(item.index + 1, chapters[index + 1]?.index ?? lines.length).join('\n').trim()
  }))
  return { sections, error: outlineError(sections) }
}

export function chapterOwnerError(owner, context) {
  const state = context?.getEditorState?.()
  return !owner || !state?.active || owner.context !== context || owner.editorId !== state.editorId || owner.documentId !== state.documentId
    ? '文章已切换，请重新确认大纲' : ''
}

export function chapterPayload(sections, index, context, tone, length) {
  const section = sections[index]
  return {
    action: 'chapter', title: context?.getTitle?.() || '', content: context?.getFullText?.() || '',
    outline: sections.map(({ title, brief }) => ({ title, brief })), chapterIndex: index,
    chapterDraft: section.draft, instruction: section.instruction.trim(), tone, length,
    selection: '', history: []
  }
}

export function chapterMarkdown(section) {
  const body = section.draft.trim()
  // 容忍模型重复输出同名章节标题，避免写入两个相同标题。
  const firstHeading = body.match(/^#{1,2}\s+(.+?)\s*#*\s*(?:\n|$)/)
  const content = firstHeading?.[1].trim() === section.title ? body.slice(firstHeading[0].length).trim() : body
  return content ? `## ${section.title}\n\n${content}` : ''
}
