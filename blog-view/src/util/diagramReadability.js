// Evaluate at an article width, independently of the narrow assistant sidebar.
export function assessDiagram({ width, height, fontSize = 16 }, availableWidth = 720) {
  if (![width, height, fontSize].every(value => Number.isFinite(value) && value > 0)) {
    return { score: -Infinity, needsReview: true, fontSize: 0, height: 0 }
  }
  const scale = Math.min(1, availableWidth / width, 420 / height)
  const displayedFont = fontSize * scale
  const ratio = Math.max(width / height, height / width)
  // Prefer legible text, then avoid very elongated diagrams. Never change edges.
  const score = Math.min(displayedFont, 16) - Math.max(0, ratio - 3) * 0.6
  return { score, needsReview: displayedFont < 12 || (ratio > 4 && (width > availableWidth || height > 420)),
    fontSize: displayedFont, height: height * Math.min(1, availableWidth / width) }
}

export function diagramLayoutCandidates(source) {
  const header = /^(\s*(?:flowchart|graph)\s+)(TD|TB|BT|LR|RL)\b/
  if (!header.test(source)) return [source]
  return [...new Set([source, ...['TD', 'LR'].map(direction => source.replace(header, `$1${direction}`))])]
}

export function preferDiagramLayout(original, candidate) {
  // Keep the author's layout unless the improvement is meaningful.
  return (original.needsReview && !candidate.needsReview && candidate.score >= original.score)
    || candidate.score > original.score + 0.75
}
