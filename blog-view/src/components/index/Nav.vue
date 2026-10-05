<template>
	<div ref="nav" class="ui fixed inverted stackable pointing menu blog-nav" :class="{'nav-open': !mobileHide}">
		<div class="ui container">
			<router-link to="/" class="blog-nav-brand">
				<h3 class="ui header item m-blue">{{ blogName }}</h3>
			</router-link>
			<router-link to="/home" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='home'}">
				<i class="home icon"></i>首页
			</router-link>
			<el-dropdown trigger="click" :class="{'m-mobile-hide': mobileHide}" @command="categoryRoute">
				<span class="el-dropdown-link item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='category'}">
					<i class="idea icon"></i>分类<i class="caret down icon"></i>
				</span>
				<el-dropdown-menu slot="dropdown" class="blog-nav-dropdown">
					<el-dropdown-item :command="category.categoryName" v-for="(category,index) in categoryList" :key="index">{{ category.categoryName }}</el-dropdown-item>
				</el-dropdown-menu>
			</el-dropdown>
			<router-link to="/archives" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='archives'}">
				<i class="clone icon"></i>归档
			</router-link>
			<router-link to="/moments" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='moments'}">
				<i class="comment alternate outline icon"></i>动态
			</router-link>
			<router-link to="/friends" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='friends'}">
				<i class="users icon"></i>友人帐
			</router-link>
			<router-link to="/about" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='about'}">
				<i class="info icon"></i>关于我
			</router-link>
			<router-link to="/docs" class="item" :class="{'m-mobile-hide': mobileHide,'active':$route.name==='docs'}">
				<i class="book icon"></i>文档站
			</router-link>
			<el-autocomplete v-model="queryString" :fetch-suggestions="debounceQuery" placeholder="Search..."
			                 class="right item m-search" :class="{'m-mobile-hide': mobileHide}"
			                 popper-class="m-search-item" @select="handleSelect">
				<i class="search icon el-input__icon" slot="suffix"></i>
				<template slot-scope="{ item }">
					<div class="title">{{ item.title }}</div>
					<span class="content">{{ item.content }}</span>
				</template>
			</el-autocomplete>
			<button class="ui menu icon button m-right-top m-mobile-show" aria-label="切换导航菜单" :aria-expanded="mobileHide ? 'false' : 'true'" @click="toggle">
				<i class="sidebar icon"></i>
			</button>
		</div>
	</div>
</template>

<script>
	import {getSearchBlogList} from "@/api/blog";
	import {isSuccess} from "@/util/gvaResponse";

	export default {
		name: "Nav",
		props: {
			blogName: {
				type: String,
				required: true
			},
			categoryList: {
				type: Array,
				required: true
			},
		},
		data() {
			return {
				mobileHide: true,
				queryString: '',
				queryResult: [],
				timer: null,
				handleDocumentClick: null
			}
		},
		watch: {
			//路由改变时，收起导航栏
			'$route.path'() {
				this.mobileHide = true
			}
		},
		mounted() {
			//监听点击事件，收起导航菜单
			this.handleDocumentClick = (e) => {
				//遍历冒泡
				let flag = this.$refs.nav.contains(e.target)
				//如果导航栏是打开状态，且点击的元素不是Nav的子元素，则收起菜单
				if (!this.mobileHide && !flag) {
					this.mobileHide = true
				}
			}
			document.addEventListener('click', this.handleDocumentClick)
		},
		beforeDestroy() {
			if (this.handleDocumentClick) {
				document.removeEventListener('click', this.handleDocumentClick)
			}
			if (this.timer) {
				clearTimeout(this.timer)
			}
		},
		methods: {
			toggle() {
				this.mobileHide = !this.mobileHide
			},
			categoryRoute(name) {
				this.$router.push({name: 'category', params: {name}}).catch(err => {
					if (err.name !== 'NavigationDuplicated') {
						throw err
					}
				})
			},
			debounceQuery(queryString, callback) {
				this.timer && clearTimeout(this.timer)
				this.timer = setTimeout(() => this.querySearchAsync(queryString, callback), 1000)
			},
			querySearchAsync(queryString, callback) {
				if (queryString == null
						|| queryString.trim() === ''
						|| queryString.indexOf('%') !== -1
						|| queryString.indexOf('_') !== -1
						|| queryString.indexOf('[') !== -1
						|| queryString.indexOf('#') !== -1
						|| queryString.indexOf('*') !== -1
						|| queryString.trim().length > 20) {
					return
				}
				getSearchBlogList(queryString).then(res => {
					if (isSuccess(res)) {
						this.queryResult = res.data
						if (this.queryResult.length === 0) {
							this.queryResult.push({title: '无相关搜索结果'})
						}
						callback(this.queryResult)
					}
				}).catch(() => {
					this.msgError("请求失败")
				})
			},
			handleSelect(item) {
				if (item.id) {
					this.$router.push({name: 'blog', params: {id: item.id}}).catch(err => {
						if (err.name !== 'NavigationDuplicated') {
							throw err
						}
					})
				}
			}
		}
	}
