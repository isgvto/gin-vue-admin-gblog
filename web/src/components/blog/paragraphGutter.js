// Measure soft wrapping using the textarea's exact typography and inner width.
export function measureParagraphGutter(textarea, paragraphs) {
  if (!textarea || !paragraphs.length || !textarea.clientWidth) return []
  const style = getComputedStyle(textarea)
  const mirror = document.createElement('div')
  for (const property of ['fontFamily', 'fontSize', 'fontWeight', 'fontStyle', 'lineHeight', 'letterSpacing', 'wordSpacing', 'textIndent', 'textTransform', 'tabSize', 'padding', 'wordBreak', 'overflowWrap']) mirror.style[property] = style[property]
  Object.assign(mirror.style, { position: 'fixed', top: '0', left: '0', visibility: 'hidden',
    pointerEvents: 'none', whiteSpace: 'pre-wrap', overflowWrap: 'break-word',
    boxSizing: 'border-box', width: `${textarea.clientWidth}px`, border: '0' })
  const markers = []
  let offset = 0
  for (const paragraph of paragraphs) {
    mirror.append(document.createTextNode(textarea.value.slice(offset, paragraph.textareaStart)))
    const marker = document.createElement('span')
    marker.textContent = '\u200b'
    mirror.append(marker); markers.push(marker); offset = paragraph.textareaStart
  }
  mirror.append(document.createTextNode(textarea.value.slice(offset)))
  document.body.append(mirror)
  try {
    const top = mirror.getBoundingClientRect().top
    const lineHeight = parseFloat(style.lineHeight) || parseFloat(style.fontSize) * 1.75
    return paragraphs.map((paragraph, index) => {
      const rect = markers[index].getBoundingClientRect()
      return { ...paragraph, top: rect.top - top - (lineHeight - rect.height) / 2 }
    })
  } finally { mirror.remove() }
}
