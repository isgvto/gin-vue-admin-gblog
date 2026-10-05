import {nearRoadRight, nearRoadLeft, traceRoadCurve} from './landscapeGeometry'

// A quiet layered landscape: distant light, a winding passage and moving air.
export function createHeroScene(canvas, header, options = {}) {
	const context = canvas.getContext('2d')
	if (!context) return {destroy() {}}
	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
	let width = 0, height = 0, time = 0, lastFrame = 0
	let frameId = null, visible = true, destroyed = false
	const pointer = {x: 0, y: 0}, target = {x: 0, y: 0}
	const tau = Math.PI * 2
	let brandX = null, brandY = null

	function positionBrand() {
		if (!options.brand) return
		const x = reducedMotion.matches ? 0 : Number((pointer.x * 6).toFixed(3))
		const y = reducedMotion.matches ? 0 : Number((pointer.y * 4).toFixed(3))
		if (x !== brandX) options.brand.style.setProperty('--brand-x', `${x}px`)
		if (y !== brandY) options.brand.style.setProperty('--brand-y', `${y}px`)
		brandX = x
		brandY = y
	}

	function breeze(position, detail = 0) {
		if (reducedMotion.matches) return 0
		const localTime = time - position * .8
		const gust = .7 + Math.sin(time * .27 - .9) * .3
		return gust * (Math.sin(localTime * .85) * .72 + Math.sin(localTime * 1.73 + detail * .15) * .2 + Math.sin(localTime * 2.91 + detail * .3) * .08)
	}
	function candlelight() {
		if (reducedMotion.matches) return .75
		return Math.max(.3, Math.min(1, .65 + Math.sin(time * 2.1) * .18 + Math.sin(time * 5.37 + 1.4) * .1 + Math.sin(time * 8.63 + .7) * .07 + Math.sin(time * 13.1) * .04))
	}

	function layer(depth, draw) {
		context.save()
		context.translate(pointer.x * depth, pointer.y * depth * .35)
		draw(context)
		context.restore()
	}
	function haze(ctx, x, y, radius, color, alpha) {
		const light = ctx.createRadialGradient(x, y, 0, x, y, radius)
		light.addColorStop(0, `rgba(${color},${alpha})`)
		light.addColorStop(1, `rgba(${color},0)`)
		ctx.fillStyle = light
		ctx.fillRect(x - radius, y - radius, radius * 2, radius * 2)
	}
	function ridgeHeight(x, level, amplitude, phase) {
		const u = (x + 40) / (width + 80)
		return height * (level + amplitude * (Math.sin(u * 7.4 + phase) + Math.cos(u * 12.6 + phase) * .28))
	}
	function ridge(ctx, level, amplitude, phase, color) {
		ctx.beginPath()
		ctx.moveTo(-40, height)
		for (let i = 0; i <= 100; i++) {
			const u = i / 100
			const x = u * (width + 80) - 40
			const y = ridgeHeight(x, level, amplitude, phase)
			ctx.lineTo(x, y)
		}
		ctx.lineTo(width + 40, height)
		ctx.closePath()
		ctx.fillStyle = color
		ctx.fill()
	}
	function paintMoon(ctx) {
		const x = width * (width < 600 ? .82 : .79), y = height * .29
		const radius = Math.min(width * .057, 72)
		haze(ctx, x, y, radius * 2.8, '176,202,224', .13)
		ctx.save()
		ctx.beginPath()
		ctx.arc(x, y, radius, 0, tau)
		ctx.clip()

		// Earthshine keeps the shadowed face visible in the pale daytime sky.
		const shadow = ctx.createLinearGradient(x - radius, y, x + radius, y + radius)
		shadow.addColorStop(0, '#d4e3ef')
		shadow.addColorStop(.6, '#e3edf5')
		shadow.addColorStop(1, '#eff5fa')
		ctx.fillStyle = shadow
		ctx.fillRect(x - radius, y - radius, radius * 2, radius * 2)

		const moonlight = ctx.createLinearGradient(x, y - radius, x + radius, y + radius)
		moonlight.addColorStop(0, '#fffdf6')
		moonlight.addColorStop(.55, '#fbfcfc')
		moonlight.addColorStop(1, '#f1f7fb')
		ctx.beginPath()
		ctx.arc(x, y, radius, -Math.PI / 2, Math.PI / 2)
		ctx.bezierCurveTo(x + radius * .62, y + radius * .5, x + radius * .62, y - radius * .5, x, y - radius)
		ctx.closePath()
		ctx.fillStyle = moonlight
		ctx.fill()

		// Soft impressions suggest a lunar surface without adding hard outlines.
		const impressions = [[-.38, -.25, .15], [-.12, .4, .12], [.16, -.48, .09], [.59, .19, .08]]
		for (const [dx, dy, size] of impressions) {
			haze(ctx, x + radius * dx, y + radius * dy, radius * size, '119,153,179', .08)
		}
		ctx.restore()

		// A broad translucent cloud softens the lower rim and follows the breeze.
		const drift = reducedMotion.matches ? 0 : Math.sin(time * .075 + .6) * radius * .22
		const veilX = x - radius * .18 + drift, veilY = y + radius * .4
		ctx.save()
		ctx.translate(veilX, veilY)
		ctx.scale(1, .16)
		haze(ctx, 0, 0, radius * 1.9, '248,251,254', .72)
		ctx.restore()
	}
	function paintSky(ctx) {
		const sky = ctx.createLinearGradient(0, 0, 0, height)
		sky.addColorStop(0, '#ffffff')
		sky.addColorStop(.48, '#f8fbfe')
		sky.addColorStop(.7, '#edf4fa')
		sky.addColorStop(1, '#ffffff')
		ctx.fillStyle = sky
		ctx.fillRect(0, 0, width, height)
		paintMoon(ctx)

		// Long, soft cloud shapes carry the wind without crossing the wordmark.
		for (let i = 0; i < 4; i++) {
			const x = width * (i % 2 ? .76 : .16) + Math.sin(time * .07 + i) * 15
			const y = height * (.19 + i * .064)
			const length = Math.min(width * .18, 230)
			const cloud = ctx.createLinearGradient(x - length, 0, x + length, 0)
			cloud.addColorStop(0, 'rgba(184,205,223,0)')
			cloud.addColorStop(.5, 'rgba(184,205,223,.22)')
			cloud.addColorStop(1, 'rgba(184,205,223,0)')
			ctx.strokeStyle = cloud
			ctx.lineWidth = 1
			ctx.beginPath()
			ctx.moveTo(x - length, y + 4)
			ctx.bezierCurveTo(x - length * .35, y - 7, x + length * .3, y + 6, x + length, y)
			ctx.stroke()
		}
		for (let i = 0; i < 3; i++) {
			const drift = Math.sin(time * .055) * width * .045
			const x = width * .24 + drift + i * (width < 600 ? 14 : 24)
			const y = height * .27 - i * 7 + Math.sin(time * .22 + i) * 3
			const flap = reducedMotion.matches ? -.35 : Math.sin(time * 4.2 + i * .85)
			const wing = (width < 600 ? 5.5 : 8.5) * (1 - Math.abs(flap) * .14)
			const lift = flap * wing * .68
			ctx.beginPath()
			ctx.moveTo(x - wing, y + lift)
			ctx.quadraticCurveTo(x - wing * .45, y + lift * .4 - 1, x, y)
			ctx.quadraticCurveTo(x + wing * .45, y + lift * .4 - 1, x + wing, y + lift)
			ctx.strokeStyle = 'rgba(49,72,94,.58)'
			ctx.lineWidth = width < 600 ? 1.15 : 1.35
			ctx.lineCap = 'round'
			ctx.stroke()
			ctx.beginPath()
			ctx.moveTo(x, y - .6)
			ctx.lineTo(x, y + 1.3)
			ctx.lineWidth = .9
			ctx.stroke()
		}
	}
	function paintSmoke(ctx, x, y, scale) {
		const smokeTime = reducedMotion.matches ? 3.8 : time
		for (let i = 0; i < 6; i++) {
			const age = ((smokeTime + i * 1.3) % 8) / 8
			const drift = age * (12 + breeze(.63) * 8) + Math.sin(age * 5 + smokeTime * .4) * age * 2
			const radius = (1.5 + age * 6) * scale
			ctx.beginPath()
			ctx.ellipse(x + drift * scale, y - age * 32 * scale, radius, radius * .7, -.3, 0, tau)
			ctx.fillStyle = `rgba(120,146,165,${Math.sin(age * Math.PI) * .14})`
			ctx.fill()
		}
	}
	function paintVillage(ctx) {
		const compact = width < 600
		const scale = compact ? .68 : 1
		const destination = {x: width * .63, y: height * .648}
		ctx.beginPath()
		ctx.ellipse(destination.x, destination.y + 2 * scale, 19 * scale, 3 * scale, 0, 0, tau)
		ctx.fillStyle = 'rgba(92,125,151,.13)'
		ctx.fill()
		const buildings = [
			{dx: -116, dy: 14, w: 22, h: 17},
			{dx: -88, dy: 10, w: 18, h: 23},
			{dx: -59, dy: 13, w: 30, h: 16},
			{dx: 0, dy: 0, w: 24, h: 29},
			{dx: 32, dy: 8, w: 18, h: 20},
			{dx: 55, dy: 13, w: 32, h: 15}
		]
		buildings.forEach((building, index) => {
			const x = destination.x + building.dx * scale
			const base = destination.y + building.dy * scale
			const w = building.w * scale, h = building.h * scale
			ctx.fillStyle = index === 3 ? '#516b82' : '#b3c7d7'
			ctx.beginPath()
			ctx.moveTo(x - w * .5, base)
			ctx.lineTo(x - w * .5, base - h * .7)
			ctx.lineTo(x, base - h)
			ctx.lineTo(x + w * .5, base - h * .7)
			ctx.lineTo(x + w * .5, base)
			ctx.closePath()
			ctx.fill()
			ctx.fillStyle = index === 3 ? '#71899c' : '#cbdbe7'
			ctx.fillRect(x + 1, base - h * .64, w * .42, h * .64)
			if (index === 3) {
				const chimneyX = x + w * .2, chimneyY = base - h * .99
				paintSmoke(ctx, chimneyX + 1.7 * scale, chimneyY, scale)
				ctx.fillStyle = '#72899b'
				ctx.fillRect(chimneyX, chimneyY, 3.4 * scale, 7 * scale)
				ctx.fillStyle = '#60798d'
				ctx.fillRect(chimneyX - .7 * scale, chimneyY, 4.8 * scale, 1.5 * scale)
				const windowY = base - h * .46
				const windowX = x - w * .28
				const flicker = candlelight()
				haze(ctx, x - w * .13, windowY, (18 + flicker * 14) * scale, '239,186,106', .12 + flicker * .4)
				ctx.fillStyle = `rgba(255,222,164,${.36 + flicker * .6})`
				ctx.fillRect(windowX, windowY - 4 * scale, 6 * scale, 8 * scale)
				const flutter = reducedMotion.matches ? 0 : Math.sin(time * 7.3) * .35 * scale
				const flameX = windowX + 3 * scale
				ctx.beginPath()
				ctx.moveTo(flameX, windowY + 2 * scale)
				ctx.bezierCurveTo(flameX - 1.3 * scale, windowY + scale, flameX - .8 * scale, windowY, flameX + flutter, windowY - (1.5 + flicker) * scale)
				ctx.bezierCurveTo(flameX + 1.2 * scale, windowY, flameX + 1.1 * scale, windowY + scale, flameX, windowY + 2 * scale)
				ctx.fillStyle = `rgba(255,245,208,${.5 + flicker * .4})`
				ctx.fill()
				ctx.strokeStyle = 'rgba(96,113,128,.5)'
				ctx.lineWidth = .65
				ctx.strokeRect(windowX, windowY - 4 * scale, 6 * scale, 8 * scale)
			} else {
				ctx.fillStyle = 'rgba(255,255,255,.6)'
				ctx.fillRect(x - w * .2, base - h * .45, 3 * scale, 4 * scale)
			}
		})
	}
	function paintPassage(ctx) {
		// The distant edge rests on the ground below the village, with a flat end.
		ctx.beginPath()
		ctx.moveTo(width * .619, height * .655)
		ctx.bezierCurveTo(width * .55, height * .658, width * .39, height * .7, width * .4, height * .77)
		traceRoadCurve(ctx, nearRoadRight, width, height)
		ctx.lineTo(width * .18, height * 1.02)
		traceRoadCurve(ctx, nearRoadLeft, width, height, true)
		ctx.bezierCurveTo(width * .32, height * .68, width * .58, height * .655, width * .629, height * .655)
		ctx.closePath()
		const passage = ctx.createLinearGradient(0, height * .64, 0, height)
		passage.addColorStop(0, '#f5f9fd')
		passage.addColorStop(.5, '#f8fbfe')
		passage.addColorStop(1, '#ffffff')
		ctx.fillStyle = passage
		ctx.fill()
		ctx.strokeStyle = 'rgba(93,130,161,.14)'
		ctx.lineWidth = 1
		ctx.stroke()
		ctx.save()
		ctx.clip()
		const reflectionY = height * (.78 + Math.sin(time * .13) * .016)
		const reflection = ctx.createLinearGradient(0, reflectionY - 40, 0, reflectionY + 40)
		reflection.addColorStop(0, 'rgba(211,226,240,0)')
		reflection.addColorStop(.5, 'rgba(211,226,240,.24)')
		reflection.addColorStop(1, 'rgba(211,226,240,0)')
		ctx.fillStyle = reflection
		ctx.fillRect(0, reflectionY - 40, width, 80)
		ctx.restore()
	}
	function paintRoadside(ctx) {
		const scale = width < 600 ? .65 : 1
		const posts = []
		for (let i = 0; i < 6; i++) {
			const t = .08 + i * .08, s = 1 - t
			// Follow the near road's right edge so the fence stays on the verge.
			const x = width * (s * s * s * .4 + 3 * s * s * t * .4 + 3 * s * t * t * .57 + t * t * t * .38 + .014)
			const y = height * (s * s * s * .77 + 3 * s * s * t * .84 + 3 * s * t * t * .85 + t * t * t * 1.02)
			const size = (7 + t * 14) * scale
			posts.push({x, y, size})
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.lineTo(x, y - size)
			ctx.strokeStyle = 'rgba(78,107,131,.32)'
			ctx.lineWidth = 1.4 * scale
			ctx.stroke()
		}
		for (const fraction of [.3, .7]) {
			ctx.beginPath()
			posts.forEach((post, index) => {
				if (index) ctx.lineTo(post.x, post.y - post.size * fraction)
				else ctx.moveTo(post.x, post.y - post.size * fraction)
			})
			ctx.strokeStyle = 'rgba(78,107,131,.23)'
			ctx.lineWidth = .9 * scale
			ctx.stroke()
		}
		for (const [u, v, size] of [[.555, .689, 10], [.427, .735, 15]]) {
			const x = width * u, y = height * v, lampHeight = size * scale
			ctx.strokeStyle = 'rgba(76,104,128,.37)'
			ctx.lineWidth = 1 * scale
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.lineTo(x, y - lampHeight)
			ctx.lineTo(x + 4 * scale, y - lampHeight)
			ctx.stroke()
			haze(ctx, x + 4 * scale, y - lampHeight + scale, 12 * scale, '236,195,133', .21)
			ctx.fillStyle = '#eed5ad'
			ctx.fillRect(x + 2 * scale, y - lampHeight, 4 * scale, 3 * scale)
		}
	}
	function paintBench(ctx) {
		const scale = width < 600 ? .72 : 1
		ctx.save()
		ctx.translate(width * .275, height * .802)
		ctx.scale(scale, scale)
		ctx.fillStyle = 'rgba(68,100,125,.08)'
		ctx.beginPath()
		ctx.ellipse(0, 3, 19, 2.3, 0, 0, tau)
		ctx.fill()
		ctx.strokeStyle = 'rgba(66,92,114,.43)'
		ctx.lineWidth = 1.3
		for (const x of [-10, 10]) {
			ctx.beginPath()
			ctx.moveTo(x, 2)
			ctx.lineTo(x, -16)
			ctx.stroke()
		}
		ctx.strokeStyle = 'rgba(105,135,156,.66)'
		ctx.lineWidth = 2.4
		for (const y of [-15, -11]) {
			ctx.beginPath()
			ctx.moveTo(-14, y)
			ctx.lineTo(12, y + .7)
			ctx.stroke()
		}
		ctx.beginPath()
		ctx.moveTo(-14, -7)
		ctx.lineTo(12, -6)
		ctx.lineTo(15, -3)
		ctx.lineTo(-12, -4)
		ctx.closePath()
		ctx.fillStyle = 'rgba(110,141,162,.65)'
		ctx.fill()
		ctx.restore()
	}
	function paintTree(ctx, x, y, size, phase, atmosphere = 1) {
		const wind = breeze(x / width, phase) * atmosphere
		const leafFlutter = reducedMotion.matches ? 0 : Math.sin(time * 2.3 + phase) * .015 * atmosphere
		ctx.save()
		ctx.globalAlpha *= atmosphere
		ctx.translate(x, y)
		// Bend above the root; canopy and trunk move together through each gust.
		ctx.transform(1, 0, -wind * .11, 1, 0, 0)
		ctx.strokeStyle = 'rgba(53,78,99,.36)'
		ctx.lineWidth = Math.max(.7, size * .028)
		ctx.beginPath()
		ctx.moveTo(0, 0)
		ctx.quadraticCurveTo(wind * size * .015, -size * .4, wind * size * .025, -size * .78)
		ctx.stroke()
		ctx.fillStyle = 'rgba(63,93,116,.18)'
		ctx.beginPath()
		ctx.ellipse(wind * size * .018, -size * .66, size * .21, size * .38, -.13 + wind * .06 + leafFlutter, 0, tau)
		ctx.fill()
		ctx.fillStyle = 'rgba(96,128,153,.19)'
		ctx.beginPath()
		ctx.ellipse(size * .09 + wind * size * .012, -size * .56, size * .2, size * .27, .24 + wind * .05 - leafFlutter, 0, tau)
		ctx.fill()
		ctx.restore()
	}
	function paintDistantTrees(ctx, level, amplitude, phase, placements, atmosphere) {
		const scale = width < 600 ? .7 : 1
		placements.forEach(([position, size], index) => {
			const x = width * position
			paintTree(ctx, x, ridgeHeight(x, level, amplitude, phase), size * scale, index, atmosphere)
		})
	}
	function paintWildflowers(ctx, scale) {
		for (const [u, v] of [[.17, .825], [.215, .854], [.735, .836], [.795, .855]]) {
			for (let flower = 0; flower < 3; flower++) {
				const x = width * u + (flower - 1) * 7 * scale
				const y = height * v + (flower % 2) * 3 * scale
				const stem = (9 + flower * 2) * scale
				const wind = breeze(u, flower)
				const tip = {x: x + wind * stem * .23, y: y - stem}
				ctx.strokeStyle = 'rgba(84,116,136,.3)'
				ctx.lineWidth = .8 * scale
				ctx.beginPath()
				ctx.moveTo(x, y)
				ctx.quadraticCurveTo(x + wind * stem * .12, y - stem * .5, tip.x, tip.y)
				ctx.stroke()
				ctx.fillStyle = flower === 1 ? '#e6ddc9' : '#e5f0f8'
				for (let petal = 0; petal < 5; petal++) {
					const angle = petal / 5 * tau + wind * .15
					ctx.beginPath()
					ctx.ellipse(tip.x + Math.cos(angle) * 1.6 * scale, tip.y + Math.sin(angle) * 1.6 * scale, 1.4 * scale, .85 * scale, angle, 0, tau)
					ctx.fill()
				}
				ctx.beginPath()
				ctx.arc(tip.x, tip.y, .75 * scale, 0, tau)
				ctx.fillStyle = '#a4bbcb'
				ctx.fill()
			}
		}
	}
	function paintForeground(ctx) {
		const scale = width < 600 ? .65 : 1
		for (const [x, y, size, phase] of [[.1, .78, 62, 0], [.15, .76, 40, 2], [.82, .78, 54, 1], [.88, .8, 73, 3], [.94, .81, 40, 4]]) {
			paintTree(ctx, width * x, height * y, size * scale, phase)
		}
		for (let i = 0; i < 32; i++) {
			const side = i < 16
			const u = (i % 16) / 15
			const x = width * (side ? .025 + u * .23 : .71 + u * .26)
			const y = height * (.88 + Math.sin(i * 2.7) * .035)
			const length = (9 + (i % 5) * 3) * scale
			const sway = breeze(x / width, i * .2) * length * .22
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.quadraticCurveTo(x + sway, y - length * .6, x + sway + 3, y - length)
			ctx.strokeStyle = `rgba(58,85,107,${.12 + i % 3 * .035})`
			ctx.lineWidth = .8
			ctx.stroke()
		}
		paintWildflowers(ctx, scale)
		for (let i = 0; i < 8; i++) {
			const x = width * (.06 + i * .115)
			const y = height * (.71 + Math.sin(i * 2.2) * .015)
			ctx.fillStyle = 'rgba(87,122,151,.17)'
			ctx.beginPath()
			ctx.ellipse(x, y, 4 + i % 3, 1.3, -.15, 0, tau)
			ctx.fill()
		}
	}
	function paint() {
		if (!width || !height || destroyed) return
		positionBrand()
		const ctx = context
		ctx.clearRect(0, 0, width, height)
		layer(4, paintSky)
		layer(7, () => {
			ridge(ctx, .615, .031, .9, '#e8f0f7')
			paintDistantTrees(ctx, .615, .031, .9, [[.075, 22], [.1, 27], [.78, 25], [.81, 20]], .38)
			ridge(ctx, .66, .036, 3.1, '#dce8f2')
			paintDistantTrees(ctx, .66, .036, 3.1, [[.34, 29], [.365, 22], [.73, 30]], .55)
		})
		layer(11, () => {
			ridge(ctx, .72, .049, 1.8, '#cedeea')
		})
		layer(17, () => {
			ridge(ctx, .81, .043, .2, '#bdd3e3')
			ridge(ctx, .88, .036, 3.8, '#a9c5da')
		})
		// The road and village share parallax; buildings always sit above the road.
		layer(11, () => {
			paintPassage(ctx)
			paintRoadside(ctx)
			paintBench(ctx)
			paintVillage(ctx)
			const mistX = width * .48 + Math.sin(time * .09) * width * .04
			haze(ctx, mistX, height * .65, width * .24, '255,255,255', .48)
		})
		layer(17, paintForeground)
		const fade = ctx.createLinearGradient(0, height * .87, 0, height)
		fade.addColorStop(0, 'rgba(255,255,255,0)')
		fade.addColorStop(1, '#ffffff')
		ctx.fillStyle = fade
		ctx.fillRect(0, height * .87, width, height * .13)
	}
	function frame(now) {
		frameId = null
		if (destroyed || !visible || document.hidden || reducedMotion.matches) return
		const delta = lastFrame ? Math.min((now - lastFrame) / 1000, .05) : 0
		lastFrame = now
		time += delta
		const follow = 1 - Math.exp(-delta * 4)
		pointer.x += (target.x - pointer.x) * follow
		pointer.y += (target.y - pointer.y) * follow
		paint()
		frameId = window.requestAnimationFrame(frame)
	}
	function sync() {
		if (destroyed) return
		if (!visible || document.hidden || reducedMotion.matches) {
			if (frameId !== null) window.cancelAnimationFrame(frameId)
			frameId = null
			lastFrame = 0
			if (reducedMotion.matches) {
				pointer.x = target.x = 0
				pointer.y = target.y = 0
				paint()
			}
		} else if (frameId === null) {
			lastFrame = 0
			frameId = window.requestAnimationFrame(frame)
		}
		header.classList.toggle('hero-paused', !visible || document.hidden || reducedMotion.matches)
	}
	function resize() {
		const bounds = header.getBoundingClientRect()
		width = bounds.width
		height = bounds.height
		const ratio = Math.min(window.devicePixelRatio || 1, 1.5)
		canvas.width = Math.round(width * ratio)
		canvas.height = Math.round(height * ratio)
		context.setTransform(ratio, 0, 0, ratio, 0, 0)
		paint()
	}
	function move(event) {
		if (event.pointerType === 'touch' || reducedMotion.matches || !width || !height) return
		const bounds = header.getBoundingClientRect()
		target.x = Math.max(-.5, Math.min(.5, (event.clientX - bounds.left) / bounds.width - .5))
		target.y = Math.max(-.5, Math.min(.5, (event.clientY - bounds.top) / bounds.height - .5))
	}
	function leave() { target.x = 0; target.y = 0 }
	const resizeObserver = new ResizeObserver(resize)
	resizeObserver.observe(header)
	const visibilityObserver = new IntersectionObserver(entries => {
		visible = entries[0].isIntersecting && entries[0].intersectionRatio > .05
		sync()
	}, {threshold: .05})
	visibilityObserver.observe(header)
	header.addEventListener('pointermove', move, {passive: true})
	header.addEventListener('pointerleave', leave)
	document.addEventListener('visibilitychange', sync)
	if (reducedMotion.addEventListener) reducedMotion.addEventListener('change', sync)
	else reducedMotion.addListener(sync)
	resize()
	sync()
	return {
		destroy() {
			destroyed = true
			if (frameId !== null) window.cancelAnimationFrame(frameId)
			if (options.brand) {
				options.brand.style.removeProperty('--brand-x')
				options.brand.style.removeProperty('--brand-y')
			}
			resizeObserver.disconnect()
			visibilityObserver.disconnect()
			header.removeEventListener('pointermove', move)
			header.removeEventListener('pointerleave', leave)
			document.removeEventListener('visibilitychange', sync)
			if (reducedMotion.removeEventListener) reducedMotion.removeEventListener('change', sync)
			else reducedMotion.removeListener(sync)
		}
	}
}
