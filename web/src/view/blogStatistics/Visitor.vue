<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="访客统计" description="查看访客来源、访问记录与访问时间。" />
    <section class="gva-table-box">
		<!--搜索-->
		<el-form inline>
			<el-form-item label="最后访问时间">
				<DateTimeRangePicker :date="queryInfo.date" :setDate="setDate"/>
			</el-form-item>
			<el-form-item>
				<el-button type="primary" size="small" icon="Search" @click="search">搜索</el-button>
			</el-form-item>
			<el-form-item>
				<el-popconfirm title="确定删除选中的访客吗？" icon="Delete" iconColor="red" @confirm="deleteSelectedVisitors">
					<template #reference>
						<el-button type="danger" size="small" icon="Delete" :disabled="multipleSelection.length === 0">批量删除</el-button>
					</template>
				</el-popconfirm>
			</el-form-item>
		</el-form>

		<el-table :data="visitorList" @selection-change="handleSelectionChange">
			<el-table-column type="selection" width="55" align="center"></el-table-column>
			<el-table-column label="序号" type="index" width="70" align="center"></el-table-column>
			<el-table-column label="访客标识" prop="uuid" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="IP" prop="ip" show-overflow-tooltip min-width="140"></el-table-column>
			<el-table-column label="IP来源" prop="ipSource" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="操作系统" prop="os" show-overflow-tooltip min-width="130"></el-table-column>
			<el-table-column label="浏览器" prop="browser" show-overflow-tooltip min-width="110"></el-table-column>
			<el-table-column label="首次访问" width="170">
				<template v-slot="scope">{{ blogDateFormat(scope.row.createTime) }}</template>
			</el-table-column>
			<el-table-column width="170">
				<template #header>
					最后访问
					<el-tooltip effect="dark" content="每日凌晨自动更新" placement="top"><i class="el-icon-question"></i></el-tooltip>
				</template>
				<template v-slot="scope">{{ blogDateFormat(scope.row.lastTime) }}</template>
			</el-table-column>
			<el-table-column prop="pv" width="70">
				<template #header>
					PV
					<el-tooltip effect="dark" content="访客总浏览量，每日凌晨自动更新" placement="top"><i class="el-icon-question"></i></el-tooltip>
				</template>
			</el-table-column>
			<el-table-column label="操作" width="200" align="center">
				<template v-slot="scope">
					<el-button type="warning" icon="View" size="small" @click="showLog(scope.row.uuid)">查看记录</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteVisitorById(scope.row)">
						<template #reference><el-button size="small" type="danger" icon="Delete" >删除</el-button></template>
					</el-popconfirm>
				</template>
			</el-table-column>
		</el-table>

		<!--分页-->
		<el-pagination @size-change="handleSizeChange" @current-change="handleCurrentChange" :current-page="queryInfo.pageNum"
		               :page-sizes="[10, 20, 30, 50]" :page-size="queryInfo.pageSize" :total="total"
		               layout="total, sizes, prev, pager, next, jumper" background>
		</el-pagination>

    </section>
  </div>
</template>

<script>
	import {getVisitorList, deleteVisitor, deleteVisitorsByIds} from "@/api/blog/visitor";
	import DateTimeRangePicker from "@/components/DateTimeRangePicker.vue";

	export default {
		name: 'BlogVisitorStats',
		components: {DateTimeRangePicker},
		data() {
			return {
				queryInfo: {
					date: [],
					pageNum: 1,
					pageSize: 10
				},
				visitorList: [],
				total: 0,
				multipleSelection: [],
			}
		},
		created() {
			this.getData()
		},
		methods: {
			getData() {
				let query = {...this.queryInfo}
				if (query.date && query.date.length === 2) {
					query.date = query.date[0] + ',' + query.date[1]
				}
				getVisitorList(query).then(res => {
					this.visitorList = res.data.list
					this.total = res.data.total
				})
			},
			handleSizeChange(newSize) {
				this.queryInfo.pageSize = newSize
				this.getData()
			},
			handleCurrentChange(newPage) {
				this.queryInfo.pageNum = newPage
				this.getData()
			},
			handleSelectionChange(selection) {
				this.multipleSelection = selection
			},
			deleteVisitorById(visitor) {
				deleteVisitor(visitor.id, visitor.uuid).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
				})
			},
			deleteSelectedVisitors() {
				const ids = this.multipleSelection.map(item => item.id)
				if (ids.length === 0) {
					return
				}
				deleteVisitorsByIds(ids).then(res => {
					this.msgSuccess(res.msg)
					if (this.visitorList.length === ids.length && this.queryInfo.pageNum > 1) {
						this.queryInfo.pageNum--
					}
					this.multipleSelection = []
					this.getData()
				})
			},
			showLog(uuid) {
				this.$router.push({
					name: 'blogVisitLog',
					query: {
						uuid
					}
				})
			},
			search() {
				this.queryInfo.pageNum = 1
				this.queryInfo.pageSize = 10
				this.getData()
			},
			setDate(value) {
				this.queryInfo.date = value
			},
		}
	}
</script>

<style scoped>
	.el-button + span {
		margin-left: 10px;
	}

	.el-form--inline .el-form-item {
		margin-bottom: 0;
	}
</style>
