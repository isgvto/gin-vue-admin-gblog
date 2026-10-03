import Vue from 'vue'
import { createMermaidRenderer } from './mermaidRenderer.js'
import './mermaid.css'

const renderer = createMermaidRenderer(() => import('mermaid').then(module => module.default))
Vue.directive('mermaid', {
  inserted: el => renderer.update(el, 0),
  componentUpdated(el, binding) {
    if (binding.value !== binding.oldValue) renderer.update(el)
  },
  unbind: el => renderer.dispose(el)
})
