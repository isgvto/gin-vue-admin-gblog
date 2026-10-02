import service from '@/utils/request'

async function requestModel(config) {
  const result = await service(config)
  if (result?.code !== 0) throw new Error(result?.msg || '模型配置请求失败')
  return result
}

export function getModelConfigList(params) {
  return requestModel({
    url: '/ai/modelConfig/list',
    method: 'GET',
    params
  })
}

export function createModelConfig(data) {
  return requestModel({
    url: '/ai/modelConfig',
    method: 'POST',
    data
  })
}

export function updateModelConfig(data) {
  return requestModel({
    url: '/ai/modelConfig',
    method: 'PUT',
    data
  })
}

export function deleteModelConfig(id) {
  return requestModel({
    url: `/ai/modelConfig/${id}`,
    method: 'DELETE'
  })
}

export function setDefaultModelConfig(id) {
  return requestModel({
    url: `/ai/modelConfig/setDefault/${id}`,
    method: 'PUT'
  })
}

export function getModelProviders() {
  return requestModel({
    url: '/ai/modelConfig/providers',
    method: 'GET'
  })
}

export const testModelConnection = (data) => requestModel({
  url: '/ai/modelConfig/testConnection', method: 'POST', data, timeout: 35000
})

export const getProviderModels = (data) => requestModel({
  url: '/ai/modelConfig/providerModels', method: 'POST', data, timeout: 35000
})

export const getErrorAnalysisConfig = () => requestModel({ url: '/ai/modelConfig/errorAnalysis', method: 'GET' })
export const saveErrorAnalysisConfig = (data) => requestModel({ url: '/ai/modelConfig/errorAnalysis', method: 'PUT', data })
export const testErrorAnalysis = (data) => requestModel({ url: '/ai/modelConfig/errorAnalysis/test', method: 'POST', data, timeout: (data.timeoutSeconds + 15) * 1000 })
