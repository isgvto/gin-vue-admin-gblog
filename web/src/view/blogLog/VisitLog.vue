<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="访问日志" description="按访客和时间查看访问请求记录。" />
    <section class="gva-table-box">
		<!--搜索-->
		<el-form inline>
			<el-form-item label="访客标识">
				<el-input v-model="queryInfo.uuid" :clearable="true" size="small" @keyup.enter="search"
				          placeholder="请输入访客标识码" style="width: 100%; max-width: 360px">
				</el-input>
			</el-form-item>
			<el-form-item label="访问时间">
				<DateTimeRangePicker :date="queryInfo.date" :setDate="setDate"/>
			</el-form-item>
			<el-form-item>
				<el-button type="primary" size="small" icon="Search" @click="search">搜索</el-button>
			</el-form-item>
			<el-form-item>
				<el-popconfirm title="确定删除选中的访问日志吗？" icon="Delete" iconColor="red" @confirm="deleteSelectedLogs">
					<template #reference>
						<el-button type="danger" size="small" icon="Delete" :disabled="multipleSelection.length === 0">批量删除</el-button>
					</template>
				</el-popconfirm>
			</el-form-item>
		</el-form>

		<el-table :data="logList" @selection-change="handleSelectionChange">
			<el-table-column type="selection" width="55" align="center"></el-table-column>
			<el-table-column type="expand">
				<template v-slot="props">
					<el-form label-position="left" class="table-expand">
						<el-form-item label="访客标识">
							<span>{{ props.row.uuid }}</span>
						</el-form-item>
						<el-form-item label="请求方式">
							<span>{{ props.row.method }}</span>
						</el-form-item>
						<el-form-item label="请求接口">
							<span>{{ props.row.uri }}</span>
						</el-form-item>
						<el-form-item label="请求参数">
							<span>{{ props.row.param }}</span>
						</el-form-item>
						<el-form-item label="备注">
							<span>{{ props.row.remark }}</span>
						</el-form-item>
					</el-form>
				</template>
			</el-table-column>
			<el-table-column label="序号" type="index" width="70" align="center"></el-table-column>
			<el-table-column label="访客标识"  show-overflow-tooltip min-width="180">
				<template v-slot="scope">
					<el-link type="primary" href="" :underline="false" @click.prevent="showThis(scope.row.uuid)">{{ scope.row.uuid }}</el-link>
				</template>
			</el-table-column>
			<el-table-column label="访问行为" prop="behavior" min-width="110" show-overflow-tooltip></el-table-column>
			<el-table-column label="访问内容" prop="content" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="IP" prop="ip" min-width="140" show-overflow-tooltip></el-table-column>
			<el-table-column label="IP来源" prop="ipSource" show-overflow-tooltip min-width="180"></el-table-column>
			<el-table-column label="操作系统" prop="os" show-overflow-tooltip min-width="130"></el-table-column>
			<el-table-column label="浏览器" prop="browser" show-overflow-tooltip min-width="110"></el-table-column>
			<el-table-column label="访问时间" width="170">
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
	import {getVisitLogList, deleteVisitLogById, deleteVisitLogsByIds} from "@/api/blog/visitLog";
	import DateTimeRangePicker from "@/components/DateTimeRangePicker.vue";

	export default {
		name: 'BlogVisitLog',
		components: {DateTimeRangePicker},
		data() {
			return {
				queryInfo: {
					uuid: '',
					date: [],
					pageNum: 1,
					pageSize: 10
				},
				logList: [],
				total: 0,
				multipleSelection: [],
			}
		},
		created() {
			if (this.$route.query.uuid) {
				this.queryInfo.uuid = this.$route.query.uuid
			}
			this.getData()
		},
		methods: {
			getData() {
				let query = {...this.queryInfo}
				if (query.date && query.date.length === 2) {
					query.date = query.date[0] + ',' + query.date[1]
				}
				getVisitLogList(query).then(res => {
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
			handleSelectionChange(selection) {
				this.multipleSelection = selection
			},
			deleteLogById(id) {
				deleteVisitLogById(id).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
				})
			},
			deleteSelectedLogs() {
				const ids = this.multipleSelection.map(item => item.id)
				if (ids.length === 0) {
					return
				}
				deleteVisitLogsByIds(ids).then(res => {
					this.msgSuccess(res.msg)
					if (this.logList.length === ids.length && this.queryInfo.pageNum > 1) {
						this.queryInfo.pageNum--
					}
					this.multipleSelection = []
					this.getData()
				})
			},
			search() {
				this.queryInfo.pageNum = 1
				this.queryInfo.pageSize = 10
				this.getData()
			},
			showThis(uuid) {
				this.queryInfo.uuid = uuid
				this.search()
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
