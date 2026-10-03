export function placeSelectionToolbar(point, viewport, size = { width: 184, height: 38 }) {
  const margin = 8
  const left = Math.max(margin, Math.min(point.x - size.width / 2, viewport.width - size.width - margin))
  const above = point.y - size.height - margin
  const top = above >= margin ? above : Math.min(point.y + point.height + margin, viewport.height - size.height - margin)
  return { left, top: Math.max(margin, top) }
}

// textarea 的选区没有 DOM Range；用同样的排版测量选区活动端的位置。
export function textareaSelectionPoint(textarea, allowOffscreen = false) {
  const style = window.getComputedStyle(textarea)
  const mirror = document.createElement('div')
  for (const property of ['fontFamily', 'fontSize', 'fontWeight', 'fontStyle', 'lineHeight', 'letterSpacing', 'wordSpacing', 'textTransform', 'textIndent', 'tabSize', 'padding', 'wordBreak', 'overflowWrap']) {
    mirror.style[property] = style[property]
  }
  Object.assign(mirror.style, {
    position: 'fixed', top: '0', left: '0', visibility: 'hidden', pointerEvents: 'none',
    whiteSpace: 'pre-wrap', overflowWrap: 'break-word', boxSizing: 'border-box',
    width: `${textarea.clientWidth}px`, border: '0'
  })
  const index = textarea.selectionDirection === 'backward' ? textarea.selectionStart : textarea.selectionEnd
  mirror.textContent = textarea.value.slice(0, index)
  const marker = document.createElement('span')
  marker.textContent = '\u200b'
  mirror.append(marker)
  document.body.append(mirror)
  try {
    const rect = textarea.getBoundingClientRect()
    const markerRect = marker.getBoundingClientRect()
    const height = parseFloat(style.lineHeight) || parseFloat(style.fontSize) * 1.5
    const x = rect.left + textarea.clientLeft + markerRect.left - textarea.scrollLeft
    const y = rect.top + textarea.clientTop + markerRect.top - textarea.scrollTop
    if (!allowOffscreen && (x < rect.left || x > rect.right || y + height < rect.top || y > rect.bottom)) return null
    return { x, y, height }
  } finally {
    mirror.remove()
  }
}
