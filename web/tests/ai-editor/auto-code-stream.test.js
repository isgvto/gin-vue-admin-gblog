import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import vm from 'node:vm'
import { test } from 'node:test'

// Exercise the production request module with only its network/store imports stubbed.
const source = (await readFile(new URL('../../src/api/autoCode.js', import.meta.url), 'utf8'))
  .replaceAll('\r\n', '\n')
  .replace(/^import .*\n/gm, '').replace(/export const /g, 'const ')
  .replaceAll("import.meta.env.VITE_BASE_API", "'/api'")
const load = (reply) => {
  let request
  const context = vm.createContext({
    TextDecoder, AbortController, setTimeout, clearTimeout,
    console: { debug() {}, warn() {} },
    service: async config => config,
    useUserStore: () => ({ token: 'test-token', userInfo: { ID: 7 } }),
    fetch: async (url, options) => { request = { url, ...options }; return reply() }
  })
  vm.runInContext(source + '\nglobalThis.api = { analyzeRequirementByAISSEStream, generatePromptFlowByAI };', context)
  return { api: context.api, request: () => request }
}
const sseResponse = text => {
  const bytes = new TextEncoder().encode(text)
  return new Response(new ReadableStream({ start(controller) {
    for (let i = 0; i < bytes.length; i += 17) controller.enqueue(bytes.slice(i, i + 17))
    controller.close()
  } }), { headers: { 'Content-Type': 'text/event-stream' } })
}
test('workflow uses the authenticated endpoint and repeated deltas are retained', async () => {
  const harness = load(() => sseResponse('data: {"event":"message","delta":"好"}\n\ndata: {"event":"message","delta":"好"}\n\n'))
  const result = await harness.api.analyzeRequirementByAISSEStream({ query: '需求' })
  assert.equal(result.answer, '好好')
  assert.equal(harness.request().url, '/api/autoCode/aiWorkflowChat')
  assert.equal(harness.request().headers['x-token'], 'test-token')
})
test('SSE business errors reject instead of becoming successful text', async () => {
  const harness = load(() => sseResponse('data: {"event":"error","message":"模型结果无效"}\n\n'))
  await assert.rejects(harness.api.analyzeRequirementByAISSEStream({ query: '需求' }), /模型结果无效/)
})
test('final validated answer replaces the draft, including a final event without newline', async () => {
  const harness = load(() => sseResponse('data: {"event":"message","delta":"draft"}\n\ndata: {"event":"done","answer":"有效结果","structured":{"summary":"结果"}}'))
  const result = await harness.api.analyzeRequirementByAISSEStream({ query: '需求' })
  assert.equal(result.answer, '有效结果')
  assert.equal(result.structured.summary, '结果')
  const config = await harness.api.generatePromptFlowByAI({ mode: 'wrong', query: '需求' })
  assert.equal(config.url, '/autoCode/aiWorkflowChat')
  assert.equal(config.data.mode, 'workflowPromptChat')
  assert.equal(config.data.response_mode, 'blocking')
})
