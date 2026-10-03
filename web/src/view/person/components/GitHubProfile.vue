<template>
  <div class="github-panel">
    <section class="github-card">
      <div class="github-heading"><div><h2>GitHub 开源名片</h2><p>展示公开账号的数据，不代表已验证账号所有权。</p></div><el-button :loading="loading" :disabled="saving" @click="load">刷新</el-button></div>
      <form class="github-binding" @submit.prevent="save">
        <el-input v-model="username" aria-label="GitHub 用户名" maxlength="39" placeholder="GitHub 用户名，例如 octocat" :disabled="loading || saving" />
        <el-button type="primary" native-type="submit" :loading="saving" :disabled="loading">保存账号</el-button>
        <el-button :disabled="loading || saving || !boundUsername" @click="unbind">解除展示</el-button>
      </form>
      <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
      <p v-if="loading" class="github-note">正在读取 GitHub 数据…</p>
      <el-empty v-else-if="!boundUsername && !error" description="填写 GitHub 用户名，加载你的开源名片" :image-size="60" />
      <div v-if="data?.profile" class="github-identity">
        <img v-if="avatarURL" :src="avatarURL" alt="GitHub 头像" referrerpolicy="no-referrer" width="64" height="64" />
        <div><a :href="profileURL" target="_blank" rel="noopener noreferrer">{{ data.profile.name || data.profile.login }} <span>@{{ data.profile.login }}</span></a><p>{{ data.profile.bio || '尚未填写 GitHub 简介' }}</p><small v-if="data.profile.location || data.profile.company">{{ [data.profile.location, data.profile.company].filter(Boolean).join(' · ') }}</small></div>
      </div>
      <div v-if="data?.profile" class="github-stats">
        <div><strong>{{ data.profile.public_repos }}</strong><span>公开仓库</span></div>
        <div><strong>{{ data.profile.followers }}</strong><span>关注者</span></div>
        <div><strong>{{ data.profile.following }}</strong><span>正在关注</span></div>
        <div><strong>{{ date(data.profile.created_at) }}</strong><span>加入 GitHub</span></div>
      </div>
      <el-alert v-for="warning in data?.warnings || []" :key="warning" :title="warning" type="warning" :closable="false" show-icon class="github-warning" />
      <p v-if="data?.fetchedAt" class="github-note">数据获取时间：{{ dateTime(data.fetchedAt) }} · 正常数据缓存 10 分钟，部分失败缓存 2 分钟。</p>
    </section>

    <section v-if="data?.profile" class="github-card">
      <div class="github-heading"><div><h2>近期公开提交</h2><p>最近 5 条公开提交，搜索仅覆盖默认分支，可能有索引延迟。</p></div></div>
      <p v-if="!data.commits?.length" class="github-note">{{ warningFor('近期公开提交') ? '提交信息暂不可用，请查看上方提示。' : '暂无可读取的公开提交。' }}</p>
      <ul v-else class="github-records"><li v-for="commit in data.commits" :key="`${commit.repository}:${commit.sha}`"><a :href="`${repoURL(commit.repository)}/commit/${encodeURIComponent(commit.sha)}`" target="_blank" rel="noopener noreferrer">{{ commit.message || '代码提交' }}</a><p>{{ commit.repository }} · {{ commit.sha.slice(0, 7) }} · {{ dateTime(commit.date) }}</p></li></ul>
    </section>
    <section v-if="data?.profile" class="github-card">
      <div class="github-heading"><div><h2>公开仓库</h2><p>按更新时间排列，展示最近 12 个公开仓库。</p></div><a :href="`${profileURL}?tab=repositories`" target="_blank" rel="noopener noreferrer">全部仓库 ↗</a></div>
      <p v-if="!data.repositories?.length" class="github-note">{{ warningFor('公开仓库') ? '仓库列表暂不可用，请查看上方提示。' : '暂无公开仓库。' }}</p>
      <div v-else class="github-repositories"><article v-for="repo in data.repositories" :key="repo.full_name"><a :href="repoURL(repo.full_name)" target="_blank" rel="noopener noreferrer">{{ repo.name }}</a><el-tag v-if="repo.archived" size="small" type="info">已归档</el-tag><el-tag v-else-if="repo.fork" size="small" type="info">Fork</el-tag><p>{{ repo.description || '暂无描述' }}</p><footer><span>{{ repo.language || '未标注语言' }}</span><span>☆ {{ repo.stargazers_count }} · Fork {{ repo.forks_count }}</span></footer></article></div>
    </section>
    <section v-if="data?.profile" class="github-card">
      <div class="github-heading"><div><h2>近期公开动态</h2><p>展示近 30 天内最近 10 条公开事件，可能延迟。</p></div></div>
      <p v-if="!data.events?.length" class="github-note">{{ warningFor('近期公开动态') ? '动态暂不可用，请查看上方提示。' : '近期暂无公开动态。' }}</p>
      <ul v-else class="github-records"><li v-for="event in data.events" :key="event.id"><strong>{{ eventType(event.type) }}</strong><a :href="repoURL(event.repository)" target="_blank" rel="noopener noreferrer">{{ event.repository }}</a><p>{{ dateTime(event.createdAt) }}</p></li></ul>
    </section>
  </div>
