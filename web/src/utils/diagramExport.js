// Render the vector at export resolution, independently of sidebar/CSS width.
export function diagramExportSize(width, height) {
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) throw new Error('图示绘图尺寸无效')
  const scale = Math.min(Math.max(3, 3200 / Math.max(width, height)), 4096 / width, 4096 / height, Math.sqrt(16 * 1024 * 1024 / (width * height)))
  return { width: Math.max(1, Math.floor(width * scale)), height: Math.max(1, Math.floor(height * scale)) }
}

export async function exportDiagramPNG(image) {
  if (!image?.complete || !image.naturalWidth || !image.src.startsWith('blob:')) throw new Error('请等待图示渲染完成')
  const response = await fetch(image.src)
  const svg = new DOMParser().parseFromString(await response.text(), 'image/svg+xml').documentElement
  const box = svg.getAttribute('viewBox')?.trim().split(/[\s,]+/).map(Number)
  if (svg.localName !== 'svg' || box?.length !== 4) throw new Error('图示缺少有效的矢量尺寸')
  const size = diagramExportSize(box[2], box[3])
  svg.setAttribute('width', String(size.width)); svg.setAttribute('height', String(size.height)); svg.style.maxWidth = 'none'
  const url = URL.createObjectURL(new Blob([new XMLSerializer().serializeToString(svg)], { type: 'image/svg+xml' }))
  try {
    await document.fonts?.ready
    const vector = new Image()
    await new Promise((resolve, reject) => { vector.onload = resolve; vector.onerror = () => reject(new Error('图示高清渲染失败')); vector.src = url })
    const canvas = document.createElement('canvas'); canvas.width = size.width; canvas.height = size.height
    const ctx = canvas.getContext('2d'); if (!ctx) throw new Error('浏览器不支持图片导出')
    ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, canvas.width, canvas.height); ctx.drawImage(vector, 0, 0, canvas.width, canvas.height)
    return canvas.toDataURL('image/png')
  } finally { URL.revokeObjectURL(url) }
}
