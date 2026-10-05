<template>
  <section class="visual-panel" aria-label="图示与配图">
    <div class="visual-intro"><strong>把文章变成图</strong><p>图示表达关系，配图呈现场景；先预览，满意后采用。</p></div>
    <el-alert v-if="statusError" :title="statusError" type="warning" :closable="false" />
    <div class="visual-fields">
      <div><label>内容来源</label><el-select v-model="sourceMode" aria-label="配图内容来源" :disabled="busy || blocked">
        <el-option label="自动：优先选区，否则正文" value="auto" /><el-option label="当前选区" value="selection" :disabled="!selectionText" /><el-option label="文章正文" value="body" /><el-option label="标题与摘要" value="summary" />
      </el-select></div>
      <div><label>图片类型</label><el-select v-model="kind" aria-label="图片类型" :disabled="busy || blocked">
        <el-option v-for="option in kinds" :key="option.value" :label="option.label" :value="option.value" />
      </el-select></div>
    </div>
    <details class="source-preview"><summary>使用内容：{{ sourceLabel }} · {{ sourceText.length }} 字符</summary><p v-if="sourceTruncated" class="warning">正文较长，本次仅使用前 12000 字符；建议选中相关段落。</p><pre>{{ sourceText || '请先输入文章内容、标题或摘要。' }}</pre><el-button link size="small" :disabled="busy" @click="refreshSelection">重新读取选区</el-button></details>
    <el-alert v-if="recommendation" :title="recommendation" type="info" :closable="false" />
    <template v-if="raster">
      <label>图片方向 <span>实际像素由图片接口类型确定</span></label><el-select v-model="size" aria-label="图片尺寸" :disabled="busy || blocked"><el-option label="横图" value="1536x1024" /><el-option label="方图" value="1024x1024" /><el-option label="竖图" value="1024x1536" /></el-select>
      <p v-if="!status.imageEnabled" class="warning">图片模型未启用，请在「AI 模型配置 → 图片生成模型」中配置。流程图与结构图可继续使用写作模型。</p>
    </template>
    <label for="visual-prompt">配图要求 <span>填写画面内容、风格与限制，也可让 AI 根据文章整理</span></label>
    <el-input id="visual-prompt" v-model="prompt" type="textarea" :rows="4" :maxlength="6000" :disabled="busy || blocked" placeholder="例如：根据正文画登录流程图，使用蓝灰配色，保留成功与失败分支，文字简短。" />
    <div class="visual-controls">
      <el-button v-if="busy" size="small" @click="cancel">停止等待</el-button>
      <template v-else><el-button :disabled="blocked || !hasSource" @click="plan">AI 整理配图方案</el-button><el-button type="primary" :disabled="blocked || !hasSource || kind === 'auto' || (raster && !status.imageEnabled)" @click="generate">{{ result ? '重新生成预览' : '生成预览' }}</el-button></template>
    </div>
    <p class="help">{{ kind === 'auto' ? '先让 AI 整理方案，确认图片类型和要求后再生成。' : '生成会调用模型；预览暂存一小时，采用才上传。' }} 当前存储：{{ status.storage || '未检测' }}。<el-button link size="small" :disabled="busy" @click="loadStatus">刷新状态</el-button></p>
    <div v-if="error" class="visual-error" role="alert">{{ error }}</div>

    <section v-if="result" class="visual-result">
      <div class="result-heading"><strong>{{ kinds.find(option => option.value === result.kind)?.label }}预览</strong><el-tag size="small" :type="result.url ? 'success' : 'info'">{{ result.url ? '已上传' : '未上传' }}</el-tag></div>
      <template v-if="diagram">
        <details class="source-preview"><summary>编辑 Mermaid 源码</summary>
          <label for="visual-mermaid">Mermaid 源码 <span>修改后自动刷新预览</span></label>
          <el-input id="visual-mermaid" v-model="mermaidSource" type="textarea" :rows="7" :maxlength="20000" :disabled="busy || blocked || !!result.url" />
        </details>
        <div ref="previewRef" class="diagram-preview" v-mermaid="diagramHtml" v-html="diagramHtml" @load.capture="diagramReady = true" @diagram-layout="diagramNeedsReview = $event.detail.needsReview" />
        <div v-if="diagramNeedsReview && result.kind === 'flowchart'" class="layout-help">
          <el-button size="small" :disabled="busy || blocked" @click="prepareStagePlan">准备分阶段方案</el-button>
          <p class="help">将要求填入上方，可编辑后重新生成。当前预览与正文会保留。</p>
        </div>
        <div class="visual-controls"><el-button type="primary" :disabled="busy || blocked || !diagramReady || inserted" @click="insertDiagram">{{ inserted ? '已插入正文' : '插入可编辑图示' }}</el-button><el-button :disabled="busy || blocked || !diagramReady || !!result.url" @click="uploadDiagram">导出图片并上传</el-button></div>
      </template>
      <template v-else>
        <img class="generated-image" :src="result.preview || result.url" alt="生成的文章配图预览" />
        <el-button v-if="!result.url" type="primary" :disabled="busy || blocked" @click="upload()">采用并上传</el-button>
      </template>
      <template v-if="result.url">
        <div class="uploaded-link">{{ result.url }}</div>
        <div class="visual-controls"><el-button :disabled="busy || blocked || inserted" @click="insertImage">{{ inserted ? '已插入正文' : '插入正文' }}</el-button><el-button :disabled="busy || blocked || coverSet || !context?.fillCover" @click="setCover">{{ coverSet ? '已设为封面' : '设为封面' }}</el-button><el-button @click="copyLink">复制链接</el-button></div>
      </template>
      <p class="help">插入时追加到正文末尾，可用编辑器的「撤销 AI 修改」撤回；保存文章后才会持久保留。</p>
      <el-button link size="small" :disabled="busy" @click="discard">{{ result.url ? '关闭结果（保留附件）' : '丢弃预览' }}</el-button>
    </section>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useAiStore } from '@/pinia/modules/ai'
