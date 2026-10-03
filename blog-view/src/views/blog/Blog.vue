<template>
	<div>
		<div class="ui padded attached segment m-padded-tb-large blog-detail-card" v-loading="loading">
			<div class="ui large red right corner label" v-if="blog.top">
				<i class="arrow alternate circle up icon"></i>
			</div>
			<header class="blog-header">
				<div class="blog-header-top">
					<div class="blog-header-left">
						<button class="back-button" type="button" @click="goBack" title="返回">
							<i class="home icon"></i>
						</button>
						<div class="header-tags" v-if="blog.tags && blog.tags.length">
							<span class="breadcrumb-separator">/</span>
							<template v-for="(tag, index) in blog.tags">
								<router-link
									:key="tag.id || tag.tagName"
									:to="`/tag/${tag.tagName}`"
									class="header-tag"
									:class="tag.color"
								>{{ tag.tagName }}</router-link>
								<span class="breadcrumb-separator" :key="`${tag.id || tag.tagName}-separator`" v-if="index < blog.tags.length - 1">/</span>
							</template>
						</div>
					</div>
					<router-link :to="`/category/${blog.category.categoryName}`" class="header-category" v-if="blog.category">
						{{ blog.category.categoryName }}
					</router-link>
				</div>
				<h1 class="blog-title">{{ blog.title }}</h1>
				<div class="blog-meta">
					<span><i class="user outline icon"></i>作者：{{ authorName }}</span>
					<span><i class="calendar outline icon"></i>{{ blog.createTime | dateFormat('YYYY-MM-DD') }}</span>
					<span><i class="eye icon"></i>{{ blog.views || 0 }}</span>
					<span><i class="book icon"></i>字数 {{ blog.words || 0 }}</span>
					<span><i class="clock outline icon"></i>阅读时长 {{ blog.readTime || 1 }} 分钟</span>
					<button class="meta-action" type="button" @click.prevent="bigFontSize=!bigFontSize" title="切换字体大小">
						<i class="font icon"></i>
					</button>
					<button class="meta-action" type="button" @click.prevent="changeFocusMode" title="专注模式">
						<i class="book reader icon"></i>
					</button>
				</div>
			</header>
			<div class="ui middle aligned mobile reversed stackable">
				<div class="ui grid m-margin-lr">
					<!--文章Markdown正文-->
					<div class="typo js-toc-content m-padded-tb-small match-braces rainbow-braces" v-viewer v-mermaid="blog.content" :class="{'m-big-fontsize':bigFontSize}" v-html="blog.content"></div>
					<!--赞赏-->
					<div style="margin: 2em auto">
						<el-popover placement="top" width="220" trigger="click" v-if="blog.appreciation">
							<div class="ui orange basic label" style="width: 100%">
								<div class="image">
									<div style="font-size: 12px;text-align: center;margin-bottom: 5px;">一毛是鼓励</div>
									<img :src="$store.state.siteInfo.reward" alt="" class="ui rounded bordered image" loading="lazy" decoding="async" style="width: 100%">
									<div style="font-size: 12px;text-align: center;margin-top: 5px;">一块是真爱</div>
								</div>
							</div>
							<el-button slot="reference" class="ui orange inverted circular button m-text-500">赞赏</el-button>
						</el-popover>
					</div>
				</div>
			</div>
		</div>
		<!--评论-->
		<div class="ui bottom teal attached segment threaded comments" v-if="!loading && blog.id">
			<CommentList :page="0" :blogId="blogId" v-if="blog.commentEnabled"/>
			<h3 class="ui header" v-else>评论已关闭</h3>
		</div>
	</div>
</template>

