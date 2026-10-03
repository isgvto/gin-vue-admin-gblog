<template>
	<div class="admin-page admin-page--cards dashboard-page">
    <AdminPageHeading title="工作台" description="查看内容统计、访问趋势与最近动态。" />
		<!-- 统计卡片 -->
		<el-row :gutter="16" class="stat-row">
			<el-col v-for="item in statCards" :key="item.label" :xs="12" :sm="12" :md="6">
				<div class="stat-card" :class="item.theme">
					<div class="stat-icon">
						<el-icon :size="26">
							<component :is="item.icon" />
						</el-icon>
					</div>
					<div class="stat-info">
						<div class="stat-label">{{ item.label }}</div>
						<div class="stat-num">{{ item.value }}</div>
					</div>
				</div>
			</el-col>
		</el-row>

		<!-- 图表区 -->
		<el-row :gutter="16" class="chart-row">
			<el-col :xs="24" :lg="8">
				<el-card shadow="hover" class="chart-card">
					<template #header>
						<div class="chart-title">
							<el-icon><Collection /></el-icon>
							<span>分类下文章数量</span>
						</div>
					</template>
					<div ref="categoryEcharts" class="chart-body"></div>
				</el-card>
			</el-col>
			<el-col :xs="24" :lg="8">
				<el-card shadow="hover" class="chart-card">
					<template #header>
						<div class="chart-title">
							<el-icon><PriceTag /></el-icon>
							<span>标签下文章数量</span>
						</div>
					</template>
					<div ref="tagEcharts" class="chart-body"></div>
				</el-card>
			</el-col>
			<el-col :xs="24" :lg="8">
				<el-card shadow="hover" class="chart-card">
					<template #header>
						<div class="chart-title">
							<el-icon><Location /></el-icon>
							<span>访客地图</span>
						</div>
					</template>
					<div ref="mapEcharts" class="chart-body"></div>
				</el-card>
			</el-col>
		</el-row>

		<!-- 访问趋势 -->
		<el-card shadow="hover" class="chart-card trend-card">
			<template #header>
				<div class="chart-title">
					<el-icon><TrendCharts /></el-icon>
					<span>近 30 天访问趋势</span>
				</div>
			</template>
			<div ref="visitRecordEcharts" class="chart-body trend-body"></div>
		</el-card>
	</div>
</template>

