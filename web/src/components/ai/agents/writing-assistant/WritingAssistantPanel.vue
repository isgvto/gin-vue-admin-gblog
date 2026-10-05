<template>
  <div class="writing-assistant">
    <div class="assistant-sections" role="group" aria-label="助手功能">
      <el-button :type="activeSection === 'writing' ? 'primary' : 'default'" :aria-pressed="activeSection === 'writing'" :disabled="streaming || chapterBusy || visualBusy" @click="activeSection = 'writing'">文字写作</el-button>
      <el-button :type="activeSection === 'visual' ? 'primary' : 'default'" :aria-pressed="activeSection === 'visual'" :disabled="streaming || chapterBusy || visualBusy || editorDiffOpen" @click="activeSection = 'visual'">图示配图</el-button>
    </div>
    <div v-show="activeSection === 'writing'" class="writing-section">
      <div v-if="!aiEnabled" class="ai-disabled-tip">
        <el-empty description="AI 功能不可用" :image-size="72" />
        <p>{{ disabledReason }}</p><el-button size="small" :loading="checkingStatus" @click="checkStatus(true)">重新检测</el-button>
      </div>
      <template v-else>
        <div class="writing-modes">
          <el-button size="small" :type="!chapterMode ? 'primary' : 'default'" :disabled="streaming || chapterBusy || editorDiffOpen" @click="chapterMode = false">写作对话</el-button>
          <el-button size="small" :type="chapterMode ? 'primary' : 'default'" :disabled="streaming || chapterBusy || editorDiffOpen" @click="chapterMode = true">大纲到章节</el-button>
          <el-popover trigger="click" width="280">
            <template #reference><el-button link size="small" :disabled="streaming || chapterBusy || editorDiffOpen">偏好</el-button></template>
            <div class="preference-fields">
              <label>语言风格</label><el-select v-model="tone" size="small"><el-option v-for="item in writingTones" :key="item.value" :label="item.label" :value="item.value" /></el-select>
              <label>目标篇幅</label><el-select v-model="outputLength" size="small"><el-option v-for="item in writingLengths" :key="item.value" :label="item.label" :value="item.value" /></el-select>
              <label>修改力度</label><el-select v-model="editStrength" size="small"><el-option v-for="item in writingStrengths" :key="item.value" :label="item.label" :value="item.value" /></el-select>
              <small>明确的对话要求优先于偏好。</small><el-button link size="small" @click="checkStatus(true)">{{ modelName || '检测模型连接' }}</el-button>
            </div>
          </el-popover>
        </div>
        <WritingChat v-show="!chapterMode" :blocked="chapterBusy || visualBusy || editorDiffOpen" :preferences="{tone, length:outputLength, editStrength}" @busy="streaming = $event" @outline="startChapters" />
        <ChapterWritingWorkflow v-show="chapterMode" ref="chapterWorkflow" v-model:tone="tone" v-model:length="outputLength" :blocked="streaming || editorDiffOpen || !!aiStore.selectionAction" @busy="chapterBusy = $event" />
      </template>
    </div>
    <VisualAssistantPanel v-show="activeSection === 'visual'" :visible="activeSection === 'visual'" :blocked="streaming || chapterBusy || editorDiffOpen || !!aiStore.selectionAction" @busy="visualBusy = $event" />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useAiStore } from '@/pinia/modules/ai'
import { getAiStatus } from '@/api/blog/ai'
import WritingChat from './WritingChat.vue'
import ChapterWritingWorkflow from './ChapterWritingWorkflow.vue'
import VisualAssistantPanel from './VisualAssistantPanel.vue'
import { writingTones, writingLengths, writingStrengths } from './writingPreferences.js'
const aiStore = useAiStore()
const activeSection = ref('writing'), chapterMode = ref(false), chapterWorkflow = ref()
const streaming = ref(false), chapterBusy = ref(false), visualBusy = ref(false)
const aiEnabled = ref(false), disabledReason = ref('正在检查模型…'), checkingStatus = ref(false), modelName = ref('')
const tone = ref('natural'), outputLength = ref('original'), editStrength = ref('standard')
const editorDiffOpen = computed(() => aiStore.diff.active)
let disposed = false, statusCheck = null
watch([streaming, chapterBusy, visualBusy], values => { aiStore.writingBusy = values.some(Boolean) }, {flush:'sync'})
async function checkStatus(connection = false) {
  if (statusCheck) return statusCheck
  checkingStatus.value = true
  statusCheck = (async () => {
    try {
      const res = await getAiStatus(connection)
      if (disposed) return
      aiEnabled.value = !!res.data?.enabled && !res.data?.quotaError
      modelName.value = res.data?.model || ''
      disabledReason.value = res.data?.quotaError || res.data?.reason || '请在 AI 模型配置中设置默认写作模型。'
      if (connection && aiEnabled.value) ElMessage.success('模型连接正常')
    } catch (error) {
      if (disposed) return
      aiEnabled.value = false
      disabledReason.value = error?.response?.status === 401 ? '登录已过期，请重新登录' : error?.response?.status === 403 ? '当前角色没有 AI 接口权限，请联系管理员。' : error?.response?.data?.msg || error.message || '暂时无法连接 AI 服务'
    } finally { checkingStatus.value = false; statusCheck = null }
  })()
  return statusCheck
}
function startChapters(outline) { chapterWorkflow.value?.loadOutline(outline); chapterMode.value = true }
watch(() => aiStore.dockVisible, visible => { if (visible && !aiEnabled.value) checkStatus() })
watch(() => aiStore.selectionAction, request => { if (request) { activeSection.value = 'writing'; chapterMode.value = false } }, {immediate:true})
watch(() => aiStore.contexts.editor?.getEditorState?.()?.documentId, () => { activeSection.value = 'writing'; chapterMode.value = false })
checkStatus()
onBeforeUnmount(() => { disposed = true; aiStore.writingBusy = false; aiStore.selectionAction = null })
</script>

<style scoped>
.writing-assistant{display:flex;flex-direction:column;height:100%;min-height:0;gap:16px}
.assistant-sections{display:flex;flex-shrink:0;gap:8px}.assistant-sections .el-button{flex:1;margin-left:0}
.writing-section{display:flex;flex-direction:column;flex:1;min-height:0;gap:14px}
.writing-modes{display:flex;flex-shrink:0;align-items:center;gap:8px}.writing-modes .el-button{margin-left:0}
.preference-fields{display:flex;flex-direction:column;gap:8px}.preference-fields label{font-size:12px;font-weight:600}.preference-fields small{font-size:12px;color:var(--el-text-color-secondary)}
.ai-disabled-tip{margin:auto;text-align:center;font-size:13px;line-height:1.7}
</style>
