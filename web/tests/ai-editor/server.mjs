import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
import vue from '@vitejs/plugin-vue'

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
      server.middlewares.use(async (req, res, next) => {
        if (req.url === '/test-api/blog/ai/status') {
          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify({ code: 0, data: { enabled: true } }))
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
