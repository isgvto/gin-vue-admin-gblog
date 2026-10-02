<template>
  <div class="admin-page admin-page--list model-config-page">
    <PageHeading title="AI 模型配置" description="管理供应商、访问凭据与默认模型，验证模型连接是否可用。">
      <el-button type="primary" icon="Plus" @click="openCreateDialog">新增模型</el-button>
    </PageHeading>
    <div class="gva-table-box">
    <el-form inline class="admin-filter-form" @submit.prevent="getData">
      <el-form-item label="名称">
        <el-input
          v-model="queryInfo.name"
          placeholder="按名称搜索"
          clearable
          style="width: 200px"
          @keyup.enter="getData"
          @clear="getData"
        />
      </el-form-item>
      <el-form-item label="供应商">
        <el-select
          v-model="queryInfo.provider"
          placeholder="全部供应商"
          clearable
          style="width: 180px"
          @change="getData"
        >
          <el-option
            v-for="p in providers"
            :key="p.value"
            :label="p.label"
            :value="p.value"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="modelList">
      <el-table-column label="名称" prop="name" min-width="130" />
      <el-table-column label="供应商" width="130">
        <template #default="{ row }">{{ providerLabel(row.provider) }}</template>
      </el-table-column>
      <el-table-column label="模型" prop="model" min-width="130" />
      <el-table-column label="API Key" width="110">
        <template #default="{ row }">
          <span v-if="row.hasKey">****{{ row.keyTail }}</span>
          <el-tag v-else type="danger" size="small">未配置</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status ? 'success' : 'info'" size="small">
            {{ row.status ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="默认" width="90" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.isDefault" type="warning" size="small">默认</el-tag>
          <el-button
            v-else
            link
            type="primary"
            size="small"
            :disabled="!row.status"
            @click="setDefault(row.id)"
          >
            设为默认
          </el-button>
        </template>
      </el-table-column>
      <el-table-column label="备注" prop="remark" min-width="120" show-overflow-tooltip />
      <el-table-column label="操作" width="180" :fixed="compactTable ? false : 'right'" align="center">
        <template #default="{ row }">
          <el-button type="primary" link @click="openEditDialog(row)">编辑</el-button>
          <el-button link type="primary" :loading="testingId === row.id" :disabled="testingId !== null" @click="testSaved(row)">测试</el-button>
          <el-popconfirm title="确定删除该模型配置吗？" @confirm="remove(row.id)">
            <template #reference>
              <el-button type="danger" link :disabled="row.isDefault">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="pagination"
      :current-page="queryInfo.page"
      :page-sizes="[10, 20, 30, 50]"
      :page-size="queryInfo.pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
      background
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />

    </div>

    <ErrorAnalysisSettings ref="errorAnalysisSettings" />

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? '编辑模型' : '新增模型'"
      width="min(640px, 94vw)"
      class="admin-config-dialog"
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="显示名称" prop="name">
          <el-input v-model="form.name" placeholder="如 DeepSeek-V3" />
        </el-form-item>
        <el-form-item label="供应商" prop="provider">
          <el-select v-model="form.provider" class="full-width" @change="providerChanged">
            <el-option
              v-for="p in providers"
              :key="p.value"
              :label="p.label"
              :value="p.value"
            />
          </el-select>
          <div v-if="providerHint" class="provider-hint">{{ providerHint }}</div>
        </el-form-item>
        <el-form-item v-if="needBaseUrl" label="Base URL" prop="baseUrl">
          <el-input autocomplete="off" name="ai-config-base-url" v-model="form.baseUrl" :placeholder="baseUrlPlaceholder" @blur="autoLoadProviderModels" />
        </el-form-item>
        <el-form-item label="API Key" prop="apiKey">
          <el-input
            v-model="form.apiKey"
            type="password"
            autocomplete="new-password"
            name="ai-config-api-key"
            show-password
            clearable
            :placeholder="form.id ? '留空表示不修改' : 'sk-...'"
            @blur="autoLoadProviderModels"
          />
        </el-form-item>
        <el-form-item label="模型标识" prop="model">
          <el-select v-model="form.model" filterable allow-create default-first-option class="full-width" placeholder="读取后选择，或手动输入模型 / ep-xxx">
            <el-option v-for="item in providerModels" :key="item.id" :label="item.name === item.id ? item.id : `${item.name} (${item.id})`" :value="item.id" />
          </el-select>
          <el-button size="small" :loading="loadingModels" @click="loadProviderModels">读取供应商模型列表</el-button>
          <div class="provider-hint">列表不代表模型一定可调用，选择后请测试连接。</div>
          <div class="provider-hint">测试连接会发送一次简短请求，供应商可能计费。</div>
        </el-form-item>
        <el-form-item label="温度" prop="temperature">
          <el-slider v-model="form.temperature" :min="0" :max="2" :step="0.1" show-input />
        </el-form-item>
        <el-form-item label="最大输出" prop="maxTokens">
          <el-input-number v-model="form.maxTokens" :min="256" :max="65536" :step="256" />
        </el-form-item>
        <el-form-item label="启用" prop="status">
          <el-switch v-model="form.status" :disabled="form.isDefault" />
          <span v-if="form.isDefault" class="provider-hint">请先将其他模型设为默认，再停用此模型</span>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :loading="testingDraft" @click="testDraft">测试连接</el-button>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import {
    getModelConfigList,
    createModelConfig,
    updateModelConfig,
    deleteModelConfig,
    setDefaultModelConfig,
    getModelProviders,
    getProviderModels,
    testModelConnection
  } from '@/api/ai/modelConfig'
  import { ElMessage } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import PageHeading from '@/components/admin/PageHeading.vue'

  import ErrorAnalysisSettings from '@/components/ai/ErrorAnalysisSettings.vue'

  export default {
    name: 'AiModelConfig',
    components: { PageHeading, ErrorAnalysisSettings },
    data() {
      return {
        loading: false,
        submitting: false,
        testingId: null,
        testingDraft: false,
        loadingModels: false,
        providerModels: [],
        modelListRequest: 0,
        modelList: [],
        total: 0,
        providers: [],
        queryInfo: {
          name: '',
          provider: '',
          page: 1,
          pageSize: 10
        },
        dialogVisible: false,
        form: this.createEmptyForm(),
        rules: {
          name: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
          provider: [{ required: true, message: '请选择供应商', trigger: 'change' }],
          model: [{ required: true, message: '请输入模型标识', trigger: 'blur' }],
          apiKey: [
            {
              validator: (rule, value, callback) => {
                if (!this.form.id && !value) {
                  callback(new Error('新增时 API Key 不能为空'))
                  return
                }
                callback()
              },
              trigger: 'blur'
            }
          ]
        }
      }
    },
    computed: {
      compactTable() { return useAppStore().device === 'mobile' },
      currentProvider() {
        return this.providers.find((p) => p.value === this.form.provider)
      },
      needBaseUrl() {
        return this.currentProvider ? this.currentProvider.needBaseUrl : true
      },
      providerHint() {
        return this.currentProvider?.hint || ''
      },
      baseUrlPlaceholder() {
        return this.currentProvider?.baseURLPlaceholder || 'https://...'
      }
    },
    created() {
      this.getProviders()
      this.getData()
    },
    watch: {
      'form.provider': 'invalidateProviderModels',
      'form.baseUrl': 'invalidateProviderModels',
      'form.apiKey': 'invalidateProviderModels'
    },
    methods: {
      autoLoadProviderModels() {
        if (this.loadingModels || this.providerModels.length || !(this.form.apiKey || this.form.id)) return
        if (this.form.provider === 'openai' && !this.form.baseUrl) return
        this.loadProviderModels()
      },
      invalidateProviderModels() {
        this.modelListRequest++
        this.providerModels = []
      },
      async loadProviderModels() {
        const request = ++this.modelListRequest
        this.loadingModels = true
        try {
          const res = await getProviderModels({ ...this.form })
          if (request !== this.modelListRequest) return
          this.providerModels = res.data || []
          if (!this.providerModels.length) ElMessage.info('未返回可用模型，请手动输入模型标识')
        } catch (_) {
          // 请求层展示供应商错误，保留手动输入。
        } finally {
          this.loadingModels = false
        }
      },
      async testDraft() {
        this.testingDraft = true
        try {
          const res = await testModelConnection({ ...this.form })
          ElMessage.success(res.msg)
        } catch (_) {
          // 请求层展示连接错误。
        } finally { this.testingDraft = false }
      },
      async testSaved(row) {
        this.testingId = row.id
        try {
          const res = await testModelConnection({ ...row, apiKey: '' })
          ElMessage.success(res.msg)
        } catch (_) {
          // 请求层展示连接错误。
        } finally { this.testingId = null }
      },
      createEmptyForm() {
        return {
          id: 0,
          name: '',
          provider: 'openai',
          baseUrl: '',
          apiKey: '',
          model: '',
          temperature: 0.7,
          maxTokens: 4096,
          status: true,
          remark: ''
        }
      },
      async getProviders() {
        try {
          const res = await getModelProviders()
          this.providers = res.data || []
        } catch (error) {
          this.providers = []
        }
      },
      providerLabel(value) {
        const provider = this.providers.find((p) => p.value === value)
        return provider ? provider.label : value
      },
      async getData() {
        this.loading = true
        try {
          const res = await getModelConfigList(this.queryInfo)
          this.modelList = (res.data?.list || []).map(row => ({ ...row, id: row.id ?? row.ID }))
          this.total = res.data?.total || 0
          this.$refs.errorAnalysisSettings?.refreshModels()
        } finally {
          this.loading = false
        }
      },
      handleSizeChange(size) {
        this.queryInfo.pageSize = size
        this.getData()
      },
      handleCurrentChange(page) {
        this.queryInfo.page = page
        this.getData()
      },
      providerChanged() {
        this.form.baseUrl = ''
        this.form.model = ''
        this.$nextTick(() => this.autoLoadProviderModels())
      },
      openCreateDialog() {
        this.invalidateProviderModels()
        this.form = this.createEmptyForm()
        this.dialogVisible = true
      },
      openEditDialog(row) {
        this.invalidateProviderModels()
        this.form = {
          id: row.id,
          isDefault: row.isDefault,
          name: row.name,
          provider: row.provider,
          baseUrl: row.baseUrl,
          apiKey: '',
          model: row.model,
          temperature: row.temperature,
          maxTokens: row.maxTokens,
          status: row.status,
          remark: row.remark || ''
        }
        this.dialogVisible = true
      },
      resetForm() {
        this.$refs.formRef?.clearValidate?.()
      },
      submit() {
        this.$refs.formRef.validate(async (valid) => {
          if (!valid) return
          this.submitting = true
          try {
            if (this.form.id) {
              await updateModelConfig(this.form)
              ElMessage.success('更新成功')
            } else {
              await createModelConfig(this.form)
              ElMessage.success('创建成功')
            }
            this.dialogVisible = false
            this.getData()
          } finally {
            this.submitting = false
          }
        })
      },
      async setDefault(id) {
        await setDefaultModelConfig(id)
        ElMessage.success('已设为默认，AI 功能将即时使用该模型')
        this.getData()
      },
      async remove(id) {
        await deleteModelConfig(id)
        ElMessage.success('删除成功')
        this.getData()
      }
    }
  }
</script>

<style scoped lang="scss">
.model-config-page {
  min-width: 0;
}

.pagination {
  margin-top: 12px;
}

.full-width {
  width: 100%;
}

.provider-hint {
  width: 100%;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}
</style>