import { getVisualStatus, planVisual, generateVisual, adoptVisual, discardVisual } from '@/api/blog/ai'
import { renderSafeMarkdown } from '@/utils/safeMarkdown'
import { vMermaid } from '@/utils/mermaid'
import { exportDiagramPNG } from '@/utils/diagramExport'
const props = defineProps({ visible: Boolean, blocked: Boolean })
const emit = defineEmits(['busy'])
const store = useAiStore(), context = computed(() => store.contexts.editor)
const kinds = [{ value: 'auto', label: '自动推荐' }, { value: 'flowchart', label: '流程图' }, { value: 'structure', label: '结构图' }, { value: 'cover', label: '封面图' }, { value: 'illustration', label: '正文插图' }]
const sourceMode = ref('auto'), selectionText = ref(''), kind = ref('auto'), size = ref('1536x1024')
const prompt = ref(''), recommendation = ref(''), error = ref(''), busy = ref(false)
const status = ref({ imageEnabled: false }), statusError = ref(''), result = shallowRef(null), owner = shallowRef(null)
const mermaidSource = ref(''), previewRef = ref(), diagramReady = ref(false), inserted = ref(false), coverSet = ref(false)
const diagramNeedsReview = ref(false)
const raster = computed(() => kind.value === 'cover' || kind.value === 'illustration')
watch(kind, () => { recommendation.value = '' }, { flush: 'sync' })
const diagram = computed(() => result.value?.kind === 'flowchart' || result.value?.kind === 'structure')
const sourceLabel = computed(() => sourceMode.value === 'summary' ? '标题与摘要' : sourceMode.value === 'selection' || (sourceMode.value === 'auto' && selectionText.value) ? '当前选区' : '文章正文')
const rawSource = computed(() => sourceLabel.value === '当前选区' ? selectionText.value : sourceLabel.value === '标题与摘要' ? context.value?.getDescription?.() || '' : context.value?.getFullText?.() || '')
const sourceText = computed(() => rawSource.value.slice(0, 12000)), sourceTruncated = computed(() => rawSource.value.length > 12000)
const hasSource = computed(() => Boolean(context.value?.getEditorState?.()?.active && (sourceText.value.trim() || context.value?.getTitle?.()?.trim())))
const diagramHtml = computed(() => renderSafeMarkdown('```mermaid\n' + mermaidSource.value + '\n```'))
watch(diagramHtml, () => { diagramReady.value = false; diagramNeedsReview.value = false })
watch(busy, value => emit('busy', value), { flush: 'sync' })
let controller = null, disposed = false, statusSeq = 0
function refreshSelection() { selectionText.value = context.value?.getSelection?.()?.text || '' }
function prepareStagePlan() {
  const requirement = '请将本次完整流程改为阶段概览图：按真实阶段组织节点，保留关键判断、分支和阶段之间的反馈关系，细节步骤由正文说明。reason中说明各阶段包含哪些步骤，并建议哪些阶段适合另行生成详细流程图。不要编造步骤，不要自动改动文章。'
  if (!prompt.value.includes(requirement)) {
    const combined = [prompt.value, requirement].filter(Boolean).join('\n')
    if (combined.length > 6000) return ElMessage.warning('配图要求接近长度上限，请先缩短再准备分阶段方案')
    prompt.value = combined
  }
  document.getElementById('visual-prompt')?.focus()
}
function captureOwner() { const ctx = context.value, state = ctx?.getEditorState?.(); return { context: ctx, editorId: state?.editorId, documentId: state?.documentId, cover: ctx?.getCover?.() || '' } }
function owns(value = owner.value) { const state = context.value?.getEditorState?.(); return !disposed && value && value.context === context.value && state?.active && value.editorId === state.editorId && value.documentId === state.documentId }
function payload() { return { kind: kind.value, source: sourceText.value, title: (context.value?.getTitle?.() || '').slice(0, 500), summary: (context.value?.getDescription?.() || '').slice(0, 3000), prompt: prompt.value, size: size.value } }
async function loadStatus() { const seq = ++statusSeq; statusError.value = ''; try { const res = await getVisualStatus(); if (seq === statusSeq && !disposed) status.value = res.data } catch (e) { if (seq === statusSeq && !disposed) statusError.value = e?.message || '配图状态暂不可用' } }
watch(() => props.visible, visible => { if (visible) { refreshSelection(); loadStatus() } })
function cancel() { controller?.abort(); controller = null; busy.value = false }
async function run(work, captured = captureOwner()) {
  if (busy.value || props.blocked || !owns(captured)) return
  const handle = new AbortController(); controller = handle; busy.value = true; error.value = ''
  try { await work(handle.signal, captured) }
  catch (e) { if (!handle.signal.aborted && owns(captured)) error.value = e?.response?.data?.msg || e?.message || '操作失败，可重试' }
  finally { if (controller === handle) { controller = null; busy.value = false } }
}
async function plan() { await run(async (signal, captured) => { const res = await planVisual(payload(), signal); if (!owns(captured) || signal.aborted) return; kind.value = res.data.kind; prompt.value = res.data.prompt; recommendation.value = res.data.reason || '已整理描述，请确认类型和内容后生成预览。' }) }
async function generate() {
  if (!hasSource.value || kind.value === 'auto') return
  await run(async (signal, captured) => {
    const res = await generateVisual(payload(), signal)
    if (!owns(captured) || signal.aborted) return
    const previous = result.value
    result.value = res.data; owner.value = captured; mermaidSource.value = res.data.mermaid || ''; inserted.value = false; coverSet.value = false
    if (previous?.id && !previous.url) discardVisual(previous.id).catch(() => {})
  })
}
async function upload(png = '') { if (!owns()) return ElMessage.warning('文章已切换，请重新生成'); await run(async (signal, captured) => { const current = result.value; const res = await adoptVisual({ id: current.id, png }, signal); if (owns(captured) && !signal.aborted && result.value === current) { result.value = { ...current, ...res.data }; ElMessage.success('图片已上传，可插入正文或设为封面') } }, owner.value) }
async function uploadDiagram() {
  await run(async (signal, captured) => {
    const current = result.value
    const image = previewRef.value?.querySelector('.markdown-diagram > img')
    const png = await exportDiagramPNG(image)
    if (signal.aborted || !owns(captured) || result.value !== current) return
    const res = await adoptVisual({ id: current.id, png }, signal)
    if (!signal.aborted && owns(captured) && result.value === current) {
      result.value = { ...current, ...res.data }
      ElMessage.success('高清图示已上传，可插入正文或设为封面')
    }
  }, owner.value)
}
function insert(text) {
  if (!owns() || props.blocked || busy.value || inserted.value) return
  const expected = context.value.getEditorState()
  const applied = context.value?.appendChapter?.(text, expected)
  if (!applied?.ok) return ElMessage.warning(applied?.message || '当前页面不支持插入，请复制内容')
  inserted.value = true; ElMessage.success('已追加到正文，请检查并保存文章')
}
function insertDiagram() { if (diagramReady.value) insert('```mermaid\n' + mermaidSource.value + '\n```') }
function insertImage() { if (result.value?.url) { const url = result.value.url.replaceAll(' ', '%20').replaceAll('(', '%28').replaceAll(')', '%29'); insert(`![文章配图](${url})`) } }
function setCover() { if (!owns() || busy.value || props.blocked || !result.value?.url) return; const applied = context.value?.fillCover?.(result.value.url, owner.value.cover); if (!applied?.ok) return ElMessage.warning(applied?.message || '当前页面不支持设置封面'); coverSet.value = true; ElMessage.success('已填写首图链接，请保存文章') }
async function copyLink() { try { await navigator.clipboard.writeText(result.value.url); ElMessage.success('链接已复制') } catch (_) { ElMessage.error('复制失败，请手动复制链接') } }
async function discard() { if (!result.value || busy.value) return; const current = result.value; try { if (!current.url) await discardVisual(current.id); if (result.value === current) { result.value = null; owner.value = null } } catch (e) { error.value = e?.message || '丢弃失败' } }
watch([() => context.value, () => context.value?.getEditorState?.()?.documentId, () => context.value?.getEditorState?.()?.editorId, () => context.value?.getEditorState?.()?.active], () => {
  cancel(); result.value = null; owner.value = null; selectionText.value = ''; prompt.value = ''; recommendation.value = ''; error.value = ''; sourceMode.value = 'auto'; kind.value = 'auto'
}, { flush: 'sync' })
onBeforeUnmount(() => { disposed = true; statusSeq++; cancel() })
</script>

