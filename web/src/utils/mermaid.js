import { createMermaidRenderer } from './mermaidRenderer.js'
import './mermaid.css'

const renderer = createMermaidRenderer(() => import('mermaid').then(module => module.default))
export const vMermaid = {
  mounted: el => renderer.update(el, 0),
  updated(el, binding) {
    if (binding.value !== binding.oldValue) renderer.update(el)
  },
  beforeUnmount: el => renderer.dispose(el)
}
