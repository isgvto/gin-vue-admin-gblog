<template>
  <div class="admin-page admin-page--editor write-blog-page">
    <el-form :model="form" :rules="formRules" ref="formRef" label-position="top" class="writer-form">
      <div class="writer-shell">
        <main class="writer-main">
          <div class="writer-header">
            <div>
              <div class="writer-title">{{ pageTitle }}</div>
              <div class="writer-subtitle">
                {{ contentStats.words }} 字 · {{ form.readTime || 0 }} 分钟阅读
                <span v-if="draftStatusText"> · {{ draftStatusText }}</span>
              </div>
            </div>
            <div class="writer-actions">
              <el-button icon="Back" @click="$router.back()">返回</el-button>
              <el-button icon="DocumentAdd" @click="saveDraft">保存草稿</el-button>
              <el-button icon="Delete" @click="clearDraftCache">清空缓存</el-button>
              <el-button type="primary" icon="Check" @click="openPublishDialog">保存</el-button>
            </div>
          </div>

          <section class="section-block">
            <el-row :gutter="16">
              <el-col :xs="24" :sm="24" :md="14">
                <el-form-item label="文章标题" prop="title">
                  <el-input v-model="form.title" placeholder="请输入标题" maxlength="120" show-word-limit />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="24" :md="10">
                <el-form-item label="文章首图 URL" prop="firstPicture">
                  <el-input v-model="form.firstPicture" placeholder="文章首图，用于随机文章展示" clearable />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="文章摘要" prop="description">
              <MarkdownEditor
                v-model="form.description"
                height="260px"
                placeholder="请输入文章摘要，支持 Markdown"
              />
            </el-form-item>
          </section>

          <section class="section-block content-section">
            <div class="section-title">
              <span>文章正文</span>
              <el-button size="small" icon="Refresh" @click="recalculateWords">重新统计</el-button>
            </div>
            <el-form-item prop="content">
              <MarkdownEditor
                ref="contentEditorRef"
                :document-id="draftKey"
                v-model="form.content"
                height="680px"
                enable-ai-diff
                placeholder="从这里开始写作，支持标题、引用、列表、代码块、表格、图片和实时预览"
              />
            </el-form-item>
          </section>
        </main>

        <aside class="writer-side">
          <section class="side-panel">
            <div class="side-title">发布设置</div>
            <el-radio-group v-model="radio" class="visibility-group">
              <el-radio :label="1">公开</el-radio>
              <el-radio :label="2">私密</el-radio>
              <el-radio :label="3">密码保护</el-radio>
            </el-radio-group>
            <el-form-item v-if="radio === 3" label="访问密码" prop="password">
              <el-input v-model="form.password" show-password clearable />
            </el-form-item>
            <div v-if="radio !== 2" class="switch-grid">
              <el-checkbox v-model="form.appreciation">赞赏</el-checkbox>
              <el-checkbox v-model="form.recommend">推荐</el-checkbox>
              <el-checkbox v-model="form.commentEnabled">评论</el-checkbox>
              <el-checkbox v-model="form.top">置顶</el-checkbox>
            </div>
          </section>

          <section class="side-panel">
            <div class="side-title">分类与标签</div>
            <el-form-item label="分类" prop="cate">
              <el-select
                v-model="form.cate"
                placeholder="请选择分类（输入可添加）"
                allow-create
                filterable
                clearable
                class="full-width"
              >
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.categoryName"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="标签" prop="tagList">
              <el-select
                v-model="form.tagList"
                placeholder="请选择标签（输入可添加）"
                allow-create
                filterable
                multiple
                class="full-width"
              >
                <el-option
                  v-for="item in tagList"
                  :key="item.id"
                  :label="item.tagName"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </section>

          <section class="side-panel">
            <div class="side-title">文章数据</div>
            <el-form-item label="字数" prop="words">
              <el-input-number
                v-model="form.words"
                :min="0"
                :controls="false"
                class="full-width"
                @change="autoWords = false"
              />
            </el-form-item>
            <el-form-item label="阅读时长(分钟)" prop="readTime">
              <el-input-number v-model="form.readTime" :min="0" :controls="false" class="full-width" />
            </el-form-item>
            <el-form-item label="浏览次数" prop="views">
              <el-input-number v-model="form.views" :min="0" :controls="false" class="full-width" />
            </el-form-item>
          </section>
        </aside>
      </div>
    </el-form>

    <el-dialog title="确认保存" width="460px" v-model="dialogVisible">
      <div class="publish-summary">
        <div>
          <span>标题</span>
          <strong>{{ form.title || '未填写' }}</strong>
        </div>
        <div>
          <span>可见性</span>
          <strong>{{ visibilityText }}</strong>
        </div>
        <div>
          <span>统计</span>
          <strong>{{ form.words || 0 }} 字 · {{ form.readTime || 0 }} 分钟</strong>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" icon="Check" :loading="isSubmitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import {
    getCategoryAndTag,
    saveBlog as createBlog,
    getBlogById,
    updateBlog
  } from '@/api/blog/article'
  import MarkdownEditor from '@/components/blog/MarkdownEditor.vue'
  import { useAiStore } from '@/pinia/modules/ai'
  import { titleFillError } from '@/components/ai/agents/writing-assistant/writingTask.js'
  import { buildSuggestionPatch, cursorContext } from '@/components/ai/agents/writing-assistant/suggestion.js'

  const createEmptyForm = () => ({
    title: '',
    firstPicture: '',
    description: '',
    content: '',
    cate: null,
    tagList: [],
    words: 0,
    readTime: 0,
    views: 0,
    appreciation: false,
    recommend: false,
    commentEnabled: false,
    top: false,
    published: false,
    password: '',
  })

  const countMarkdownWords = (content = '') => {
    const words = content
      .replace(/```[\s\S]*?```/g, ' ')
      .replace(/[#>*_`~\-[\]()!|]/g, ' ')
      .match(/[\u4e00-\u9fa5]|[a-zA-Z0-9]+/g)
    return words ? words.length : 0
  }

  export default {
    name: 'BlogWriteArticle',
    components: {
      MarkdownEditor
    },
    data() {
      return {
        categoryList: [],
        tagList: [],
        dialogVisible: false,
        lastDraftSavedAt: null,
        isSubmitting: false,
        radio: 1,
        autoWords: true,
        form: createEmptyForm(),
        formRules: {
          title: [{ required: true, message: '请输入标题', trigger: 'change' }],
          firstPicture: [{ required: true, message: '请输入首图链接', trigger: 'change' }],
          description: [{ required: true, message: '请输入文章摘要', trigger: 'change' }],
          content: [{ required: true, message: '请输入文章正文', trigger: 'change' }],
          cate: [{ required: true, message: '请选择分类', trigger: 'change' }],
          tagList: [{ required: true, message: '请选择标签', trigger: 'change' }],
          words: [{ required: true, message: '请输入文章字数', trigger: 'change' }],
          password: [{ validator: (rule, value, callback) => {
            if (this.radio === 3 && !value) {
              callback(new Error('密码保护模式必须填写密码'))
              return
            }
            callback()
          }, trigger: 'change' }],
        },
      }
    },
    computed: {
      pageTitle() {
        return this.$route.params.id ? '编辑文章' : '写新文章'
      },
      contentStats() {
        const content = this.form.content || ''
        return {
          words: countMarkdownWords(content),
          characters: content.length
        }
      },
      visibilityText() {
        if (this.radio === 1) return '公开'
        if (this.radio === 2) return '私密'
        return '密码保护'
      },
      draftKey() {
        return `blog-write-draft:${this.$route.params.id || 'new'}`
      },
      draftStatusText() {
        if (!this.lastDraftSavedAt) return ''
        const time = new Date(this.lastDraftSavedAt).toLocaleTimeString('zh-CN', {
          hour: '2-digit',
          minute: '2-digit'
        })
        return `草稿已保存 ${time}`
      }
    },
    watch: {
      'form.content'() {
        if (this.autoWords) {
          this.recalculateWords()
        }
      },
      'form.words'(newValue) {
        this.form.readTime = newValue ? Math.max(1, Math.round(newValue / 200)) : 0
      },
    },
    created() {
      this.getData()
      if (this.$route.params.id) {
        this.getBlog(this.$route.params.id)
      } else {
        this.restoreDraft()
      }
    },
    mounted() {
      this.registerAiEditor()
    },
    activated() {
      this.registerAiEditor()
    },
    deactivated() {
      this.aiStore?.unregisterContext?.('editor', this.aiEditorContext)
    },
    beforeUnmount() {
      this.aiStore?.unregisterContext?.('editor', this.aiEditorContext)
    },
    methods: {
      registerAiEditor() {
        this.aiStore = useAiStore()
        this.aiEditorContext ||= this.buildEditorContext()
        this.aiStore.registerContext('editor', this.aiEditorContext)
      },
      buildEditorContext() {
        return {
          getEditorState: () => this.$refs.contentEditorRef?.getEditorState?.(),
          captureSelection: () => this.$refs.contentEditorRef?.captureSelection?.(),
          applySelectionSnapshot: (snapshot, text) => this.$refs.contentEditorRef?.applySelectionSnapshot?.(snapshot, text),
          getSelection: () =>
            this.$refs.contentEditorRef?.getSelection?.() || { text: '', start: 0, end: 0 },
          replaceSelection: (text) => this.$refs.contentEditorRef?.replaceSelection?.(text),
          insertAtCursor: (text) => this.$refs.contentEditorRef?.insertAtCursor?.(text),
          appendChapter: (text, expected) => this.$refs.contentEditorRef?.appendChapter?.(text, expected),
          getFullText: () => this.form.content || '',
          getCursorContext: () => {
            const sel = this.$refs.contentEditorRef?.getSelection?.()
            const content = this.form.content || ''
            return cursorContext(content, sel?.start)
          },
          getTitle: () => this.form.title || '',
          fillTitle: (title, expectedTitle) => {
            const message = titleFillError(title, expectedTitle, this.form.title || '')
            if (message) return { ok: false, message }
            this.form.title = title.trim()
            return { ok: true }
          },
          fillDescription: (text) => {
            if (!text) return false
            this.form.description = text
            return true
          },
          applySuggestion: (suggestion) => {
            const result = buildSuggestionPatch(this.form, suggestion, this.categoryList, this.tagList)
            if (result.ok) Object.assign(this.form, result.patch)
            return result
          }
        }
      },
      getData() {
        getCategoryAndTag().then(res => {
          this.categoryList = res.data.categories
          this.tagList = res.data.tags
        })
      },
      getBlog(id) {
        getBlogById(id).then(res => {
          this.form = this.normalizeBlogForm(res.data)
          this.autoWords = false
          this.radio = this.form.published ? (this.form.password !== '' ? 3 : 1) : 2
          this.restoreDraft()
        })
      },
      normalizeBlogForm(blog = {}) {
        return {
          ...createEmptyForm(),
          ...blog,
          cate: blog.category?.id ?? blog.cate ?? blog.categoryId ?? null,
          tagList: Array.isArray(blog.tags)
            ? blog.tags.map(item => item.id)
            : (Array.isArray(blog.tagList) ? blog.tagList : [])
        }
      },
      recalculateWords() {
        this.autoWords = true
        this.form.words = this.contentStats.words
      },
      openPublishDialog() {
        if (this.autoWords || !this.form.words) {
          this.recalculateWords()
        }
        this.$refs.formRef.validate(valid => {
          if (!valid) {
            return this.msgError('请填写必要的表单项')
          }
          this.dialogVisible = true
        })
      },
      submit() {
        this.$refs.formRef.validate(valid => {
          if (valid) {
            this.isSubmitting = true
            const payload = this.buildSubmitPayload()
            const request = this.$route.params.id ? updateBlog(payload) : createBlog(payload)
            request.then(res => {
              this.clearDraftCache(false)
              this.msgSuccess(res.msg)
              this.$router.push('/layout/gblog/list')
            }).finally(() => {
              this.isSubmitting = false
            })
          } else {
            this.dialogVisible = false
            return this.msgError('请填写必要的表单项')
          }
        })
      },
      buildSubmitPayload() {
        const payload = {
          ...this.form,
          category: null,
          tags: null
        }

        if (this.radio === 2) {
          payload.appreciation = false
          payload.recommend = false
          payload.commentEnabled = false
          payload.top = false
          payload.published = false
        } else {
          payload.published = true
        }

        if (this.radio !== 3) {
          payload.password = ''
        }

        return payload
      },
      serializeDraftForm() {
        const source = this.form || {}
        return {
          ...createEmptyForm(),
          id: source.id,
          title: source.title,
          firstPicture: source.firstPicture,
          description: source.description,
          content: source.content,
          cate: source.cate,
          tagList: Array.isArray(source.tagList) ? source.tagList : [],
          words: source.words,
          readTime: source.readTime,
          views: source.views,
          appreciation: source.appreciation,
          recommend: source.recommend,
          commentEnabled: source.commentEnabled,
          top: source.top,
          published: source.published,
          password: source.password,
        }
      },
      saveDraft() {
        if (!this.hasDraftContent()) {
          return this.msgError('没有可保存的草稿内容')
        }
        const payload = {
          savedAt: Date.now(),
          radio: this.radio,
          autoWords: this.autoWords,
          form: this.serializeDraftForm()
        }

        localStorage.setItem(this.draftKey, JSON.stringify(payload))
        this.lastDraftSavedAt = payload.savedAt
        this.msgSuccess('草稿已保存到本地缓存')
      },
      restoreDraft() {
        const rawDraft = localStorage.getItem(this.draftKey)
        if (rawDraft) {
          try {
            const draft = JSON.parse(rawDraft)
            if (draft?.form && this.hasDraftContent(draft.form)) {
              this.form = {
                ...createEmptyForm(),
                ...this.form,
                ...draft.form
              }
              this.radio = draft.radio || this.radio
              this.autoWords = draft.autoWords ?? this.autoWords
              this.lastDraftSavedAt = draft.savedAt || null
              this.$message.success('已恢复本地草稿')
            }
          } catch (error) {
            localStorage.removeItem(this.draftKey)
          }
        }
      },
      clearDraftCache(showMessage = true) {
        localStorage.removeItem(this.draftKey)
        this.lastDraftSavedAt = null
        if (showMessage) {
          this.msgSuccess('本地缓存已清空')
        }
      },
      hasDraftContent(form = this.form) {
        if (!form) return false
        const textFields = [
          form.title,
          form.firstPicture,
          form.description,
          form.content,
          form.password
        ]
        return textFields.some(value => String(value || '').trim()) ||
          Boolean(form.cate) ||
          (Array.isArray(form.tagList) && form.tagList.length > 0)
      },
    }
  }
</script>

<style scoped lang="scss">
.write-blog-page {
  min-height: calc(100vh - 100px);
  padding: 16px;
  background: #f5f7fa;
}

.writer-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 16px;
  align-items: start;
}

.writer-main,
.writer-side {
  min-width: 0;
}

.writer-header,
.section-block,
.side-panel {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #fff;
}

.writer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  padding: 14px 16px;
}

