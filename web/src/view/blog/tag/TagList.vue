<template>
	<div class="admin-page admin-page--legacy-list">
    <AdminPageHeading title="标签管理" description="维护标签名称、颜色与内容归类。" />
    <section class="gva-table-box">
		<!--添加-->
		<el-row :gutter="10">
			<el-col :span="6">
				<el-button type="primary" size="small" icon="Plus" @click="addDialogVisible=true">添加标签</el-button>
			</el-col>
		</el-row>

		<el-table :data="tagList">
			<el-table-column label="序号" type="index" width="100" align="center"></el-table-column>
			<el-table-column label="名称" prop="tagName" min-width="280"></el-table-column>
			<el-table-column label="颜色">
				<template v-slot="scope">
					<div class="tag-color-view" v-if="scope.row.color">
						<span class="tag-color-swatch" :style="{ backgroundColor: getColorValue(scope.row.color) }"></span>
						<span>{{ getColorLabel(scope.row.color) }}</span>
						<span class="tag-color-key">{{ scope.row.color }}</span>
					</div>
					<span v-else class="tag-color-empty">未设置</span>
				</template>
			</el-table-column>
			<el-table-column label="操作" align="center">
				<template v-slot="scope">
					<el-button type="primary" icon="Edit" size="small" @click="showEditDialog(scope.row)">编辑</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteTagById(scope.row.id)">
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

		<!--添加标签对话框-->
		<el-dialog title="添加标签" width="min(760px, 94vw)" v-model="addDialogVisible" :close-on-click-modal="false" @close="addDialogClosed">
			<!--内容主体-->
			<el-form :model="addForm" :rules="formRules" ref="addFormRef" label-width="80px">
				<el-form-item label="标签名称" prop="tagName">
					<el-input v-model="addForm.tagName"></el-input>
				</el-form-item>
				<el-form-item label="标签颜色">
					<el-select v-model="addForm.color" placeholder="请选择颜色" :clearable="true" style="width: 100%">
						<el-option v-for="item in colors" :key="item.value" :label="item.label" :value="item.value">
							<div class="tag-color-option">
								<span class="tag-color-swatch" :style="{ backgroundColor: item.color }"></span>
								<span>{{ item.label }}</span>
								<span class="tag-color-key">{{ item.value }}</span>
							</div>
						</el-option>
					</el-select>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="addDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="addTag">确 定</el-button>
			</template>
		</el-dialog>

		<!--编辑标签对话框-->
		<el-dialog title="编辑标签" width="min(760px, 94vw)" v-model="editDialogVisible" :close-on-click-modal="false" @close="editDialogClosed">
			<!--内容主体-->
			<el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="80px">
				<el-form-item label="标签名称" prop="tagName">
					<el-input v-model="editForm.tagName"></el-input>
				</el-form-item>
				<el-form-item label="标签颜色" prop="color">
					<el-select v-model="editForm.color" placeholder="请选择颜色" :clearable="true" style="width: 100%">
						<el-option v-for="item in colors" :key="item.value" :label="item.label" :value="item.value">
							<div class="tag-color-option">
								<span class="tag-color-swatch" :style="{ backgroundColor: item.color }"></span>
								<span>{{ item.label }}</span>
								<span class="tag-color-key">{{ item.value }}</span>
							</div>
						</el-option>
					</el-select>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="editDialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="editTag">确 定</el-button>
			</template>
		</el-dialog>

    </section>
  </div>
</template>

<script>
	import {
		getData as fetchTagData,
		addTag as createTag,
		editTag as updateTag,
		deleteTagById as removeTag
	} from '@/api/blog/tag'

	export default {
		name: 'BlogTagList',
		components: {},
		data() {
			return {
				queryInfo: {
					pageNum: 1,
					pageSize: 10
				},
				tagList: [],
				total: 0,
				addDialogVisible: false,
				editDialogVisible: false,
				addForm: {
					tagName: '',
					color: ''
				},
				editForm: {},
				formRules: {
					tagName: [{required: true, message: '请输入标签名称', trigger: 'blur'}]
				},
				colors: [
					{label: '红色', value: 'red', color: '#DD3C3C'},
					{label: '橘黄', value: 'orange', color: '#F27E31'},
					{label: '黄色', value: 'yellow', color: '#FAC21F'},
					{label: '橄榄绿', value: 'olive', color: '#BBCF2D'},
					{label: '纯绿', value: 'green', color: '#36BF56'},
					{label: '水鸭蓝', value: 'teal', color: '#18BBB3'},
					{label: '纯蓝', value: 'blue', color: '#368FD3'},
					{label: '紫罗兰', value: 'violet', color: '#7248CD'},
					{label: '紫色', value: 'purple', color: '#AB46CC'},
					{label: '粉红', value: 'pink', color: '#E14BA0'},
					{label: '棕色', value: 'brown', color: '#AC7551'},
					{label: '灰色', value: 'grey', color: '#828282'},
					{label: '黑色', value: 'black', color: '#303132'},
				],
			}
		},
		created() {
			this.getData()
		},
		methods: {
			getData() {
				fetchTagData(this.queryInfo).then(res => {
					this.tagList = res.data.list
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
				this.addForm.color = ''
				this.$refs.addFormRef.resetFields()
			},
			editDialogClosed() {
				this.editForm = {}
				this.$refs.editFormRef.resetFields()
			},
			addTag() {
				this.$refs.addFormRef.validate(valid => {
					if (valid) {
						createTag(this.addForm).then(res => {
							this.msgSuccess(res.msg)
							this.addDialogVisible = false
							this.getData()
						})
					}
				})
			},
			editTag() {
				this.$refs.editFormRef.validate(valid => {
					if (valid) {
						updateTag(this.editForm).then(res => {
							this.msgSuccess(res.msg)
							this.editDialogVisible = false
							this.getData()
						})
					}
				})
			},
			showEditDialog(row) {
				this.editForm = {...row}
				this.editDialogVisible = true
			},
			getColorLabel(value) {
				const item = this.colors.find(color => color.value === value)
				return item ? item.label : value
			},
			getColorValue(value) {
				const item = this.colors.find(color => color.value === value)
				return item ? item.color : value
			},
			deleteTagById(id) {
				removeTag(id).then(res => {
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

	.tag-color-view,
	.tag-color-option {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.tag-color-option {
		justify-content: space-between;
		width: 100%;
	}

	.tag-color-swatch {
		display: inline-block;
		flex: 0 0 auto;
		width: 56px;
		height: 22px;
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 4px;
		box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.18);
	}

	.tag-color-key {
		color: #909399;
		font-size: 13px;
	}

	.tag-color-empty {
		color: #c0c4cc;
	}
</style>
