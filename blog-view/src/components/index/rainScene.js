// Both views share the same quiet rain; only the foreground has water impacts.
export function createRainScene(canvas, surface, options = {}) {
	const ctx = canvas.getContext('2d')
	if (!ctx) return {destroy() {}}
	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
	let width = 0, height = 0, time = 0, lastFrame = 0, frameId = null
	let visible = false, destroyed = false
	const seed = n => {const value = Math.sin(n * 127.1 + 311.7) * 43758.5453; return value - Math.floor(value)}
	const wrap = (value, max) => ((value % max) + max) % max

	function paint() {
		ctx.clearRect(0, 0, width, height)
		const count = Math.min(130, Math.max(26, Math.round(width * height / 13500)))
		for (let i = 0; i < count; i++) {
			const depth = .4 + seed(i + 2) * .6
			const x = wrap(seed(i + 31) * (width + 40) - time * (8 + depth * 10), width + 40) - 20
			const y = wrap(seed(i + 61) * (height + 60) + time * (75 + depth * 65), height + 60) - 30
			const length = 5 + depth * 10
			const nearBrand = options.hero && x > width * .27 && x < width * .73 && y > height * .22 && y < height * .53
			ctx.strokeStyle = `rgba(80,116,145,${(.07 + depth * .11) * (nearBrand ? .4 : 1)})`
			ctx.lineWidth = .45 + depth * .35
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.lineTo(x - length * .14, y + length)
			ctx.stroke()
		}
		const puddle = options.getPuddle ? options.getPuddle() : null
		if (puddle) paintWater(puddle)
	}
	function paintWater(p) {
		ctx.save()
		ctx.translate(p.x, p.y)
		ctx.rotate(p.angle)
		// Perspective keeps the ripples flat on the road instead of upright circles.
		ctx.save()
		ctx.beginPath()
		ctx.ellipse(0, 0, p.rx * .92, p.ry * .65, 0, 0, Math.PI * 2)
		ctx.clip()
		const impacts = []
		for (let i = 0; i < 5; i++) {
			const period = 2.7 + seed(i + 130) * .8
			const phase = (time + i * .73) % period
			const x = (seed(i + 150) - .5) * p.rx * 1.1
			const y = (seed(i + 170) - .5) * p.ry * .55
			if (phase > 1.5) continue
			const age = phase / 1.5
			const radius = 2 + age * p.rx * .28
			ctx.strokeStyle = `rgba(86,132,168,${(1-age) * .34})`
			ctx.lineWidth = .7
			ctx.beginPath()
			ctx.ellipse(x, y, radius, radius * .23, 0, 0, Math.PI * 2)
			ctx.stroke()
			ctx.strokeStyle = `rgba(255,255,255,${(1-age) * .6})`
			ctx.beginPath()
			ctx.ellipse(x, y + .7, radius + 1, radius * .23, 0, 0, Math.PI)
			ctx.stroke()
			if (phase < .3) impacts.push({x, y, age: phase / .3})
		}
		ctx.restore()
		if (!reducedMotion.matches) for (const impact of impacts) {
			const lift = Math.sin(impact.age * Math.PI) * (p.rx < 40 ? 3 : 5)
			ctx.strokeStyle = `rgba(109,148,179,${.4 * (1-impact.age)})`
			ctx.lineWidth = .65
			ctx.beginPath()
			ctx.moveTo(impact.x - 3, impact.y)
			ctx.quadraticCurveTo(impact.x - 4, impact.y - lift, impact.x - 1, impact.y - lift * .6)
			ctx.moveTo(impact.x + 3, impact.y)
			ctx.quadraticCurveTo(impact.x + 4, impact.y - lift, impact.x + 1, impact.y - lift * .6)
			ctx.stroke()
			ctx.fillStyle = `rgba(109,148,179,${.38 * (1-impact.age)})`
			for (const side of [-1, 1]) {
				ctx.beginPath()
				ctx.arc(impact.x + side * (2 + impact.age * 3), impact.y - lift * 1.35, .65, 0, Math.PI * 2)
				ctx.fill()
			}
		}
		ctx.restore()
	}
	function frame(now) {
		frameId = null
		if (destroyed || !visible || document.hidden || reducedMotion.matches) return
		if (!lastFrame || now - lastFrame >= 1000 / 30) {
			time += lastFrame ? Math.min((now - lastFrame) / 1000, .08) : 0
			lastFrame = now
			paint()
		}
		frameId = window.requestAnimationFrame(frame)
	}
	function sync() {
		if (frameId !== null) window.cancelAnimationFrame(frameId)
		frameId = null
		lastFrame = 0
		if (destroyed) return
		if (reducedMotion.matches) paint()
		else if (visible && !document.hidden) frameId = window.requestAnimationFrame(frame)
	}
	function resize() {
		const bounds = surface.getBoundingClientRect()
		width = bounds.width
		height = bounds.height
		const ratio = Math.min(window.devicePixelRatio || 1, 1.5)
		canvas.width = Math.round(width * ratio)
		canvas.height = Math.round(height * ratio)
		ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
		paint()
	}
	const resizeObserver = new ResizeObserver(resize)
	resizeObserver.observe(surface)
	const visibilityObserver = new IntersectionObserver(entries => {visible = entries[0].isIntersecting; sync()})
	visibilityObserver.observe(surface)
	document.addEventListener('visibilitychange', sync)
	reducedMotion.addEventListener('change', sync)
	resize()
	return {destroy() {
		destroyed = true
		if (frameId !== null) window.cancelAnimationFrame(frameId)
		resizeObserver.disconnect()
		visibilityObserver.disconnect()
		document.removeEventListener('visibilitychange', sync)
		reducedMotion.removeEventListener('change', sync)
	}}
}
