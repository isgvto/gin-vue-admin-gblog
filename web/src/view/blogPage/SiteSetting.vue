<template>
	<div class="admin-page admin-page--cards">
    <AdminPageHeading title="站点设置" description="分组维护站点资料、文档站与显示设置。" />
		<el-row :gutter="20">
			<el-col :xs="24" :lg="12">
				<el-card>
					<template #header>
						<span>基础设置</span>
					</template>
					<el-form label-position="right" label-width="100px">
						<el-form-item :label="item.nameZh" v-for="item in typeMap.type1" :key="item.id || item.key || item.nameEn">
							<el-input v-model="item.value" size="small"></el-input>
						</el-form-item>
					</el-form>
				</el-card>
			</el-col>
			<el-col :xs="24" :lg="12">
				<el-card>
					<template #header>
						<span>资料卡</span>
					</template>
					<el-form label-position="right" label-width="100px">
						<el-form-item :label="item.nameZh" v-for="item in typeMap.type2" :key="item.id || item.key || item.nameEn">
							<div v-if="item.nameEn=='favorite'" class="admin-favorite-row">
								<el-col :span="20">
									<el-input v-model="item.value" size="small"></el-input>
								</el-col>
								<el-col :span="4">
									<el-button type="danger" size="small" icon="Delete" @click="deleteFavorite(item)">删除</el-button>
								</el-col>
							</div>
							<div v-else class="admin-setting-field">
								<el-input v-model="item.value" size="small"></el-input>
							</div>
						</el-form-item>
						<el-button type="primary" size="small" icon="Plus" @click="addFavorite">添加自定义</el-button>
					</el-form>
				</el-card>
			</el-col>
		</el-row>

		<el-row style="margin-top: 20px">
			<el-card>
				<template #header>
					<div class="card-header">
						<span>文档站设置</span>
						<el-button type="success" size="small" icon="Refresh" :loading="syncingDocs" @click="handleSyncDocs">同步文档</el-button>
					</div>
				</template>
				<el-form label-position="right" label-width="130px">
					<el-form-item :label="item.nameZh" v-for="item in typeMap.type4" :key="item.id || item.key || item.nameEn">
						<el-input v-model="item.value" size="small" :placeholder="docPlaceholder(item.nameEn)"></el-input>
					</el-form-item>
				</el-form>
			</el-card>
		</el-row>

		<el-row style="margin-top: 20px">
			<el-card>
				<template #header>
					<span>页脚徽标</span>
				</template>
				<el-form :inline="true" v-for="(badge, index) in typeMap.type3" :key="badge.id || badge.key || index">
					<el-form-item label="title">
						<el-input v-model="badge.value.title" size="small"></el-input>
					</el-form-item>
					<el-form-item label="url">
						<el-input v-model="badge.value.url" size="small"></el-input>
					</el-form-item>
					<el-form-item label="subject">
						<el-input v-model="badge.value.subject" size="small"></el-input>
					</el-form-item>
					<el-form-item label="value">
						<el-input v-model="badge.value.value" size="small"></el-input>
					</el-form-item>
					<el-form-item>
						<el-button type="danger" size="small" icon="Delete" @click="deleteBadge(badge)">删除</el-button>
					</el-form-item>
				</el-form>
				<el-button type="primary" size="small" icon="Plus" @click="addBadge">添加 badge</el-button>
			</el-card>
		</el-row>

		<div style="text-align: right;margin-top: 30px">
			<el-button type="primary" icon="Check" @click="submit">保存</el-button>
		</div>
	</div>
</template>

