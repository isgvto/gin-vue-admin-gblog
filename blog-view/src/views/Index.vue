<template>
	<div class="site" :class="{'docs-layout': isDocsPage}">
		<!--顶部导航-->
		<Nav :blogName="siteInfo.blogName" :categoryList="categoryList"/>
		<!--首页大图 只在首页且pc端时显示-->
		<div class="m-mobile-hide">
			<Header v-if="$route.name==='home'"/>
		</div>

		<div class="main">
			<div class="m-padded-tb-big">
				<div class="ui container">
					<router-view v-if="isDocsPage"/>
					<div v-else class="ui stackable grid">
						<!--左侧-->
						<div class="three wide column m-mobile-hide">
							<Introduction :class="[{'m-display-none':focusMode}, {'blog-left-sticky': $route.name==='blog'}]"/>
						</div>
						<!--中间-->
						<div class="ten wide column">
							<keep-alive include="Home">
								<router-view/>
							</keep-alive>
						</div>
						<!--右侧-->
						<div class="three wide column m-mobile-hide">
							<RandomBlog :randomBlogList="randomBlogList" :class="{'m-display-none':focusMode}"/>
							<Tags :tagList="tagList" :class="{'m-display-none':focusMode}"/>
							<!--只在文章页面显示目录-->
							<Tocbot v-if="$route.name==='blog'"/>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!--私密文章密码对话框-->
		<BlogPasswordDialog/>

		<!--回到顶部-->
		<el-backtop style="box-shadow: none;background: none;z-index: 9999;">
			<img src="/img/paper-plane.png" loading="lazy" decoding="async" style="width: 40px;height: 40px;">
		</el-backtop>
		<!--底部footer-->
		<Footer :siteInfo="siteInfo" :badges="badges" :newBlogList="newBlogList" :hitokoto="hitokoto"/>
	</div>
</template>

<script>
	import {getHitokoto, getSite} from '@/api/index'
	import Nav from "@/components/index/Nav";
	import Header from "@/components/index/Header";
	import Footer from "@/components/index/Footer";
	import Introduction from "@/components/sidebar/Introduction";
	import Tags from "@/components/sidebar/Tags";
	import RandomBlog from "@/components/sidebar/RandomBlog";
	import Tocbot from "@/components/sidebar/Tocbot";
	import BlogPasswordDialog from "@/components/index/BlogPasswordDialog";
	import {mapState} from 'vuex'
	import {SAVE_CLIENT_SIZE, SAVE_INTRODUCTION, SAVE_SITE_INFO, RESTORE_COMMENT_FORM} from "@/store/mutations-types";
	import {isSuccess, normalizeSite} from "@/util/gvaResponse";

	export default {
		name: "Index",
		components: {Header, BlogPasswordDialog, Tocbot, RandomBlog, Tags, Nav, Footer, Introduction},
		data() {
			return {
				siteInfo: {
					blogName: '',
					webTitleSuffix: ''
				},
				categoryList: [],
				tagList: [],
				randomBlogList: [],
				badges: [],
				newBlogList: [],
				siteStats: {
					articleCount: 0,
					categoryCount: 0,
					tagCount: 0
				},
				hitokoto: {},
				handleResize: null,
			}
		},
		computed: {
			...mapState(['focusMode']),
			isDocsPage() {
				return this.$route.name === 'docs'
			}
		},
		watch: {
			//路由改变时，页面滚动至顶部
			'$route.path'() {
				this.scrollToTop()
			}
		},
		created() {
			this.getSite()
			this.getHitokoto()
			//从localStorage恢复之前的评论信息
			this.$store.commit(RESTORE_COMMENT_FORM)
		},
		mounted() {
			//保存可视窗口大小
			this.$store.commit(SAVE_CLIENT_SIZE, {clientHeight: document.body.clientHeight, clientWidth: document.body.clientWidth})
			this.handleResize = () => {
				this.$store.commit(SAVE_CLIENT_SIZE, {clientHeight: document.body.clientHeight, clientWidth: document.body.clientWidth})
			}
			window.addEventListener('resize', this.handleResize)
		},
		beforeDestroy() {
			if (this.handleResize) {
				window.removeEventListener('resize', this.handleResize)
			}
		},
		methods: {
			getSite() {
				return getSite().then(res => {
					if (isSuccess(res)) {
						const site = normalizeSite(res.data)
						this.siteInfo = site.siteInfo
						this.badges = site.badges
						this.newBlogList = site.newBlogList
						this.categoryList = site.categoryList
						this.tagList = site.tagList
						this.randomBlogList = site.randomBlogList
						this.siteStats = site.siteStats
						this.$store.commit(SAVE_SITE_INFO, this.siteInfo)
						this.$store.commit(SAVE_INTRODUCTION, site.introduction)
						document.title = this.$route.meta.title + this.siteInfo.webTitleSuffix
					} else {
						this.msgError(res.msg || '站点信息加载失败，请刷新重试')
					}
				}).catch(error => {
					this.msgError(error.code === 'ECONNABORTED'
						? '站点信息加载超时，请稍后刷新重试'
						: '站点信息加载失败，请检查网络后重试')
				})
			},
			//获取一言
			getHitokoto() {
				return getHitokoto().then(res => {
					this.hitokoto = res
				}).catch(() => {
					// 一言为可选的外部服务，不让失败阻断页面阅读。
					this.hitokoto = {}
				})
			}
		}
	}
</script>

<style scoped>
	.site {
		display: flex;
		min-height: 100vh; /* 没有元素时，也把页面撑开至100% */
		flex-direction: column;
	}

	.main {
		margin-top: 40px;
		flex: 1;
	}

	.main .ui.container {
		width: clamp(960px, 75vw, 1800px) !important;
		margin-left: auto !important;
		margin-right: auto !important;
	}

	.ui.grid .three.column {
		padding: 0;
	}

	.ui.grid .ten.column {
		padding-top: 0;
	}

	.m-display-none {
		display: none !important;
	}

	.blog-left-sticky {
		position: sticky;
		top: 60px;
		z-index: 10;
	}
	.docs-layout { background: #f6f8fb; }
	.docs-layout .main .ui.container {
		width: calc(100% - 40px) !important;
		max-width: 1440px;
	}
	@media (max-width: 768px) {
		.docs-layout .main .ui.container { width: calc(100% - 24px) !important; }
	}
</style>
