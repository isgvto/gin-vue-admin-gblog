<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="动态管理" description="查看和管理已发布的日常动态。" />
    <section class="gva-table-box">
		<el-table :data="momentList">
			<el-table-column label="序号" type="index" width="100" align="center"></el-table-column>
			<el-table-column label="内容" prop="content" show-overflow-tooltip></el-table-column>
			<el-table-column label="发布状态" width="80">
				<template v-slot="scope">
					<el-switch v-model="scope.row.isPublished" @change="momentPublishedChanged(scope.row)"></el-switch>
				</template>
			</el-table-column>
			<el-table-column label="点赞数" prop="likes" width="80"></el-table-column>
			<el-table-column label="创建时间" width="170">
				<template v-slot="scope">{{ blogDateFormat(scope.row.createTime) }}</template>
			</el-table-column>
			<el-table-column label="操作" width="200" align="center">
				<template v-slot="scope">
					<el-button type="primary" icon="Edit" size="small" @click="goEditMomentPage(scope.row.id)">编辑</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteMomentById(scope.row.id)">
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
	import {
		getMomentListByQuery,
		updatePublished,
		deleteMomentById as removeMoment
	} from "@/api/blog/moment";

	export default {
		name: 'BlogMomentList',
		components: {},
		data() {
			return {
				queryInfo: {
					pageNum: 1,
					pageSize: 10
				},
				momentList: [],
				total: 0,
			}
		},
		created() {
			this.getMomentList()
		},
		methods: {
			getMomentList() {
				getMomentListByQuery(this.queryInfo).then(res => {
					this.momentList = res.data.list
					this.total = res.data.total
				})
			},
			//监听 pageSize 改变事件
			handleSizeChange(newSize) {
				this.queryInfo.pageSize = newSize
				this.getMomentList()
			},
			//监听页码改变的事件
			handleCurrentChange(newPage) {
				this.queryInfo.pageNum = newPage
				this.getMomentList()
			},
			momentPublishedChanged(row) {
				updatePublished(row.id, row.isPublished).then(res => {
					this.msgSuccess(res.msg)
				})
			},
			goEditMomentPage(id) {
				this.$router.push(`/layout/gblog/moment/edit/${id}`)
			},
			deleteMomentById(id) {
				removeMoment(id).then(res => {
					this.msgSuccess(res.msg)
					this.getMomentList()
				})
			}
		}
	}
</script>

<style scoped>
	.el-button + span {
		margin-left: 10px;
	}
</style>