<script>
	import {getSiteSettingData, update} from "@/api/blog/siteSetting";
	import {syncDocs} from "@/api/blog/docs";

	const emptyTypeMap = () => ({
		type1: [],
		type2: [],
		type3: [],
		type4: []
	})

	const parseBadgeValue = (value) => {
		const fallback = {
			subject: "",
			title: "",
			url: "",
			value: ""
		}

		if (!value) {
			return fallback
		}

		try {
			const parsed = typeof value === 'object' ? value : JSON.parse(value)
			return Object.fromEntries(Object.keys(fallback).map(key => [key, typeof parsed?.[key] === 'string' ? parsed[key] : '']))
		} catch (e) {
			console.warn('Invalid badge site setting value:', value, e)
			return fallback
		}
	}

	const cloneTypeMap = (typeMap) => JSON.parse(JSON.stringify(typeMap))
	// Old servers may still return these fields during a rolling upgrade.
	const retiredSettings = new Set(['bg1', 'bg2', 'bg3', 'playlistServer', 'playlistId'])

	export default {
		name: 'BlogSiteSetting',
		components: {},
		data() {
			return {
				deleteIds: [],
				typeMap: emptyTypeMap(),
				syncingDocs: false,
			}
		},
		created() {
			this.getData()
		},
		methods: {
			getData() {
				getSiteSettingData().then(res => {
					const data = res.data || {}
					const nextTypeMap = {
						type1: Array.isArray(data.type1) ? data.type1.filter(item => !retiredSettings.has(item.nameEn)) : [],
						type2: Array.isArray(data.type2) ? data.type2 : [],
						type3: Array.isArray(data.type3) ? data.type3 : [],
						type4: Array.isArray(data.type4) ? data.type4 : []
					}

					nextTypeMap.type1.forEach(item => {
						item.value = item.value || ''
					})
					nextTypeMap.type2.forEach(item => {
						item.value = item.value || ''
					})
					nextTypeMap.type3.forEach(item => {
						item.value = parseBadgeValue(item.value)
					})
					nextTypeMap.type4.forEach(item => {
						item.value = item.value || ''
					})
					this.typeMap = nextTypeMap
				})
			},
			addFavorite() {
				this.typeMap.type2.push({
					key: Date.now(),
					nameEn: "favorite",
					nameZh: "自定义",
					type: 2,
					value: "{\"title\":\"\",\"content\":\"\"}"
				})
			},
			addBadge() {
				this.typeMap.type3.push({
					key: Date.now(),
					nameEn: "badge",
					nameZh: "徽标",
					type: 3,
					value: {
						subject: "",
						title: "",
						url: "",
						value: ""
					}
				})
			},
			deleteFavorite(favorite) {
				let arr = this.typeMap.type2
				if (favorite.id) {
					this.deleteIds.push(favorite.id)
					arr.forEach((item, index) => {
						if (item.id === favorite.id) {
							arr.splice(index, 1)
							return
						}
					})
				} else {
					arr.forEach((item, index) => {
						if (item.key === favorite.key) {
							arr.splice(index, 1)
							return
						}
					})
				}
			},
			deleteBadge(badge) {
				let arr = this.typeMap.type3
				if (badge.id) {
					this.deleteIds.push(badge.id)
					arr.forEach((item, index) => {
						if (item.id === badge.id) {
							arr.splice(index, 1)
							return
						}
					})
				} else {
					arr.forEach((item, index) => {
						if (item.key === badge.key) {
							arr.splice(index, 1)
							return
						}
					})
				}
			},
			submit() {
				const result = cloneTypeMap(this.typeMap)
				result.type3.forEach(item => {
					item.value = JSON.stringify(parseBadgeValue(item.value))
				})
				let updateArr = []
				updateArr.push(...result.type1)
				updateArr.push(...result.type2)
				updateArr.push(...result.type3)
				updateArr.push(...result.type4)
				update(updateArr, this.deleteIds).then(res => {
					this.deleteIds = []
					this.getData()
					this.msgSuccess(res.msg)
				})
			},
			docPlaceholder(nameEn) {
				const map = {
					docsGithubRepo: '例如：https://github.com/Percygu/GolangGuide',
					docsGithubBranch: '例如：main',
					docsGithubRoot: '例如：src 或 docs',
					docsGithubWebhookSecret: '可选：与 GitHub Webhook Secret 保持一致'
				}
				return map[nameEn] || ''
			},
			handleSyncDocs() {
				this.syncingDocs = true
				syncDocs().then(res => {
					if (res && res.code === 0) {
						this.msgSuccess(res.msg || '同步文档成功')
					}
				}).finally(() => {
					this.syncingDocs = false
				})
			}
		}
	}
</script>

<style scoped>
	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
</style>
