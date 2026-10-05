import {nearRoadLeft, nearRoadRight, roadJoin} from './landscapeGeometry'

// A static continuation below the hero: no new horizon, just the nearby verge.
export function createFooterScene(canvas, footer) {
	const ctx = canvas.getContext('2d')
	if (!ctx) return {destroy() {}}
	const left = roadJoin(nearRoadLeft), right = roadJoin(nearRoadRight)
	let puddle = null
	function draw() {
		const {width: w, height: h} = footer.getBoundingClientRect()
		if (!w || !h) return
		const ratio = Math.min(window.devicePixelRatio || 1, 1.5)
		canvas.width = Math.round(w * ratio)
		canvas.height = Math.round(h * ratio)
		ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
		const ground = ctx.createLinearGradient(0, 0, 0, h)
		ground.addColorStop(0, '#a9c5da')
		ground.addColorStop(1, '#b7cddd')
		ctx.fillStyle = ground
		ctx.fillRect(0, 0, w, h)
		ctx.fillStyle = 'rgba(121,160,187,.14)'
		ctx.beginPath()
		ctx.moveTo(w * .53, h)
		ctx.bezierCurveTo(w * .8, h * .84, w * .7, h * .37, w, h * .16)
		ctx.lineTo(w, h)
		ctx.fill()

		const heroHeight = document.querySelector('.home-hero')?.getBoundingClientRect().height || Math.max(window.innerHeight, 600)
		const tangentScale = h * .18 / heroHeight
		const leftEdge = [[left.x * w, 0], [(left.x + left.slope * tangentScale) * w, h * .18], [-w * .09, h * .6], [-w * .22, h]]
		const rightEdge = [[right.x * w, 0], [(right.x + right.slope * tangentScale) * w, h * .18], [w * .14, h * .6], [w * .19, h]]
		ctx.beginPath()
		ctx.moveTo(right.x * w, 0)
		ctx.bezierCurveTo((right.x + right.slope * tangentScale) * w, h * .18, w * .14, h * .6, w * .19, h)
		ctx.lineTo(-w * .22, h)
		ctx.bezierCurveTo(-w * .09, h * .6, (left.x + left.slope * tangentScale) * w, h * .18, left.x * w, 0)
		ctx.closePath()
		const road = ctx.createLinearGradient(0, 0, 0, h)
		road.addColorStop(0, '#fff')
		road.addColorStop(1, '#f4f8fc')
		ctx.fillStyle = road
		ctx.fill()
		ctx.strokeStyle = 'rgba(93,130,161,.14)'
		ctx.lineWidth = 1
		ctx.stroke()
		// A shallow puddle rests inside the road, separate from the text on the verge.
		ctx.save()
		ctx.clip()
		const a = curvePoint(leftEdge, .56), b = curvePoint(rightEdge, .56)
		const radius = Math.min(76, (b.x - a.x) * .32)
		puddle = {x: Math.max(radius + 8, (a.x + b.x) / 2), y: a.y, rx: radius, ry: Math.max(6, radius * .23), angle: -.09}
		paintPuddle(puddle)
		ctx.restore()

		// Sparse, larger grass blades stay near the edges, away from reading lines.
		for (let i = 0; i < 42; i++) {
			const u = i / 41
			const x = w * (i < 14 ? .01 + u * .22 : .72 + (u - .34) * .42)
			const y = h * (.81 + Math.sin(i * 2.7) * .13)
			const length = 13 + i % 5 * 4
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.quadraticCurveTo(x - 4, y - length * .6, x + (i % 2 ? -8 : 7), y - length)
			ctx.strokeStyle = 'rgba(58,85,107,.19)'
			ctx.lineWidth = .9
			ctx.stroke()
		}
		const compact = w < 768
		for (const [x, y, size] of compact ? [[1.02, .74, 116], [.015, .3, 60]] : [[.955, .7, 185], [.035, .28, 100], [.99, .82, 125]]) {
			tree(x * w, y * h, size)
		}
		for (const [u, v] of [[.045, .72], [.92, .91], [.96, .88]]) {
			for (let i = 0; i < 3; i++) flower(u * w + i * 9, v * h + i % 2 * 6, 19 + i * 3)
		}
		// Match the white mist at the bottom of the hero; the scene continues beneath it.
		const mist = ctx.createLinearGradient(0, 0, 0, Math.min(h * .24, 80))
		mist.addColorStop(0, '#fff')
		mist.addColorStop(1, 'rgba(255,255,255,0)')
		ctx.fillStyle = mist
		ctx.fillRect(0, 0, w, Math.min(h * .24, 80))
	}
	function curvePoint(points, t) {
		const coordinate = axis => (1-t)**3 * points[0][axis] + 3*(1-t)**2*t * points[1][axis] + 3*(1-t)*t*t * points[2][axis] + t**3 * points[3][axis]
		return {x: coordinate(0), y: coordinate(1)}
	}
	function paintPuddle(p) {
		ctx.save()
		ctx.translate(p.x, p.y)
		ctx.rotate(p.angle)
		ctx.scale(p.rx, p.ry)
		ctx.beginPath()
		ctx.moveTo(-1, .06)
		ctx.bezierCurveTo(-.91, -.5, -.43, -.84, -.04, -.7)
		ctx.bezierCurveTo(.34, -.62, .62, -.93, .93, -.25)
		ctx.bezierCurveTo(1.14, .2, .63, .67, .17, .68)
		ctx.bezierCurveTo(-.19, .84, -.97, .71, -1, .06)
		ctx.closePath()
		const water = ctx.createLinearGradient(0, -1, 0, 1)
		water.addColorStop(0, '#c8dce9')
		water.addColorStop(.48, '#e5eff6')
		water.addColorStop(1, '#bfd5e5')
		ctx.fillStyle = water
		ctx.fill()
		ctx.strokeStyle = 'rgba(106,143,173,.3)'
		ctx.lineWidth = .018
		ctx.stroke()
		ctx.clip()
		ctx.fillStyle = 'rgba(91,127,156,.09)'
		ctx.fillRect(.25, -.8, .07, 1.45)
		ctx.fillRect(.38, -.8, .12, 1.2)
		ctx.strokeStyle = 'rgba(255,255,255,.65)'
		ctx.lineWidth = .035
		for (const [x, y, length] of [[-.68, -.15, .38], [.12, .17, .52], [-.34, .46, .31]]) {
			ctx.beginPath()
			ctx.moveTo(x, y)
			ctx.quadraticCurveTo(x + length * .5, y - .06, x + length, y)
			ctx.stroke()
		}
		ctx.restore()
	}
	function tree(x, y, size) {
		ctx.strokeStyle = 'rgba(53,78,99,.36)'
		ctx.lineWidth = Math.max(1, size * .028)
		ctx.beginPath()
		ctx.moveTo(x, y)
		ctx.quadraticCurveTo(x, y - size * .4, x, y - size * .78)
		ctx.stroke()
		ctx.fillStyle = 'rgba(63,93,116,.18)'
		ctx.beginPath()
		ctx.ellipse(x, y - size * .66, size * .21, size * .38, -.13, 0, Math.PI * 2)
		ctx.fill()
		ctx.fillStyle = 'rgba(96,128,153,.19)'
		ctx.beginPath()
		ctx.ellipse(x + size * .09, y - size * .56, size * .2, size * .27, .24, 0, Math.PI * 2)
		ctx.fill()
	}
	function flower(x, y, stem, scale = 1) {
		ctx.strokeStyle = 'rgba(84,116,136,.4)'
		ctx.lineWidth = Math.max(1, scale * .75)
		ctx.beginPath()
		ctx.moveTo(x, y)
		ctx.quadraticCurveTo(x - 1, y - stem * .5, x + 3, y - stem)
		ctx.stroke()
		if (scale > 1.2) {
			ctx.fillStyle = 'rgba(84,116,136,.26)'
			for (const side of [-1, 1]) {
				ctx.beginPath()
				ctx.ellipse(x + side * 4 * scale, y - stem * (.3 + (side + 1) * .12), 5 * scale, 1.6 * scale, side * -.65, 0, Math.PI * 2)
				ctx.fill()
			}
		}
		ctx.fillStyle = '#edf4fa'
		for (let i = 0; i < 5; i++) {
			const angle = i / 5 * Math.PI * 2
			ctx.beginPath()
			ctx.ellipse(x + 3 + Math.cos(angle) * 2.6 * scale, y - stem + Math.sin(angle) * 2.6 * scale, 2.4 * scale, 1.5 * scale, angle, 0, Math.PI * 2)
			ctx.fill()
		}
		ctx.fillStyle = '#c6bda7'
		ctx.beginPath()
		ctx.arc(x + 3, y - stem, 1.4 * scale, 0, Math.PI * 2)
		ctx.fill()
	}
	const observer = new ResizeObserver(draw)
	observer.observe(footer)
	window.addEventListener('resize', draw)
	draw()
	return {getPuddle() {return puddle}, destroy() {observer.disconnect(); window.removeEventListener('resize', draw)}}
}
