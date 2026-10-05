<template>
	<div class="article-list">
		<article class="article-item" v-for="item in blogList" :key="item.id">
			<div class="top-badge" v-if="item.top"><i class="arrow alternate circle up icon"></i></div>
			<div class="article-cover">
				<img v-if="getCoverUrl(item)" :src="getCoverUrl(item)" alt="" loading="lazy" decoding="async">
				<div v-else class="article-cover-placeholder">{{ getTitleInitial(item.title) }}</div>
			</div>

			<div class="article-main">
				<div class="article-category" v-if="item.category">
					<router-link :to="`/category/${item.category.categoryName}`">{{ item.category.categoryName }}</router-link>
				</div>

				<h3 class="article-title">
					<a href="" @click.prevent="toBlog(item)">{{ item.title }}</a>
				</h3>

				<div class="article-meta">
					<span><i class="calendar outline icon"></i>{{ item.createTime | dateFormat('YYYY-MM-DD') }}</span>
					<span><i class="eye icon"></i>{{ item.views || 0 }}</span>
					<span><i class="book icon"></i>字数 {{ formatCount(item.words) }}</span>
					<span><i class="clock outline icon"></i>阅读时长 {{ item.readTime || 1 }} 分钟</span>
				</div>

				<p class="article-preview">{{ item.privacy ? '当前文章已加密' : (item.preview || '暂无预览内容') }}</p>

				<div class="article-bottom">
					<div class="article-tags">
						<router-link
							v-for="tag in item.tags"
							:key="tag.id || tag.tagName"
							:to="`/tag/${tag.tagName}`"
						>{{ tag.tagName }}</router-link>
					</div>
					<a href="" class="read-more" @click.prevent="toBlog(item)">
						阅读全文 <i class="arrow right icon"></i>
					</a>
				</div>
			</div>
		</article>
	</div>
</template>

<script>
	export default {
		name: "BlogItem",
		props: {
			blogList: {
				type: Array,
				required: true
			}
		},
		methods: {
			toBlog(blog) {
				this.$store.dispatch('goBlogPage', blog)
			},
			formatCount(value) {
				const num = Number(value) || 0
				if (num >= 1000) {
					const short = Math.round(num / 100) / 10
					return `${short}k`
				}
				return String(num)
			},
			getTitleInitial(title) {
				return (title || '文').trim().slice(0, 1)
			},
			getCoverUrl(item) {
				const html = item && item.description
				if (!html) {
					return ''
				}
				const match = String(html).match(/<img\b[^>]*\bsrc=["']([^"']+)["'][^>]*>/i)
				return match ? match[1] : ''
			}
		}
	}
</script>

<style scoped>
	.article-list {
		display: grid;
		gap: 16px;
	}

	.article-item {
		position: relative;
		display: grid;
		grid-template-columns: 220px minmax(0, 1fr);
		gap: 22px;
		padding: 16px;
		overflow: hidden;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		background: #fff;
		box-shadow: 0 2px 8px rgba(15, 23, 42, .025);
		transition: border-color .2s ease, box-shadow .2s ease, transform .2s ease;
	}

	.article-item:hover {
		border-color: var(--blog-accent-border-strong);
		box-shadow: 0 10px 26px rgba(50, 111, 168, .09);
		transform: translateY(-1px);
	}

	.article-cover {
		position: relative;
		display: block;
		overflow: hidden;
		min-height: 174px;
		border-radius: 8px;
		background: #eef2f7;
	}

	.article-cover img {
		display: block;
		width: 100%;
		height: 100%;
		min-height: 174px;
		object-fit: cover;
		transition: transform .25s ease;
	}

	.article-item:hover .article-cover img {
		transform: scale(1.04);
	}

	.article-cover-placeholder {
		display: flex;
		width: 100%;
		height: 100%;
		min-height: 174px;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, var(--blog-accent-soft-hover), var(--blog-accent-soft));
		color: var(--blog-accent);
		font-size: 40px;
		font-weight: 700;
	}

	.top-badge {
		position: absolute;
		top: 0;
		right: 0;
		z-index: 2;
		width: 54px;
		height: 54px;
		overflow: hidden;
		border-top-right-radius: 8px;
		color: #fff;
	}

	.top-badge::before {
		position: absolute;
		top: 0;
		right: 0;
		width: 0;
		height: 0;
		border-top: 54px solid #ef4444;
		border-left: 54px solid transparent;
		content: '';
	}

	.top-badge i {
		position: absolute;
		top: 9px;
		right: 8px;
		z-index: 1;
		display: flex;
		width: 18px;
		height: 18px;
		align-items: center;
		justify-content: center;
		margin: 0 !important;
		font-size: 17px;
		line-height: 1 !important;
	}

	.article-main {
		display: flex;
		min-width: 0;
		flex-direction: column;
	}

	.article-category a {
		display: inline-flex;
		padding: 5px 10px;
		border-radius: 999px;
		background: var(--blog-accent-soft);
		color: var(--blog-accent);
		font-size: 13px;
		font-weight: 600;
	}

	.article-category a:hover,
	.article-category a:focus-visible {
		background: var(--blog-accent-soft-hover);
		color: var(--blog-accent-hover);
	}

	.article-title {
		margin: 14px 0 10px;
		font-size: 24px;
		line-height: 1.28;
		letter-spacing: 0;
	}

	.article-title a {
		color: #111827;
	}

	.article-title a:hover {
		color: var(--blog-accent);
	}

	.article-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
		margin-bottom: 10px;
		color: #667085;
		font-size: 14px;
	}

	.article-meta span {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		line-height: 1;
	}

	.article-meta i {
		display: inline-flex;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		margin: 0 !important;
		line-height: 1 !important;
		vertical-align: middle;
	}

	.article-preview {
		display: -webkit-box;
		margin: 0 0 14px;
		overflow: hidden;
		color: #4b5563;
		font-size: 15px;
		line-height: 1.65;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
	}

	.article-bottom {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 14px;
		margin-top: auto;
	}

	.article-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.article-tags a {
		padding: 6px 12px;
		border-radius: 999px;
		background: var(--blog-accent-soft);
		color: var(--blog-accent);
		font-size: 13px;
	}

	.article-tags a:hover,
	.article-tags a:focus-visible {
		background: var(--blog-accent-soft-hover);
		color: var(--blog-accent-hover);
	}

	.read-more {
		display: inline-flex;
		flex: 0 0 auto;
		align-items: center;
		justify-content: center;
		gap: 4px;
		padding: 7px 12px;
		border: 1px solid var(--blog-accent-border-strong);
		border-radius: 999px;
		color: var(--blog-accent);
		font-size: 14px;
		font-weight: 700;
	}

	.read-more:hover,
	.read-more:focus-visible {
		background: var(--blog-accent-soft);
		color: var(--blog-accent-hover);
	}

	.read-more i {
		display: inline-flex;
		width: 14px;
		height: 14px;
		align-items: center;
		justify-content: center;
		margin: 0 !important;
		line-height: 1 !important;
		vertical-align: middle;
	}

	@media only screen and (max-width: 760px) {
		.article-item {
			grid-template-columns: 1fr;
		}

		.article-cover,
		.article-cover img,
		.article-cover-placeholder {
			min-height: 176px;
		}

		.article-bottom {
			align-items: flex-start;
			flex-direction: column;
		}
	}
</style>
