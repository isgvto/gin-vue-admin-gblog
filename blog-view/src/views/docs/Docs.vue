<template>
	<div class="docs-page" data-prismjs-copy="复制" data-prismjs-copy-success="已复制" data-prismjs-copy-error="请选中代码手动复制">
		<div class="docs-mobile-picker">
			<el-select :value="activePath" placeholder="选择文档" filterable size="small" @change="selectDocPath">
				<el-option
					v-for="item in flatDocs"
					:key="item.path"
					:label="item.label"
					:value="item.path"
				/>
			</el-select>
		</div>
		<aside class="docs-sidebar m-mobile-hide">
			<div class="docs-panel">
				<div class="docs-panel-title">
					<i class="book icon"></i>
					<span>文档目录</span>
				</div>
				<div v-if="loadingTree" class="docs-empty">目录加载中...</div>
				<div v-else-if="tree.length === 0" class="docs-empty">暂无文档</div>
				<DocTree v-else :nodes="tree" :active-path="activePath" @select="selectDoc"/>
			</div>
		</aside>

		<main class="docs-content">
			<div class="ui padded attached segment docs-article">
				<div v-if="loadingContent" class="docs-empty">文档加载中...</div>
				<div v-else-if="error" class="docs-empty error">{{ error }}</div>
				<template v-else>
					<header v-if="doc.path" class="blog-header">
						<div class="blog-header-top">
							<div class="blog-header-left">
								<span class="docs-path">{{ doc.path }}</span>
							</div>
							<a v-if="doc.editUrl" :href="doc.editUrl" target="_blank" rel="noopener noreferrer" class="header-category">
								<i class="github icon"></i>在 GitHub 上编辑
							</a>
						</div>
						<h1 class="blog-title">{{ docFileTitle }}</h1>
						<div class="blog-meta">
							<span><i class="user outline icon"></i>作者：Gvto</span>
							<span><i class="book icon"></i>字数 {{ docWordCount }}</span>
							<span><i class="clock outline icon"></i>阅读时长 {{ docReadMinutes }} 分钟</span>
							<button class="meta-action" type="button" @click.prevent="bigFontSize=!bigFontSize" title="切换字体大小">
								<i class="font icon"></i>
							</button>
						</div>
					</header>
					<div class="ui middle aligned mobile reversed stackable">
						<div class="ui grid m-margin-lr">
							<div
								class="typo js-toc-content m-padded-tb-small match-braces rainbow-braces"
								v-viewer
								:class="{'m-big-fontsize':bigFontSize}"
								v-html="doc.content"
							></div>
						</div>
					</div>
				</template>
			</div>
		</main>

		<aside class="docs-toc">
			<button class="docs-toc-trigger" type="button" :aria-expanded="mobileTocOpen ? 'true' : 'false'" aria-controls="docs-outline" @click="mobileTocOpen = !mobileTocOpen">
				本文目录 <span aria-hidden="true">{{ mobileTocOpen ? '−' : '+' }}</span>
			</button>
			<div id="docs-outline" :class="{'docs-outline-open': mobileTocOpen}">
			<Tocbot
				title="本页内容"
				empty-text="暂无目录"
				content-selector=".docs-page .js-toc-content"
				:refresh-key="tocRevision"
				:enable-sticky="false"
				:number-content-headings="false"
			/>
			</div>
		</aside>
	</div>
</template>

