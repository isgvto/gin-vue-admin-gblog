<template>
  <div>
    <!-- 顶部栏 AI 按钮 -->
    <div class="ai-entry" @click="openDock()">
      <el-badge is-dot :hidden="!anyAgentAvailable">
        <el-icon :size="20"><MagicStick /></el-icon>
      </el-badge>
      <span v-if="!isMobile" class="ai-entry-text">AI</span>
    </div>

    <!-- 统一 AI 抽屉 -->
    <el-drawer
      v-model="aiStore.dockVisible"
      title="AI 助手"
      direction="rtl"
      :size="isMobile ? '100%' : '460px'"
      class="ai-dock-drawer"
    >
      <div class="ai-dock">
        <div v-if="agentRegistry.length > 1" class="agent-tabs">
          <div
            v-for="agent in agentRegistry"
            :key="agent.id"
            class="agent-tab"
            :class="{ active: aiStore.activeAgentId === agent.id }"
            @click="aiStore.activeAgentId = agent.id"
          >
            <el-icon><component :is="agent.icon" /></el-icon>
            <span>{{ agent.name }}</span>
          </div>
        </div>

        <div class="agent-panel">
          <component :is="activeAgent.component" />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { MagicStick } from '@element-plus/icons-vue'
  import { storeToRefs } from 'pinia'
  import { useAiStore } from '@/pinia/modules/ai'
  import { useAppStore } from '@/pinia'
  import { agentRegistry } from './agents/registry'

  const aiStore = useAiStore()
  const appStore = useAppStore()
  const { activeAgentId } = storeToRefs(aiStore)

  const isMobile = computed(() => appStore.device === 'mobile')
  const anyAgentAvailable = computed(() => agentRegistry.length > 0)

  const activeAgent = computed(
    () =>
      agentRegistry.find((agent) => agent.id === activeAgentId.value) ||
      agentRegistry[0]
  )

  const openDock = () => {
    aiStore.openDock()
  }
</script>

<style scoped lang="scss">
.ai-entry {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 100%;
  padding: 0 12px;
  cursor: pointer;
  color: #64748b;

  &:hover {
    color: var(--el-color-primary);
  }
}

.ai-entry-text {
  font-weight: 600;
  font-size: 14px;
}

.ai-dock {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}

.agent-tabs {
  display: flex;
  gap: 8px;
}

.agent-tab {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 8px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  cursor: pointer;
  color: #64748b;
  font-size: 13px;
  transition: all 0.2s;

  &:hover {
    border-color: var(--el-color-primary-light-5);
  }

  &.active {
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }
}

.agent-panel {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
</style>
