<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="友情链接" description="维护友链信息与友链页面内容。" />
    <section class="gva-table-box">
		<!--添加-->
		<el-form inline>
			<el-form-item>
				<el-button type="primary" size="small" icon="Plus" @click="addDialogVisible=true">添加友链</el-button>
			</el-form-item>
			<el-form-item style="margin-left: 20px">
				<el-switch v-model="infoForm.commentEnabled" active-text="页面评论" @change="commentEnabledChanged"></el-switch>
			</el-form-item>
		</el-form>

		<el-table :data="friendList">
			<el-table-column label="序号" type="index" width="100" align="center"></el-table-column>
			<el-table-column label="头像" width="80">
				<template v-slot="scope">
					<el-avatar shape="square" :size="50" fit="contain" :src="scope.row.avatar"></el-avatar>
				</template>
			</el-table-column>
			<el-table-column label="昵称" prop="nickname"></el-table-column>
			<el-table-column label="描述" prop="description"></el-table-column>
			<el-table-column label="站点" prop="website"></el-table-column>
			<el-table-column label="是否公开" width="100" align="center">
				<template v-slot="scope">
					<el-switch v-model="scope.row.isPublished" @change="friendPublishedChanged(scope.row)"></el-switch>
				</template>
			</el-table-column>
			<el-table-column label="浏览次数" prop="views" width="100"></el-table-column>
			<el-table-column label="创建时间" width="170">
				<template v-slot="scope">{{ blogDateFormat(scope.row.createTime) }}</template>
			</el-table-column>
			<el-table-column label="操作" width="200" align="center">
				<template v-slot="scope">
					<el-button type="primary" icon="Edit" size="small" @click="showEditDialog(scope.row)">编辑</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteFriendById(scope.row.id)">
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

		<!--友链页面信息-->
		<el-form label-position="top">
			<el-form-item label="友链页面信息">
				<mavon-editor v-model="infoForm.content"/>
			</el-form-item>
			<el-form-item style="text-align: right;">
				<el-button type="primary" icon="Check" @click="updateContent">保存</el-button>
			</el-form-item>
		</el-form>

		<!--添加友链对话框-->
		<el-dialog title="添加友链" width="min(760px, 94vw)" v-model="addDialogVisible" :close-on-click-modal="false" @close="addDialogClosed">
			<!--内容主体-->
			<el-form :model="addForm" :rules="formRules" ref="addFormRef" label-width="80px">
				<el-form-item label="昵称" prop="nickname">
					<el-input v-model="addForm.nickname"></el-input>
				</el-form-item>
				<el-form-item label="描述" prop="description">
					<el-input v-model="addForm.description"></el-input>
				</el-form-item>
				<el-form-item label="网站" prop="website">
					<el-input v-model="addForm.website"></el-input>
				</el-form-item>
				<el-form-item label="头像URL" prop="avatar">
					<el-input v-model="addForm.avatar"></el-input>
				</el-form-item>
				<el-form-item label="是否公开" prop="isPublished">
					<el-switch v-model="addForm.isPublished"></el-switch>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="addDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="saveFriend">确 定</el-button>
			</template>
		</el-dialog>

		<!--编辑友链对话框-->
		<el-dialog title="编辑友链" width="min(760px, 94vw)" v-model="editDialogVisible" :close-on-click-modal="false" @close="editDialogClosed">
			<!--内容主体-->
			<el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="80px">
				<el-form-item label="昵称" prop="nickname">
					<el-input v-model="editForm.nickname"></el-input>
				</el-form-item>
				<el-form-item label="描述" prop="description">
					<el-input v-model="editForm.description"></el-input>
				</el-form-item>
				<el-form-item label="网站" prop="website">
					<el-input v-model="editForm.website"></el-input>
				</el-form-item>
				<el-form-item label="头像URL" prop="avatar">
					<el-input v-model="editForm.avatar"></el-input>
				</el-form-item>
				<el-form-item label="是否公开" prop="isPublished">
					<el-switch v-model="editForm.isPublished"></el-switch>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="editDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="editFriend">确 定</el-button>
			</template>
		</el-dialog>

    </section>
  </div>
</template>

<script>
	import {
		getFriendsByQuery,
		updatePublished,
		saveFriend as createFriend,
		updateFriend,
		deleteFriendById as removeFriend,
		getFriendInfo,
		updateContent as updateFriendContent,
		updateCommentEnabled
	} from "@/api/blog/friend";

	export default {
		name: 'BlogFriendList',
		components: {},
		data() {
			return {
				infoForm: {
					content: '',
					commentEnabled: true,
				},
				queryInfo: {
					pageNum: 1,
					pageSize: 10
				},
				friendList: [],
				total: 0,
				addDialogVisible: false,
				editDialogVisible: false,
				addForm: {
					nickname: '',
					description: '',
					website: '',
					avatar: '',
					isPublished: true
				},
				editForm: {
					nickname: '',
					description: '',
					website: '',
					avatar: '',
					isPublished: true
				},
				formRules: {
					nickname: [{required: true, message: '请输入昵称', trigger: 'blur'}],
					description: [{required: true, message: '请输入描述', trigger: 'blur'}],
					website: [{required: true, message: '请输入网站', trigger: 'blur'}],
					avatar: [{required: true, message: '请输入头像URL', trigger: 'blur'}],
				}
			}
		},
		created() {
			this.getFriendList()
			this.getInfo()
		},
		methods: {
			getInfo() {
				getFriendInfo().then(res => {
					this.infoForm = res.data
				})
			},
			updateContent() {
				updateFriendContent(this.infoForm.content).then(res => {
					this.msgSuccess(res.msg)
					this.getInfo()
				})
			},
			commentEnabledChanged() {
				updateCommentEnabled(this.infoForm.commentEnabled).then(res => {
					this.msgSuccess(res.msg)
				})
			},
			getFriendList() {
				getFriendsByQuery(this.queryInfo).then(res => {
					this.friendList = res.data.list
					this.total = res.data.total
				})
			},
			handleSizeChange(newSize) {
				this.queryInfo.pageSize = newSize
				this.getFriendList()
			},
			handleCurrentChange(newPage) {
				this.queryInfo.pageNum = newPage
				this.getFriendList()
			},
			friendPublishedChanged(row) {
				updatePublished(row.id, row.isPublished).then(res => {
					this.msgSuccess(res.msg)
				})
			},
			deleteFriendById(id) {
				removeFriend(id).then(res => {
					this.getFriendList()
					this.msgSuccess(res.msg)
				})
			},
			showEditDialog(row) {
				this.editForm = {...row}
				this.editDialogVisible = true
			},
			addDialogClosed() {
				this.$refs.addFormRef.resetFields()
			},
			editDialogClosed() {
				this.editForm = {}
				this.$refs.editFormRef.resetFields()
			},
			saveFriend() {
				this.$refs.addFormRef.validate(valid => {
					if (valid) {
						createFriend(this.addForm).then(res => {
							this.getFriendList()
							this.msgSuccess(res.msg)
							this.addDialogVisible = false
						})
					}
				})
			},
			editFriend() {
				this.$refs.editFormRef.validate(valid => {
					if (valid) {
						updateFriend(this.editForm).then(res => {
							this.getFriendList()
							this.msgSuccess(res.msg)
							this.editDialogVisible = false
						})
					}
				})
			}
		}
	}
</script>

<style scoped>
	.el-button + span {
		margin-left: 10px;
	}

	.el-form {
		margin-top: 15px !important;
	}

	.el-form--inline .el-form-item {
		margin-bottom: 0;
	}
</style>
