<template>
  <section class="gva-table-box image-settings" v-loading="loading" aria-labelledby="independent-image-title">
    <div class="settings-heading"><div><h3 id="independent-image-title">图片生成 · 独立模型配置</h3><p>封面与插图使用此配置，独立于上方文本模型列表。</p></div><el-tag :type="dirty ? 'warning' : form.enabled ? 'success' : 'info'">{{ dirty ? '未保存' : form.enabled ? '已启用' : '未启用' }}</el-tag></div>
    <el-alert v-if="loadError" :title="loadError" type="error" :closable="false"><el-button link @click="load">重新加载</el-button></el-alert>
    <el-form v-else label-position="top" class="image-settings-form" @submit.prevent="save">
      <el-form-item label="启用图片生成"><el-switch v-model="form.enabled" /></el-form-item>
      <el-form-item label="接口类型"><el-select v-model="form.provider"><el-option label="OpenAI 兼容" value="openai" /><el-option label="火山方舟 Ark" value="ark" /></el-select></el-form-item>
      <el-form-item label="接口地址"><el-input v-model="form.baseUrl" :placeholder="form.provider === 'ark' ? '留空使用 Ark 默认地址' : '例如 https://api.openai.com/v1'" autocomplete="off" /></el-form-item>
      <el-form-item label="图片模型"><el-input v-model="form.model" placeholder="填写图片模型标识或 ep-xxx" /></el-form-item>
      <el-form-item label="API Key"><el-input v-model="form.apiKey" type="password" show-password autocomplete="new-password" :placeholder="hasKey ? '已配置，留空保留原 Key' : '填写图片接口 Key'" /><el-checkbox v-if="hasKey" v-model="form.clearKey">清除已保存的 Key</el-checkbox></el-form-item>
      <el-form-item label="生成超时（秒）"><el-input-number v-model="form.timeoutSeconds" :min="10" :max="600" :step="10" /></el-form-item>
      <div class="settings-actions"><el-button :loading="testing" :disabled="saving" @click="test">测试图片生成</el-button><el-button type="primary" native-type="submit" :loading="saving" :disabled="testing">保存图片配置</el-button></div>
    </el-form>
    <p class="settings-note">测试会实际调用供应商，可能计费；不会上传图片。流程图、推荐类型与描述使用默认文本模型。配图先预览，采用后才上传。</p>
  </section>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getFeatureModelConfig, saveFeatureModelConfig, testFeatureModelConfig } from '@/api/ai/modelConfig'
const form = ref({enabled:false, provider:'openai', baseUrl:'', model:'', apiKey:'', clearKey:false, timeoutSeconds:180})
const loading=ref(false), saving=ref(false), testing=ref(false), hasKey=ref(false), loadError=ref(''), saved=ref('')
const dirty=computed(()=>saved.value && saved.value!==JSON.stringify(form.value))
async function load() { loading.value=true;try { const res=await getFeatureModelConfig('image'); const cfg=res.data.config;form.value={enabled:cfg.enabled,provider:cfg.provider || 'openai',baseUrl:cfg.baseUrl || '',model:cfg.model || '',timeoutSeconds:cfg.timeoutSeconds || 180,apiKey:'',clearKey:false};hasKey.value=!!res.data.hasKey;saved.value=JSON.stringify(form.value);loadError.value='' } catch(e) { loadError.value=e.message || '读取图片配置失败' } finally {loading.value=false} }
async function save() {saving.value=true;try {await saveFeatureModelConfig('image',form.value);await load();ElMessage.success('独立图片配置已保存')} catch(_) {} finally {saving.value=false} }
async function test() {testing.value=true;try {const res=await testFeatureModelConfig('image',form.value);ElMessage.success(res.msg)} catch(_) {} finally {testing.value=false} }
onMounted(load)
</script>
<style scoped>
.image-settings { margin-top:16px; }
.settings-heading { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px; }
h3 { margin:0 0 6px; font-size:16px; }
p { margin:0; font-size:13px; color:var(--el-text-color-secondary); line-height:1.7; }
.image-settings-form { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:0 24px; max-width:1000px; }
.el-select { width:100%; }
.settings-actions { grid-column:1 / -1; display:flex; gap:8px; margin-bottom:16px; }
.settings-actions .el-button { margin:0; }
@media(max-width:700px) { .image-settings-form { grid-template-columns:1fr; } }
</style>
