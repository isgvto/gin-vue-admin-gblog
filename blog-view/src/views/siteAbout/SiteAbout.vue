<template>
	<div class="site-about-page">
		<div class="ui top attached segment m-padded-lr-big site-about-header">
			<div>
				<div class="ui teal label">GitHub</div>
				<h2 class="m-text-500">关于本站</h2>
				<p>展示 LZMclear/gin-vue-admin-gblog main 分支的提交记录。</p>
			</div>
			<a class="ui basic teal button" :href="repositoryUrl" target="_blank" rel="noopener noreferrer">
				<i class="github icon"></i>
				查看仓库
			</a>
		</div>

		<div class="ui bottom attached segment site-about-content">
			<div v-if="loading && commits.length === 0" class="ui active centered inline loader"></div>
			<div v-else-if="commits.length === 0" class="empty-state">
				<i class="github icon"></i>
				<span>暂无提交记录</span>
			</div>

			<template v-else>
				<el-timeline class="commit-timeline">
					<el-timeline-item
						v-for="item in commits"
						:key="item.sha"
						:timestamp="formatDate(item.date)"
						placement="top"
					>
						<div class="commit-card">
							<div class="commit-card-header">
								<a class="commit-message" :href="item.htmlUrl" target="_blank" rel="noopener noreferrer">
									{{ item.message }}
								</a>
								<a class="commit-sha" :href="item.htmlUrl" target="_blank" rel="noopener noreferrer">
									{{ item.shortSha }}
								</a>
							</div>
							<div class="commit-meta">
								<span>
									<i class="user outline icon"></i>
									{{ item.author }}
								</span>
								<span>
									<i class="calendar alternate outline icon"></i>
									{{ item.date | dateFromNow }}
								</span>
							</div>
						</div>
					</el-timeline-item>
				</el-timeline>

				<div class="load-more-wrap">
					<el-button
						type="primary"
						plain
						:loading="loading"
						:disabled="finished"
						@click="loadMore"
					>
						{{ finished ? '没有更多了' : '加载更多' }}
					</el-button>
				</div>
			</template>
		</div>
	</div>
</template>

<script>
	import {getRepositoryCommits} from '@/api/github'

	export default {
		name: 'SiteAbout',
		data() {
			return {
				repositoryUrl: 'https://github.com/LZMclear/gin-vue-admin-gblog',
				commits: [],
				page: 1,
				perPage: 20,
				loading: false,
				finished: false
			}
		},
		created() {
			this.loadCommits()
		},
		methods: {
			loadMore() {
				if (this.loading || this.finished) {
					return
				}
				this.page += 1
				this.loadCommits()
			},
			loadCommits() {
				this.loading = true
				getRepositoryCommits(this.page, this.perPage).then(res => {
					const list = Array.isArray(res) ? res : []
					const normalized = list.map(item => {
						const commit = item.commit || {}
						const author = commit.author || {}
						const message = (commit.message || '').split('\n')[0]
						return {
							sha: item.sha,
							shortSha: item.sha ? item.sha.slice(0, 7) : '',
							htmlUrl: item.html_url,
							author: author.name === 'key' ? 'Gvto' : (author.name || 'Gvto'),
							date: author.date || '',
							message: message || 'No commit message'
						}
					})

					this.commits = this.commits.concat(normalized)
					this.finished = normalized.length < this.perPage
				}).catch(() => {
					this.msgError('提交记录加载失败')
					if (this.page > 1) {
						this.page -= 1
					}
				}).finally(() => {
					this.loading = false
				})
			},
			formatDate(date) {
				if (!date) {
					return ''
				}
				const value = new Date(date)
				const year = value.getFullYear()
				const month = String(value.getMonth() + 1).padStart(2, '0')
				const day = String(value.getDate()).padStart(2, '0')
				return `${year}-${month}-${day}`
			}
		}
	}
</script>

<style scoped>
	.site-about-page {
		border-radius: 8px;
		overflow: hidden;
		box-shadow: 0 8px 24px rgba(15, 23, 42, .08);
	}

	.site-about-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 18px;
		background: #fff !important;
	}

	.site-about-header h2 {
		margin: 14px 0 8px;
		color: #1f2937;
	}

	.site-about-header p {
		margin: 0;
		color: #64748b;
		line-height: 1.7;
	}

	.site-about-content {
		min-height: 520px;
		padding: 28px 30px 30px !important;
		background: #fff !important;
	}

	.commit-timeline {
		padding-left: 4px;
	}

	.empty-state {
		display: flex;
		min-height: 260px;
		align-items: center;
		justify-content: center;
		gap: 10px;
		color: #94a3b8;
		font-size: 16px;
		font-weight: 600;
	}

	.empty-state i {
		margin: 0 !important;
		font-size: 24px;
	}

	.commit-card {
		padding: 16px 18px;
		border: 1px solid #eef2f7;
		border-radius: 8px;
		background: #fff;
		box-shadow: 0 6px 18px rgba(15, 23, 42, .06);
	}

	.commit-card-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 14px;
	}

	.commit-message {
		flex: 1;
		color: #1f2937;
		font-size: 16px;
		font-weight: 700;
		line-height: 1.5;
		word-break: break-word;
	}

	.commit-message:hover {
		color: #00a7e0;
	}

	.commit-sha {
		flex: 0 0 auto;
		padding: 3px 8px;
		border-radius: 6px;
		background: #f1f5f9;
		color: #475569;
		font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
		font-size: 12px;
		line-height: 1.5;
	}

	.commit-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 10px 18px;
		margin-top: 12px;
		color: #64748b;
		font-size: 13px;
	}

	.commit-meta span {
		display: inline-flex;
		align-items: center;
		height: 20px;
		line-height: 20px;
	}

	.commit-meta i {
		display: inline-flex !important;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		margin-right: 5px !important;
		line-height: 1 !important;
	}

	.load-more-wrap {
		display: flex;
		justify-content: center;
		margin-top: 18px;
	}

	@media screen and (max-width: 767px) {
		.site-about-header {
			display: block;
			padding-left: 1.2em !important;
			padding-right: 1.2em !important;
		}

		.site-about-header .button {
			margin-top: 16px;
		}

		.site-about-content {
			padding: 20px 16px 24px !important;
		}

		.commit-card-header {
			display: block;
		}

		.commit-sha {
			display: inline-block;
			margin-top: 10px;
		}
	}
</style>
