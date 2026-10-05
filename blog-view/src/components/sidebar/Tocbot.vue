<template>
	<div class="toc-sticky-holder" :style="holderStyle">
		<div
			ref="toc"
			class="ui segments m-toc toc-wrapper m-box"
			:class="{'toc-fixed': tocFixed}"
			:style="tocStyle"
		>
			<div class="ui secondary segment">
				<i class="list ul icon"></i><span class="toc-title">{{ title }}</span>
				<div class="toc-actions" v-if="collapsibleTocItems.length">
					<button
						type="button"
						:title="allTocExpanded ? '全部收起' : '全部展开'"
						:aria-label="allTocExpanded ? '全部收起' : '全部展开'"
						:aria-expanded="allTocExpanded ? 'true' : 'false'"
						@click="allTocExpanded ? collapseAll() : expandAll()"
					>
						<i :class="allTocExpanded ? 'compress arrows alternate icon' : 'expand arrows alternate icon'"></i>
					</button>
				</div>
			</div>
			<div class="ui segment" :class="{'toc-scroll-body': limitHeight}">
				<div v-if="fallbackTocItems.length === 0" class="toc-empty">{{ emptyText }}</div>
				<ul class="toc-list fallback-toc" v-else>
					<li
						v-for="item in visibleTocItems"
						:key="item.id"
						:class="[`toc-list-item level-${item.level}`, {active: item.id === activeHeadingId, collapsed: isCollapsed(item)}]"
						:style="{'--toc-depth': item.depth}"
					>
						<div class="toc-row">
							<button
								v-if="hasChildren(item)"
								type="button"
								class="toc-toggle"
								:title="isCollapsed(item) ? '展开' : '折叠'"
								:aria-expanded="isCollapsed(item) ? 'false' : 'true'"
								:aria-label="`${isCollapsed(item) ? '展开' : '折叠'} ${item.text}`"
								@click.stop="toggleCollapse(item)"
							>
								<i :class="isCollapsed(item) ? 'caret right icon' : 'caret down icon'"></i>
							</button>
							<span v-else class="toc-toggle-placeholder"></span>
							<a class="toc-link" :href="`#${item.id}`" :aria-current="item.id === activeHeadingId ? 'location' : null" @click.prevent="scrollToHeading(item.id)">
								{{ item.text }}
							</a>
						</div>
					</li>
				</ul>
			</div>
		</div>
	</div>
</template>

