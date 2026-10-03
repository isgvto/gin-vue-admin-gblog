import service from '@/utils/request'
import { useUserStore } from '@/pinia/modules/user'
import { consumeAiStream } from './aiStream.js'

async function requestAi(url, payload, signal) {
  const result = await service({
    url: `/blog/ai/${url}`, method: payload ? 'POST' : 'GET',
    data: payload, signal, donNotShowLoading: true, silentError: true,
    ...(url === 'visual/generate' ? { timeout: 15 * 60 * 1000 } : {})
  })
  if (result?.code !== 0) throw new Error(result?.msg || 'AI 请求失败')
  return result
}

export const getAiStatus = (checkConnection = false) => requestAi(checkConnection ? 'status?checkConnection=true' : 'status')
export const generateSummary = (payload, signal) => requestAi('summary', payload, signal)
export const suggestTags = (payload, signal) => requestAi('suggest-tags', payload, signal)
export const getVisualStatus = () => requestAi('visual/status')
export const planVisual = (payload, signal) => requestAi('visual/plan', payload, signal)
export const generateVisual = (payload, signal) => requestAi('visual/generate', payload, signal)
export const adoptVisual = (payload, signal) => requestAi('visual/adopt', payload, signal)
export const discardVisual = (id) => requestAi('visual/discard', { id })

export function streamAiChat(payload, handlers = {}) {
  const controller = new AbortController()
  const userStore = useUserStore()
  const promise = (async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_BASE_API || ''}/blog/ai/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'x-token': userStore.token },
        body: JSON.stringify(payload), signal: controller.signal
      })
      const token = response.headers.get('new-token')
      if (token) userStore.setToken(token)
      const result = await consumeAiStream(response, handlers)
      return { ...result, aborted: false, errorMessage: null }
    } catch (error) {
      if (controller.signal.aborted) return { aborted: true, completed: false, errorMessage: null }
      const message = error?.message || '网络错误'
      handlers.onError?.(message)
      return { aborted: false, completed: false, errorMessage: message }
    }
  })()
  return { abort: () => controller.abort(), promise }
}
