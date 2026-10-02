<template>
  <div class="admin-page admin-page--cards state-page">
    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="page-title">
        <el-icon class="title-icon"><Monitor /></el-icon>
        <div>
          <h2>服务器状态</h2>
          <p class="subtitle">
            每 10 秒自动刷新
            <span v-if="lastUpdated" class="update-time">
              · 上次更新 {{ lastUpdated }}
            </span>
          </p>
        </div>
      </div>
      <el-button :loading="refreshing" @click="handleManualRefresh">
        <el-icon v-if="!refreshing" class="mr-1"><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- Runtime 概览卡片 -->
    <div v-if="state.os" class="runtime-card">
      <div class="runtime-item" v-for="item in runtimeItems" :key="item.label">
        <div class="runtime-label">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </div>
        <div class="runtime-value">{{ item.value }}</div>
      </div>
    </div>

    <el-row :gutter="16" class="card-row">
      <!-- CPU -->
      <el-col :xs="24" :md="12">
        <div v-if="state.cpu" class="stat-card">
          <div class="card-header">
            <div class="card-title">
              <el-icon><Cpu /></el-icon>
              <span>CPU</span>
            </div>
            <el-tag size="small" type="info" effect="plain">
              {{ state.cpu.cores }} 核
            </el-tag>
          </div>
          <div class="card-body scroll-area">
            <div
              v-for="(item, index) in state.cpu.cpus"
              :key="index"
              class="core-row"
            >
              <span class="core-label">核心 {{ index }}</span>
              <el-progress
                class="core-progress"
                type="line"
                :percentage="+item.toFixed(0)"
                :color="colors"
                :stroke-width="8"
              />
            </div>
          </div>
        </div>
      </el-col>

      <!-- 内存 -->
      <el-col :xs="24" :md="12">
        <div v-if="state.ram" class="stat-card">
          <div class="card-header">
            <div class="card-title">
              <el-icon><MagicStick /></el-icon>
              <span>内存</span>
            </div>
            <el-tag
              size="small"
              effect="plain"
              :type="ramTagType(state.ram.usedPercent)"
            >
              {{ state.ram.usedPercent.toFixed(1) }}%
            </el-tag>
          </div>
          <div class="card-body ram-body">
            <el-progress
              type="dashboard"
              :percentage="state.ram.usedPercent"
              :color="colors"
              :width="120"
            >
              <template #default="{ percentage }">
                <span class="progress-inner">{{ percentage }}%</span>
              </template>
            </el-progress>
            <div class="ram-info">
              <div class="info-row">
                <span class="info-label">总内存</span>
                <span class="info-value">{{
                  formatGb(state.ram.totalMb)
                }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">已使用</span>
                <span class="info-value">{{
                  formatGb(state.ram.usedMb)
                }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">剩余</span>
                <span class="info-value">{{
                  formatGb(state.ram.totalMb - state.ram.usedMb)
                }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="card-row">
      <!-- 磁盘 -->
      <el-col :span="24">
        <div v-if="state.disk" class="stat-card">
          <div class="card-header">
            <div class="card-title">
              <el-icon><Files /></el-icon>
              <span>磁盘</span>
            </div>
            <el-tag size="small" type="info" effect="plain">
              {{ state.disk.length }} 个挂载点
            </el-tag>
          </div>
          <div class="disk-grid">
            <div
              v-for="(item, index) in state.disk"
              :key="index"
              class="disk-item"
            >
              <div class="disk-head">
                <span class="disk-mount">{{ item.mountPoint }}</span>
                <span class="disk-percent" :style="{ color: diskColor(item.usedPercent) }">
                  {{ item.usedPercent.toFixed(1) }}%
                </span>
              </div>
              <el-progress
                :percentage="item.usedPercent"
                :color="colors"
                :stroke-width="8"
                :show-text="false"
              />
              <div class="disk-detail">
                <span>已用 {{ formatGb(item.usedMb) }}</span>
                <span>共 {{ formatGb(item.totalMb) }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
  import { computed, onUnmounted, ref } from 'vue'
  import {
    Cpu,
    Files,
    MagicStick,
    Memo,
    Monitor,
    Refresh,
    Setting
  } from '@element-plus/icons-vue'
  import { getSystemState } from '@/api/system'

  defineOptions({
    name: 'State'
  })

  const timer = ref(null)
  const state = ref({})
  const lastUpdated = ref('')
  const refreshing = ref(false)

  const colors = ref([
    { color: '#5cb87a', percentage: 20 },
    { color: '#e6a23c', percentage: 40 },
    { color: '#f56c6c', percentage: 80 }
  ])

  const runtimeItems = computed(() => {
    const os = state.value.os
    if (!os) return []
    return [
      { label: '操作系统', value: os.goos, icon: Monitor },
      { label: 'CPU 数量', value: os.numCpu, icon: Cpu },
      { label: '编译器', value: os.compiler, icon: Setting },
      { label: 'Go 版本', value: os.goVersion, icon: Memo },
      { label: 'Goroutine 数', value: os.numGoroutine, icon: MagicStick }
    ]
  })

  const formatGb = (mb) => {
    if (mb == null || isNaN(mb)) return '-'
    const gb = mb / 1024
    return gb >= 10 ? `${gb.toFixed(1)} GB` : `${gb.toFixed(2)} GB`
  }

  const ramTagType = (percent) => {
    if (percent >= 80) return 'danger'
    if (percent >= 40) return 'warning'
    return 'success'
  }

  const diskColor = (percent) => {
    if (percent >= 80) return '#f56c6c'
    if (percent >= 40) return '#e6a23c'
    return '#5cb87a'
  }

  const reload = async () => {
    try {
      refreshing.value = true
      const { data } = await getSystemState()
      state.value = data.server
      lastUpdated.value = new Date().toLocaleTimeString()
    } finally {
      refreshing.value = false
    }
  }

  const handleManualRefresh = () => {
    reload()
  }

  reload()
  timer.value = setInterval(() => {
    reload()
  }, 1000 * 10)

  onUnmounted(() => {
    clearInterval(timer.value)
    timer.value = null
  })
</script>

<style scoped>
  .state-page {
    @apply p-1;
  }

  .page-header {
    @apply flex items-center justify-between mb-4 rounded-lg bg-white dark:bg-slate-800 px-5 py-4;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }

  .page-title {
    @apply flex items-center gap-3;
  }

  .title-icon {
    @apply text-2xl text-blue-500;
  }

  .page-title h2 {
    @apply m-0 text-lg font-semibold text-slate-800 dark:text-slate-100;
  }

  .subtitle {
    @apply m-0 mt-0.5 text-xs text-slate-400 dark:text-slate-500;
  }

  .update-time {
    @apply text-slate-500 dark:text-slate-400;
  }

  /* Runtime 概览条 */
  .runtime-card {
    @apply grid grid-cols-2 md:grid-cols-5 gap-3 mb-4;
  }

  .runtime-item {
    @apply rounded-lg bg-white dark:bg-slate-800 px-4 py-3;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .runtime-item:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .runtime-label {
    @apply flex items-center gap-1.5 text-xs text-slate-400 dark:text-slate-500;
  }

  .runtime-value {
    @apply mt-1.5 text-base font-medium text-slate-700 dark:text-slate-200 truncate;
  }

  /* 统计卡片 */
  .card-row {
    @apply mb-4;
  }

  .stat-card {
    @apply h-full rounded-lg bg-white dark:bg-slate-800 p-5;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }

  .card-header {
    @apply flex items-center justify-between mb-4 pb-3 border-b border-slate-100 dark:border-slate-700;
  }

  .card-title {
    @apply flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-200;
  }

  .card-title .el-icon {
    @apply text-blue-500;
  }

  .card-body {
    @apply text-sm;
  }

  .scroll-area {
    @apply overflow-y-auto pr-1;
    max-height: 240px;
  }

  .scroll-area::-webkit-scrollbar {
    width: 4px;
  }

  .scroll-area::-webkit-scrollbar-thumb {
    @apply bg-slate-200 dark:bg-slate-600 rounded;
  }

  /* CPU 核心行 */
  .core-row {
    @apply flex items-center gap-3 mb-3;
  }

  .core-label {
    @apply shrink-0 w-14 text-xs text-slate-500 dark:text-slate-400;
  }

  .core-progress {
    @apply flex-1;
  }

  /* 内存 */
  .ram-body {
    @apply flex items-center gap-6 flex-wrap;
  }

  .progress-inner {
    @apply text-lg font-semibold text-slate-700 dark:text-slate-200;
  }

  .ram-info {
    @apply flex-1 min-w-[160px];
  }

  .info-row {
    @apply flex items-center justify-between py-2 border-b border-dashed border-slate-100 dark:border-slate-700 last:border-0;
  }

  .info-label {
    @apply text-xs text-slate-400 dark:text-slate-500;
  }

  .info-value {
    @apply text-sm font-medium text-slate-700 dark:text-slate-200;
  }

  /* 磁盘网格 */
  .disk-grid {
    @apply grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4;
  }

  .disk-item {
    @apply rounded-lg p-4 bg-slate-50 dark:bg-slate-700/40 border border-slate-100 dark:border-slate-700;
  }

  .disk-head {
    @apply flex items-center justify-between mb-2;
  }

  .disk-mount {
    @apply text-sm font-medium text-slate-700 dark:text-slate-200 truncate;
  }

  .disk-percent {
    @apply text-sm font-semibold shrink-0 ml-2;
  }

  .disk-detail {
    @apply flex items-center justify-between mt-2 text-xs text-slate-400 dark:text-slate-500;
  }
</style>
