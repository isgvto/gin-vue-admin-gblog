<template>
  <section class="database-status" aria-labelledby="database-status-title">
    <div class="database-heading"><div><h3 id="database-status-title">数据库状态 <span>{{ database.type?.toUpperCase() || '未配置' }}</span></h3><p>当前后台连接的数据库 · 每 10 秒检测</p></div><el-tag :type="database.healthy ? 'success' : 'danger'">{{ database.healthy ? '连接正常' : '连接异常' }}</el-tag></div>
    <div class="health-row"><span>{{ database.message }}</span><span>连接检测耗时 <strong>{{ number(database.latencyMs, 2) }} ms</strong></span></div>
    <div class="database-sections">
      <div v-if="database.pool" class="metric-section"><h4>应用连接池 <small>仅本项目，非 MySQL 全局连接</small></h4>
        <dl class="metric-grid"><div v-for="metric in poolMetrics" :key="metric.label"><dt>{{ metric.label }}</dt><dd>{{ metric.value }}</dd></div></dl>
        <p class="metric-note">等待次数和时长为当前后台进程的累计值。</p>
      </div>
      <div v-if="database.mysql" class="metric-section"><h4>MySQL 基本指标 <small>数据库实例全局统计</small></h4>
        <el-alert v-if="database.mysql.warning" :title="database.mysql.warning" type="warning" :closable="false" show-icon />
        <dl class="metric-grid"><div v-for="metric in mysqlMetrics" :key="metric.label"><dt>{{ metric.label }}</dt><dd>{{ metric.value }}</dd></div></dl>
        <p class="metric-note">QPS 按连续采样计算，包含监控查询；查询数和慢查询数为累计值，重启或计数重置后重新采样。</p>
      </div>
    </div>
  </section>
</template>
<script setup>
import { computed } from 'vue'
const props = defineProps({ database: { type: Object, required: true } })
const number = (value, digits = 0) => value == null
  ? '—'
  : Number(value).toLocaleString('zh-CN', { maximumFractionDigits: digits })

function duration(seconds) {
  if (seconds == null) return '—'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor(seconds % 86400 / 3600)
  const minutes = Math.floor(seconds % 3600 / 60)
  return days ? `${days} 天 ${hours} 小时` : `${hours} 小时 ${minutes} 分钟`
}

const poolMetrics = computed(() => {
  const p = props.database.pool
  if (!p) return []
  return [
    { label: '已打开连接', value: number(p.openConnections) },
    { label: '正在使用', value: number(p.inUse) },
    { label: '空闲连接', value: number(p.idle) },
    { label: '连接池上限', value: p.maxOpenConnections === 0 ? '不限制' : number(p.maxOpenConnections) },
    { label: '累计等待次数', value: number(p.waitCount) },
    { label: '累计等待时长', value: `${number(p.waitDurationMs, 2)} ms` }
  ]
})

const mysqlMetrics = computed(() => {
  const m = props.database.mysql
  if (!m) return []
  return [
    { label: '版本', value: m.version || '—' },
    { label: '运行时长', value: duration(m.uptime) },
    { label: '当前连接 / 上限', value: `${number(m.threadsConnected)} / ${number(m.maxConnections)}` },
    { label: '正在执行的线程', value: number(m.threadsRunning) },
    { label: '每秒查询（QPS）', value: m.qps == null && m.questions != null && m.uptime != null ? '等待下一次采样' : number(m.qps, 2) },
    { label: '累计查询数', value: number(m.questions) },
    { label: '累计慢查询数', value: number(m.slowQueries) },
    { label: '全局只读设置', value: m.readOnly == null ? '—' : m.readOnly ? '开启' : '关闭' }
  ]
})
</script>
<style scoped>
.database-status { padding:20px; margin-bottom:16px; background:var(--admin-surface,var(--el-bg-color)); border:1px solid var(--admin-border,var(--el-border-color-lighter)); border-radius:10px; color:var(--admin-text,var(--el-text-color-primary)); }
.database-heading { display:flex; justify-content:space-between; align-items:center; gap:16px; }
h3 { margin:0; font-size:16px; font-weight:600; } h3 span { font-size:12px; color:var(--el-text-color-secondary); margin-left:8px; }
p { margin:6px 0 0; font-size:12px; color:var(--el-text-color-secondary); line-height:1.7; }
.health-row { display:flex; justify-content:space-between; flex-wrap:wrap; gap:8px; margin:16px 0; font-size:13px; color:var(--el-text-color-secondary); }
.health-row strong { color:var(--admin-text,var(--el-text-color-primary)); font-weight:500; }
.database-sections { display:grid; grid-template-columns:1fr 1fr; gap:24px; }
h4 { margin:0 0 14px; font-size:14px; font-weight:600; } small { display:block; font-size:12px; color:var(--el-text-color-secondary); margin-top:5px; font-weight:400; }
.metric-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:16px 24px; margin:16px 0; }
dt { font-size:12px; color:var(--el-text-color-secondary); margin-bottom:6px; } dd { margin:0; font-size:16px; font-variant-numeric:tabular-nums; overflow-wrap:anywhere; }
@media(max-width:800px) { .database-sections { grid-template-columns:1fr; gap:16px; } }
</style>
