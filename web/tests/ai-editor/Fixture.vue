<template>
  <div v-if="showBrand" class="gva-admin-shell brand-fixture">
    <button type="button" @click="toggleTheme">切换测试主题</button>
    <div class="admin-header" style="height:64px;display:flex;align-items:center;margin-top:24px;">
      <AdminBrand @navigate="brandNavigation++" />
      <span style="font-size:14px;color:var(--admin-muted);margin-left:16px;">工作台</span>
    </div>
    <p style="padding:24px;color:var(--admin-muted);">点击返回首页次数：{{ brandNavigation }}</p>
    <div class="admin-header" style="height:64px;display:flex;align-items:center;">
      <AdminBrand compact @navigate="brandNavigation++" />
      <span style="font-size:14px;color:var(--admin-muted);margin-left:16px;">移动端</span>
    </div>
  </div>
  <div v-else-if="showProfile" class="gva-admin-shell">
    <button type="button" @click="toggleTheme">切换测试主题</button>
    <button type="button" @click="setProfileScenario('limited')">模拟GitHub限流</button>
    <button type="button" @click="setProfileScenario('healthy')">恢复GitHub数据</button>
    <Person />
  </div>
  <div v-else-if="showDashboard" class="gva-admin-shell">
    <button type="button" @click="toggleTheme">切换测试主题</button>
    <button type="button" @click="setDashboardScenario('empty')">模拟空博客</button>
    <button type="button" @click="setDashboardScenario('failure')">模拟统计失败</button>
    <button type="button" @click="setDashboardScenario('healthy')">恢复内容数据</button>
    <Dashboard />
  </div>
  <div v-else-if="showState" class="gva-admin-shell">
    <button type="button" @click="toggleTheme">切换测试主题</button>
    <button type="button" @click="setStatusScenario('down')">模拟数据库断开</button>
    <button type="button" @click="setStatusScenario('denied')">模拟指标权限不足</button>
    <button type="button" @click="setStatusScenario('healthy')">模拟数据库恢复</button>
    <ServerState />
  </div>
  <div v-else-if="showWriter" class="gva-admin-shell writer-fixture">
    <button type="button" @click="toggleTheme">切换测试主题</button>
    <WriteBlog ref="writer" />
  </div>
  <ModelConfig v-else-if="showModelConfig" />
  <div v-else class="fixture">
    <MarkdownEditor ref="editor" v-model="content" :document-id="documentId" enable-ai-diff />
    <AiDock v-if="selectionFixture" />
    <WritingAssistantPanel v-else />
    <button v-if="chapterFixture" type="button" @click="documentId = 'article-b'">切换测试文章</button>
  </div>
</template>

