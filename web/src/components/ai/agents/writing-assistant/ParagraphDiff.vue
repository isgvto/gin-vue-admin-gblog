<template>
  <div class="paragraph-diff">
    <div class="diff-toolbar">
      <el-radio-group v-model="viewMode" size="small">
        <el-radio-button value="inline">内联</el-radio-button>
        <el-radio-button value="split">分栏</el-radio-button>
      </el-radio-group>
      <div class="diff-actions">
        <el-button size="small" :disabled="disabled" @click="setAll(true)">采用剩余 AI 修改</el-button>
        <el-button size="small" :disabled="disabled" @click="setAll(false)">保留剩余原文</el-button>
      </div>
    </div>
    <div class="diff-summary">剩余 {{ stats.pending }} 处待处理 · 已处理 {{ stats.total - stats.pending }}/{{ stats.total }} 处</div>
    <div v-if="blocks.some(block => block.fallback)" class="diff-summary">部分内容的段落或结构发生变化，已合并为一组供你选择。</div>

    <div class="diff-list">
      <div
        v-for="(block, index) in blocks"
        :key="index"
        class="diff-block"
        :class="[ `is-${block.type}`, { 'is-reviewed': block.reviewed } ]"
      >
        <div class="block-tag">
          {{ block.type !== 'equal' && block.reviewed ? (block.takeRevised ? '已采用 AI' : '已保留原文') : typeLabel(block.type) }}
        </div>

        <template v-if="block.type === 'equal' || block.reviewed">
          <pre class="block-text">{{ block.takeRevised ? block.revised : block.original }}</pre>
        </template>

        <template v-else-if="block.type === 'modified' && viewMode === 'split'">
          <div class="split-pane">
            <pre class="block-text original">{{ block.original }}</pre>
            <pre class="block-text revised">{{ block.revised }}</pre>
          </div>
        </template>

        <template v-else-if="viewMode === 'split' && (block.type === 'added' || block.type === 'removed')">
          <div class="split-pane">
            <pre class="block-text original">{{ block.original || '（无）' }}</pre>
            <pre class="block-text revised">{{ block.revised || '（已删除）' }}</pre>
          </div>
        </template>

        <template v-else>
          <pre v-if="block.type !== 'added'" class="block-text original">{{ block.original }}</pre>
          <pre v-if="block.type !== 'removed'" class="block-text revised">{{ block.revised }}</pre>
        </template>

        <div v-if="block.type !== 'equal' && !block.reviewed" class="block-actions">
          <el-button
            size="small"
            type="primary"
            :disabled="disabled"
            @click="choose(index, true)"
          >
            {{ block.type === 'added' ? '采用' : '采用 AI 版' }}
          </el-button>
          <el-button
            size="small"
            :disabled="disabled"
            @click="choose(index, false)"
          >
            {{ block.type === 'removed' ? '保留' : '保留原文' }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'

  const props = defineProps({
    blocks: {
      type: Array,
      required: true
    },
    disabled: Boolean
  })

  const emit = defineEmits(['change'])

  const viewMode = ref('inline')

  const stats = computed(() => {
    const selectable = props.blocks.filter(
      (b) => b.type === 'modified' || b.type === 'added' || b.type === 'removed'
    )
    return {
      total: selectable.length,
      pending: selectable.filter((b) => !b.reviewed).length
    }
  })

  const typeLabel = (type) => {
    const labels = {
      equal: '未改动',
      modified: '修改',
      added: '新增',
      removed: '删除'
    }
    return labels[type] || type
  }

  const choose = (index, takeRevised) => {
    emit('change', { index, takeRevised })
  }

  const setAll = (takeRevised) => {
    emit('change', { takeRevised })
  }
</script>

<style scoped lang="scss">
.paragraph-diff {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.diff-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.diff-actions {
  display: flex;
  gap: 8px;
}

.diff-summary {
  margin: 6px 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.diff-list {
  flex: 1;
  min-height: 0;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.diff-block {
  position: relative;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: var(--admin-surface, var(--el-bg-color));
}

.diff-block.is-modified {
  border-color: var(--el-color-warning);
}

.diff-block.is-added {
  border-color: var(--el-color-success);
}

.diff-block.is-removed {
  border-color: var(--el-color-danger);
}

.block-tag {
  display: inline-block;
  margin-bottom: 6px;
  padding: 1px 8px;
  border-radius: 4px;
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.is-modified .block-tag {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
}

.is-added .block-tag {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.is-removed .block-tag {
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.block-text {
  margin: 0;
  padding: 6px 8px;
  border-radius: 4px;
  background: var(--el-fill-color-lighter);
  color: var(--admin-text, var(--el-text-color-primary));
  font-family: inherit;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.block-text.original {
  background: var(--el-color-danger-light-9);
}

.block-text.revised {
  background: var(--el-color-success-light-9);
}

.diff-block.is-reviewed {
  border-color: var(--el-border-color-lighter);
}

.is-reviewed .block-tag {
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
}

.split-pane {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.block-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

@media (max-width: 768px) {
  .split-pane {
    grid-template-columns: 1fr;
  }
}
</style>
