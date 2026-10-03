<template>
  <div class="admin-page admin-page--cards dashboard-page">
    <section class="dashboard-welcome">
      <div>
        <p class="dashboard-date">{{ today }} · 内容管理</p>
        <h1>欢迎回来</h1>
        <p>回顾内容积累，继续下一篇创作。</p>
      </div>
      <div class="dashboard-actions">
        <el-button :loading="loading" @click="loadData">刷新数据</el-button>
        <el-button v-if="canOpen('/layout/gblog/edit')" type="primary" :icon="EditPen" @click="open('/layout/gblog/edit')">写文章</el-button>
      </div>
    </section>

    <el-alert v-if="summaryError" :title="summaryError" type="warning" :closable="false" show-icon class="dashboard-alert" />
    <div class="dashboard-stats" aria-label="内容概览">
      <div v-for="item in statCards" :key="item.label" class="dashboard-stat">
        <span class="stat-icon"><el-icon :size="22"><component :is="item.icon" /></el-icon></span>
        <div><p>{{ item.label }}</p><strong>{{ item.value }}</strong></div>
      </div>
    </div>

    <div class="dashboard-grid">
      <section class="dashboard-panel recent-panel">
        <div class="panel-heading">
          <div><h2>最近创建的文章</h2><p>按创建时间排列，展示最近 5 篇。</p></div>
          <el-button v-if="canOpen('/layout/gblog/list')" link type="primary" @click="open('/layout/gblog/list')">全部文章 <el-icon><ArrowRight /></el-icon></el-button>
        </div>
        <el-alert v-if="articlesError" :title="articlesError" type="warning" :closable="false" show-icon />
        <p v-else-if="loading && !articles.length" class="panel-empty">正在加载文章…</p>
        <el-empty v-else-if="!articles.length" description="还没有文章，开始你的第一篇创作吧" :image-size="70" />
        <ul v-else class="article-list">
          <li v-for="article in articles" :key="article.id">
            <span class="article-icon"><el-icon><Document /></el-icon></span>
            <div class="article-info">
              <button v-if="canOpen(`/layout/gblog/edit/${article.id}`)" type="button" class="article-title" @click="open(`/layout/gblog/edit/${article.id}`)">{{ article.title || '未命名文章' }}</button>
              <span v-else class="article-title">{{ article.title || '未命名文章' }}</span>
              <p>{{ article.category?.categoryName || '未分类' }} · {{ formatDate(article.createTime) }}</p>
            </div>
            <el-tag :type="article.published ? 'success' : 'info'" size="small" effect="plain">{{ article.published ? (article.password ? '密码保护' : '公开') : '私密' }}</el-tag>
          </li>
        </ul>
      </section>
      <section class="dashboard-panel">
        <div class="panel-heading"><div><h2>分类分布</h2><p>文章数量最多的 6 个分类</p></div></div>
        <p v-if="!summary" class="panel-empty">{{ loading ? '正在加载分类…' : '分类统计暂不可用' }}</p>
        <p v-else-if="!topCategories.length" class="panel-empty">暂无分类</p>
        <ul v-else class="category-list">
          <li v-for="item in topCategories" :key="item.name">
            <div><span>{{ item.name }}</span><strong>{{ item.value }} 篇</strong></div>
            <div class="category-track"><span :style="{ width: categoryWidth(item.value) }"></span></div>
          </li>
        </ul>
      </section>
    </div>

    <section class="dashboard-panel tag-panel">
      <div class="panel-heading"><div><h2>标签概览</h2><p>使用最多的 12 个标签，数字表示关联文章数。</p></div></div>
      <p v-if="!summary" class="panel-empty">{{ loading ? '正在加载标签…' : '标签统计暂不可用' }}</p>
      <p v-else-if="!topTags.length" class="panel-empty">暂无标签</p>
      <div v-else class="tag-list"><el-tag v-for="item in topTags" :key="item.name" effect="plain">{{ item.name }} <span>{{ item.value }}</span></el-tag></div>
    </section>
    <section class="dashboard-panel" aria-labelledby="dashboard-shortcuts-title">
      <div class="panel-heading"><div><h2 id="dashboard-shortcuts-title">管理入口</h2><p>常用的系统管理功能</p></div></div>
      <GvaQuickLink />
    </section>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ArrowRight, ChatDotRound, Collection, Document, EditPen, PriceTag } from '@element-plus/icons-vue'
  import { getDashboard } from '@/api/blog/dashboard'
  import { getArticlePage } from '@/api/blog/article'
  import GvaQuickLink from './components/quickLinks.vue'

  defineOptions({ name: 'Dashboard' })
  const router = useRouter()
  const summary = ref(null)
  const articles = ref([])
  const loading = ref(false)
  const summaryError = ref('')
  const articlesError = ref('')
  const today = new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
  const canOpen = path => router.resolve(path).matched.some(route => !route.path.includes(':catchAll') && !route.path.includes(':pathMatch'))
  const open = path => router.push(path)
  const number = value => value == null ? '—' : Number(value).toLocaleString('zh-CN')
  const categoryItems = computed(() => summary.value?.category?.series || [])
  const tagItems = computed(() => summary.value?.tag?.series || [])
  const sorted = items => [...items].sort((a, b) => b.value - a.value)
  const topCategories = computed(() => sorted(categoryItems.value).slice(0, 6))
  const topTags = computed(() => sorted(tagItems.value).slice(0, 12))
  const categoryWidth = value => `${Math.max(0, Math.min(100, value / Math.max(1, ...categoryItems.value.map(item => item.value)) * 100))}%`
  const statCards = computed(() => [
    { label: '文章总数', value: number(summary.value?.blogCount), icon: Document },
    { label: '评论总数', value: number(summary.value?.commentCount), icon: ChatDotRound },
    { label: '文章分类', value: summary.value ? number(categoryItems.value.length) : '—', icon: Collection },
    { label: '文章标签', value: summary.value ? number(tagItems.value.length) : '—', icon: PriceTag }
  ])
  function formatDate(value) {
    const date = new Date(value)
    return value && Number.isFinite(date.getTime()) ? date.toLocaleDateString('zh-CN') : '日期未知'
  }
  async function loadData() {
    if (loading.value) return
    loading.value = true
    const results = await Promise.allSettled([getDashboard(), getArticlePage({ page: 1, pageSize: 5 })])
    const [stats, recent] = results
    if (stats.status === 'fulfilled' && stats.value.code === 0 && stats.value.data) {
      summary.value = stats.value.data
      summaryError.value = ''
    } else {
      summary.value = null
      summaryError.value = '内容概览暂时无法加载，请检查访问权限或稍后刷新。'
    }
    if (recent.status === 'fulfilled' && recent.value.code === 0 && Array.isArray(recent.value.data?.list)) {
      articles.value = recent.value.data.list
      articlesError.value = ''
    } else {
      articles.value = []
      articlesError.value = '最近文章暂时无法加载，请检查访问权限或稍后刷新。'
    }
    loading.value = false
  }
  onMounted(loadData)