.writer-title {
  color: #303133;
  font-size: 20px;
  font-weight: 700;
}

.writer-subtitle {
  margin-top: 4px;
  color: #909399;
  font-size: 13px;
}

.writer-actions {
  display: flex;
  flex-shrink: 0;
  gap: 8px;
}

.section-block {
  margin-bottom: 16px;
  padding: 16px;
}

.content-section {
  padding-bottom: 8px;
}

.section-title,
.side-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  color: #303133;
  font-size: 15px;
  font-weight: 700;
}

.writer-side {
  position: sticky;
  top: 16px;
}

.side-panel {
  margin-bottom: 16px;
  padding: 16px;
}

.visibility-group {
  display: grid;
  gap: 8px;
  margin-bottom: 14px;
}

.switch-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 12px;
}

.full-width {
  width: 100%;
}

.publish-summary {
  display: grid;
  gap: 12px;

  div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  span {
    color: #909399;
  }

  strong {
    max-width: 280px;
    overflow: hidden;
    color: #303133;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

:deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

:deep(.el-form-item__content > .markdown-editor) {
  width: 100%;
}

@media (max-width: 1200px) {
  .writer-shell {
    grid-template-columns: 1fr;
  }

  .writer-side {
    position: static;
  }
}

@media (max-width: 768px) {
  .write-blog-page {
    padding: 10px;
  }

  .writer-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .writer-actions {
    width: 100%;

    .el-button {
      flex: 1;
    }
  }
}
</style>
