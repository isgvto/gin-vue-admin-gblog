<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="文章分类" description="维护文章分类，组织内容主题。" />
    <section class="gva-table-box">
		<!--添加-->
		<el-row :gutter="10">
			<el-col :span="6">
				<el-button type="primary" size="small" icon="Plus" @click="addDialogVisible=true">添加分类</el-button>
			</el-col>
		</el-row>

		<el-table :data="categoryList">
			<el-table-column label="序号" type="index" width="100" align="center"></el-table-column>
			<el-table-column label="名称" prop="categoryName" min-width="280"></el-table-column>
			<el-table-column label="操作" align="center">
				<template v-slot="scope">
					<el-button type="primary" icon="Edit" size="small" @click="showEditDialog(scope.row)">编辑</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteCategoryById(scope.row.id)">
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

		<!--添加分类对话框-->
		<el-dialog title="添加分类" width="min(760px, 94vw)" v-model="addDialogVisible" :close-on-click-modal="false" @close="addDialogClosed">
			<!--内容主体-->
			<el-form :model="addForm" :rules="formRules" ref="addFormRef" label-width="80px">
				<el-form-item label="分类名称" prop="categoryName">
					<el-input v-model="addForm.categoryName"></el-input>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="addDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="addCategory">确 定</el-button>
			</template>
		</el-dialog>

		<!--编辑分类对话框-->
		<el-dialog title="编辑分类" width="min(760px, 94vw)" v-model="editDialogVisible" :close-on-click-modal="false" @close="editDialogClosed">
			<!--内容主体-->
			<el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="80px">
				<el-form-item label="分类名称" prop="categoryName">
					<el-input v-model="editForm.categoryName"></el-input>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="editDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="editCategory">确 定</el-button>
			</template>
		</el-dialog>

    </section>
  </div>
</template>

<script>
	import {
		getData as fetchCategoryData,
		addCategory as createCategory,
		editCategory as updateCategory,
		deleteCategoryById as removeCategory
	} from '@/api/blog/category'

	export default {
		name: 'BlogCategoryList',
		components: {},
		data() {
			return {
				queryInfo: {
					pageNum: 1,
					pageSize: 10
				},
				categoryList: [],
				total: 0,
				addDialogVisible: false,
				editDialogVisible: false,
				addForm: {
					categoryName: ''
				},
				editForm: {},
				formRules: {
					categoryName: [{required: true, message: '请输入分类名称', trigger: 'blur'}]
				}
			}
		},
		created() {
			this.getData()
		},
		methods: {
			getData() {
				fetchCategoryData(this.queryInfo).then(res => {
					this.categoryList = res.data.list
					this.total = res.data.total
				})
			},
			//监听 pageSize 改变事件
			handleSizeChange(newSize) {
				this.queryInfo.pageSize = newSize
				this.getData()
			},
			//监听页码改变事件
			handleCurrentChange(newPage) {
				this.queryInfo.pageNum = newPage
				this.getData()
			},
			addDialogClosed() {
				this.$refs.addFormRef.resetFields()
			},
			editDialogClosed() {
				this.editForm = {}
				this.$refs.editFormRef.resetFields()
			},
			addCategory() {
				this.$refs.addFormRef.validate(valid => {
					if (valid) {
						createCategory(this.addForm).then(res => {
							this.msgSuccess(res.msg)
							this.addDialogVisible = false
							this.getData()
						})
					}
				})
			},
			editCategory() {
				this.$refs.editFormRef.validate(valid => {
					if (valid) {
						updateCategory(this.editForm).then(res => {
							this.msgSuccess(res.msg)
							this.editDialogVisible = false
							this.getData()
						})
					}
				})
			},
			showEditDialog(row) {
				//row 中没有对象(blogs是表单不需要的属性)，直接拓展运算符深拷贝一份(拓展运算符不能深拷贝对象，只能拷贝引用)
				//如果直接赋值，则为引用，表格上的数据也会随对话框中数据的修改而实时改变
				this.editForm = {...row}
				this.editDialogVisible = true
			},
			deleteCategoryById(id) {
				removeCategory(id).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
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