<script>
	import * as echarts from 'echarts'
	import {
		Collection,
		ChatDotRound,
		Document,
		Location,
		PriceTag,
		TrendCharts,
		User,
		View
	} from '@element-plus/icons-vue'
	import { getDashboard } from '@/api/blog/dashboard'
	import chinaMap from '@/assets/map/china.json'
	import geoCoordMap from '@/assets/map/city2coord.json'

	echarts.registerMap('china', chinaMap)

	export default {
		name: 'BlogDashboard',
		components: {},
		data() {
			return {
				pv: 0,
				uv: 0,
				blogCount: 0,
				commentCount: 0,
				categoryEcharts: null,
				tagEcharts: null,
				mapEcharts: null,
				visitRecordEcharts: null,
				categoryOption: {
					tooltip: {
						trigger: 'item',
						formatter: '{a} <br/>{b} : {c} ({d}%)'
					},
					legend: {
						left: 'center',
						top: 'bottom',
						data: []
					},
					series: [
						{
							name: '文章数量',
							type: 'pie',
							radius: ['35%', '70%'],
							center: ['50%', '45%'],
							roseType: 'area',
							itemStyle: {
								borderRadius: 6,
								borderColor: '#fff',
								borderWidth: 2
							},
							data: []
						}
					]
				},
				tagOption: {
					tooltip: {
						trigger: 'item',
						formatter: '{a} <br/>{b} : {c} ({d}%)'
					},
					legend: {
						left: 'center',
						top: 'bottom',
						data: []
					},
					series: [
						{
							name: '文章数量',
							type: 'pie',
							radius: ['35%', '70%'],
							center: ['50%', '45%'],
							roseType: 'area',
							itemStyle: {
								borderRadius: 6,
								borderColor: '#fff',
								borderWidth: 2
							},
							data: []
						}
					]
				},
				//地图效果 reference https://www.jianshu.com/p/028525cbd080
				//reference https://echarts.apache.org/examples/zh/editor.html?c=map-polygon
				mapOption: {
					tooltip: {
						show: false
					},
					geo: {
						map: 'china',
						roam: false, //关闭拖拽
						zoom: 1.24,
						center: [104.2, 36], //调整地图位置
						label: {
							show: false,
							fontSize: 10,
							color: 'rgba(0,0,0,0.7)'
						},
						emphasis: {
							label: { show: false },
							areaColor: '#184cff',
							shadowOffsetX: 0,
							shadowOffsetY: 0,
							shadowBlur: 5,
							borderWidth: 0,
							shadowColor: 'rgba(0, 0, 0, 0.5)'
						},
						itemStyle: {
							areaColor: '#0d0059',
							borderColor: '#389dff',
							borderWidth: 1, //设置外层边框
							shadowBlur: 5,
							shadowOffsetY: 8,
							shadowOffsetX: 0,
							shadowColor: '#01012a'
						}
					},
					series: [
						{
							type: 'map',
							map: 'china',
							roam: false,
							zoom: 1.24,
							center: [104.2, 36],
							showLegendSymbol: false,
							label: {
								show: false
							},
							emphasis: {
								label: { show: false }
							},
							itemStyle: {
								areaColor: '#0d0059',
								borderColor: '#389dff',
								borderWidth: 0.5
							},
							emphasis: {
								areaColor: '#17008d',
								shadowOffsetX: 0,
								shadowOffsetY: 0,
								shadowBlur: 5,
								borderWidth: 0,
								shadowColor: 'rgba(0, 0, 0, 0.5)'
							}
						},
						{
							name: '',
							type: 'scatter',
							coordinateSystem: 'geo',
							data: [],
							symbol: 'circle',
							symbolSize: 5,
							hoverSymbolSize: 10,
							tooltip: {
								formatter(value) {
									return value.data.name + '<br/>' + '访客数：' + value.data.uv
								},
								show: true
							},
							encode: {
								value: 2
							},
							label: {
								formatter: '{b}',
								position: 'right',
								show: false
							},
							itemStyle: {
								color: '#0efacc'
							},
							emphasis: {
								label: {
									show: false
								}
							}
						},
						{
							name: 'Top 5',
							type: 'effectScatter',
							coordinateSystem: 'geo',
							data: [],
							symbol: 'circle',
							symbolSize: 12,
							tooltip: {
								formatter(value) {
									return value.data.name + '<br/>' + '访客数：' + value.data.uv
								},
								show: true
							},
							encode: {
								value: 2
							},
							showEffectOn: 'render',
							rippleEffect: {
								brushType: 'stroke',
								color: '#0efacc',
								period: 9,
								scale: 5
							},
							hoverAnimation: true,
							label: {
								formatter: '{b}',
								position: 'right',
								show: true
							},
							itemStyle: {
								color: '#0efacc',
								shadowBlur: 2,
								shadowColor: '#333'
							},
							zlevel: 1
						}
					]
				},
				visitRecordOption: {
					xAxis: {
						data: [],
						boundaryGap: false,
						axisLabel: {
							formatter: value => String(value).slice(5) || value
						},
						axisTick: {
							show: false
						}
					},
					grid: {
						left: 10,
						right: 20,
						top: 30,
						bottom: 0,
						containLabel: true
					},
					tooltip: {
						trigger: 'axis',
						axisPointer: {
							type: 'cross'
						},
						padding: [5, 10]
					},
					yAxis: {
						minInterval: 1,
						axisTick: {
							show: false
						}
					},
					legend: {
						data: ['访问量(PV)', '独立访客(UV)']
					},
					series: [
						{
							name: '访问量(PV)',
							smooth: true,
							type: 'line',
							itemStyle: {
								color: '#FF005A'
							},
							lineStyle: {
								color: '#FF005A',
								width: 2
							},
							areaStyle: {
								color: 'rgba(255, 0, 90, 0.08)'
							},
							data: [],
							animationDuration: 2800,
							animationEasing: 'cubicInOut'
						},
						{
							name: '独立访客(UV)',
							smooth: true,
							type: 'line',
							itemStyle: {
								color: '#3888fa'
							},
							lineStyle: {
								color: '#3888fa',
								width: 2
							},
							areaStyle: {
								color: 'rgba(56, 136, 250, 0.12)'
							},
							data: [],
							animationDuration: 2800,
							animationEasing: 'quadraticOut'
						}
					]
				}
			}
		},
		computed: {
			statCards() {
				return [
					{
						label: '今日 PV',
						value: this.pv,
						icon: View,
						theme: 'theme-blue'
					},
					{
						label: '今日 UV',
						value: this.uv,
						icon: User,
						theme: 'theme-green'
					},
					{
						label: '文章数',
						value: this.blogCount,
						icon: Document,
						theme: 'theme-orange'
					},
					{
						label: '评论数',
						value: this.commentCount,
						icon: ChatDotRound,
						theme: 'theme-purple'
					}
				]
			}
		},
		mounted() {
			this.getData()
			window.addEventListener('resize', this.handleResize)
		},
		beforeUnmount() {
			window.removeEventListener('resize', this.handleResize)
			;[
				this.categoryEcharts,
				this.tagEcharts,
				this.mapEcharts,
				this.visitRecordEcharts
			].forEach(chart => {
				if (chart && !chart.isDisposed()) {
					chart.dispose()
				}
			})
		},
		methods: {
			handleResize() {
				;[
					this.categoryEcharts,
					this.tagEcharts,
					this.mapEcharts,
					this.visitRecordEcharts
				].forEach(chart => {
					if (chart && !chart.isDisposed()) {
						chart.resize()
					}
				})
			},
			getData() {
				getDashboard().then(res => {
					this.pv = res.data.pv
					this.uv = res.data.uv
					this.blogCount = res.data.blogCount
					this.commentCount = res.data.commentCount
					//渲染分类数据
					this.categoryOption.legend.data = res.data.category.legend
					this.categoryOption.series[0].data = res.data.category.series
					this.initCategoryEcharts()
					//渲染标签数据
					this.tagOption.legend.data = res.data.tag.legend
					this.tagOption.series[0].data = res.data.tag.series
					this.initTagEcharts()
					//渲染访客地图数据
					const mapData = this.convertData(res.data.cityVisitor || [])
					this.mapOption.series[1].data = mapData
					this.mapOption.series[2].data = mapData.slice(0, 5)
					this.initMapEcharts()
					//渲染一周访问量数据
					this.visitRecordOption.xAxis.data = res.data.visitRecord.date
					this.visitRecordOption.series[0].data = res.data.visitRecord.pv
					this.visitRecordOption.series[1].data = res.data.visitRecord.uv
					this.initVisitRecordEcharts()
				})
			},
			initCategoryEcharts() {
				if (!this.categoryEcharts || this.categoryEcharts.isDisposed()) {
					this.categoryEcharts = echarts.init(this.$refs.categoryEcharts)
				}
				this.categoryEcharts.setOption(this.categoryOption)
			},
			initTagEcharts() {
				if (!this.tagEcharts || this.tagEcharts.isDisposed()) {
					this.tagEcharts = echarts.init(this.$refs.tagEcharts)
				}
				this.tagEcharts.setOption(this.tagOption)
			},
			initMapEcharts() {
				if (!this.mapEcharts || this.mapEcharts.isDisposed()) {
					this.mapEcharts = echarts.init(this.$refs.mapEcharts)
				}
				this.mapEcharts.setOption(this.mapOption)
			},
			convertData(data) {
				const result = []
				for (let i = 0; i < data.length; i++) {
					const geoCoord = this.resolveCityCoord(data[i].city)
					if (geoCoord) {
						result.push({
							name: data[i].city,
							value: [+geoCoord[0], +geoCoord[1]],
							uv: data[i].uv
						})
					}
				}
				return result
			},
			// 后端 city 是 ip2region 归属地串（如 "中国 广东省 深圳市 电信"），
			// 坐标表键是城市全名（如 "深圳市"），需要容错匹配
			resolveCityCoord(city) {
				if (!city) return null
				if (geoCoordMap[city]) return geoCoordMap[city]
				const tokens = String(city).split(/\s+/)
				for (let i = tokens.length - 1; i >= 0; i--) {
					const token = tokens[i]
					if (geoCoordMap[token]) return geoCoordMap[token]
					if (geoCoordMap[token + '市']) return geoCoordMap[token + '市']
				}
				for (const key of Object.keys(geoCoordMap)) {
					const bare = key.replace(/市$/, '')
					if (city.includes(key) || city.includes(bare)) {
						return geoCoordMap[key]
					}
				}
				return null
			},
			initVisitRecordEcharts() {
				if (!this.visitRecordEcharts || this.visitRecordEcharts.isDisposed()) {
					this.visitRecordEcharts = echarts.init(this.$refs.visitRecordEcharts)
				}
				this.visitRecordEcharts.setOption(this.visitRecordOption)
			}
		}
	}
