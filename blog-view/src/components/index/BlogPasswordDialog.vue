<template>
	<!--私密文章密码对话框-->
	<el-dialog title="请输入受保护文章密码" width="30%" custom-class="blog-password-dialog" :visible.sync="blogPasswordDialogVisible"
	           :lock-scroll="false" :before-close="blogPasswordDialogClosed" @opened="focusPasswordInput">
		<!--内容主体-->
		<el-form :model="blogPasswordForm" :rules="formRules" ref="formRef" label-width="80px">
			<el-form-item label="密码" prop="password">
				<el-input
					ref="passwordInput"
					v-model="blogPasswordForm.password"
					show-password
					autocomplete="current-password"
					inputmode="text"
					@keyup.native.enter="submitBlogPassword"
				></el-input>
			</el-form-item>
		</el-form>
		<!--底部-->
		<span slot="footer">
			<el-button @click="blogPasswordDialogClosed">取 消</el-button>
			<el-button type="primary" @click="submitBlogPassword">确 定</el-button>
		</span>
	</el-dialog>
</template>

<script>
	import {mapState} from "vuex";
	import {SET_BLOG_PASSWORD_DIALOG_VISIBLE} from "../../store/mutations-types";
	import {checkBlogPassword} from "@/api/blog";
	import {isSuccess} from "@/util/gvaResponse";

	export default {
		name: "BlogPasswordDialog",
		computed: {
			...mapState(['blogPasswordDialogVisible', 'blogPasswordForm'])
		},
		data() {
			return {
				formRules: {
					password: [{required: true, message: '请输入密码', trigger: 'change'}]
				}
			}
		},
		methods: {
			focusPasswordInput() {
				this.$nextTick(() => {
					if (this.$refs.passwordInput) {
						this.$refs.passwordInput.focus()
					}
				})
			},
			blogPasswordDialogClosed() {
				this.$refs.formRef.resetFields()
				this.$store.commit(SET_BLOG_PASSWORD_DIALOG_VISIBLE, false)
			},
			submitBlogPassword() {
				this.$refs.formRef.validate(valid => {
					if (valid) {
						checkBlogPassword(this.blogPasswordForm).then(res => {
							if (isSuccess(res)) {
								const blogId = this.blogPasswordForm.blogId
								this.msgSuccess(res.msg)
								window.localStorage.setItem(`blog${blogId}`, res.data)
								if (this.isCurrentBlogRoute(blogId)) {
									window.dispatchEvent(new CustomEvent('blog-password-verified', {detail: {blogId}}))
								} else {
									this.$router.push({name: 'blog', params: {id: blogId}}).catch(err => {
										if (err.name !== 'NavigationDuplicated') {
											this.msgError(err.message || '璺宠浆澶辫触')
										}
									})
								}
								this.blogPasswordDialogClosed()
							} else {
								this.msgError(res.msg)
							}
						}).catch(() => {
							this.msgError("请求失败")
						})
					}
				})
			},
			isCurrentBlogRoute(blogId) {
				return this.$route.name === 'blog' && String(this.$route.params.id) === String(blogId)
			}
		}
	}
</script>

<style>
	.blog-password-dialog {
		max-width: 420px;
		border-radius: 10px;
		overflow: hidden;
	}

	.blog-password-dialog .el-dialog__header {
		padding: 20px 24px 12px;
		border-bottom: 1px solid #f0f2f5;
	}

	.blog-password-dialog .el-dialog__body {
		padding: 24px;
	}

	.blog-password-dialog .el-dialog__footer {
		padding: 12px 24px 20px;
		border-top: 1px solid #f0f2f5;
	}

	.blog-password-dialog .el-form-item {
		margin-bottom: 0;
	}

	.blog-password-dialog .el-input,
	.blog-password-dialog .el-input__inner {
		width: 100%;
	}

	@media screen and (max-width: 600px) {
		.blog-password-dialog.el-dialog {
			width: calc(100% - 28px) !important;
			margin: 12vh auto 0 !important;
		}

		.blog-password-dialog .el-dialog__header {
			padding: 18px 18px 12px;
		}

		.blog-password-dialog .el-dialog__body {
			padding: 18px 18px 12px;
		}

		.blog-password-dialog .el-dialog__footer {
			padding: 10px 18px 16px;
		}

		.blog-password-dialog .el-form-item__label {
			display: block;
			float: none;
			width: auto !important;
			padding: 0 0 7px;
			line-height: 1.4;
			text-align: left;
		}

		.blog-password-dialog .el-form-item__content {
			margin-left: 0 !important;
		}

		.blog-password-dialog .el-button {
			min-width: 72px;
		}
	}
</style>
