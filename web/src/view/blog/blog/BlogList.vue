<template>
	<div class="admin-page admin-page--list">
		<PageHeading title="文章管理" description="管理文章内容、分类与可见性，集中查看发布和更新状态。">
      <el-button type="primary" icon="Plus" @click="goBlogEditPage()">新建文章</el-button>
    </PageHeading>
    <div class="gva-table-box">
      <el-form inline class="admin-filter-form" @submit.prevent="search">
        <el-form-item label="标题">
          <el-input placeholder="搜索文章标题" v-model="queryInfo.title" clearable @clear="search" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="queryInfo.categoryId" placeholder="全部分类" clearable @change="search">
            <el-option v-for="item in categoryList" :key="item.id" :label="item.categoryName" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button icon="Search" @click="search">查询</el-button><el-button @click="resetSearch">重置</el-button></el-form-item>
      </el-form>

		<el-table :data="blogList">
			<el-table-column label="序号" type="index" width="70" align="center"></el-table-column>
			<el-table-column label="标题" prop="title" show-overflow-tooltip min-width="220"></el-table-column>
			<el-table-column label="分类" prop="category.categoryName" width="100"></el-table-column>
			<el-table-column label="置顶" width="60" align="center">
				<template v-slot="scope">
					<el-switch size="small" :aria-label="`置顶 ${scope.row.title}`" v-model="scope.row.top" @change="blogTopChanged(scope.row)"></el-switch>
				</template>
			</el-table-column>
			<el-table-column label="推荐" width="60" align="center">
				<template v-slot="scope">
					<el-switch size="small" :aria-label="`推荐 ${scope.row.title}`" v-model="scope.row.recommend" @change="blogRecommendChanged(scope.row)"></el-switch>
				</template>
			</el-table-column>
			<el-table-column label="可见性" width="100">
				<template v-slot="scope">
					<el-link icon="Edit" :underline="false" @click="editBlogVisibility(scope.row)">
						{{ scope.row.published ? (scope.row.password !== '' ? '密码保护' : '公开') : '私密' }}
					</el-link>
				</template>
			</el-table-column>
			<el-table-column label="更新时间" width="160">
        <template #default="{ row }">
          <div class="admin-date-cell">{{ blogDateFormat(row.updateTime) }}</div>
          <div class="admin-date-cell admin-secondary-text">创建 {{ blogDateFormat(row.createTime) }}</div>
        </template>
      </el-table-column>
			<el-table-column label="操作" width="140" :fixed="compactTable ? false : 'right'" align="center">
				<template v-slot="scope">
					<el-button type="primary" link icon="Edit" @click="goBlogEditPage(scope.row.id)">编辑</el-button>
					<el-popconfirm title="确定删除吗？" icon="Delete" iconColor="red" @confirm="deleteBlogById(scope.row.id)">
						<template #reference><el-button link type="danger" icon="Delete" >删除</el-button></template>
					</el-popconfirm>
				</template>
			</el-table-column>
		</el-table>

		<!--分页-->
		<el-pagination @size-change="handleSizeChange" @current-change="handleCurrentChange" :current-page="queryInfo.pageNum"
		               :page-sizes="[10, 20, 30, 50]" :page-size="queryInfo.pageSize" :total="total"
		               layout="total, sizes, prev, pager, next, jumper" background>
		</el-pagination>

    </div>

		<!--编辑可见性状态对话框-->
		<el-dialog title="文章可见性" width="min(620px, 94vw)" class="admin-config-dialog" v-model="dialogVisible">
			<!--内容主体-->
			<el-form label-width="50px" @submit.prevent>
				<el-form-item>
					<el-radio-group v-model="radio">
						<el-radio :label="1">公开</el-radio>
						<el-radio :label="2">私密</el-radio>
						<el-radio :label="3">密码保护</el-radio>
					</el-radio-group>
				</el-form-item>
				<el-form-item label="密码" v-if="radio===3">
					<el-input v-model="visForm.password"></el-input>
				</el-form-item>
				<el-form-item v-if="radio!==2">
					<el-row>
						<el-col :span="6">
							<el-switch v-model="visForm.appreciation" active-text="赞赏"></el-switch>
						</el-col>
						<el-col :span="6">
							<el-switch v-model="visForm.recommend" active-text="推荐"></el-switch>
						</el-col>
						<el-col :span="6">
							<el-switch v-model="visForm.commentEnabled" active-text="评论"></el-switch>
						</el-col>
						<el-col :span="6">
							<el-switch v-model="visForm.top" active-text="置顶"></el-switch>
						</el-col>
					</el-row>
				</el-form-item>
			</el-form>
			<!--底部-->
			<template #footer>
				<el-button @click="dialogVisible=false">取 消</el-button>
				<el-button type="primary" @click="saveVisibility">保存</el-button>
			</template>
		</el-dialog>
	</div>
