<template>
  <section class="chapter-workflow" aria-label="大纲到章节写作">
    <div class="workflow-heading">
      <div><strong>{{ confirmed ? '逐节写作' : '先确定文章结构' }}</strong><p>{{ confirmed ? `已采纳 ${adoptedCount} / ${sections.length} 章` : '确认大纲后，每次生成一章，修改满意再采纳。' }}</p></div>
      <el-button v-if="confirmed" size="small" :disabled="busy || blocked" @click="restart">重新开始</el-button>
    </div>
    <p class="draft-note">章节草稿仅保留在当前页面，离开文章会清空；已采纳内容请随文章保存。</p>

    <template v-if="!confirmed">
      <label class="field-label" for="chapter-outline">文章大纲</label>
      <el-input id="chapter-outline" v-model="outlineText" type="textarea" :rows="9" :maxlength="12000"
                placeholder="## 背景与问题&#10;- 介绍背景和读者面临的问题&#10;&#10;## 解决方案&#10;- 说明步骤和示例"
                :disabled="blocked" />
      <p class="field-help">每章用同级 Markdown 标题；子标题和列表作为本章要点。最多20章。</p>
      <ol v-if="outlineText.trim() && !parsed.error" class="outline-preview">
        <li v-for="(section, index) in parsed.sections" :key="index">{{ section.title }}</li>
      </ol>
      <p v-else-if="outlineText.trim()" class="workflow-error" role="status">{{ parsed.error }}</p>
      <el-button type="primary" :disabled="blocked || !hasEditor || !!parsed.error" @click="confirmOutline">确认大纲，开始逐节写作</el-button>
      <p v-if="!hasEditor" class="field-help">请在文章编辑页使用章节写作。</p>
    </template>

    <template v-else>
      <nav class="chapter-list" aria-label="章节列表">
        <button v-for="(section, index) in sections" :key="index" type="button" class="chapter-step"
                :class="{ current: index === activeIndex, adopted: section.adopted }" :aria-current="index === activeIndex ? 'step' : undefined"
                :disabled="busy" @click="activeIndex = index">
          <span class="step-number">{{ index + 1 }}</span><span class="step-title">{{ section.title }}</span>
          <small>{{ section.adopted ? '已采纳' : section.ready ? '待采纳' : '待生成' }}</small>
        </button>
      </nav>
      <div v-if="current" class="chapter-detail">
        <strong>{{ activeIndex + 1 }}. {{ current.title }}</strong>
        <p v-if="current.brief" class="chapter-brief">{{ current.brief }}</p>
        <p v-if="current.adopted" class="adopted-note">本章已追加到正文，可在正文编辑器中继续修改。</p>
        <p v-else-if="activeIndex !== nextIndex" class="field-help">请先完成并采纳前面的章节，再生成本章。</p>
        <template v-if="!current.adopted">
          <div class="chapter-preferences">
            <el-select :model-value="tone" size="small" aria-label="章节写作语气" :disabled="busy || blocked" @update:model-value="emit('update:tone', $event)">
              <el-option v-for="option in writingTones" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
            <el-select :model-value="length" size="small" aria-label="章节篇幅" :disabled="busy || blocked" @update:model-value="emit('update:length', $event)">
              <el-option label="标准篇幅" value="original" /><el-option label="精简篇幅" value="shorter" /><el-option label="扩展篇幅" value="longer" />
            </el-select>
          </div>
          <label class="field-label" for="chapter-instruction">本章补充或修改要求</label>
          <el-input id="chapter-instruction" v-model="current.instruction" type="textarea" :rows="2" :maxlength="2000"
                    placeholder="例如：补充一个实际案例，减少术语。" :disabled="busy || blocked" />
          <div class="chapter-controls">
            <el-button v-if="busy" type="danger" plain size="small" @click="stop">停止生成</el-button>
            <el-button v-else type="primary" size="small" :disabled="blocked || activeIndex !== nextIndex" @click="generate">
              {{ current.ready ? '按要求重新生成本章' : '生成本章' }}
            </el-button>
          </div>
        </template>
        <p v-if="notice" class="field-help" role="status">{{ notice }}</p>
        <p v-if="current.error" class="workflow-error" role="status">{{ current.error }}</p>
        <div v-if="busy || (!current.ready && current.output)" class="chapter-preview">
          <div v-if="current.output" v-mermaid="current.output" v-html="renderSafeMarkdown(current.output)" />
          <span v-else>正在生成本章…</span>
        </div>
        <template v-if="current.ready && !busy">
          <label class="field-label" for="chapter-draft">{{ current.adopted ? '采纳时的草稿' : '本章草稿（可直接修改 Markdown）' }}</label>
          <el-input id="chapter-draft" v-model="current.draft" type="textarea" :rows="10" :maxlength="16000" :disabled="current.adopted || blocked" />
          <details class="draft-preview"><summary>预览本章</summary><div class="chapter-preview" v-mermaid="chapterMarkdown(current)" v-html="renderSafeMarkdown(chapterMarkdown(current))" /></details>
          <el-button v-if="!current.adopted" type="success" :disabled="blocked || !current.draft.trim() || activeIndex !== nextIndex" @click="adopt">
            采纳本章到正文末尾
          </el-button>
        </template>
        <p v-if="adoptedCount === sections.length" class="complete-note" role="status">全部章节已采纳，请检查正文并保存文章。</p>
      </div>
    </template>
  </section>