<script>
	import {mapState} from 'vuex'

	export default {
		name: "Tocbot",
		props: {
			title: {
				type: String,
				default: '本文目录'
			},
			emptyText: {
				type: String,
				default: '暂无目录'
			},
			contentSelector: {
				type: String,
				default: '.js-toc-content'
			},
			headingSelector: {
				type: String,
				default: 'h1,h2,h3,h4'
			},
			refreshKey: {
				type: [String, Number],
				default: ''
			},
			enableSticky: {
				type: Boolean,
				default: true
			},
			stickyTop: {
				type: Number,
				default: 60
			},
			numberContentHeadings: {
				type: Boolean,
				default: false
			},
			limitHeight: {
				type: Boolean,
				default: true
			}
		},
		data() {
			return {
				tocFixed: false,
				tocTop: 0,
				tocLeft: 0,
				tocWidth: 0,
				tocHeight: 0,
				tocFixedTop: 0,
				stickyTimer: null,
				resizeObserver: null,
				fallbackTocItems: [],
				collapsedTocIds: {},
				activeHeadingId: ''
			}
		},
		computed: {
			...mapState(['isBlogRenderComplete']),
			holderStyle() {
				return this.enableSticky && this.tocFixed ? {height: `${this.tocHeight}px`} : {}
			},
			tocStyle() {
				return this.enableSticky && this.tocFixed ? {top: `${this.tocFixedTop}px`, left: `${this.tocLeft}px`, width: `${this.tocWidth}px`} : {}
			},
			visibleTocItems() {
				const hiddenLevels = []
				return this.fallbackTocItems.filter(item => {
					while (hiddenLevels.length && item.level <= hiddenLevels[hiddenLevels.length - 1]) {
						hiddenLevels.pop()
					}
					const visible = hiddenLevels.length === 0
					if (visible && this.isCollapsed(item)) {
						hiddenLevels.push(item.level)
					}
					return visible
				})
			},
			collapsibleTocItems() {
				return this.fallbackTocItems.filter(item => this.hasChildren(item))
			},
			allTocExpanded() {
				return this.collapsibleTocItems.length > 0
					&& this.collapsibleTocItems.every(item => !this.isCollapsed(item))
			}
		},
		mounted() {
			this.initTocbot()
			this.$nextTick(this.bindSticky)
		},
		beforeDestroy() {
			window.removeEventListener('scroll', this.updateSticky)
			window.removeEventListener('scroll', this.updateActiveHeading)
			window.removeEventListener('resize', this.refreshSticky)
			if (this.resizeObserver) {
				this.resizeObserver.disconnect()
			}
			if (this.stickyTimer) {
				clearTimeout(this.stickyTimer)
			}
			if (window.tocbot && typeof window.tocbot.destroy === 'function') {
				window.tocbot.destroy()
			}
		},
		watch: {
			isBlogRenderComplete() {
				if (this.isBlogRenderComplete) {
					this.initTocbot()
					this.$nextTick(this.refreshSticky)
				}
			},
			refreshKey() {
				this.$nextTick(() => {
					this.initTocbot()
					this.refreshSticky()
				})
			}
		},
		methods: {
			initTocbot() {
				const content = window.document.querySelector(this.contentSelector)
				if (!content || !content.querySelector(this.headingSelector)) {
					this.fallbackTocItems = []
					return
				}
				if (window.tocbot && typeof window.tocbot.destroy === 'function') {
					window.tocbot.destroy()
				}
				this.buildFallbackToc(content)
			},
			buildFallbackToc(content) {
				const headings = Array.from(content.querySelectorAll(this.headingSelector))
				const counters = [0, 0, 0, 0]
				const levels = []
				this.fallbackTocItems = headings.map((heading, index) => {
					if (!heading.id) {
						heading.id = `blog-heading-${index + 1}`
					}
					const level = Number(heading.tagName.slice(1))
					while (levels.length && levels[levels.length - 1] >= level) levels.pop()
					const depth = levels.length
					levels.push(level)
					const counterIndex = level - 1
					counters[counterIndex] += 1
					for (let i = counterIndex + 1; i < counters.length; i++) {
						counters[i] = 0
					}
					const number = counters.slice(0, counterIndex + 1).filter(Boolean).join('.')
					this.removeHeadingNumber(heading)
					const text = heading.textContent.trim()
					if (level <= 2) {
						this.applyHeadingNumber(heading, number)
					}
					return {
						id: heading.id,
						text,
						level,
						depth,
						number
					}
				}).filter(item => item.text)
				this.collapseAll()
				this.updateActiveHeading()
				this.$nextTick(this.refreshSticky)
			},
			bindSticky() {
				this.refreshSticky()
				window.addEventListener('scroll', this.updateSticky, {passive: true})
				window.addEventListener('scroll', this.updateActiveHeading, {passive: true})
				window.addEventListener('resize', this.refreshSticky)
				if (window.ResizeObserver && this.$el.parentElement) {
					this.resizeObserver = new ResizeObserver(this.refreshSticky)
					this.resizeObserver.observe(this.$el.parentElement)
					this.resizeObserver.observe(this.$refs.toc)
					const footer = this.$el.closest('.site')?.querySelector('footer')
					if (footer) this.resizeObserver.observe(footer)
				}
				this.stickyTimer = setTimeout(this.refreshSticky, 500)
			},
			refreshSticky() {
				if (!this.enableSticky) {
					this.tocFixed = false
					return
				}
				const toc = this.$refs.toc
				if (!toc) {
					return
				}
				this.tocFixed = false
				this.$nextTick(() => {
					const rect = toc.getBoundingClientRect()
					this.tocTop = rect.top + window.pageYOffset
					this.tocLeft = rect.left
					this.tocWidth = rect.width
					this.tocHeight = rect.height
					this.updateSticky()
				})
			},
			updateSticky() {
				if (!this.enableSticky || !this.tocTop) {
					return
				}
				this.tocFixed = window.pageYOffset + this.stickyTop >= this.tocTop
				const footer = this.$el.closest('.site')?.querySelector('footer')
				const height = this.$refs.toc ? this.$refs.toc.getBoundingClientRect().height : this.tocHeight
				const marginTop = this.$refs.toc ? parseFloat(window.getComputedStyle(this.$refs.toc).marginTop) || 0 : 0
				this.tocFixedTop = footer
					? Math.min(this.stickyTop, footer.getBoundingClientRect().top - height - marginTop - 12)
					: this.stickyTop
			},
			updateActiveHeading() {
				if (this.fallbackTocItems.length === 0) {
					return
				}
				const nav = document.querySelector('.blog-nav')
				const offset = nav ? nav.getBoundingClientRect().height + 2 : 90
				let activeId = this.fallbackTocItems[0].id
				this.fallbackTocItems.forEach(item => {
					const heading = document.getElementById(item.id)
					if (heading && heading.getBoundingClientRect().top <= offset + 1) {
						activeId = item.id
					}
				})
				this.activeHeadingId = activeId
			},
			scrollToHeading(id) {
				const heading = document.getElementById(id)
				if (!heading) {
					return
				}
				const nav = document.querySelector('.blog-nav')
				const offset = nav ? nav.getBoundingClientRect().height + 2 : 55
				const top = heading.getBoundingClientRect().top + window.pageYOffset - offset
				const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
				window.scrollTo({top, behavior: reducedMotion ? 'auto' : 'smooth'})
				this.activeHeadingId = id
			},
			removeHeadingNumber(heading) {
				const oldNumber = heading.querySelector(':scope > .docs-heading-number')
				if (oldNumber) {
					oldNumber.remove()
				}
			},
			applyHeadingNumber(heading, number) {
				this.removeHeadingNumber(heading)
				if (!this.numberContentHeadings) {
					return
				}
				const numberEl = document.createElement('span')
				numberEl.className = 'docs-heading-number'
				numberEl.textContent = number
				heading.insertBefore(numberEl, heading.firstChild)
			},
			hasChildren(item) {
				const index = this.fallbackTocItems.findIndex(tocItem => tocItem.id === item.id)
				return index >= 0
					&& this.fallbackTocItems[index + 1]
					&& this.fallbackTocItems[index + 1].level > item.level
			},
			isCollapsed(item) {
				return Boolean(this.collapsedTocIds[item.id])
			},
			toggleCollapse(item) {
				if (this.isCollapsed(item)) {
					this.$delete(this.collapsedTocIds, item.id)
				} else {
					this.$set(this.collapsedTocIds, item.id, true)
				}
				this.$nextTick(this.refreshSticky)
			},
			collapseAll() {
				const next = {}
				this.collapsibleTocItems.forEach(item => {
					next[item.id] = true
				})
				this.collapsedTocIds = next
				this.$nextTick(this.refreshSticky)
			},
			expandAll() {
				this.collapsedTocIds = {}
				this.$nextTick(this.refreshSticky)
			}
		}
	}
