// Absolute page URLs remain stable when their navigation folder changes.
export function joinMenuPath(segments, currentPath) {
  return [...segments, currentPath].filter(Boolean).reduce((result, segment) => {
    const value = String(segment)
    return value.startsWith('/') ? value : [result, value].filter(Boolean).join('/')
  }, '').replace(/\/+/g, '/')
}

export function firstMenuPage(items, segments = []) {
  for (const item of items || []) {
    if (item.hidden || /^(https?:)?\/\//.test(item.path || '')) continue
    const fullPath = joinMenuPath(segments, item.path)
    if (item.children?.length) {
      const target = firstMenuPage(item.children, [fullPath])
      if (target) return target
    } else if (!fullPath.includes(':')) {
      return fullPath.startsWith('/') ? fullPath : `/layout/${fullPath}`
    }
  }
  return null
}
