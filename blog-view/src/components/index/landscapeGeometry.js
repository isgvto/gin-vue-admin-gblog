// Shared road edges keep the hero's far view and the footer's near view aligned.
export const nearRoadRight = [[.4, .77], [.4, .84], [.57, .85], [.38, 1.02]]
export const nearRoadLeft = [[.3, .76], [.27, .83], [.46, .85], [.18, 1.02]]

export function traceRoadCurve(ctx, points, width, height, reverse = false) {
	const p = reverse ? points.slice().reverse() : points
	ctx.bezierCurveTo(p[1][0] * width, p[1][1] * height, p[2][0] * width, p[2][1] * height, p[3][0] * width, p[3][1] * height)
}

export function roadJoin(points) {
	const value = (t, axis) => (1-t)**3 * points[0][axis] + 3*(1-t)**2*t * points[1][axis] + 3*(1-t)*t*t * points[2][axis] + t**3 * points[3][axis]
	const derivative = (t, axis) => 3*(1-t)**2*(points[1][axis]-points[0][axis]) + 6*(1-t)*t*(points[2][axis]-points[1][axis]) + 3*t*t*(points[3][axis]-points[2][axis])
	let t = .96
	for (let i = 0; i < 8; i++) t -= (value(t, 1) - 1) / derivative(t, 1)
	return {x: value(t, 0), slope: derivative(t, 0) / derivative(t, 1)}
}