<script>
	import DocTree from '@/components/docs/DocTree'
	import Tocbot from '@/components/sidebar/Tocbot'
	import {getDocContent, getDocsTree} from '@/api/docs'
	import {isSuccess} from '@/util/gvaResponse'

	export default {
		name: 'Docs',
		components: {DocTree, Tocbot},
		data() {
			return {
				tree: [],
				doc: {
					title: '',
					path: '',
					content: '',
					editUrl: ''
				},
				loadingTree: false,
				loadingContent: false,
				error: '',
				bigFontSize: false,
				mobileTocOpen: false,
				tocRevision: 0,
				contentRequestSeq: 0
			}
		},
		computed: {
			activePath() {
				return this.$route.query.path || ''
			},
			flatDocs() {
				const result = []
				const walk = (nodes, parents = []) => {
					nodes.forEach(node => {
						const nextParents = node.type === 'dir' ? [...parents, node.title] : parents
						if (node.type === 'file' && node.path) {
							result.push({
								path: node.path,
								label: [...parents, node.title].join(' / ')
							})
						}
						if (node.children && node.children.length) {
							walk(node.children, nextParents)
						}
					})
				}
				walk(this.tree)
				return result
			},
			docFileTitle() {
				const filename = (this.doc.path || '').split('/').pop() || this.doc.title || 'README'
				return filename.replace(/\.md$/i, '') || 'README'
			},
			docPlainText() {
				const wrapper = document.createElement('div')
				wrapper.innerHTML = this.doc.content || ''
				return (wrapper.textContent || wrapper.innerText || '').trim()
			},
			docWordCount() {
				const text = this.docPlainText
				if (!text) {
					return 0
				}
				const cjkCount = (text.match(/[\u4e00-\u9fa5]/g) || []).length
				const wordCount = (text.replace(/[\u4e00-\u9fa5]/g, ' ').match(/[A-Za-z0-9_]+(?:[-'][A-Za-z0-9_]+)*/g) || []).length
				return cjkCount + wordCount
			},
			docReadMinutes() {
				return Math.max(1, Math.ceil(this.docWordCount / 400))
			}
		},
		watch: {
			'$route.query.path'(path) {
				if (path) {
					this.loadContent(path)
				}
			}
		},
		created() {
			this.loadTree()
		},
		methods: {
			loadTree() {
				this.loadingTree = true
				getDocsTree().then(res => {
					if (!isSuccess(res)) {
						this.error = res.msg || '获取文档目录失败'
						return
					}
					this.tree = Array.isArray(res.data) ? res.data : []
					const currentPath = this.activePath
					const firstPath = currentPath || this.findReadmeDocPath(this.tree) || this.findFirstDocPath(this.tree)
					if (firstPath && firstPath !== currentPath) {
						this.$router.replace({name: 'docs', query: {path: firstPath}})
					} else if (firstPath) {
						this.loadContent(firstPath)
					}
				}).catch(() => {
					this.error = '获取文档目录失败'
				}).finally(() => {
					this.loadingTree = false
				})
			},
			loadContent(path) {
				const requestSeq = ++this.contentRequestSeq
				this.loadingContent = true
				this.mobileTocOpen = false
				this.$nextTick(() => { this.tocRevision += 1 })
				this.error = ''
				getDocContent(path).then(res => {
					if (requestSeq !== this.contentRequestSeq) {
						return
					}
					if (!isSuccess(res)) {
						this.error = res.msg || '获取文档内容失败'
						this.loadingContent = false
						return
					}
					this.doc = res.data || {}
					document.title = `${this.doc.title || '文档'}${this.$store.state.siteInfo.webTitleSuffix || ''}`
					this.loadingContent = false
					this.$nextTick(() => {
						if (window.Prism && typeof window.Prism.highlightAll === 'function') {
							const content = this.$el.querySelector('.js-toc-content')
							content.querySelectorAll('pre > code').forEach(code => {
								const language = Array.from(code.classList).find(name => name.startsWith('language-'))
								code.parentElement.setAttribute('data-label', language ? language.slice(9) : 'text')
							})
							window.Prism.highlightAllUnder(content)
						}
						this.tocRevision += 1
					})
				}).catch(() => {
					if (requestSeq !== this.contentRequestSeq) {
						return
					}
					this.error = '获取文档内容失败'
					this.loadingContent = false
				})
			},
			selectDoc(node) {
				if (node.path === this.activePath) {
					return
				}
				this.$router.push({name: 'docs', query: {path: node.path}})
			},
			selectDocPath(path) {
				if (!path || path === this.activePath) {
					return
				}
				this.$router.push({name: 'docs', query: {path}})
			},
			findFirstDocPath(nodes) {
				for (const node of nodes) {
					if (node.type === 'file' && node.path) {
						return node.path
					}
					if (node.children && node.children.length) {
						const found = this.findFirstDocPath(node.children)
						if (found) {
							return found
						}
					}
				}
				return ''
			},
			findReadmeDocPath(nodes) {
				for (const node of nodes) {
					if (node.type === 'file' && node.path && /(^|\/)README\.md$/i.test(node.path)) {
						return node.path
					}
					if (node.children && node.children.length) {
						const found = this.findReadmeDocPath(node.children)
						if (found) {
							return found
						}
					}
				}
				return ''
			},
		}
	}
</script>

<style scoped>
	.docs-page {
		display: grid;
		grid-template-columns: 240px minmax(0, 1fr) 216px;
		gap: 24px;
		align-items: start;
		max-width: 1280px;
		margin: 0 auto;
	}

	.docs-mobile-picker {
		display: none;
	}

	.docs-sidebar {
		position: sticky;
		top: 76px;
		max-height: calc(100vh - 82px);
		overflow: auto;
		scrollbar-width: thin;
	}

	.docs-toc {
		position: relative;
	}

	.docs-panel {
		background: #fff;
		border: 1px solid #e7ecf2;
		border-radius: 12px;
		padding: 16px 12px;
		box-shadow: 0 10px 28px rgba(15, 23, 42, .045);
	}

	.docs-panel-title {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 4px 14px;
		margin-bottom: 12px;
		border-bottom: 1px solid #edf1f5;
		color: #1f2937;
		font-weight: 700;
		font-size: 14px;
	}

	.docs-content {
		min-width: 0;
	}

	.docs-article {
		position: relative;
		min-height: 520px;
		overflow: hidden;
		background: #fff !important;
		border: 1px solid #e7ecf2 !important;
		border-radius: 14px !important;
		box-shadow: 0 14px 40px rgba(15, 23, 42, .055);
	}

	.blog-header {
		margin: 0 1.5rem 26px;
		padding: 0 0 22px;
		border-bottom: 1px solid #edf1f5;
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
		color: #8491a2;
		font-size: 13px;
		font-weight: 500;
	}

	.docs-path {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.header-category {
		display: inline-flex;
		flex: 0 0 auto;
		min-height: 30px;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 7px 12px;
		border: 1px solid #dbe7ff;
		border-radius: 999px;
		background: #f5f8ff;
		color: #3568d4;
		font-size: 13px;
		font-weight: 700;
		line-height: 1;
		white-space: nowrap;
	}

	.header-category:hover {
		background: #dbeafe;
		color: #1d4ed8;
	}

	.header-category i,
	.blog-meta i,
	.meta-action i {
		display: inline-flex;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		margin: 0 !important;
		line-height: 1 !important;
	}

	.blog-title {
		margin: 0;
		color: #0f172a;
		font-size: 32px;
		font-weight: 800;
		letter-spacing: 0;
		line-height: 1.25;
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
		display: inline-flex;
		width: 28px;
		height: 28px;
		align-items: center;
		justify-content: center;
		border: 0;
		border-radius: 6px;
		background: transparent;
		color: #64748b;
		cursor: pointer;
		font-family: inherit;
	}

	.meta-action:hover {
		background: #dbeafe;
		color: #2563eb;
	}

	.docs-article .typo {
		width: 100%;
	}

	.docs-article .ui.grid.m-margin-lr {
		margin-right: 0 !important;
		margin-left: 0 !important;
	}

	.docs-empty {
		padding: 30px 10px;
		color: #6b7280;
		text-align: center;
	}

	.docs-empty.error {
		color: #dc2626;
	}

	.docs-article > .ui.middle.aligned {
		padding: 0 1.5rem 2rem;
	}

	.docs-article .typo {
		max-width: 820px;
		margin: 0 auto;
		color: #334155;
		line-height: 1.75;
	}

	h1::before, h2::before, h3::before, h4::before, h5::before, h6::before {
		display: block;
		content: " ";
		height: 55px;
		margin-top: -55px;
		visibility: hidden;
	}

	@media (max-width: 768px) {
		.docs-page {
			display: block;
		}

		.docs-mobile-picker {
			display: block;
			margin-bottom: 12px;
		}

		.docs-mobile-picker .el-select {
			width: 100%;
		}

		.docs-content {
			width: 100%;
		}

		.docs-article {
			border-radius: 0 !important;
		}

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

		.docs-article > .ui.middle.aligned {
			padding: 0 1rem 1.25rem;
		}

		.docs-article .typo {
			font-size: 15px;
			line-height: 1.75;
		}
	}
</style>

<style>
	.docs-page .js-toc-content h1,
	.docs-page .js-toc-content h2,
	.docs-page .js-toc-content h3,
	.docs-page .js-toc-content h4,
	.docs-page .js-toc-content h5,
	.docs-page .js-toc-content h6 {
		padding-bottom: 0;
		border-bottom: 0;
		color: #4b5563;
	}

	.docs-page .js-toc-content h1,
	.docs-page .js-toc-content h2 {
		display: flex;
		align-items: baseline;
		gap: 10px;
	}

	.docs-page .docs-heading-number {
		flex: 0 0 auto;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: auto;
		padding: 0;
		border-radius: 0;
		background: none;
		color: #4b5563;
		font-size: .88em;
		font-weight: 400;
		font-variant-numeric: tabular-nums;
	}
</style>


<style>
/* Document-only presentation: shared blog typography and TOC remain unchanged. */
.docs-page { --docs-accent: #3568d4; --docs-border: #e7ecf2; max-width: 1440px; grid-template-columns: 240px minmax(0, 1fr) 210px; }
.docs-page .docs-content { grid-column: 2; grid-row: 1; }
.docs-page .docs-toc { grid-column: 3; grid-row: 1; position: sticky; top: 76px; min-width: 0; }
.docs-page .docs-toc-trigger { display: none; }
.docs-page .docs-article.ui.segment { margin: 0; width: 100%; padding: 32px; box-shadow: 0 4px 18px rgba(15,23,42,.025); border-radius: 10px !important; }
.docs-page .blog-header { margin: 0 0 28px; }
.docs-page .blog-header-top { flex-wrap: wrap; }
.docs-page .header-category { padding: 4px 0; background: transparent; border: 0; font-weight: 500; }
.docs-page .blog-meta { font-size: 12px; gap: 10px 16px; margin-top: 16px; color: #64748b; }
.docs-page .docs-article > .ui.middle.aligned { padding: 0; }
.docs-page .blog-title { overflow-wrap: anywhere; }
.docs-page .typo { text-align: left; }
.docs-page .typo p, .docs-page .typo li { line-height: 1.75; text-align: left; }
.docs-page .typo h1, .docs-page .typo h2, .docs-page .typo h3, .docs-page .typo h4 { color: #1f2937; line-height: 1.45; scroll-margin-top: 80px; overflow-wrap: anywhere; }
.docs-page .typo h1 { font-size: 28px; }
.docs-page .typo h2 { font-size: 23px; margin-top: 2em; border-left: 3px solid var(--docs-accent); padding-left: 12px; }
.docs-page .typo h3 { font-size: 19px; margin-top: 1.6em; }
.docs-page .typo blockquote { background: #f5f8ff; border-color: #a8c2f5; color: #52647b; }
.docs-page .typo table { display: block; max-width: 100%; overflow-x: auto; }
.docs-page .typo table th { background: #f6f8fb; }
.docs-page .typo table td, .docs-page .typo table th { border-color: var(--docs-border); padding: 10px 14px; }
.docs-page .typo :not(pre) > code { background: #f1f5f9; color: #335278; border: 0; border-radius: 4px; padding: 2px 5px; }
.docs-page .typo pre { background: #202936; color: #e2e8f0; overflow-x: auto; padding: 44px 20px 20px; line-height: 1.7; }
.docs-page .typo pre code { font-family: Consolas, monospace; font-size: 14px; background: transparent; text-shadow: none; }
.docs-page .code-toolbar > .toolbar { opacity: 1; top: 8px; right: 12px; left: 12px; display: flex; justify-content: space-between; }
.docs-page .code-toolbar > .toolbar > .toolbar-item > span,
.docs-page .code-toolbar > .toolbar button { color: #d5dfec; background: transparent; box-shadow: none; font-size: 12px; padding: 2px 8px; }
.docs-page .copy-to-clipboard-button span { background: transparent !important; }
.docs-page a:focus-visible, .docs-page button:focus-visible { outline: 2px solid var(--docs-accent); outline-offset: 3px; }
.docs-page .m-toc.ui.segments { border: 0; border-radius: 0; box-shadow: none !important; background: transparent; }
.docs-page .m-toc > .ui.segment { background: transparent; border: 0 !important; padding: 12px 0; }
.docs-page .m-toc > .secondary.segment { color: #475569; font-size: 13px; font-weight: 600; }
.docs-page .m-toc > .secondary.segment > i { display: none; }
.docs-page .m-toc .fallback-toc { border-left: 1px solid var(--docs-border); padding-left: 10px; }
.docs-page .m-toc .toc-link { font-size: 13px; overflow-wrap: anywhere; }
.docs-page .m-toc .toc-number { display: none; }
.docs-page .m-toc .toc-list li a:hover,
.docs-page .m-toc .fallback-toc .active .toc-link { color: var(--docs-accent) !important; }
.docs-page .m-toc .active > .toc-row { border-left: 2px solid var(--docs-accent); margin-left: -12px; padding-left: 10px; }
.docs-page .m-toc .toc-actions button:hover, .docs-page .m-toc .toc-toggle:hover { background: #edf3ff; color: var(--docs-accent); }
.docs-page .m-toc .toc-scroll-body::-webkit-scrollbar-thumb { background: #cbd5e1; }
.docs-page .doc-tree { padding-left: 0; }
.docs-page .doc-tree .doc-tree { padding-left: 16px; }
.docs-page .doc-tree-item { font-size: 14px; border-radius: 6px; }
.docs-page .doc-tree-item.folder { color: #475569; }
.docs-page .doc-tree-item:not(.folder) > i { opacity: .6; }
@media (max-width: 1199px) {
  .docs-page { grid-template-columns: 220px minmax(0, 1fr); }
  .docs-page .docs-toc { grid-column: 2; grid-row: 1; position: static; }
  .docs-page .docs-content { grid-row: 2; }
  .docs-page .docs-sidebar { grid-row: 1 / 3; }
  .docs-page .docs-toc-trigger { display: flex; justify-content: space-between; width: 100%; padding: 12px 16px; color: #475569; background: white; border: 1px solid var(--docs-border); border-radius: 8px; cursor: pointer; }
  .docs-page #docs-outline { display: none; padding: 0 16px; }
  .docs-page #docs-outline.docs-outline-open { display: block; }
}
@media (max-width: 768px) {
  .docs-page { display: flex; flex-direction: column; gap: 16px; }
  .docs-page .docs-sidebar { display: none !important; }
  .docs-page .docs-mobile-picker { display: block; order: 0; width: 100%; margin: 0; }
  .docs-page .docs-mobile-picker .el-select { width: 100%; }
  .docs-page .docs-toc { order: 1; width: 100%; }
  .docs-page .docs-content { order: 2; width: 100%; }
  .docs-page .docs-article.ui.segment { padding: 22px 18px; }
  .docs-page .blog-header { margin: 0 0 22px; }
  .docs-page .blog-title { font-size: 28px; }
  .docs-page .typo h1 { font-size: 25px; }
  .docs-page .typo h2 { font-size: 21px; }
}
@media (prefers-reduced-motion: reduce) {
  .docs-page .doc-tree-item { transition: none; transform: none; }
}

/* Reading rhythm for rendered Markdown */
.docs-page .typo > :first-child { margin-top: 0; }
.docs-page .typo > p:first-child { color: #52647b; font-size: 17px; }
.docs-page .typo p { margin: .8em 0; }
.docs-page .typo h1,
.docs-page .typo h2,
.docs-page .typo h3,
.docs-page .typo h4,
.docs-page .typo h5,
.docs-page .typo h6 { border: 0; padding-bottom: 0; }
.docs-page .typo h1 { margin-top: 1.2em; margin-bottom: .7em; font-size: 30px; }
.docs-page .typo h2 { margin-top: 1.6em; margin-bottom: .65em; }
.docs-page .typo h3 { margin-top: 1.8em; margin-bottom: .65em; }
.docs-page .typo ul,
.docs-page .typo ol { margin: .7em 0 .9em; padding-left: 1.5em; }
.docs-page .typo li { margin: .2em 0; padding-left: .2em; }
.docs-page .typo li::marker { color: #7d9ee9; }
.docs-page .typo a { color: var(--docs-accent); text-decoration: none; border-bottom: 1px solid #c8d7f6; }
.docs-page .typo a:hover { border-bottom-color: var(--docs-accent); }
.docs-page .typo img { border: 1px solid #e7ecf2; border-radius: 8px; box-shadow: 0 8px 24px rgba(15,23,42,.06); }
.docs-page .typo hr { height: 0; margin: 2em 0; border-bottom: 1px solid #e7ecf2; }
.docs-page .typo blockquote { margin: 1.5em 0; padding: 12px 16px; border-left-width: 3px; border-radius: 0 8px 8px 0; }
.docs-page .typo table { width: 100%; margin: 1.5em 0; border: 1px solid var(--docs-border); border-radius: 8px; border-collapse: separate; border-spacing: 0; overflow: hidden; }
.docs-page .typo table th,
.docs-page .typo table td { padding: 11px 14px; border-width: 0 1px 1px 0; color: #475569; }
.docs-page .typo table tr:last-child td { border-bottom: 0; }
.docs-page .typo table th:last-child,
.docs-page .typo table td:last-child { border-right: 0; }
.docs-page .typo table tbody tr:nth-child(even) { background: #fafbfc; }
.docs-page .typo pre { margin: 1.5em 0; border: 1px solid #2c3949; border-radius: 10px; box-shadow: 0 10px 24px rgba(15,23,42,.12); }
.docs-page .typo pre code { white-space: pre; }
.docs-page .typo dl { margin: 1.2em 0; }
.docs-page .typo dt { margin-top: 1em; color: #334155; font-weight: 700; }
.docs-page .typo dd { margin: .35em 0 0 1.2em; color: #64748b; }
@media (max-width: 768px) {
  .docs-page .typo > p:first-child { font-size: 16px; }
  .docs-page .typo h1 { font-size: 26px; }
  .docs-page .typo pre { margin-right: -2px; margin-left: -2px; padding-right: 14px; padding-left: 14px; border-radius: 8px; }
  .docs-page .typo table { font-size: 14px; }
}
</style>

<style>
/* Modern documentation shell: a quiet canvas, dense navigation rails, and a wide reading column. */
.site.docs-layout {
	background: #fff;
}

.site.docs-layout > .ui.fixed.inverted.pointing.menu {
	background: rgba(255, 255, 255, .96) !important;
	border-bottom: 1px solid #e8edf3;
	box-shadow: 0 1px 10px rgba(15, 23, 42, .06);
	backdrop-filter: blur(12px);
}

.site.docs-layout > .ui.fixed.inverted.pointing.menu .item,
.site.docs-layout > .ui.fixed.inverted.pointing.menu .ui.header {
	color: #475569 !important;
}

.site.docs-layout > .ui.fixed.inverted.pointing.menu .item:hover,
.site.docs-layout > .ui.fixed.inverted.pointing.menu .item.active {
	background: #f5f8ff !important;
	color: #2563eb !important;
}

.site.docs-layout > .ui.fixed.inverted.pointing.menu .item.active:after {
	background: #2563eb !important;
}

.site.docs-layout > .ui.fixed.inverted.pointing.menu .m-search input {
	background: #f8fafc !important;
	border: 1px solid #e2e8f0 !important;
	color: #334155 !important;
}

.site.docs-layout > .main {
	margin-top: 40px;
	background: #fff;
}

.site.docs-layout > .main > .m-padded-tb-big {
	padding-top: 20px !important;
	padding-bottom: 36px !important;
}

.docs-page {
	min-height: calc(100vh - 100px);
	grid-template-columns: 248px minmax(0, 1fr) 208px;
	gap: 32px;
	max-width: 1480px;
	align-items: stretch;
}

.docs-page .docs-sidebar {
	top: 76px;
	max-height: calc(100vh - 92px);
}

.docs-page .docs-panel {
	min-height: calc(100vh - 120px);
	padding: 22px 14px;
	border: 0;
	border-right: 1px solid #e8edf3;
	border-radius: 0;
	box-shadow: none;
}

.docs-page .docs-panel-title {
	padding: 0 10px 16px;
	margin-bottom: 14px;
	color: #1e293b;
	font-size: 15px;
}

.docs-page .docs-content {
	min-width: 0;
}

.docs-page .docs-article.ui.segment {
	min-height: calc(100vh - 120px);
	padding: 34px 44px 72px;
	border: 0 !important;
	border-radius: 0 !important;
	box-shadow: none;
}

.docs-page .blog-header {
	max-width: 960px;
	margin: 0 auto 24px;
	padding-bottom: 20px;
}

.docs-page .blog-title {
	font-size: clamp(28px, 3vw, 40px);
	letter-spacing: -.025em;
	line-height: 1.18;
}

.docs-page .docs-article .typo {
	max-width: 960px;
	color: #334155;
	font-size: 16px;
	line-height: 1.72;
}

.docs-page .docs-toc {
	top: 76px;
	align-self: start;
	padding-top: 24px;
}

.docs-page .m-toc > .ui.segment {
	padding-left: 0;
}

.docs-page .m-toc .toc-scroll-body,
.docs-page .m-toc .fallback-toc {
	max-height: calc(100vh - 150px);
	margin: 0;
	padding-top: 4px;
	overflow: auto;
}

.site.docs-layout > footer {
	display: none;
}

@media (max-width: 1199px) {
	.docs-page {
		grid-template-columns: 220px minmax(0, 1fr);
		gap: 24px;
	}
	.docs-page .docs-article.ui.segment {
		padding-right: 30px;
		padding-left: 30px;
	}
}

@media (max-width: 768px) {
	.site.docs-layout > .main {
		margin-top: 40px;
	}
	.site.docs-layout > .main > .m-padded-tb-big {
		padding-top: 12px !important;
		padding-bottom: 20px !important;
	}
	.docs-page {
		min-height: 0;
		gap: 12px;
	}
	.docs-page .docs-article.ui.segment {
		min-height: 0;
		padding: 24px 18px 44px;
	}
	.docs-page .blog-title {
		font-size: 28px;
	}
	.docs-page .docs-toc {
		padding-top: 0;
	}
	.docs-page .docs-article .typo {
		font-size: 15px;
	}
}
</style>
<style>
/* Compact technical-document rhythm */
.docs-page .typo { line-height: 1.62; }
.docs-page .typo p { margin: .55em 0; line-height: 1.62; }
.docs-page .typo > h1 { margin: 24px 0 10px; font-size: 28px; line-height: 1.3; }
.docs-page .typo > h2 { margin: 22px 0 8px; padding-left: 0; font-size: 22px; line-height: 1.35; }
.docs-page .typo > h3 { margin: 1.1em 0 .35em; font-size: 18px; line-height: 1.4; }
.docs-page .typo > h4,
.docs-page .typo > h5,
.docs-page .typo > h6 { margin: .9em 0 .3em; line-height: 1.4; }
.docs-page .typo ul,
.docs-page .typo ol { margin: .4em 0 .75em; line-height: 1.62; }
.docs-page .typo li { margin: .08em 0; line-height: 1.62; }
.docs-page .typo li p { margin: .25em 0; }
.docs-page .typo > :first-child { margin-top: 0; }
.docs-page .typo > h1 + h2,
.docs-page .typo > h2 + h3,
.docs-page .typo > h3 + h4 { margin-top: 12px; }
.docs-page .typo > h1 + p,
.docs-page .typo > h2 + p,
.docs-page .typo > h3 + p,
.docs-page .typo > h4 + p { margin-top: 4px; }
.docs-page .typo li > p:first-child { margin-top: 0; }
.docs-page .typo li > p:last-child { margin-bottom: 0; }
.docs-page .typo li > ul,
.docs-page .typo li > ol { margin-top: 4px; margin-bottom: 4px; }
.docs-page .blog-header { padding-bottom: 16px; margin-bottom: 18px; }
.docs-page .blog-header-top { margin-bottom: 12px; }
.docs-page .blog-meta { margin-top: 12px; }
.docs-page .typo blockquote { margin: .9em 0; }
.docs-page .typo pre,
.docs-page .typo table { margin: 1em 0; }
.docs-page .typo hr { margin: 1.3em 0; }
@media (max-width: 768px) {
  .docs-page .typo { line-height: 1.58; }
  .docs-page .typo p,
  .docs-page .typo li { line-height: 1.58; }
  .docs-page .typo > h1 { font-size: 25px; }
  .docs-page .typo > h2 { font-size: 20px; }
}
</style>