</template>

<script setup>
  import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { streamAiChat } from '@/api/blog/ai'
  import { useAiStore } from '@/pinia/modules/ai'
  import { renderSafeMarkdown } from '@/utils/safeMarkdown'
  import { vMermaid } from '@/utils/mermaid'
  import { chapterMarkdown, chapterOwnerError, chapterPayload, parseChapterOutline } from './chapterWorkflow.js'
  import { writingTones } from './writingPreferences.js'

  const props = defineProps({ blocked: Boolean, tone: String, length: String })
  const emit = defineEmits(['busy', 'update:tone', 'update:length'])
  const store = useAiStore()
  const context = computed(() => store.contexts.editor)
  const hasEditor = computed(() => !!context.value?.getEditorState?.()?.active && !!context.value?.appendChapter)
  const outlineText = ref('')
  const parsed = computed(() => parseChapterOutline(outlineText.value))
  const confirmed = ref(false)
  const sections = ref([])
  const owner = shallowRef(null)
  const activeIndex = ref(0)
  const current = computed(() => sections.value[activeIndex.value])
  const nextIndex = computed(() => sections.value.findIndex(section => !section.adopted))
  const adoptedCount = computed(() => sections.value.filter(section => section.adopted).length)
  const busy = ref(false)
  const notice = ref('')
  let activeRequest = null
  watch(busy, value => emit('busy', value), { flush: 'sync' })

  const loadOutline = (text) => {
    if (confirmed.value || outlineText.value.trim()) {
      ElMessage.warning('章节流程已有大纲，请先在该页签清空大纲或重新开始')
      return
    }
    outlineText.value = text
  }
  defineExpose({ loadOutline })

  const confirmOutline = () => {
    if (props.blocked || !hasEditor.value || parsed.value.error) return
    const state = context.value.getEditorState()
    owner.value = { context: context.value, editorId: state.editorId, documentId: state.documentId }
    sections.value = parsed.value.sections.map(section => ({ ...section, draft: '', instruction: '', ready: false, adopted: false, output: '', error: '' }))
    activeIndex.value = 0
    confirmed.value = true
  }
  const stop = () => {
    const request = activeRequest
    activeRequest = null
    request?.handle?.abort()
    if (request) request.section.error = request.section.ready ? '已停止，上次完整草稿已保留。' : '已停止，未完成的内容不能采纳，请重新生成。'
    busy.value = false
  }
  const reset = () => {
    stop()
    outlineText.value = ''
    confirmed.value = false
    sections.value = []
    owner.value = null
    activeIndex.value = 0
    notice.value = ''
  }
  const restart = async () => {
    const savedOwner = owner.value
    try {
      await ElMessageBox.confirm('重新开始会清除本流程的章节草稿，已采纳的正文会保留。', '重新开始章节写作', { type: 'warning', confirmButtonText: '重新开始', cancelButtonText: '保留当前流程' })
      if (owner.value !== savedOwner || busy.value || props.blocked) return
      reset()
    } catch { /* 取消时保留当前流程。 */ }
  }
  const generate = async () => {
    if (busy.value || props.blocked || activeIndex.value !== nextIndex.value || !current.value) return
    const message = chapterOwnerError(owner.value, context.value)
    if (message) return ElMessage.warning(message)
    const section = current.value
    if ([...section.draft].length > 8000) return ElMessage.warning('章节草稿超过8000字，请精简后再生成')
    const payload = chapterPayload(sections.value, activeIndex.value, context.value, props.tone, props.length)
    const request = { section, handle: null }
    activeRequest = request
    section.output = ''
    section.error = ''
    notice.value = ''
    busy.value = true
    let completed = false
    try {
      request.handle = streamAiChat(payload, {
        onDelta: delta => { if (activeRequest === request) section.output += delta },
        onContext: info => { if (activeRequest === request) notice.value = info.notice || '' },
        onDone: () => { if (activeRequest === request) completed = true }
      })
      const result = await request.handle.promise
      if (activeRequest !== request) return
      if (result.aborted || result.errorMessage || !completed || !section.output.trim()) {
        throw new Error(result.errorMessage || '本次生成未完整结束，请重新生成')
      }
      const error = chapterOwnerError(owner.value, context.value)
      if (error) throw new Error(error)
      section.draft = section.output
      section.ready = true
    } catch (error) {
      if (activeRequest === request) section.error = `${error.message || '生成失败'}。${section.ready ? '上次完整草稿已保留。' : '未完成内容不能采纳。'}`
    } finally {
      if (activeRequest === request) { activeRequest = null; busy.value = false }
    }
  }
  const adopt = () => {
    if (busy.value || props.blocked || activeIndex.value !== nextIndex.value || !current.value?.ready || current.value.adopted) return
    const message = chapterOwnerError(owner.value, context.value)
    if (message) return ElMessage.warning(message)
    const markdown = chapterMarkdown(current.value)
    if (!markdown) return ElMessage.warning('请先填写本章正文')
    const result = context.value.appendChapter(markdown, context.value.getEditorState())
    if (!result?.ok) return ElMessage.warning(result?.message || '当前编辑器无法采纳章节')
    current.value.adopted = true
    ElMessage.success('本章已追加到正文末尾')
    notice.value = ''
    if (nextIndex.value >= 0) activeIndex.value = nextIndex.value
  }

  watch([context, () => context.value?.getEditorState?.()?.editorId, () => context.value?.getEditorState?.()?.documentId,
    () => context.value?.getEditorState?.()?.active], reset, { flush: 'sync' })
  onBeforeUnmount(stop)
