import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import '../../src/style/admin-workspace.scss'
import 'virtual:uno.css'
import Fixture from './Fixture.vue'
import { createRouter, createWebHashHistory } from 'vue-router'

const app = createApp(Fixture).use(createPinia()).use(ElementPlus)
if (new URLSearchParams(window.location.search).has('dashboard')) {
  const component = { template: '<div />' }
  app.use(createRouter({ history: createWebHashHistory(), routes: [
    { path: '/', component },
    { path: '/layout/gblog/edit/:id?', component },
    { path: '/layout/gblog/list', component },
    ...['menu', 'api', 'authority', 'user', 'autoPkg', 'autoCode'].map(name => ({ name, path: `/layout/admin/${name}`, component }))
  ] }))
} else {
  app.config.globalProperties.$route = { params: {} }
  app.config.globalProperties.$router = { back() {}, push() {} }
}
app.mount('#app')
