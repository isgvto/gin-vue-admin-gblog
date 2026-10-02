<template>
  <section class="gva-table-box analysis-settings" v-loading="loading" aria-labelledby="analysis-setting-title">
    <div class="settings-heading">
      <div><h3 id="analysis-setting-title">功能模型分配 · 错误日志分析</h3><p>指定用于分析系统错误日志的模型，独立控制启用状态。</p></div>
      <el-tag :type="form.enabled ? 'success' : 'info'">{{ form.enabled ? '已启用' : '未启用' }}</el-tag>
    </div>
    <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" show-icon>
      <el-button link @click="load">重新加载</el-button>
    </el-alert>
    <el-form v-else label-position="top" class="settings-form" @submit.prevent="save">
      <el-form-item label="启用错误分析"><el-switch v-model="form.enabled" /></el-form-item>
      <el-form-item label="分析模型">
        <el-select v-model="form.modelId" class="model-picker" filterable>
          <el-option :value="0" label="使用数据库中的全局默认模型" />
          <el-option v-if="missingModel" :value="form.modelId" :label="`已失效的模型（ID ${form.modelId}），请重新选择`" disabled />
          <el-option v-for="model in models" :key="model.id" :value="model.id" :label="`${model.name} · ${model.model}`" />
        </el-select>
      </el-form-item>
      <el-form-item label="分析超时（秒）"><el-input-number v-model="form.timeoutSeconds" :min="10" :max="180" :step="10" /></el-form-item>
      <div class="settings-actions"><el-button :loading="testing" :disabled="saving || missingModel" @click="test">测试示例分析</el-button><el-button type="primary" native-type="submit" :loading="saving" :disabled="testing">保存分配</el-button></div>
    </el-form>
    <p class="settings-note">分析会将脱敏后的日志发送至所选供应商，测试也会产生一次模型请求。指定模型失效时会提示错误，不会自动切换到其他供应商。测试使用当前选择，保存后才对日志分析生效。</p>
    <el-dialog v-model="resultVisible" title="示例错误分析" width="min(820px, 94vw)">
      <el-tag>{{ resultModel }}</el-tag>
      <pre class="analysis-result">{{ result }}</pre>
    </el-dialog>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getErrorAnalysisConfig, saveErrorAnalysisConfig, testErrorAnalysis } from '@/api/ai/modelConfig'

const form = ref({ enabled: false, modelId: 0, timeoutSeconds: 60 })
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
    const response = await getErrorAnalysisConfig()
    form.value = response.data.config
    models.value = response.data.models || []
    loadError.value = ''
  } catch (error) { loadError.value = error.message || '读取错误分析配置失败' }
  finally { loading.value = false }
}
async function save() {
  saving.value = true
  try { await saveErrorAnalysisConfig(form.value); ElMessage.success('错误分析模型分配已保存') }
  catch (_) { /* 请求层展示错误 */ }
  finally { saving.value = false }
}
async function refreshModels() {
  try { const response = await getErrorAnalysisConfig(); models.value = response.data.models || [] }
  catch (_) { /* 保留当前编辑内容 */ }
}
async function test() {
  testing.value = true
  try {
    const response = await testErrorAnalysis(form.value)
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