</script>

<style>
	.ui.fixed.inverted.pointing.menu.blog-nav {
		padding: 0 0 16px !important;
		border: 0 !important;
		border-radius: 0 !important;
		box-shadow: none !important;
		pointer-events: none;
		background: linear-gradient(180deg, #41698f 0px, #5e8bb3 26px, #9bbdd9 46px, rgba(198,220,237,.6) 56px, rgba(233,242,249,.24) 64px, rgba(255,255,255,0) 68px) !important;
	}
	.ui.fixed.menu.blog-nav .container {
		position: relative;
		width: min(1320px, calc(100% - 80px)) !important;
		min-height: 52px;
		margin: 0 auto !important;
		align-items: center;
		pointer-events: auto;
	}
	.ui.inverted.menu.blog-nav .item, .ui.inverted.menu.blog-nav .item.m-blue, .ui.inverted.menu.blog-nav .item > i.icon {
		color: #fff !important;
		text-shadow: 0 1px 2px rgba(31,65,97,.18);
	}
	.ui.menu.blog-nav .container > a.item, .ui.menu.blog-nav .el-dropdown-link.item {
		height: 32px;
		padding: 0 13px;
		margin: 0 3px;
		border-radius: 9px;
		font-size: 14px;
		font-weight: 500;
		line-height: 1;
		align-self: center;
		display: inline-flex;
		align-items: center;
	}
	.ui.menu.blog-nav .item > i.icon {
		opacity: .85;
		font-size: 13px;
	}
	.ui.menu.blog-nav .blog-nav-brand {
		display: flex;
		align-items: center;
		margin-right: 22px;
	}
	.ui.menu.blog-nav .blog-nav-brand .item {
		margin: 0;
		padding: 0;
		height: 52px;
		font-size: 19px;
		letter-spacing: -.3px;
		background: transparent !important;
	}
	.ui.inverted.menu.blog-nav .item:before, .ui.inverted.pointing.menu.blog-nav .active.item:after {
		display: none;
	}
	.ui.inverted.menu.blog-nav .item:hover {
		background: rgba(255,255,255,.08) !important;
	}
	.ui.inverted.menu.blog-nav .active.item {
		background: rgba(255,255,255,.13) !important;
		box-shadow: inset 0 0 0 1px rgba(255,255,255,.14);
	}
	.ui.menu.blog-nav .m-search {
		min-width: 190px;
		margin: 0;
		padding: 0 !important;
	}
	.ui.menu.blog-nav .m-search input {
		height: 32px;
		background: rgba(255,255,255,.08);
		border-radius: 9px;
		box-shadow: inset 0 0 0 1px rgba(255,255,255,.16);
		padding-left: 13px;
	}
	.ui.menu.blog-nav .m-search input::placeholder {
		color: rgba(255,255,255,.82);
	}
	.ui.menu.blog-nav .m-right-top {
		top: 10px;
		right: 0;
		width: 32px;
		height: 32px;
		min-height: 0;
		padding: 0;
		margin: 0;
		border-radius: 8px;
		background: rgba(255,255,255,.1);
		box-shadow: inset 0 0 0 1px rgba(255,255,255,.16);
	}
	.ui.menu.blog-nav .m-right-top, .ui.menu.blog-nav .m-right-top > i.icon {
		color: #fff !important;
	}
	@media(max-width:900px) {
		.ui.fixed.menu.blog-nav .container {
			width: calc(100% - 32px) !important;
			min-height: 52px;
			align-items: stretch;
		}
		.ui.menu.blog-nav .blog-nav-brand {
			margin-right: 0;
		}
		.ui.menu.blog-nav .blog-nav-brand .item {
			height: 52px;
		}
		.ui.menu.blog-nav .container > a.item, .ui.menu.blog-nav .el-dropdown-link.item {
			width: 100%;
			height: 38px;
			padding: 0 14px;
			margin: 3px 0;
		}
		.ui.menu.blog-nav .m-search {
			min-width: 0;
			margin: 8px 0 10px;
		}
		.ui.fixed.inverted.pointing.menu.blog-nav.nav-open {
			background: linear-gradient(180deg, #41698f 0px, #5e8bb3 calc(100% - 16px), rgba(198,220,237,.6) calc(100% - 10px), rgba(233,242,249,.24) calc(100% - 4px), rgba(255,255,255,0) 100%) !important;
		}
	}
	.ui.menu.blog-nav .el-dropdown-link {
		outline: none;
		cursor: pointer;
	}
	.ui.menu.blog-nav .m-search input {
		color: #fff !important;
		border: 0 !important;
		padding-right: 34px;
	}
	.ui.menu.blog-nav .m-search i {
		color: #fff !important;
	}
	.ui.menu.blog-nav a:focus-visible, .ui.menu.blog-nav button:focus-visible, .ui.menu.blog-nav .el-dropdown-link:focus-visible {
		outline: 2px solid #fff;
		outline-offset: -3px;
	}
	.el-dropdown-menu.blog-nav-dropdown {
		margin: 8px 0 0 !important;
		padding: 6px !important;
		border: 1px solid rgba(255,255,255,.2) !important;
		border-radius: 12px;
		background: linear-gradient(180deg, #4a749b, #5e8bb3) !important;
		box-shadow: 0 12px 30px rgba(62,101,138,.14);
	}
	.blog-nav-dropdown .el-dropdown-menu__item {
		padding: 0 14px !important;
		color: #fff !important;
		border-radius: 7px;
	}
	.blog-nav-dropdown .el-dropdown-menu__item:hover, .blog-nav-dropdown .el-dropdown-menu__item:focus {
		background: rgba(255,255,255,.12) !important;
	}
	.blog-nav-dropdown .popper__arrow {
		display: none !important;
	}
	.m-search-item {
		min-width: min(350px, calc(100vw - 32px)) !important;
	}
	.m-search-item li {
		line-height: normal !important;
		padding: 8px 10px !important;
	}
	.m-search-item li .title {
		text-overflow: ellipsis;
		overflow: hidden;
		color: rgba(0,0,0,.87);
	}
	.m-search-item li .content {
		text-overflow: ellipsis;
		font-size: 12px;
		color: rgba(0,0,0,.7);
	}
	@media(max-width:1100px) and (min-width:901px) {
		.ui.menu.blog-nav .m-search {
			display: none !important;
		}
	}
	@media(max-width:900px) {
		.ui.menu.blog-nav .container {
			flex-direction: column;
		}
		.ui.menu.blog-nav .container > a, .ui.menu.blog-nav .el-dropdown, .ui.menu.blog-nav .container > .item {
			width: 100%;
		}
		.ui.menu.blog-nav .m-mobile-hide {
			display: none !important;
		}
		.ui.menu.blog-nav .m-mobile-show {
			display: block !important;
		}
	}
</style>
