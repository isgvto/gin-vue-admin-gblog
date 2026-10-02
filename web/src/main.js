import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'uno.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { setupVueRootValidator } from 'vite-check-multiple-dom/client';

import 'element-plus/dist/index.css'
// 引入gin-vue-admin前端初始化相关内容
import './core/gin-vue-admin'
// 引入封装的router
import router from '@/router/index'
import '@/permission'
import run from '@/core/gin-vue-admin.js'
import auth from '@/directive/auth'
import clickOutSide from '@/directive/clickOutSide'
import { store } from '@/pinia'
import App from './App.vue'
import '@/core/error-handel'
import AdminPageHeading from '@/components/admin/PageHeading.vue'
import '@/style/admin-controls.scss'
import MarkdownEditor from '@/components/blog/MarkdownEditor.vue'

const app = createApp(App)

app.config.productionTip = false
app.config.globalProperties.msgSuccess = (msg) => ElMessage.success(msg)
app.config.globalProperties.msgError = (msg) => ElMessage.error(msg)
app.config.globalProperties.$confirm = ElMessageBox.confirm
app.config.globalProperties.blogDateFormat = (value) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (num) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}
app.component('mavon-editor', MarkdownEditor)
app.component('AdminPageHeading', AdminPageHeading)

setupVueRootValidator(app, {
    lang: 'zh'
  })

app
  .use(run)
  .use(ElementPlus)
  .use(store)
  .use(auth)
  .use(clickOutSide)
  .use(router)
  .mount('#app')
export default app
