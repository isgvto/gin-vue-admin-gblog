<template>
	<header ref="header" class="home-hero">
		<canvas ref="scene" class="hero-scene" aria-hidden="true"></canvas>
		<canvas ref="rain" class="hero-scene hero-rain" aria-hidden="true"></canvas>
		<div ref="brandContent" class="hero-content">
			<h1 class="hero-brand" :aria-label="blogName">
				<svg ref="brandMark" class="brand-mark" viewBox="0 0 560 176" role="img" aria-label="Gvto，归途">
					<defs>
						<linearGradient id="hero-letter-blue" x1="0" y1="0" x2="560" y2="130" gradientUnits="userSpaceOnUse">
							<stop offset="0" stop-color="#17202d"/>
							<stop offset="0.26" stop-color="#293e5a"/>
							<stop offset="0.58" stop-color="#307eed"/>
							<stop offset="1" stop-color="#69b5ff"/>
						</linearGradient>
					</defs>
					<g class="letter-tracks" aria-hidden="true"><path v-for="(path, index) in letterPaths" :key="index" :d="path"/></g>
					<g class="letter-ink" aria-hidden="true"><path v-for="(path, index) in letterPaths" :key="index" :d="path" pathLength="1" :style="{'--letter-delay': `${index * 150}ms`}"/></g>
				</svg>
			</h1>
			<p v-if="heroCaption" class="hero-caption">
				<span class="hero-caption-line">
					<i class="caption-ornament" aria-hidden="true"></i>
					<span class="hero-caption-text">{{ heroCaption }}</span>
					<i class="caption-ornament caption-ornament-right" aria-hidden="true"></i>
				</span>
			</p>
		</div>
		<button type="button" class="hero-scroll" aria-label="进入博客列表" @click="scrollToMain">
			<span class="scroll-circle"><span aria-hidden="true">↓</span></span>
		</button>
		<div class="hero-signature"><span>归·途</span></div>
	</header>
</template>

<script>
	import {mapState} from 'vuex'
	import {createHeroScene} from './heroScene'
	import {createRainScene} from './rainScene'

	export default {
		name: 'Header',
		data() {
			return {
				letterPaths: [
					'M 139 43 C 125 26 105 18 83 18 C 44 18 18 44 18 84 C 18 125 44 152 83 152 C 106 152 125 144 140 129 L 140 92 L 91 92',
					'M 183 61 L 224 141 Q 229 151 235 139 L 274 61',
					'M 331 23 L 331 119 Q 331 147 359 147 L 378 147 M 303 61 L 378 61',
					'M 535 104 C 535 76 515 56 487 56 C 459 56 439 76 439 104 C 439 132 459 152 487 152 C 515 152 535 132 535 104'
				]
			}
		},
		computed: {
			...mapState(['siteInfo']),
			blogName() { return this.siteInfo && this.siteInfo.blogName || "Gvto's Blog" },
			heroCaption() {
				const text = this.siteInfo && this.siteInfo.malfunctionText
				return typeof text === 'string' ? text.trim() : ''
			}
		},
		mounted() {
			this._heroScene = createHeroScene(this.$refs.scene, this.$refs.header, {brand: this.$refs.brandContent})
			this._rainScene = createRainScene(this.$refs.rain, this.$refs.header, {hero: true})
		},
		beforeDestroy() {
			if (this._heroScene) this._heroScene.destroy()
			if (this._rainScene) this._rainScene.destroy()
		},
		methods: {
			async scrollToMain() {
				this.cancelScrollToTop()
				// Measure after the click has bubbled and collapsed the mobile menu.
				await new Promise(resolve => window.requestAnimationFrame(resolve))
				const target = document.getElementById('blog-list')
				if (!target) return
				const nav = document.querySelector('.site .ui.fixed.menu')
				const navHeight = nav ? nav.getBoundingClientRect().height : 0
				const top = Math.max(0, window.scrollY + target.getBoundingClientRect().top - navHeight - 2)
				const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
				window.scrollTo({top, behavior: reducedMotion ? 'auto' : 'smooth'})
			}
		}
	}
</script>