<script>
	import {getBlogById} from "@/api/blog";
	import CommentList from "@/components/comment/CommentList";
	import {mapState} from "vuex";
	import {
		SET_BLOG_PASSWORD_DIALOG_VISIBLE,
		SET_BLOG_PASSWORD_FORM,
		SET_FOCUS_MODE,
		SET_IS_BLOG_RENDER_COMPLETE
	} from '@/store/mutations-types';
	import {isSuccess, normalizeBlog} from "@/util/gvaResponse";

	export default {
		name: "Blog",
		components: {CommentList},
		data() {
			return {
				blog: this.emptyBlog(),
				bigFontSize: false,
				loading: false,
				blogRequestSeq: 0,
			}
		},
		computed: {
			blogId() {
				return parseInt(this.$route.params.id)
			},
			...mapState(['siteInfo', 'focusMode', 'introduction']),
			authorName() {
				return this.introduction && this.introduction.name ? this.introduction.name : 'Gvto'
			}
		},
		beforeRouteEnter(to, from, next) {
			//路由到博客文章页面之前，应将文章的渲染完成状态置为 false
			next(vm => {
				// 当 beforeRouteEnter 钩子执行前，组件实例尚未创建
				// vm 就是当前组件的实例，可以在 next 方法中把 vm 当做 this用
				vm.$store.commit(SET_IS_BLOG_RENDER_COMPLETE, false)
			})
		},
		beforeRouteLeave(to, from, next) {
			this.$store.commit(SET_FOCUS_MODE, false)
			// 从文章页面路由到其它页面时，销毁当前组件的同时，要销毁tocbot实例
			// 否则tocbot一直在监听页面滚动事件，而文章页面的锚点已经不存在了，会报"Uncaught TypeError: Cannot read property 'className' of null"
			if (window.tocbot && typeof window.tocbot.destroy === 'function') {
				window.tocbot.destroy()
			}
			next()
		},
		beforeRouteUpdate(to, from, next) {
			// 一般有两种情况会触发这个钩子
			// ①当前文章页面跳转到其它文章页面
			// ②点击目录跳转锚点时，路由hash值会改变，导致当前页面会重新加载，这种情况是不希望出现的
			// 在路由 beforeRouteUpdate 中判断路径是否改变
			// 如果跳转到其它页面，to.path!==from.path 就放行 next()
			// 如果是跳转锚点，path不会改变，hash会改变，to.path===from.path, to.hash!==from.path 不放行路由跳转，就能让锚点正常跳转
			if (to.path !== from.path) {
				this.$store.commit(SET_FOCUS_MODE, false)
				//在当前组件内路由到其它博客文章时，要重新获取文章
				this.getBlog(to.params.id)
				//只要路由路径有改变，且停留在当前Blog组件内，就把文章的渲染完成状态置为 false
				this.$store.commit(SET_IS_BLOG_RENDER_COMPLETE, false)
				next()
			} else {
				next(false)
			}
		},
		created() {
			this.getBlog()
		},
		mounted() {
			window.addEventListener('blog-password-verified', this.handleBlogPasswordVerified)
		},
		beforeDestroy() {
			window.removeEventListener('blog-password-verified', this.handleBlogPasswordVerified)
		},
		methods: {
			emptyBlog() {
				return {
					id: 0,
					title: '',
					content: '',
					tags: [],
					category: null,
					commentEnabled: false,
				}
			},
			goBack() {
				if (window.history.length > 1) {
					this.$router.back()
					return
				}
				this.$router.push({name: 'home'})
			},
			getBlog(id = this.blogId) {
				const requestSeq = ++this.blogRequestSeq
				this.loading = true
				this.blog = this.emptyBlog()
				this.$store.commit(SET_IS_BLOG_RENDER_COMPLETE, false)
				if (window.tocbot && typeof window.tocbot.destroy === 'function') {
					window.tocbot.destroy()
				}
				//密码保护的文章，需要发送密码验证通过后保存在localStorage的Token
				const blogToken = window.localStorage.getItem(`blog${id}`)
				//如果有则发送博主身份Token
				const adminToken = window.localStorage.getItem('adminToken')
				const token = adminToken ? adminToken : (blogToken ? blogToken : '')
				getBlogById(token, id).then(res => {
					if (requestSeq !== this.blogRequestSeq) {
						return
					}
					if (isSuccess(res)) {
						this.blog = normalizeBlog(res.data)
						document.title = this.blog.title + this.siteInfo.webTitleSuffix
						//v-html渲染完毕后，渲染代码块样式
						this.$nextTick(() => {
							if (window.Prism && typeof window.Prism.highlightAll === 'function') {
								window.Prism.highlightAll()
							}
							//将文章渲染完成状态置为 true
							this.$store.commit(SET_IS_BLOG_RENDER_COMPLETE, true)
						})
					} else {
						if (this.shouldRequestBlogPassword(res)) {
							this.openBlogPasswordDialog(id)
							return
						}
						this.msgError(res.msg)
					}
				}).catch(() => {
					if (requestSeq !== this.blogRequestSeq) {
						return
					}
					this.msgError("请求失败")
				}).finally(() => {
					if (requestSeq === this.blogRequestSeq) {
						this.loading = false
					}
				})
			},
			changeFocusMode() {
				this.$store.commit(SET_FOCUS_MODE, !this.focusMode)
			},
			shouldRequestBlogPassword(res) {
				return res && res.code !== 0 && res.msg === '文章受密码保护'
			},
			openBlogPasswordDialog(id) {
				window.localStorage.removeItem(`blog${id}`)
				this.blog = this.emptyBlog()
				this.$store.commit(SET_BLOG_PASSWORD_FORM, {blogId: id, password: ''})
				this.$store.commit(SET_BLOG_PASSWORD_DIALOG_VISIBLE, true)
			},
			handleBlogPasswordVerified(event) {
				const blogId = event && event.detail ? event.detail.blogId : null
				if (String(blogId) === String(this.blogId)) {
					this.getBlog(blogId)
				}
			}
		}
	}
