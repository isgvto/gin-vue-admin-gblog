<template>
	<footer class="journey-footer">
		<FooterLandscape/>
		<div class="footer-inner">
			<div class="footer-content" :class="{'has-qr': siteInfo.footerImgUrl}">
				<section v-if="siteInfo.footerImgUrl" class="footer-visit">
					<h4>{{ siteInfo.footerImgTitle || '扫码手机访问' }}</h4>
					<img :src="siteInfo.footerImgUrl" :alt="siteInfo.footerImgTitle || '博客访问二维码'" width="88" height="88" loading="lazy" decoding="async">
				</section>
				<section class="footer-recent">
					<h4>最新博客</h4>
					<ul>
						<li v-for="item in newBlogList" :key="item.id">
							<a :href="`/blog/${item.id}`" @click.prevent="toBlog(item)">{{ item.title }}</a>
						</li>
					</ul>
				</section>
				<blockquote v-if="hitokoto.hitokoto" class="footer-quote">
					<span class="quote-decoration" aria-hidden="true">“</span>
					<p id="hitokotoText">{{ hitokoto.hitokoto }}</p>
					<cite v-if="hitokoto.from" id="hitokotoFrom">——《{{ hitokoto.from }}》</cite>
				</blockquote>
			</div>
		</div>
		<div class="footer-bottom">
				<p class="footer-copyright">
					<template v-if="siteInfo.copyright">
						<span>{{ siteInfo.copyright.title }}</span>
						<router-link to="/">{{ siteInfo.copyright.siteName }}</router-link>
					</template>
					<span v-if="siteInfo.copyright && siteInfo.beian" class="copyright-dot" aria-hidden="true">·</span>
					<a v-if="siteInfo.beian" rel="external nofollow noopener" href="https://beian.miit.gov.cn/" target="_blank">{{ siteInfo.beian }}</a>
				</p>
				<div v-if="badges.length" class="footer-badges">
					<div class="github-badge" v-for="(item,index) in badges" :key="index">
						<a rel="external nofollow noopener" :href="item.url" target="_blank" :title="item.title">
							<span class="badge-subject">{{ item.subject }}</span>
							<span class="badge-value">{{ item.value }}</span>
						</a>
					</div>
				</div>
		</div>
	</footer>
</template>

<script>
	import FooterLandscape from './FooterLandscape'
	export default {
		name: "Footer",
		components: {FooterLandscape},
		props: {
			siteInfo: {
				type: Object,
				required: true
			},
			badges: {
				type: Array,
				required: true
			},
			newBlogList: {
				type: Array,
				required: true
			},
			hitokoto: {
				type: Object,
				required: true
			}
		},
		methods: {
			toBlog(blog) {
				this.$store.dispatch('goBlogPage', blog)
			}
		}
	}
</script>

<style scoped>
	.journey-footer { position: relative; isolation: isolate; overflow: hidden; flex-shrink: 0; background: #b7cddd; color: #243f55; font-family: var(--blog-reading-font, sans-serif); }
	.footer-inner { position: relative; width: min(1000px, calc(64% - 72px)); margin-left: calc(36% + 24px); padding-top: 68px; }
	.footer-content { display: grid; grid-template-columns: 1fr 1.2fr; gap: 40px; padding: 6px 0 30px; }
	.footer-content.has-qr { grid-template-columns: 160px minmax(0, 1fr) minmax(0, 1.2fr); }
	.footer-content h4 { margin: 0 0 12px; font: inherit; font-size: 13px; font-weight: 600; color: #2c455a; letter-spacing: .08em; }
	.footer-visit img { display: block; width: 88px; height: 88px; padding: 5px; border: 1px solid #d0dfe9; border-radius: 8px; background: #fff; }
	.footer-recent ul { list-style: none; margin: 0; padding: 0; }
	.footer-recent li + li { margin-top: 5px; }
	.footer-recent a { display: inline-block; font-size: 13px; line-height: 1.65; overflow-wrap: anywhere; }
	.journey-footer a { color: #243f55; transition: color .18s; }
	.journey-footer a:hover { color: var(--blog-accent-hover); }
	.journey-footer a:focus-visible { outline: 2px solid var(--blog-accent); outline-offset: 3px; border-radius: 2px; }
	.footer-quote { position: relative; align-self: center; margin: 0; padding: 8px 0 8px 28px; }
	.quote-decoration { position: absolute; top: -23px; left: 22px; color: rgba(56,87,112,.3); font: 64px Georgia, serif; line-height: 1; }
	.footer-quote p { position: relative; margin: 0; font-size: 15px; line-height: 1.95; letter-spacing: .045em; }
	.footer-quote cite { display: block; margin-top: 12px; color: #3c5a72; font-size: 12px; line-height: 1.7; font-style: normal; text-align: right; }
	.footer-bottom { position: relative; width: min(1120px, calc(100% - 64px)); margin: 0 auto; padding: 18px 0 34px; text-align: center; }
	.footer-copyright { display: flex; justify-content: center; flex-wrap: wrap; gap: 6px 10px; margin: 0; font-size: 12px; line-height: 1.8; color: #3c5a72; }
	.copyright-dot { color: #8ba7ba; }
	.footer-badges { display: flex; justify-content: center; flex-wrap: wrap; gap: 6px; margin-top: 12px; }
	.github-badge { margin: 0; background: transparent; border: 1px solid rgba(72,107,134,.18); border-radius: 4px; font-size: 10px; line-height: 14px; overflow: hidden; }
	.github-badge a { display: flex; color: #243f55; }
	.github-badge .badge-subject { background: rgba(255,255,255,.32); padding: 3px 5px; }
	.github-badge .badge-value { background: rgba(255,255,255,.12); padding: 3px 5px; }
	@media (max-width: 1100px) {
		.footer-content, .footer-content.has-qr { grid-template-columns: 120px minmax(0, 1fr); gap: 32px; padding: 8px 0 22px; }
		.footer-content:not(.has-qr) .footer-recent { grid-column: 1 / -1; }
		.footer-content h4 { font-size: 12px; letter-spacing: 0; }
		.footer-visit img { width: 80px; height: 80px; }
		.footer-quote { grid-column: 1 / -1; padding: 22px 0 0; margin-top: 4px; }
		.quote-decoration { top: 13px; left: -5px; font-size: 48px; opacity: .55; }
		.footer-quote p { font-size: 14px; }
		.footer-bottom { padding: 16px 0 30px; }
	}
	@media (max-width: 600px) {
		.footer-inner { width: calc(66% - 40px); margin-left: calc(34% + 16px); padding-top: 56px; }
		.footer-content, .footer-content.has-qr { grid-template-columns: minmax(0, 1fr); gap: 36px; }
		.footer-bottom { width: calc(100% - 48px); }
		.footer-visit { display: flex; flex-direction: row-reverse; align-items: center; justify-content: flex-end; gap: 12px; }
		.footer-visit h4 { flex: 1; margin: 0; line-height: 1.7; }
		.footer-visit img { flex-shrink: 0; }
		.footer-quote { padding-top: 8px; margin-top: 0; }
		.quote-decoration { top: -1px; }
	}
</style>