</script>

<style scoped>
.chapter-workflow { flex: 1; min-height: 0; overflow-y: auto; display: flex; flex-direction: column; gap: 12px; padding: 2px 4px 12px 0; }
.workflow-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 10px; }
.workflow-heading strong { font-size: 16px; color: var(--el-text-color-primary); }
.workflow-heading p, .draft-note, .field-help { margin: 4px 0 0; font-size: 12px; line-height: 1.6; color: var(--el-text-color-secondary); }
.draft-note { border-left: 2px solid var(--el-border-color); padding-left: 9px; }
.field-label { display: block; font-size: 13px; font-weight: 600; }
.outline-preview { margin: 0; padding: 12px 12px 12px 32px; background: var(--el-fill-color-light); border-radius: 6px; font-size: 13px; line-height: 1.9; overflow-wrap: anywhere; }
.chapter-list { display: flex; flex-direction: column; gap: 4px; }
.chapter-step { display: flex; align-items: center; gap: 9px; width: 100%; padding: 9px; text-align: left; border: 1px solid transparent; border-radius: 6px; background: var(--el-fill-color-light); color: var(--el-text-color-regular); cursor: pointer; }
.chapter-step.current { border-color: var(--el-color-primary); background: var(--el-color-primary-light-9); }
.chapter-step:disabled { cursor: wait; opacity: .65; }
.step-number { width: 23px; height: 23px; line-height: 23px; text-align: center; flex-shrink: 0; background: var(--el-bg-color); border-radius: 50%; font-size: 12px; }
.step-title { flex: 1; min-width: 0; overflow-wrap: anywhere; font-size: 13px; }
.chapter-step small { white-space: nowrap; color: var(--el-text-color-secondary); }
.chapter-step.adopted small, .adopted-note, .complete-note { color: var(--el-color-success); }
.chapter-detail { display: flex; flex-direction: column; gap: 12px; padding-top: 12px; border-top: 1px solid var(--el-border-color-light); }
.chapter-preferences { display: flex; gap: 8px; }
.chapter-brief { white-space: pre-wrap; overflow-wrap: anywhere; margin: 0; font-size: 12px; line-height: 1.6; color: var(--el-text-color-secondary); }
.workflow-error { margin: 0; font-size: 12px; line-height: 1.6; color: var(--el-color-danger); }
.chapter-preview { padding: 10px; border: 1px solid var(--el-border-color-light); border-radius: 5px; overflow-wrap: anywhere; font-size: 13px; line-height: 1.7; }
.chapter-preview :deep(pre) { overflow: auto; }
.draft-preview summary { cursor: pointer; font-size: 12px; margin-bottom: 8px; }
.complete-note, .adopted-note { margin: 0; font-size: 12px; line-height: 1.6; }
</style>