</script>

<style scoped lang="scss">
.dashboard-page { padding: 20px; color: var(--admin-text, var(--el-text-color-primary)); }
.dashboard-welcome, .dashboard-panel, .dashboard-stat {
  border: 1px solid var(--admin-border, var(--el-border-color-lighter));
  border-radius: 10px;
  background: var(--admin-surface, var(--el-bg-color));
}
.dashboard-welcome { padding: 24px; margin-bottom: 16px; display: flex; align-items: center; justify-content: space-between; gap: 20px;
  h1 { margin: 8px 0; font-size: 24px; font-weight: 600; }
  p { margin: 0; font-size: 14px; color: var(--admin-muted, var(--el-text-color-secondary)); }
  .dashboard-date { font-size: 12px; }
}
.dashboard-actions { display: flex; flex-wrap: wrap; gap: 8px; .el-button + .el-button { margin-left: 0; } }
.dashboard-alert { margin-bottom: 16px; }
.dashboard-stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; margin-bottom: 16px; }
.dashboard-stat { display: flex; align-items: center; gap: 16px; padding: 20px;
  p { margin: 0 0 6px; font-size: 13px; color: var(--admin-muted, var(--el-text-color-secondary)); }
  strong { font-size: 26px; line-height: 1.2; font-weight: 600; font-variant-numeric: tabular-nums; }
}
.stat-icon, .article-icon { display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; border-radius: 8px; color: var(--el-color-primary); background: var(--admin-selected, var(--el-color-primary-light-9)); }
.stat-icon { width: 44px; height: 44px; }
.dashboard-grid { display: grid; grid-template-columns: minmax(0, 1.7fr) minmax(0, 1fr); gap: 16px; margin-bottom: 16px; .dashboard-panel { margin-bottom: 0; } }
.dashboard-panel { padding: 24px; margin-bottom: 16px; min-width: 0; }
.panel-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 20px;
  h2 { margin: 0; font-size: 16px; font-weight: 600; }
  p { margin: 6px 0 0; font-size: 12px; line-height: 1.6; color: var(--admin-muted, var(--el-text-color-secondary)); }
}
.panel-empty { padding: 24px 0; font-size: 13px; color: var(--admin-muted, var(--el-text-color-secondary)); }
.article-list, .category-list { padding: 0; margin: 0; list-style: none; }
.article-list li { display: flex; align-items: center; gap: 12px; padding: 15px 0; border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter)); &:last-child { border-bottom: 0; } }
.article-icon { width: 36px; height: 36px; }
.article-info { flex: 1; min-width: 0; p { margin: 6px 0 0; font-size: 12px; color: var(--admin-muted, var(--el-text-color-secondary)); } }
.article-title { display: block; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding: 0; border: 0; background: transparent; text-align: left; color: inherit; font: inherit; font-size: 14px; }
button.article-title { cursor: pointer; &:hover { color: var(--el-color-primary); } &:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 4px; } }
.category-list li { margin-bottom: 20px; > div:first-child { display: flex; justify-content: space-between; gap: 12px; font-size: 13px; margin-bottom: 8px; } strong { white-space: nowrap; font-weight: 500; color: var(--admin-muted, var(--el-text-color-secondary)); } }
.category-track { height: 6px; border-radius: 4px; overflow: hidden; background: var(--el-fill-color-light); span { display: block; height: 100%; border-radius: inherit; background: var(--el-color-primary); opacity: .75; } }
.tag-list { display: flex; flex-wrap: wrap; gap: 10px; span { margin-left: 8px; opacity: .65; } }
@media (max-width: 1000px) { .dashboard-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); } .dashboard-grid { grid-template-columns: 1fr; } }
@media (max-width: 640px) { .dashboard-page { padding: 12px; } .dashboard-welcome, .dashboard-panel { padding: 18px; } .dashboard-welcome { flex-direction: column; align-items: flex-start; } .dashboard-stat { padding: 14px; gap: 10px; } .stat-icon { width: 32px; height: 32px; } .dashboard-stat strong { font-size: 22px; } }
</style>
