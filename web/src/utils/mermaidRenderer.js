import { assessDiagram, diagramLayoutCandidates, preferDiagramLayout } from './diagramReadability.js'

// Both frontends render only generated SVG as an image. Raw Markdown HTML never
// gains permission to inject inline SVG, scripts or Mermaid configuration.
export function createMermaidRenderer(loadMermaid) {
  let engine
  let sequence = 0
  const roots = new WeakMap()
  const cache = new Map()
  const getEngine = () => {
    if (!engine) engine = loadMermaid().then(mermaid => {
      mermaid.initialize({
        startOnLoad: false, securityLevel: 'strict', suppressErrorRendering: true,
        theme: 'default', htmlLabels: false, flowchart: { htmlLabels: false, nodeSpacing: 24, rankSpacing: 28, padding: 8, diagramPadding: 8, wrappingWidth: 180 },
        fontFamily: 'sans-serif', maxTextSize: 20000, maxEdges: 200,
        secure: ['secure', 'securityLevel', 'startOnLoad', 'maxTextSize', 'maxEdges',
          'suppressErrorRendering', 'htmlLabels', 'flowchart', 'themeCSS']
      })
      return mermaid
    }).catch(error => { engine = null; throw error })
    return engine
  }

  function measure(svg) {
    const element = new DOMParser().parseFromString(svg, 'image/svg+xml').documentElement
    const box = element.getAttribute('viewBox')?.trim().split(/[\s,]+/).map(Number)
    const sizes = [...element.querySelectorAll('text, tspan')].map(text => {
      const inline = text.getAttribute('font-size') || text.style.fontSize
      return parseFloat(inline) || 16
    })
    return assessDiagram({ width: box?.[2], height: box?.[3], fontSize: sizes.length ? Math.min(...sizes) : 16 })
  }

  async function layout(mermaid, id, source, current) {
    const svg = (await mermaid.render(id, source)).svg
    if (svg.length > 2 * 1024 * 1024) throw new Error('limit')
    let best = { svg, assessment: measure(svg), adjusted: false }
    // Only re-render when the original is difficult to read. At most 3 layouts.
    if (best.assessment.needsReview && /^(?:flowchart|graph)\s/.test(source)) {
      for (const [index, candidate] of diagramLayoutCandidates(source).entries()) {
        if (candidate === source) continue
        if (!current()) return best
        try {
          const alternative = (await mermaid.render(`${id}-layout-${index}`, candidate)).svg
          if (alternative.length > 2 * 1024 * 1024) continue
          const assessment = measure(alternative)
          if (preferDiagramLayout(best.assessment, assessment)) best = { svg: alternative, assessment, adjusted: true }
        } catch (_) { /* A failed alternative must not hide a valid original. */ }
      }
    }
    return best
  }

  function dispose(root) {
    const state = roots.get(root)
    if (!state) return
    state.cancelled = true
    clearTimeout(state.timer)
    state.viewers.forEach(close => close())
    state.urls.forEach(url => URL.revokeObjectURL(url))
    roots.delete(root)
  }

  function update(root, delay = 180) {
    dispose(root)
    const state = { cancelled: false, urls: [], viewers: [], timer: null }
    roots.set(root, state)
    state.timer = setTimeout(() => render(root, state), delay)
  }

  async function render(root, state) {
    const codes = [...root.querySelectorAll('pre > code.language-mermaid')]
    for (let index = 0; index < codes.length; index++) {
      const code = codes[index]
      const source = code.textContent.trim()
      const pre = code.parentElement
      const current = () => !state.cancelled && root.isConnected && root.contains(pre)
      if (!current()) return
      const id = `blog-mermaid-${++sequence}`
      try {
        if (index >= 12 || source.length > 20000) throw new Error('limit')
        // Document directives/frontmatter can override rendering configuration.
        // Leave them as readable source instead of accepting per-article config.
        if (/%%\s*\{|^---(?:\r?\n|$)/m.test(source)) throw new Error('config')
        let rendered = cache.get(source)
        if (!rendered) {
          const mermaid = await getEngine()
          if (!current()) return
          rendered = await layout(mermaid, id, source, current)
          if (!current()) return
          cache.set(source, rendered)
          if (cache.size > 32) cache.delete(cache.keys().next().value)
        }
        if (!current()) return
        const { svg, assessment, adjusted } = rendered
        const url = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml' }))
        state.urls.push(url)
        const figure = document.createElement('figure')
        figure.className = 'markdown-diagram'
        figure.dataset.diagramReadability = assessment.needsReview ? 'review' : 'ready'
        figure.dataset.layoutAdjusted = String(adjusted)
        const image = document.createElement('img')
        image.src = url
        image.alt = 'Mermaid 图示；完整定义可在下方查看源码'
        const details = document.createElement('details')
        const summary = document.createElement('summary')
        summary.textContent = '查看 Mermaid 源码'
        const sourcePre = document.createElement('pre')
        const sourceCode = document.createElement('code')
        sourceCode.textContent = source
        sourcePre.append(sourceCode)
        details.append(summary, sourcePre)
        const expand = document.createElement('button')
        expand.type = 'button'; expand.className = 'diagram-expand'; expand.textContent = '查看完整图示'
        expand.addEventListener('click', () => {
          if (state.cancelled || !root.contains(figure)) return
          const dialog = document.createElement('dialog')
          dialog.className = 'markdown-diagram-viewer'; dialog.setAttribute('aria-label', '完整图示')
          dialog.innerHTML = '<div class="diagram-viewer-tools"><strong>完整图示</strong><button type="button" data-action="out" aria-label="缩小图示">−</button><button type="button" data-action="in" aria-label="放大图示">＋</button><button type="button" data-action="fit">适应窗口</button><button type="button" data-action="close">关闭图示</button></div><div class="diagram-viewer-stage"></div>'
          const stage = dialog.querySelector('.diagram-viewer-stage'), full = document.createElement('img')
          full.src = url; full.alt = image.alt; stage.append(full)
          const box = new DOMParser().parseFromString(svg, 'image/svg+xml').documentElement.getAttribute('viewBox')?.trim().split(/[\s,]+/).map(Number)
          const width = box?.[2] > 0 ? box[2] : image.naturalWidth || 600
          const height = box?.[3] > 0 ? box[3] : image.naturalHeight || 400
          let scale = 1
          const draw = () => { full.style.width = `${width * scale}px`; full.style.height = `${height * scale}px` }
          const fit = () => { scale = Math.min((stage.clientWidth - 24) / width, (stage.clientHeight - 24) / height, 1); draw(); stage.scrollTop = 0; stage.scrollLeft = 0 }
          const close = () => { if (dialog.open) dialog.close(); dialog.remove() }
          state.viewers.push(close)
          dialog.addEventListener('close', () => { dialog.remove(); if (!state.cancelled) expand.focus() })
          dialog.addEventListener('click', event => {
            const action = event.target.closest('button')?.dataset.action
            if (action === 'close') close()
            else if (action === 'fit') fit()
            else if (action === 'in' || action === 'out') { scale = Math.max(0.1, Math.min(4, scale * (action === 'in' ? 1.4 : 1 / 1.4))); draw() }
          })
          document.body.append(dialog); dialog.showModal(); fit()
        })
        figure.append(image)
        if (assessment.needsReview || adjusted) {
          const note = document.createElement('p')
          note.className = 'diagram-readability'
          note.setAttribute('role', 'status')
          note.textContent = assessment.needsReview
            ? '已尝试调整布局，正文缩略图仍可能难以阅读。建议按阶段拆成多张图，或用概览图配合正文说明；完整流程已保留。'
            : '已自动调整图示方向以改善可读性；节点、连线与原始源码均保留。'
          figure.append(note)
        }
        figure.append(expand, details)
        pre.replaceWith(figure)
        figure.dispatchEvent(new CustomEvent('diagram-layout', { bubbles: true, detail: { needsReview: assessment.needsReview, adjusted } }))
      } catch (error) {
        if (!current()) return
        const note = document.createElement('div')
        note.className = 'mermaid-error'
        note.setAttribute('role', 'status')
        note.textContent = error.message === 'config'
          ? '图示配置指令暂不支持，请移除配置后重试。'
          : error.message === 'limit'
            ? '图示过大或数量过多，请拆分后预览。'
            : '图示暂时无法渲染，请检查 Mermaid 语法；源码已保留。'
        pre.after(note)
      }
    }
  }

  return { update, dispose }
}
