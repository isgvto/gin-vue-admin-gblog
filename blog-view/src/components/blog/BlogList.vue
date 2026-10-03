<template>
	<section class="latest-card">
		<header class="latest-header">
			<h2>最新文章</h2>
		</header>
		<BlogItem :blogList="blogList" v-viewer/>
		<Pagination :getBlogList="getBlogList" :totalPage="totalPage"/>
	</section>
</template>

<script>
	import BlogItem from "./BlogItem";
	import Pagination from "./Pagination";

	export default {
		name: "BlogList",
		components: {Pagination, BlogItem},
		props: {
			getBlogList: {
				type: Function,
				required: true
			},
			blogList: {
				type: Array,
				required: true
			},
			totalPage: {
				type: Number,
				required: true
			}
		}
	}
</script>

<style scoped>
	.latest-card {
		padding: 24px;
		border: 1px solid rgba(229, 231, 235, .92);
		border-radius: 12px;
		background: rgba(255, 255, 255, .94);
		box-shadow: 0 12px 32px rgba(15, 23, 42, .06);
		backdrop-filter: blur(8px);
	}

	.latest-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 18px;
	}

	.latest-header h2 {
		margin: 0;
		color: #111827;
		font-size: 22px;
		line-height: 1.2;
		letter-spacing: 0;
	}

	.latest-header h2::before {
		display: inline-block;
		width: 4px;
		height: 20px;
		margin-right: 10px;
		border-radius: 999px;
		background: #409eff;
		content: '';
		vertical-align: -2px;
	}

	@media only screen and (max-width: 640px) {
		.latest-card {
			padding: 16px;
		}
	}
</style>
