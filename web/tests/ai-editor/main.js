import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import '../../src/style/admin-workspace.scss'
import 'virtual:uno.css'
import Fixture from './Fixture.vue'

const app = createApp(Fixture).use(createPinia()).use(ElementPlus)
app.config.globalProperties.$route = { params: {} }
app.config.globalProperties.$router = { back() {}, push() {} }
app.mount('#app')