</script>

<style scoped>
	.blog-detail-card {
		position: relative;
		overflow: hidden;
		border-radius: 8px 8px 0 0 !important;
		background: #fff !important;
	}

	.blog-header {
		margin: 0 1rem 24px;
		padding: 0 0 24px;
		border-bottom: 1px solid #e5e7eb;
		background: #fff;
	}

	.blog-header-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 18px;
	}

	.blog-header-left {
		display: flex;
		min-width: 0;
		align-items: center;
		gap: 8px;
		color: #6b7280;
		font-size: 13px;
		font-weight: 500;
	}

	.back-button,
	.meta-action {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: 0;
		background: transparent;
		cursor: pointer;
		font-family: inherit;
	}

	.back-button {
		width: 18px;
		height: 18px;
		padding: 0;
		border-radius: 6px;
		color: #64748b;
		font-size: 14px;
		line-height: 1;
		transition: background-color .2s ease, color .2s ease;
	}

	.back-button:hover {
		background: transparent;
		color: #2563eb;
	}

	.back-button i,
	.meta-action i,
	.blog-meta i {
		display: inline-flex;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		margin: 0 !important;
		line-height: 1 !important;
	}

	.header-tags {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}

	.breadcrumb-separator {
		color: #c0c7d2;
	}

	.header-tag,
	.header-category {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: #64748b;
		font-size: 13px;
		font-weight: 500;
		line-height: 1;
	}

	.header-tag:hover {
		color: #1d4ed8;
	}

	.header-category {
		flex: 0 0 auto;
		min-height: 30px;
		padding: 7px 12px;
		border-radius: 6px;
		background: #eff6ff;
		color: #2563eb;
		font-weight: 700;
	}

	.header-category:hover {
		background: #dbeafe;
		color: #1d4ed8;
	}

	.blog-title {
		margin: 0;
		color: #0f172a;
		font-size: 34px;
		font-weight: 800;
		letter-spacing: 0;
		line-height: 1.25;
	}

	.blog-detail-card .ui.grid.m-margin-lr {
		margin-right: 0 !important;
		margin-left: 0 !important;
	}

	.blog-detail-card .typo {
		width: 100%;
	}

	.blog-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 14px;
		margin-top: 20px;
		color: #6b7280;
		font-size: 14px;
		font-weight: 400;
	}

	.blog-meta span {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		line-height: 1;
	}

	.meta-action {
		width: 28px;
		height: 28px;
		border-radius: 6px;
		color: #64748b;
	}

	.meta-action:hover {
		background: #dbeafe;
		color: #2563eb;
	}

	.el-divider {
		margin: 1rem 0 !important;
	}

	h1::before, h2::before, h3::before, h4::before, h5::before, h6::before {
		display: block;
		content: " ";
		height: 55px;
		margin-top: -55px;
		visibility: hidden;
	}

	@media only screen and (max-width: 760px) {
		.blog-header {
			margin: 0 1rem 20px;
			padding: 0 0 20px;
		}

		.blog-header-top,
		.blog-header-left {
			align-items: flex-start;
			flex-direction: column;
		}

		.header-category {
			align-self: flex-start;
		}

		.blog-title {
			font-size: 28px;
		}
	}
</style>
