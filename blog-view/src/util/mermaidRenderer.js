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
        theme: 'default', htmlLabels: false, flowchart: { htmlLabels: false },
        fontFamily: 'sans-serif', maxTextSize: 20000, maxEdges: 200,
        secure: ['secure', 'securityLevel', 'startOnLoad', 'maxTextSize', 'maxEdges',
          'suppressErrorRendering', 'htmlLabels', 'flowchart', 'themeCSS']
      })
      return mermaid
    }).catch(error => { engine = null; throw error })
    return engine
  }

  function dispose(root) {
    const state = roots.get(root)
    if (!state) return
    state.cancelled = true
    clearTimeout(state.timer)
    state.urls.forEach(url => URL.revokeObjectURL(url))
    roots.delete(root)
  }

  function update(root, delay = 180) {
    dispose(root)
    const state = { cancelled: false, urls: [], timer: null }
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
        let svg = cache.get(source)
        if (!svg) {
          const mermaid = await getEngine()
          if (!current()) return
          svg = (await mermaid.render(id, source)).svg
          if (svg.length > 2 * 1024 * 1024) throw new Error('limit')
          cache.set(source, svg)
          if (cache.size > 32) cache.delete(cache.keys().next().value)
        }
        if (!current()) return
        const url = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml' }))
        state.urls.push(url)
        const figure = document.createElement('figure')
        figure.className = 'markdown-diagram'
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
        figure.append(image, details)
        pre.replaceWith(figure)
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
