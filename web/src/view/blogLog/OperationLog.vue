<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="博客操作日志" description="查看博客相关操作与请求详情。" />
    <section class="gva-table-box">
		<!--搜索-->
		<el-form inline>
			<el-form-item label="操作时间">
				<DateTimeRangePicker :date="queryInfo.date" :setDate="setDate"/>
			</el-form-item>
			<el-form-item>
				<el-button type="primary" size="small" icon="Search" @click="search">搜索</el-button>
			</el-form-item>
		</el-form>

		<el-table :data="logList">
			<el-table-column type="expand">
				<template v-slot="props">
					<el-form label-position="left" class="table-expand">
						<el-form-item label="请求接口">
							<span>{{ props.row.uri }}</span>
						</el-form-item>
						<el-form-item label="请求参数">
							<span>{{ props.row.param }}</span>
						</el-form-item>
					</el-form>
				</template>
			</el-table-column>
			<el-table-column label="序号" type="index" width="70" align="center"></el-table-column>
			<el-table-column label="操作者" prop="username" min-width="120" show-overflow-tooltip></el-table-column>
			<el-table-column label="请求方式" prop="method" width="80"></el-table-column>
			<el-table-column label="描述" prop="description" show-overflow-tooltip min-width="220"></el-table-column>
			<el-table-column label="IP" prop="ip" min-width="140" show-overflow-tooltip></el-table-column>
			<el-table-column label="IP来源" prop="ipSource" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="操作系统" prop="os" min-width="130" show-overflow-tooltip></el-table-column>
			<el-table-column label="浏览器" prop="browser" show-overflow-tooltip min-width="110"></el-table-column>
			<el-table-column label="操作耗时" width="110">
				<template v-slot="scope">
					<el-tag size="small">{{ scope.row.times }}ms</el-tag>
				</template>
			</el-table-column>
			<el-table-column label="操作时间" width="170">
				<template v-slot="scope">{{ blogDateFormat(scope.row.createTime) }}</template>
			</el-table-column>
			<el-table-column label="操作" width="120" align="center">
				<template v-slot="scope">
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteLogById(scope.row.id)">
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
	import {getOperationLogList, deleteOperationLogById} from "@/api/blog/operationLog";
	import DateTimeRangePicker from "@/components/DateTimeRangePicker.vue";

	export default {
		name: 'BlogOperationLog',
		components: {DateTimeRangePicker},
		data() {
			return {
				queryInfo: {
					date: [],
					pageNum: 1,
					pageSize: 10
				},
				logList: [],
				total: 0,
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
				getOperationLogList(query).then(res => {
					this.logList = res.data.list
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
			deleteLogById(id) {
				deleteOperationLogById(id).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
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
	.el-form--inline .el-form-item {
		margin-bottom: 0;
	}
</style>
