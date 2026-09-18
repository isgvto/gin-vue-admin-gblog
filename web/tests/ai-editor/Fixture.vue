<template>
  <ModelConfig v-if="showModelConfig" />
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
  import { useAiStore } from '../../src/pinia/modules/ai.js'
  import { renderSafeMarkdown } from '../../src/utils/safeMarkdown.js'
  import { buildSuggestionPatch, cursorContext } from '../../src/components/ai/agents/writing-assistant/suggestion.js'
  import { titleFillError } from '../../src/components/ai/agents/writing-assistant/writingTask.js'

  const editor = ref()
  const showModelConfig = new URLSearchParams(window.location.search).has('models')
  const selectionFixture = new URLSearchParams(window.location.search).has('selection')
  const chapterFixture = new URLSearchParams(window.location.search).has('chapters')
  const content = ref('前文\n\n选中内容\n\n后文')
  const documentId = ref('article-a')
  const description = ref('原摘要')
  const title = ref('回归测试文章')
  const suggestion = ref(null)
  const metadataForm = ref({ cate: 9, tagList: [3] })
  const store = useAiStore()
  onMounted(() => {
    if (showModelConfig) return
    store.registerContext('editor', {
      getSelection: () => editor.value.getSelection(),
      getEditorState: () => editor.value.getEditorState(),
      captureSelection: () => editor.value.captureSelection(),
      applySelectionSnapshot: (snapshot, text) => editor.value.applySelectionSnapshot(snapshot, text),
      getFullText: () => content.value,
      getCursorContext: () => cursorContext(content.value, editor.value.getSelection().start),
      insertAtCursor: (text) => editor.value.insertAtCursor(text),
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
