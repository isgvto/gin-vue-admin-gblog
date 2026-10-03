<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="博客异常日志" description="查询异常信息并查看请求上下文。" />
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
			<el-table-column label="请求方式" prop="method" width="80"></el-table-column>
			<el-table-column label="描述" prop="description" min-width="220" show-overflow-tooltip></el-table-column>
			<el-table-column label="IP" prop="ip" min-width="140" show-overflow-tooltip></el-table-column>
			<el-table-column label="IP来源" prop="ipSource" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="操作系统" prop="os" min-width="130" show-overflow-tooltip></el-table-column>
			<el-table-column label="浏览器" prop="browser" show-overflow-tooltip min-width="110"></el-table-column>
			<el-table-column label="操作时间" width="170">
				<template v-slot="scope">{{ blogDateFormat(scope.row.createTime) }}</template>
			</el-table-column>
			<el-table-column label="操作" width="200" align="center">
				<template v-slot="scope">
					<el-button type="warning" icon="View" size="small" @click="showDetail(scope.row.error)">查看详情</el-button>
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

		<!-- 异常信息 -->
		<el-dialog title="异常信息" append-to-body top="20px" width="min(760px, 94vw)" v-model="detailDialogVisible" destroy-on-close>
			<div class="match-braces rainbow-braces">
				<pre>
					<code class="language-java">{{ detail }}</code>
				</pre>
			</div>
			<template #footer>
				<el-button @click="detailDialogVisible=false">关 闭</el-button>
			</template>
		</el-dialog>

    </section>
  </div>
</template>

<script>
	import {getExceptionLogList, deleteExceptionLogById} from "@/api/blog/exceptionLog";
	import DateTimeRangePicker from "@/components/DateTimeRangePicker.vue";

	export default {
		name: 'BlogExceptionLog',
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
				detailDialogVisible: false,
				detail: ''
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
				getExceptionLogList(query).then(res => {
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
				deleteExceptionLogById(id).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
				})
			},
			showDetail(error) {
				this.detail = '\n' + error
				this.detailDialogVisible = true
				this.$nextTick(() => {
					Prism.highlightAll()
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
