<template>
	<div class="footer-landscape" aria-hidden="true">
		<canvas ref="scene"></canvas>
		<canvas ref="rain" class="footer-rain"></canvas>
	</div>
</template>

<script>
	import {createFooterScene} from './footerScene'
	import {createRainScene} from './rainScene'
	export default {
		name: 'FooterLandscape',
		mounted() {
			this.scene = createFooterScene(this.$refs.scene, this.$el.parentElement)
			this.rain = createRainScene(this.$refs.rain, this.$el.parentElement, {getPuddle: () => this.scene.getPuddle ? this.scene.getPuddle() : null})
		},
		beforeDestroy() {
			if (this.scene) this.scene.destroy()
			if (this.rain) this.rain.destroy()
		}
	}
</script>

<style scoped>
	.footer-landscape { position: absolute; inset: 0; display: block; width: 100%; height: 100%; pointer-events: none; }
	.footer-landscape canvas { position: absolute; inset: 0; display: block; width: 100%; height: 100%; }
</style>