</template>

<script>
  import { useAppStore } from '@/pinia'
  import PageHeading from '@/components/admin/PageHeading.vue'
	import {
		getDataByQuery,
		deleteBlogById as removeBlog,
		updateTop,
		updateRecommend,
		updateVisibility
	} from '@/api/blog/article'

	export default {
		name: 'BlogArticleList',
		components: { PageHeading },
		data() {
			return {
				queryInfo: {
					title: '',
					categoryId: null,
					pageNum: 1,
					pageSize: 10
				},
				blogList: [],
				categoryList: [],
				total: 0,
				dialogVisible: false,
				blogId: 0,
				radio: 1,
				visForm: {
					appreciation: false,
					recommend: false,
					commentEnabled: false,
					top: false,
					published: false,
					password: '',
				}
			}
		},
    computed: { compactTable() { return useAppStore().device === 'mobile' } },
		created() {
			this.getData()
		},
		methods: {
			getData() {
				getDataByQuery(this.queryInfo).then(res => {
					this.blogList = res.data.blogs.list
					this.categoryList = res.data.categories
					this.total = res.data.blogs.total
				})
			},
      resetSearch() {
        this.queryInfo.title = ''
        this.queryInfo.categoryId = null
        this.search()
      },
			search() {
				this.queryInfo.pageNum = 1
				this.queryInfo.pageSize = 10
				this.getData()
			},
			//切换博客置顶状态
			blogTopChanged(row) {
				updateTop(row.id, row.top).then(res => {
					this.msgSuccess(res.msg);
				})
			},
			//切换博客推荐状态
			blogRecommendChanged(row) {
				updateRecommend(row.id, row.recommend).then(res => {
					this.msgSuccess(res.msg);
				})
			},
			//编辑博客可见性
			editBlogVisibility(row) {
				this.visForm = {
					appreciation: row.appreciation,
					recommend: row.recommend,
					commentEnabled: row.commentEnabled,
					top: row.top,
					published: row.published,
					password: row.password,
				}
				this.blogId = row.id
				this.radio = this.visForm.published ? (this.visForm.password !== '' ? 3 : 1) : 2
				this.dialogVisible = true
			},
			//修改博客可见性
			saveVisibility() {
				if (this.radio === 3 && (this.visForm.password === '' || this.visForm.password === null)) {
					return this.msgError("密码保护模式必须填写密码！")
				}
				if (this.radio === 2) {
					this.visForm.appreciation = false
					this.visForm.recommend = false
					this.visForm.commentEnabled = false
					this.visForm.top = false
					this.visForm.published = false
				} else {
					this.visForm.published = true
				}
				if (this.radio !== 3) {
					this.visForm.password = ''
				}
				updateVisibility(this.blogId, this.visForm).then(res => {
					this.msgSuccess(res.msg)
					this.getData()
					this.dialogVisible = false
				})
			},
			//监听 pageSize 改变事件
			handleSizeChange(newSize) {
				this.queryInfo.pageSize = newSize
				this.getData()
			},
			//监听页码改变的事件
			handleCurrentChange(newPage) {
				this.queryInfo.pageNum = newPage
				this.getData()
			},
			goBlogEditPage(id) {
				this.$router.push(id ? `/layout/gblog/edit/${id}` : '/layout/gblog/edit')
			},
			deleteBlogById(id) {
				this.$confirm('此操作将永久删除该博客<strong style="color: red">及其所有评论</strong>，是否删除?<br>建议将博客置为<strong style="color: red">私密</strong>状态！', '提示', {
					confirmButtonText: '确定',
					cancelButtonText: '取消',
					type: 'warning',
					dangerouslyUseHTMLString: true
				}).then(() => {
					removeBlog(id).then(res => {
						this.msgSuccess(res.msg)
						this.getData()
					})
				}).catch(() => {
					this.$message({
						type: 'info',
						message: '已取消删除'
					})
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
