<template>
	<div class="admin-page admin-page--legacy-form admin-about-page">
    <AdminPageHeading title="关于页面" description="编辑个人介绍、页面内容和评论设置。" />
    <section class="gva-form-box">
		<el-form :model="form" :rules="formRules" ref="formRef" label-position="top">
			<el-form-item label="标题" prop="title" style="width: 50%">
				<el-input v-model="form.title" placeholder="请输入标题"></el-input>
			</el-form-item>

			<el-form-item label="评论开关">
				<el-switch v-model="form.commentEnabled" active-text="评论"></el-switch>
			</el-form-item>

			<el-form-item label="正文" prop="content">
				<mavon-editor v-model="form.content"/>
			</el-form-item>

			<el-form-item style="text-align: right;">
				<el-button type="primary" icon="Check" @click="submit">保存</el-button>
			</el-form-item>
		</el-form>

    </section>
  </div>
</template>

<script>
	import {getAbout, updateAbout} from "@/api/blog/about";

	export default {
		name: 'BlogAbout',
		components: {},
		data() {
			return {
				form: {
					title: '',
					content: '',
					commentEnabled: true
				},
				hasLoaded: false,
				formRules: {
					title: [{required: true, message: '请输入标题', trigger: 'change'}],
				}
			}
		},
		created() {
			this.getData()
		},
		activated() {
			if (this.hasLoaded) {
				this.getData()
			}
		},
		methods: {
			getData() {
				getAbout().then(res => {
					this.form.title = res.data.title
					this.form.content = res.data.content
					this.form.commentEnabled = res.data.commentEnabled
					this.hasLoaded = true
				})
			},
			submit() {
				this.$refs.formRef.validate(valid => {
					if (valid) {
						updateAbout(this.form).then(res => {
							this.msgSuccess(res.msg)
						})
					} else {
						return this.msgError('请填写必要的表单')
					}
				})
			}
		}
	}
</script>

<style scoped>

</style>