</script>

<style scoped>
	.dashboard-page {
		padding: 4px;
	}

	.stat-row {
		margin-bottom: 16px;
	}

	.stat-row .el-col {
		margin-bottom: 8px;
	}

	/* 统计卡片 */
	.stat-card {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 20px;
		border-radius: 10px;
		background: #fff;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
		transition: transform 0.2s ease, box-shadow 0.2s ease;
	}

	.stat-card:hover {
		transform: translateY(-3px);
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
	}

	.stat-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 56px;
		height: 56px;
		border-radius: 12px;
		color: #fff;
		flex-shrink: 0;
	}

	.stat-info {
		min-width: 0;
	}

	.stat-label {
		font-size: 13px;
		color: #8a919f;
		margin-bottom: 6px;
	}

	.stat-num {
		font-size: 26px;
		font-weight: 700;
		color: #1f2d3d;
		line-height: 1.1;
	}

	.theme-blue .stat-icon {
		background: linear-gradient(135deg, #4f9dfd, #2f6fe0);
		box-shadow: 0 4px 12px rgba(79, 157, 253, 0.35);
	}

	.theme-green .stat-icon {
		background: linear-gradient(135deg, #4dc98b, #2ba471);
		box-shadow: 0 4px 12px rgba(77, 201, 139, 0.35);
	}

	.theme-orange .stat-icon {
		background: linear-gradient(135deg, #f9b04d, #ef8f36);
		box-shadow: 0 4px 12px rgba(249, 176, 77, 0.35);
	}

	.theme-purple .stat-icon {
		background: linear-gradient(135deg, #a06df5, #7c4bdd);
		box-shadow: 0 4px 12px rgba(160, 109, 245, 0.35);
	}

	/* 图表卡片 */
	.chart-row {
		margin-bottom: 16px;
	}

	.chart-row .el-col {
		margin-bottom: 8px;
	}

	.chart-card {
		border-radius: 10px;
		height: 100%;
	}

	.chart-card :deep(.el-card__header) {
		padding: 14px 20px;
		border-bottom: 1px solid #f0f1f5;
	}

	.chart-title {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #303133;
	}

	.chart-title .el-icon {
		color: #409eff;
	}

	.chart-body {
		height: 420px;
		width: 100%;
	}

	.trend-body {
		height: 380px;
	}

	.trend-card {
		margin-bottom: 16px;
	}
</style>