<script setup>
  import { nextTick, onMounted, ref } from 'vue'
  import MarkdownEditor from '../../src/components/blog/MarkdownEditor.vue'
  import WritingAssistantPanel from '../../src/components/ai/agents/writing-assistant/WritingAssistantPanel.vue'
  import AiDock from '../../src/components/ai/AiDock.vue'
  import ModelConfig from '../../src/view/ai/modelConfig/modelConfig.vue'
  import WriteBlog from '../../src/view/blog/blog/WriteBlog.vue'
  import ServerState from '../../src/view/system/state.vue'
  import Dashboard from '../../src/view/dashboard/index.vue'
  import Person from '../../src/view/person/person.vue'
  import AdminBrand from '../../src/view/layout/header/AdminBrand.vue'
  const showBrand = new URLSearchParams(window.location.search).has('brand')
  const brandNavigation = ref(0)
  import { useAiStore } from '../../src/pinia/modules/ai.js'
  import { renderSafeMarkdown } from '../../src/utils/safeMarkdown.js'
  import { buildSuggestionPatch, cursorContext } from '../../src/components/ai/agents/writing-assistant/suggestion.js'
  import { titleFillError } from '../../src/components/ai/agents/writing-assistant/writingTask.js'

  const editor = ref()
  const writer = ref()
  const showProfile = new URLSearchParams(window.location.search).has('profile')
  async function setProfileScenario(mode) { await fetch('/test-api/profileScenario', {method:'POST',body:JSON.stringify({mode})}) }
  const showDashboard = new URLSearchParams(window.location.search).has('dashboard')
  async function setDashboardScenario(mode) { await fetch('/test-api/dashboardScenario', {method:'POST',body:JSON.stringify({mode})}) }
  const showState = new URLSearchParams(window.location.search).has('state')
  async function setStatusScenario(mode) { await fetch('/test-api/statusScenario', {method:'POST',body:JSON.stringify({mode})}) }
  const showWriter = new URLSearchParams(window.location.search).has('writer')
  function toggleTheme() { const dark = !document.documentElement.classList.contains('dark'); document.documentElement.classList.toggle('dark', dark); document.documentElement.classList.toggle('light', !dark) }
  const showModelConfig = new URLSearchParams(window.location.search).has('models')
  const selectionFixture = new URLSearchParams(window.location.search).has('selection')
  const chapterFixture = new URLSearchParams(window.location.search).has('chapters')
  const content = ref('前文\n\n选中内容\n\n后文')
  const documentId = ref('article-a')
  const description = ref('原摘要')
  const cover = ref('')
  const title = ref('回归测试文章')
  const suggestion = ref(null)
  const metadataForm = ref({ cate: 9, tagList: [3] })
  const store = useAiStore()
  onMounted(() => {
    if (showState || showDashboard || showProfile || showBrand) return
    if (showWriter) {
      Object.assign(writer.value.form, { title: '文章编辑 · 主题预览', description: '## 文章摘要\n\n检查编辑区与预览区在明暗主题下的阅读效果。', content: '# 正文标题\n\n普通段落与 **强调文字**，以及 `行内代码`。\n\n> 引用内容应保持清晰可读。\n\n```js\n// 代码高亮预览\nconst title = "Hello";\nfunction greet() { return title; }\n```\n\n| 项目 | 状态 |\n| --- | --- |\n| 暗色背景 | 已适配 |\n| 编辑预览 | 清晰 |' })
      return
    }
    if (showModelConfig) return
    store.registerContext('editor', {
      getSelection: () => editor.value.getSelection(),
      getEditorState: () => editor.value.getEditorState(),
      captureSelection: () => editor.value.captureSelection(),
      selectSnapshot: snapshot => editor.value.selectSnapshot(snapshot),
      applySelectionSnapshot: (snapshot, text) => editor.value.applySelectionSnapshot(snapshot, text),
      getFullText: () => content.value,
      getDescription: () => description.value,
      getCover: () => cover.value,
      fillCover: (url, expected) => { if (cover.value !== expected) return { ok: false, message: '封面已改变' }; cover.value = url; return { ok: true } },
      getCursorContext: () => cursorContext(content.value, editor.value.getSelection().start),
      insertAtCursor: (text) => editor.value.insertAtCursor(text),
      insertAiAt: (text, expected) => editor.value.insertAiAt(text, expected),
      getTaxonomy: () => ({ categories: [{ id: 1, categoryName: '技术' }], tags: [{ id: 2, tagName: 'Vue' }, { id: 3, tagName: 'Go' }] }),
      appendChapter: (text, expected) => editor.value.appendChapter(text, expected),
      fillDescription: (text) => { description.value = text; return true },
      applySuggestion: (value) => {
        const result = buildSuggestionPatch(metadataForm.value, value,
          [{ id: 1, categoryName: '技术' }], [{ id: 2, tagName: 'Vue' }, { id: 3, tagName: 'Go' }])
        if (result.ok) { suggestion.value = value; Object.assign(metadataForm.value, result.patch) }
        return result
      },
      getTitle: () => title.value,
      fillTitle: (value, expected) => {
        const message = titleFillError(value, expected, title.value)
        if (message) return { ok: false, message }
        title.value = value
        return { ok: true }
      }
    })
    window.aiEditorTest = {
      getContent: () => content.value,
      setContent: async (text) => { content.value = text; await nextTick() },
      setDocumentId: async (id) => { documentId.value = id; await nextTick() },
      changeDocument: async () => { documentId.value = 'article-b'; await nextTick() },
      getDescription: () => description.value,
      getTitle: () => title.value,
      setTitle: (value) => { title.value = value },
      getSuggestion: () => suggestion.value,
      getMetadataForm: () => metadataForm.value,
      leaveEditor: () => store.unregisterContext('editor'),
      renderSafeMarkdown
    }
  })
</script>

<style>
  body { margin: 20px; }
  .fixture { display: grid; grid-template-columns: minmax(0, 1fr) 380px; gap: 20px; }
</style>