<style scoped>
.visual-panel { display: flex; flex-direction: column; flex: 1; gap: 10px; min-width: 0; min-height: 0; overflow-y: auto; overscroll-behavior: contain; scrollbar-gutter: stable; padding: 0 4px 12px 0; }
.visual-panel > * { flex-shrink: 0; }
.visual-intro strong { font-size: 16px; }
.visual-intro p, .help { margin: 4px 0 0; font-size: 12px; line-height: 1.7; color: var(--el-text-color-secondary); }
.visual-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
label { display: block; font-size: 12px; color: var(--el-text-color-regular); margin-bottom: 4px; }
label span { color: var(--el-text-color-secondary); font-size: 11px; }
.el-select { width: 100%; }
.source-preview { padding: 10px 12px; border-radius: 8px; background: var(--el-fill-color-light); font-size: 12px; }
summary { cursor: pointer; color: var(--el-text-color-secondary); }
pre { white-space: pre-wrap; overflow-wrap: anywhere; max-height: 180px; overflow: auto; font-family: inherit; line-height: 1.7; }
.visual-controls { display: flex; flex-wrap: wrap; gap: 8px; }
.visual-controls .el-button { margin: 0; }
.visual-result { border: 1px solid var(--el-border-color-light); border-radius: 10px; padding: 12px; display: flex; flex-direction: column; gap: 12px; }
.result-heading { display: flex; align-items: center; justify-content: space-between; }
.generated-image { width: 100%; height: auto; border-radius: 8px; background: var(--el-fill-color-light); }
.visual-error, .warning { color: var(--el-color-warning-dark-2); font-size: 12px; line-height: 1.7; margin: 0; }
.uploaded-link { padding: 8px; border-radius: 6px; background: var(--el-fill-color-light); overflow-wrap: anywhere; font-size: 12px; }
</style>