</script>

<style>
	.toc-sticky-holder {
		position: relative;
	}

	.m-toc {
		z-index: 10 !important;
	}
	.ui.segments.m-toc {
		border: 1px solid var(--blog-accent-border);
		border-radius: 10px;
		overflow: hidden;
		transition: box-shadow .2s ease !important;
	}
	.site:not(.docs-layout) .ui.segments.m-toc.m-box,
	.site:not(.docs-layout) .ui.segments.m-toc.m-box:hover {
		box-shadow: 0 3px 14px rgba(50,111,168,.045) !important;
	}
	.m-toc > .ui.segment {
		padding: 10px 8px;
		border-color: var(--blog-accent-border);
		background: #fff;
	}

	.m-toc > .secondary.segment {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		background: var(--blog-accent-soft);
		color: var(--blog-accent);
	}
	.m-toc > .secondary.segment > i.icon { margin: 0; }
	.m-toc .toc-title { flex: 1; color: #1f2937; font-size: 14px; font-weight: 600; }

	.m-toc .toc-actions {
		display: inline-flex;
		align-items: center;
		gap: 4px;
	}

	.m-toc .toc-actions button,
	.m-toc .toc-toggle {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: 0;
		background: transparent;
		color: #6b7280;
		cursor: pointer;
	}

	.m-toc .toc-actions button {
		width: 24px;
		height: 24px;
		border-radius: 4px;
		padding: 0;
	}

	.m-toc .toc-actions button:hover,
	.m-toc .toc-toggle:hover {
		background: var(--blog-accent-soft-hover);
		color: var(--blog-accent);
	}
	.m-toc button:focus-visible, .m-toc .toc-link:focus-visible { outline: 2px solid var(--blog-accent-border-strong); outline-offset: -2px; border-radius: 4px; }

	.m-toc .toc-actions button i,
	.m-toc .toc-toggle i {
		margin: 0 !important;
	}

	.m-toc.toc-fixed {
		position: fixed !important;
	}

	.m-toc .toc-scroll-body {
		max-height: calc(100vh - var(--blog-nav-clearance, 60px) - 65px);
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-width: thin;
	}

	.m-toc .toc-scroll-body::-webkit-scrollbar {
		width: 6px;
	}

	.m-toc .toc-scroll-body::-webkit-scrollbar-thumb {
		border-radius: 6px;
		background: var(--blog-accent-border-strong);
	}

	.m-toc .toc-scroll-body::-webkit-scrollbar-track {
		background: transparent;
	}

	.m-toc .toc {
		overflow-y: auto
	}

	.m-toc .toc > ul {
		overflow: hidden;
		position: relative
	}

	.m-toc .toc > ul li {
		list-style: none
	}

	.m-toc .toc-list {
		list-style-type: none;
		margin: 0;
		padding-left: 10px
	}

	.m-toc .fallback-toc {
		padding-left: 0
	}

	.m-toc .toc-row {
		display: flex;
		min-width: 0;
		align-items: flex-start;
		padding: 0 4px;
		border-left: 2px solid transparent;
		border-radius: 6px;
		transition: background-color .18s ease, border-color .18s ease;
	}
	.m-toc .toc-list-item { margin: 2px 0; padding-left: calc(var(--toc-depth, 0) * 12px); }
	.m-toc .toc-row:hover { background: var(--blog-accent-soft); }
	.m-toc .active > .toc-row { border-left-color: var(--blog-accent); background: var(--blog-accent-soft); }

	.m-toc .toc-toggle,
	.m-toc .toc-toggle-placeholder {
		flex: 0 0 18px;
		width: 18px;
		height: 24px;
		padding: 0;
		border-radius: 4px;
		margin-top: 4px;
	}

	.m-toc .toc-list li a {
		display: block;
		min-width: 0;
		flex: 1 1 auto;
		padding: 5px 2px;
		color: #374151;
		font-family: var(--blog-reading-font);
		font-size: 13px;
		font-weight: 400;
		line-height: 1.65;
		overflow-wrap: anywhere;
	}

	.m-toc .toc-list li a:hover {
		color: #111827;
	}

	.m-toc a.toc-link {
		text-decoration: none;
	}

	.m-toc .is-collapsible {
		max-height: 1000px;
		overflow: hidden;
		transition: all 300ms ease-in-out
	}

	.m-toc .is-collapsed {
		max-height: 0
	}

	.m-toc .is-active-link {
		font-weight: 600;
		color: #111827 !important;
	}

	.m-toc .fallback-toc .active .toc-link {
		font-weight: 600;
		color: #111827 !important;
	}

	.m-toc .toc-link::before {
		background-color: #EEE;
		content: ' ';
		display: inline-block;
		height: 0;
		left: 0;
		margin-top: -1px;
		position: absolute;
		width: 2px
	}

	.m-toc .is-active-link::before {
		background-color: var(--blog-accent)
	}
</style>
