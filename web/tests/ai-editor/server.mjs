import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
import vue from '@vitejs/plugin-vue'
import { randomBytes } from 'node:crypto'

// 独立夹具只加载实际编辑器/助手组件，不连接后台、不改写项目路由映射文件。
const server = await createServer({
  configFile: false,
  root: fileURLToPath(new URL('../../', import.meta.url)),
  cacheDir: 'node_modules/.vite-ai-editor-tests',
  plugins: [vue(), {
    name: 'optional-local-ai-mock',
    configureServer(server) {
      // AI_TEST_MOCK=1 用于手工验证真实抽屉首次挂载和选区交接，不调用模型。
      if (process.env.AI_TEST_MOCK !== '1') return
      const previews = new Map()
      const models = [
        { id: 1, name: '常用模型', model: 'fixture-text', provider: 'openai', status: true, isDefault: true, hasKey: true, keyTail: 'test', baseUrl: 'https://example.test/v1' },
        { id: 2, name: '另一个供应商模型', model: 'fixture-other', provider: 'gemini', status: true, isDefault: false, hasKey: true, keyTail: 'test', baseUrl: 'https://other.example.test/v1' }
      ]
      const features = new Map(['image', 'errorAnalysis', 'workflow'].map(feature => [feature, { enabled: false, modelId: 0, timeoutSeconds: feature === 'errorAnalysis' ? 60 : 180 }]))
      features.set('image', { enabled: true, provider: 'ark', baseUrl: '', model: 'fixture-image', timeoutSeconds: 180, apiKey: 'fixture-key' })
      server.middlewares.use(async (req, res, next) => {
        if (req.url === '/test-api/admin/categoryAndTag') { res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({code:0,data:{categories:[],tags:[]}})); return }
        if (req.url?.startsWith('/test-api/ai/modelConfig/')) {
          const action = req.url.split('/').pop().split('?')[0]
          let data
          if (req.url.split('?')[0] === '/test-api/ai/modelConfig/image/test') {
            res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({ code: 0, msg: '图片测试成功，未保存配置或上传图片' })); return
          }
          if (action === 'list') data = { list: models, total: models.length }
          else if (action === 'providers') data = [{ value: 'openai', label: 'OpenAI 兼容', needBaseUrl: true }, { value: 'gemini', label: 'Gemini' }]
          else if (action === 'testConnection') {
            let body = ''; for await (const chunk of req) body += chunk
            const input = JSON.parse(body)
            res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({ code: 0, msg: input.testMode === 'image' ? '图片测试成功，生图接口已返回有效图片，未上传存储' : '文本测试成功，模型已返回有效响应' })); return
          } else if (features.has(action)) {
            if (req.method === 'PUT') { let body = ''; for await (const chunk of req) body += chunk; const input=JSON.parse(body); if (action==='image') input.apiKey=input.clearKey?'':input.apiKey || features.get(action).apiKey; features.set(action, input) }
            data = { config: features.get(action), models: models.map(({ id, name, model }) => ({ id, name, model })) }
            if (action==='image') { const {apiKey,clearKey,...config}=features.get(action); data={config,hasKey:!!apiKey} }
          } else return next()
          res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({ code: 0, data, msg: '保存成功' }))
        } else if (req.url?.startsWith('/test-api/blog/ai/visual/')) {
          try {
            let body = ''; for await (const chunk of req) body += chunk
            const input = body ? JSON.parse(body) : {}
            const action = req.url.split('/').pop()
            let data
            const mermaid = 'flowchart TD\n A["选择内容"] --> B{"选择类型"}\n B --> C["生成预览"]\n C --> D["采用并上传"]\n classDef main fill:#e8f1ff,stroke:#517cf5,color:#24334a\n class A,C,D main'
            if (action === 'status') data = { imageEnabled: true, imageModel: 'fixture-image', storage: 'aws-s3' }
            else if (action === 'plan') data = { kind: input.kind === 'auto' ? 'flowchart' : input.kind, prompt: '清晰的蓝灰色流程图，圆角节点，保留内容中的步骤。', reason: '这段内容适合用流程图表达步骤。', mermaid }
            else if (action === 'generate') {
              const id = randomBytes(24).toString('hex')
              data = { id, kind: input.kind, prompt: input.prompt, expiresAt: new Date(Date.now() + 3600000).toISOString(), url: '', mermaid: ['flowchart', 'structure'].includes(input.kind) ? mermaid : '', preview: ['cover', 'illustration'].includes(input.kind) ? 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aAlsAAAAASUVORK5CYII=' : '' }
              previews.set(id, data)
            } else if (action === 'adopt') {
              data = previews.get(input.id)
              if (!data || (data.mermaid && !String(input.png).startsWith('data:image/png;base64,'))) throw new Error('invalid preview')
              if (data.mermaid) {
                const png = Buffer.from(input.png.split(',')[1], 'base64')
                const width = png.readUInt32BE(16), height = png.readUInt32BE(20)
                if (Math.max(width, height) < 3200 || Math.max(width, height) > 4096) throw new Error('diagram export resolution regression')
                console.log(`Diagram PNG export verified: ${width} x ${height}`)
              }
              data = { ...data, url: `https://images.example.test/ai-${input.id}.png`, fileId: 1 }; previews.set(input.id, data)
            } else if (action === 'discard') { previews.delete(input.id); data = {} }
            else return next()
            res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({ code: 0, data }))
          } catch { res.statusCode = 400; res.end('Invalid visual fixture request') }
        } else if (req.url === '/test-api/blog/ai/status') {
          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify({ code: 0, data: { enabled: true, model: 'fixture-text' } }))
        } else if (req.url === '/test-api/blog/ai/chat') {
          try {
            let body = ''
            for await (const chunk of req) body += chunk
            const payload = JSON.parse(body)
            res.setHeader('Content-Type', 'text/event-stream')
            let output = `${payload.action}：${payload.selection}`
            if (payload.action === 'outline') output = '# 测试文章\n\n## 背景\n- 介绍问题\n\n## 实践\n- 给出示例'
            if (payload.action === 'chapter') {
              output = `这是「${payload.outline[payload.chapterIndex].title}」的测试正文。`
              if (payload.chapterDraft) output += `\n\n根据要求「${payload.instruction}」修订：${payload.chapterDraft}`
            }
            res.write(`event: message\ndata: ${JSON.stringify({ delta: output })}\n\n`)
            if (payload.instruction === '【测试失败】') return res.end()
            const done = () => res.end('event: done\ndata: {"finishReason":"stop"}\n\n')
            if (payload.instruction === '【测试慢速】') {
              const timer = setTimeout(done, 15000)
              res.on('close', () => clearTimeout(timer))
            } else done()
          } catch {
            res.statusCode = 400
            res.end('Invalid test request')
          }
        } else next()
      })
    }
  }],
  optimizeDeps: { entries: ['tests/ai-editor/index.html'] },
  resolve: { alias: { '@': fileURLToPath(new URL('../../src', import.meta.url)) } },
  define: { 'import.meta.env.VITE_BASE_API': JSON.stringify('/test-api') },
  server: { host: '127.0.0.1', port: Number(process.env.AI_TEST_PORT || 5191), strictPort: true },
  css: { preprocessorOptions: { scss: { api: 'modern-compiler' } } }
})
await server.listen()
server.printUrls()
