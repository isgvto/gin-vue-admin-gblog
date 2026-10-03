<template>
	<div class="admin-page admin-page--legacy-form">
    <AdminPageHeading title="编辑动态" description="编写动态内容并设置发布时间与可见性。" />
    <section class="gva-form-box">
		<el-form :model="form" label-position="top">
			<el-form-item label="动态内容" prop="content">
				<mavon-editor v-model="form.content"/>
			</el-form-item>

			<el-form-item label="点赞数" prop="likes" style="width: 50%">
				<el-input v-model="form.likes" type="number" placeholder="可选，默认为 0"></el-input>
			</el-form-item>

			<el-form-item label="创建时间" prop="createTime">
				<el-date-picker v-model="form.createTime" type="datetime" placeholder="可选，默认此刻" :editable="false"></el-date-picker>
			</el-form-item>

			<el-form-item style="text-align: right;">
				<el-button type="info" @click="submit(false)">仅自己可见</el-button>
				<el-button type="primary" @click="submit(true)">发布动态</el-button>
			</el-form-item>
		</el-form>

    </section>
  </div>
</template>

<script>
	import {getMomentById, saveMoment, updateMoment} from "@/api/blog/moment";

	export default {
		name: 'BlogWriteMoment',
		components: {},
		data() {
			return {
				form: {
					content: '',
					createTime: null,
					likes: 0,
					isPublished: false
				},
			}
		},
		created() {
			if (this.$route.params.id) {
				this.getMoment(this.$route.params.id)
			}
		},
		methods: {
			getMoment(id) {
				getMomentById(id).then(res => {
					this.form = res.data
				})
			},
			submit(isPublished) {
				this.form.isPublished = isPublished
				if (this.$route.params.id) {
					updateMoment(this.form).then(res => {
						this.msgSuccess(res.msg)
						this.$router.push('/layout/gblog/moment/list')
					})
				} else {
					saveMoment(this.form).then(res => {
						this.msgSuccess(res.msg)
						this.$router.push('/layout/gblog/moment/list')
					})
				}
			}
		}
	}
</script>

<style scoped>

</style>