<style scoped>
	.home-hero {
		position: relative;
		isolation: isolate;
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100vh;
		height: 100svh;
		min-height: 600px;
		overflow: hidden;
		background: #fff;
		color: #253446;
	}
	.hero-scene { position: absolute; inset: 0; z-index: 0; width: 100%; height: 100%; pointer-events: none; }
	.hero-content { position: relative; z-index: 1; width: min(780px, 80%); margin-top: calc(-17vh - 44px); text-align: center; transform: translate3d(var(--brand-x, 0px), var(--brand-y, 0px), 0); }
	.hero-brand { margin: 0; line-height: 1; font-weight: 400; }
	.hero-caption {
		position: absolute;
		top: calc(100% + 16px);
		left: 50%;
		width: min(520px, 100%);
		margin: 0;
		transform: translateX(-50%);
		color: #526d86;
		font-family: "LXGW WenKai", "霞鹜文楷", "STKaiti", "KaiTi", serif;
		font-size: 16px;
		font-weight: 400;
		line-height: 1.8;
		letter-spacing: .18em;
		white-space: pre-line;
		overflow-wrap: anywhere;
		text-wrap: balance;
	}
	.hero-caption-line { display: inline-flex; align-items: center; max-width: 100%; gap: 16px; animation: caption-reveal .9s ease-out 1.8s both; }
	.hero-caption-text { display: block; min-width: 0; }
	.caption-ornament { position: relative; flex: 0 0 56px; width: 56px; height: 12px; }
	.caption-ornament::before {
		content: '';
		position: absolute;
		top: 50%;
		left: 0;
		right: 12px;
		height: 1px;
		background: linear-gradient(90deg, rgba(169,197,218,0), #a9c5da);
	}
	.caption-ornament::after {
		content: '';
		position: absolute;
		top: 50%;
		right: 0;
		width: 5px;
		height: 5px;
		border: 1px solid #a9c5da;
		background: rgba(255,255,255,.65);
		transform: translateY(-50%) rotate(45deg);
	}
	.caption-ornament-right { transform: rotate(180deg); }
	.brand-mark { display: block; width: min(460px, 100%); height: auto; margin: auto; overflow: visible; filter: drop-shadow(0 10px 16px rgba(43,118,226,.09)); animation: brand-light-settle 1.5s ease-in-out 1.8s both; }
	.brand-mark path { fill: none; stroke-width: 15px; stroke-linecap: round; stroke-linejoin: round; }
	.letter-tracks { stroke: #e9f2ff; }
	.letter-ink { stroke: url(#hero-letter-blue); }
	.letter-ink path {
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		animation: letter-draw 1.35s cubic-bezier(.45,0,.18,1) var(--letter-delay) forwards, letter-settle 1.35s cubic-bezier(.22,.7,.16,1) var(--letter-delay) both;
	}
	.hero-scroll { position: absolute; z-index: 1; bottom: 35px; left: 50%; transform: translateX(-50%); display: flex; align-items: center; flex-direction: column; gap: 11px; padding: 8px 18px; border: 0; background: transparent; color: #283b55; cursor: pointer; }
	.scroll-circle { display: flex; align-items: center; justify-content: center; width: 52px; height: 52px; border: 1px solid #d4dde9; border-radius: 50%; background: rgba(255,255,255,.95); box-shadow: 0 6px 22px rgba(25,43,70,.07); font-size: 27px; line-height: 1; transition: border-color .2s, box-shadow .2s; }
	.scroll-circle > span { display: block; animation: scroll-hint 3s ease-in-out infinite; }
	.hero-paused .scroll-circle > span, .hero-paused .letter-ink path, .hero-paused .brand-mark, .hero-paused .hero-caption-line { animation-play-state: paused; }
	.hero-scroll:hover .scroll-circle { border-color: #73a9f1; box-shadow: 0 8px 25px rgba(44,111,205,.14); }
	.hero-scroll:focus-visible { outline: 2px solid #3979cc; outline-offset: 4px; border-radius: 30px; }
	.hero-signature {
		position: absolute;
		z-index: 1;
		right: clamp(24px, 4vw, 80px);
		bottom: 58px;
		display: flex;
		align-items: center;
		gap: 14px;
		color: #516b82;
		font-family: "LXGW WenKai", "霞鹜文楷", "STKaiti", "KaiTi", serif;
		font-size: 15px;
		font-weight: 400;
		line-height: 1.5;
		user-select: none;
		pointer-events: none;
	}
	.hero-signature span { padding-left: .28em; letter-spacing: .28em; }
	.hero-signature::before,
	.hero-signature::after {
		content: '';
		width: 26px;
		height: 1px;
		background: linear-gradient(90deg, rgba(81,107,130,0), rgba(81,107,130,.6));
	}
	.hero-signature::after { transform: rotate(180deg); }
	@keyframes letter-draw { to { stroke-dashoffset: 0; } }
	@keyframes letter-settle { from { opacity: .3; transform: translateY(9px); } to { opacity: 1; transform: translateY(0); } }
	@keyframes caption-reveal { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
	@keyframes brand-light-settle {
		0%, 100% { filter: brightness(1) drop-shadow(0 10px 16px rgba(43,118,226,.09)) drop-shadow(0 0 0 rgba(80,150,240,0)); }
		28% { filter: brightness(1.12) drop-shadow(0 10px 16px rgba(43,118,226,.09)) drop-shadow(0 0 12px rgba(80,150,240,.18)); }
	}
	@keyframes scroll-hint { 0%, 100% { transform: translateY(-2px); } 50% { transform: translateY(3px); } }
	@media (max-width: 768px) {
		.home-hero { min-height: 580px; }
		.hero-content { width: 82%; margin-top: calc(-20svh - 36px); }
		.brand-mark { width: 90%; }
		.hero-caption { top: calc(100% + 10px); font-size: 14px; letter-spacing: .14em; }
		.hero-caption-line { gap: 10px; }
		.caption-ornament { flex-basis: 24px; width: 24px; }
		.caption-ornament::before { right: 9px; }
		.caption-ornament::after { width: 4px; height: 4px; }
		.hero-scroll { bottom: 26px; }
		.hero-signature { right: 20px; bottom: 49px; gap: 8px; font-size: 13px; }
		.hero-signature::before, .hero-signature::after { width: 12px; }
	}
	@media (max-height: 720px) and (min-width: 769px) {
		.brand-mark { width: 390px; }
	}
	@media (prefers-reduced-motion: reduce) {
		.brand-mark, .letter-ink path, .scroll-circle > span, .hero-caption-line { animation: none; }
		.hero-content { transform: none; }
		.letter-ink path { stroke-dashoffset: 0; }
	}
</style>
