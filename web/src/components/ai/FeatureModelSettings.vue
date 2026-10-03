<template>
  <section class="gva-table-box analysis-settings" v-loading="loading" :aria-labelledby="headingId">
    <div class="settings-heading">
      <div><h3 :id="headingId">功能模型分配 · {{ title }}</h3><p>{{ description }}</p></div>
      <el-tag :type="dirty ? 'warning' : (form.enabled ? 'success' : 'info')">{{ dirty ? '未保存' : (form.enabled ? '已启用' : '未启用') }}</el-tag>
    </div>
    <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" show-icon>
      <el-button link @click="load">重新加载</el-button>
    </el-alert>
    <el-form v-else label-position="top" class="settings-form" @submit.prevent="save">
      <el-form-item label="启用功能"><el-switch v-model="form.enabled" /></el-form-item>
      <el-form-item label="使用模型">
        <el-select v-model="form.modelId" class="model-picker" filterable>
          <el-option :value="0" label="使用数据库中的全局默认模型" />
          <el-option v-if="missingModel" :value="form.modelId" :label="`已失效的模型（ID ${form.modelId}），请重新选择`" disabled />
          <el-option v-for="model in models" :key="model.id" :value="model.id" :label="`${model.name} · ${model.model}`" />
        </el-select>
      </el-form-item>
      <el-form-item :label="timeoutLabel"><el-input-number v-model="form.timeoutSeconds" :min="10" :max="timeoutMax" :step="10" /></el-form-item>
      <div class="settings-actions"><el-button v-if="showTest" :loading="testing" :disabled="saving || missingModel" @click="test">测试示例</el-button><el-button type="primary" native-type="submit" :loading="saving" :disabled="testing">保存分配</el-button></div>
    </el-form>
    <p class="settings-note">{{ note }}</p>
    <el-dialog v-model="resultVisible" :title="`${title} · 示例测试`" width="min(820px, 94vw)">
      <el-tag>{{ resultModel }}</el-tag>
      <pre class="analysis-result">{{ result }}</pre>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getFeatureModelConfig, saveFeatureModelConfig, testFeatureModelConfig } from '@/api/ai/modelConfig'

const props = defineProps({
  feature: { type: String, required: true },
  title: { type: String, required: true },
  description: { type: String, required: true },
  timeoutMax: { type: Number, default: 180 },
  initialTimeout: { type: Number, default: 60 },
  showTest: { type: Boolean, default: true },
  timeoutLabel: { type: String, default: '分析超时（秒）' },
  note: { type: String, default: '本功能会将提交的内容发送至所选供应商，测试也会产生一次模型请求。指定模型失效时会提示错误，不会自动切换到其他供应商。测试使用当前选择，保存后才对该功能生效。' }
})
const headingId = computed(() => `feature-model-${props.feature}`)
const form = ref({ enabled: false, modelId: 0, timeoutSeconds: props.initialTimeout })
const savedForm = ref('')
const dirty = computed(() => savedForm.value && savedForm.value !== JSON.stringify(form.value))
const models = ref([])
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const loadError = ref('')
const result = ref('')
const resultModel = ref('')
const resultVisible = ref(false)
const missingModel = computed(() => form.value.modelId !== 0 && !models.value.some(model => model.id === form.value.modelId))
async function load() {
  loading.value = true
  try {
    const response = await getFeatureModelConfig(props.feature)
    form.value = response.data.config
    savedForm.value = JSON.stringify(form.value)
    models.value = response.data.models || []
    loadError.value = ''
  } catch (error) { loadError.value = error.message || '读取功能配置失败' }
  finally { loading.value = false }
}
async function save() {
  saving.value = true
  try { await saveFeatureModelConfig(props.feature, form.value); savedForm.value = JSON.stringify(form.value); ElMessage.success(`${props.title}模型分配已保存`) }
  catch (_) { /* 请求层展示错误 */ }
  finally { saving.value = false }
}
async function refreshModels() {
  try { const response = await getFeatureModelConfig(props.feature); models.value = response.data.models || [] }
  catch (_) { /* 保留当前编辑内容 */ }
}
async function test() {
  testing.value = true
  try {
    const response = await testFeatureModelConfig(props.feature, form.value)
    result.value = response.data.solution
    resultModel.value = response.data.model
    resultVisible.value = true
  } catch (_) { /* 请求层展示错误 */ }
  finally { testing.value = false }
}
defineExpose({ refreshModels })
onMounted(load)
</script>

<style scoped>
.analysis-settings { margin-top: 16px; }
.settings-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.settings-heading h3 { margin: 0 0 6px; font-size: 16px; font-weight: 600; }
.settings-heading p, .settings-note { margin: 0; color: var(--el-text-color-secondary); font-size: 13px; line-height: 1.7; }
.settings-form { display: flex; align-items: flex-end; flex-wrap: wrap; gap: 0 24px; }
.model-picker { width: min(420px, 75vw); }
.settings-actions { display: flex; margin-bottom: 18px; }
.analysis-result { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; line-height: 1.85; margin-top: 20px; }
</style>