</template>
<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/pinia/modules/user'
import { getGitHubProfile, setGitHubProfile } from '@/api/githubProfile'
const userStore = useUserStore()
const username = ref(userStore.userInfo.githubUsername || '')
const boundUsername = ref(username.value)
const data = ref(null)
const error = ref('')
const loading = ref(false)
const saving = ref(false)
let requestVersion = 0
const profileURL = computed(() => `https://github.com/${encodeURIComponent(data.value?.profile?.login || boundUsername.value)}`)
const avatarURL = computed(() => {
  try { const url = new URL(data.value?.profile?.avatar_url); return url.protocol === 'https:' && url.hostname === 'avatars.githubusercontent.com' ? url.href : '' } catch { return '' }
})
const repoURL = name => `https://github.com/${String(name).split('/').map(encodeURIComponent).join('/')}`
const date = value => { const item = new Date(value); return value && Number.isFinite(item.getTime()) ? item.toLocaleDateString('zh-CN') : '—' }
const dateTime = value => { const item = new Date(value); return value && Number.isFinite(item.getTime()) ? item.toLocaleString('zh-CN') : '—' }
const warningFor = prefix => data.value?.warnings?.some(item => item.startsWith(prefix))
const eventType = value => ({PushEvent:'推送代码', PullRequestEvent:'Pull Request', IssuesEvent:'Issue', IssueCommentEvent:'发表评论', CreateEvent:'创建资源', DeleteEvent:'删除分支或标签', WatchEvent:'Star 仓库', ForkEvent:'Fork 仓库', ReleaseEvent:'发布版本', PullRequestReviewEvent:'代码评审'}[value] || value)
async function load() {
  const version = ++requestVersion
  loading.value = true; error.value = ''; data.value = null
  try {
    const res = await getGitHubProfile()
    if (version !== requestVersion) return
    if (res.code !== 0) throw new Error(res.msg || 'GitHub 数据加载失败')
    data.value = res.data
    boundUsername.value = res.data.username || ''
    username.value = boundUsername.value
  } catch (err) { if (version === requestVersion) error.value = err.message || 'GitHub 数据加载失败，请稍后重试' }
  finally { if (version === requestVersion) loading.value = false }
}
async function saveAccount(value) {
  if (saving.value || loading.value) return
  if (value && (!/^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$/.test(value) || value.includes('--'))) { ElMessage.warning('请输入有效的 GitHub 用户名，不需要填写网址'); return }
  saving.value = true
  try {
    const res = await setGitHubProfile(value)
    if (res.code !== 0) throw new Error(res.msg || '保存失败')
    boundUsername.value = value; username.value = value
    userStore.ResetUserInfo({ githubUsername: value })
    ElMessage.success(value ? '已保存 GitHub 展示账号' : '已解除 GitHub 展示')
    await load()
  } catch (err) { error.value = err.message || '保存失败' }
  finally { saving.value = false }
}
const save = () => saveAccount(username.value.trim())
const unbind = () => saveAccount('')
onMounted(load)
onBeforeUnmount(() => requestVersion++)
</script>
<style scoped>
.github-panel { display: flex; flex-direction: column; gap: 20px; }
.github-card { padding: 24px; border: 1px solid var(--admin-border, var(--el-border-color-lighter)); border-radius: 10px; background: var(--admin-surface, var(--el-bg-color)); color: var(--admin-text, var(--el-text-color-primary)); min-width: 0; }
.github-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
h2 { margin: 0; font-size: 16px; font-weight: 600; }
.github-heading p, .github-note, .github-identity p, .github-identity small { font-size: 12px; line-height: 1.7; color: var(--admin-muted, var(--el-text-color-secondary)); margin: 6px 0 0; }
a { color: var(--el-color-primary); text-decoration: none; overflow-wrap: anywhere; } a:hover { text-decoration: underline; }
.github-binding { display: flex; gap: 8px; margin: 20px 0; } .github-binding .el-input { flex: 1; min-width: 120px; } .github-binding .el-button + .el-button { margin-left: 0; }
.github-identity { display: flex; align-items: center; gap: 16px; margin: 24px 0; } .github-identity img { border-radius: 50%; } .github-identity a { font-size: 18px; font-weight: 600; } .github-identity a span { margin-left: 8px; font-size: 13px; font-weight: 400; }
.github-stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin: 20px 0; } .github-stats div { padding: 16px 8px; text-align: center; background: var(--el-fill-color-light); border-radius: 8px; } .github-stats strong { display: block; font-size: 18px; } .github-stats span { display: block; margin-top: 8px; font-size: 12px; color: var(--admin-muted, var(--el-text-color-secondary)); }
.github-warning { margin-top: 12px; }
.github-records { padding: 0; margin: 18px 0 0; list-style: none; } .github-records li { padding: 14px 0; border-bottom: 1px solid var(--admin-border, var(--el-border-color-lighter)); } .github-records li:last-child { border-bottom: 0; } .github-records p { font-size: 12px; color: var(--el-text-color-secondary); margin: 7px 0 0; } .github-records strong { font-size: 13px; margin-right: 10px; } .github-records a { font-size: 14px; }
.github-repositories { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 12px; margin-top: 20px; } .github-repositories article { padding: 16px; border: 1px solid var(--admin-border, var(--el-border-color-lighter)); border-radius: 8px; } .github-repositories .el-tag { margin-left: 8px; } .github-repositories p { font-size: 13px; color: var(--el-text-color-secondary); line-height: 1.7; min-height: 42px; overflow-wrap: anywhere; } .github-repositories footer { display: flex; justify-content: space-between; gap: 8px; flex-wrap: wrap; font-size: 12px; color: var(--el-text-color-secondary); }
@media (max-width: 700px) { .github-card { padding: 18px; } .github-binding { flex-wrap: wrap; } .github-binding .el-input { flex-basis: 100%; } .github-stats { grid-template-columns: repeat(2,minmax(0,1fr)); } .github-repositories { grid-template-columns: 1fr; } }
</style>
