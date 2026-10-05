// 真实抽屉的长内容滚动回归；全部模型及上传请求使用模拟响应。
const assert = require('node:assert/strict')
const { chromium } = require('playwright')
const args = process.argv.slice(2)
const option = (name, fallback) => args.includes(name) ? args[args.indexOf(name) + 1] : fallback
const baseURL = option('--base-url', 'http://127.0.0.1:5191')
const channel = option('--channel', 'chrome')

async function wheel(page, selector, delta) {
  const area = page.locator(selector)
  await area.scrollIntoViewIfNeeded()
  const box = await area.boundingBox()
  // 在面板滚动条一侧滚动，避免 textarea、源码或图示预览的独立滚动区截获事件。
  await page.mouse.move(box.x + box.width - 3, box.y + Math.min(box.height / 2, 70))
  await page.mouse.wheel(0, delta)
}

async function wheelToBottom(page, selector, target) {
  const area = page.locator(selector)
  await area.evaluate(el => { el.scrollTop = 0 })
  assert(await area.evaluate(el => el.scrollHeight > el.clientHeight + 10), `${selector} 应产生长内容`)
  // 图示完成渲染后还可能重新布局，继续用滚轮到达最终底端。
  for (let attempt = 0; attempt < 4; attempt++) {
    await wheel(page, selector, 100000)
    try {
      await page.waitForFunction(selector => {
        const el = document.querySelector(selector)
        return el.scrollTop > 0 && el.scrollHeight - el.clientHeight - el.scrollTop < 3
      }, selector, {timeout:2000})
      break
    } catch (error) { if (attempt === 3) throw error }
  }
  const container = await area.boundingBox(), button = await target.boundingBox()
  assert(button && button.y >= container.y - 1 && button.y + button.height <= container.y + container.height + 1, '底部操作应可见')
  assert(button.y + button.height <= page.viewportSize().height, '底部操作应位于窗口内')
}

async function run() {
  const browser = await chromium.launch({ channel, headless: true })
  try {
    for (const viewport of [
      {width:1280,height:900}, {width:1280,height:600},
      {width:1280,height:480}, {width:390,height:700}, {width:844,height:390,mobile:true}
    ]) {
      const context = await browser.newContext({viewport:{width:viewport.width,height:viewport.height}})
      const page = await context.newPage(), errors = []
      page.on('pageerror', error => errors.push(String(error)))
      const longText = Array.from({length:100}, (_, i) => `第 ${i + 1} 段：这是长篇写作回归内容，需要完整阅读并能滚动到操作区域。`).join('\n\n')
      let imagePreview, visualResult
      await page.route('**/test-api/blog/ai/status*', route => route.fulfill({json:{code:0,data:{enabled:true,model:'fixture'}}}))
      await page.route('**/test-api/blog/ai/chat', route => {
        const payload = route.request().postDataJSON()
        const text = payload.action === 'chapter' ? longText : JSON.stringify({kind:'advice',content:longText})
        return route.fulfill({contentType:'text/event-stream',body:`event: message\ndata: ${JSON.stringify({delta:text})}\n\nevent: done\ndata: {"finishReason":"stop"}\n\n`})
      })
      await page.route('**/test-api/blog/ai/visual/*', route => {
        const action = route.request().url().split('/').pop()
        const payload = route.request().method() === 'POST' ? route.request().postDataJSON() : {}
        let data = {}
        if (action === 'status') data = {imageEnabled:true,storage:'local'}
        if (action === 'generate') {
          visualResult = {id:1,kind:payload.kind}
          if (payload.kind === 'flowchart') visualResult.mermaid = 'flowchart TD\n' + Array.from({length:16}, (_, i) => `N${i}[步骤 ${i + 1}] --> N${i + 1}[步骤 ${i + 2}]`).join('\n')
          else visualResult.preview = imagePreview
          data = visualResult
        }
        if (action === 'adopt') data = {...visualResult,url:'https://images.example.test/scroll.png',fileId:1}
        return route.fulfill({json:{code:0,data}})
      })
      await page.route('https://images.example.test/**', route => route.fulfill({contentType:'image/png',body:Buffer.from(imagePreview.split(',')[1], 'base64')}))
      await page.goto(`${baseURL}/tests/ai-editor/index.html?selection`)
      await page.waitForFunction(() => !!window.aiEditorTest)
      await page.evaluate(mobile => window.aiEditorTest.setDevice(mobile ? 'mobile' : 'desktop'), viewport.width < 600 || !!viewport.mobile)
      imagePreview = await page.evaluate(() => {
        const canvas = document.createElement('canvas'); canvas.width = 400; canvas.height = 900
        const ctx = canvas.getContext('2d'); ctx.fillStyle = '#dce9f3'; ctx.fillRect(0,0,400,900)
        return canvas.toDataURL('image/png')
      })
      await page.evaluate(text => window.aiEditorTest.setContent(text), longText)
      await page.getByRole('button',{name:'AI 助手',exact:true}).click()
      await page.locator('.writing-chat').waitFor({state:'visible'})

      // 聊天记录独立滚动；矮窗口仍能滚动到输入框和发送按钮。
      const chat = page.locator('.writing-chat'), send = page.getByRole('button',{name:'发送',exact:true})
      await page.getByRole('textbox',{name:'写作要求',exact:true}).fill('给出详细建议')
      if (await chat.evaluate(el => el.scrollHeight > el.clientHeight + 10)) {
        await chat.evaluate(el => { el.scrollTop = 0 })
        await wheel(page, '.chat-transcript', 100000)
        await page.waitForFunction(() => { const el = document.querySelector('.writing-chat'); return el.scrollTop > 0 && el.scrollHeight - el.clientHeight - el.scrollTop < 3 })
        await wheelToBottom(page, '.writing-chat', send)
      }
      await send.click()
      await page.locator('.chat-message.assistant .message-actions').waitFor()
      await page.waitForFunction(() => !document.querySelector('.chat-thinking'))
      assert(await page.locator('.chat-transcript').evaluate(el => el.scrollHeight - el.clientHeight - el.scrollTop < 3), '长回复应跟随到底部')
      await page.locator('.chat-transcript').scrollIntoViewIfNeeded()
      await wheel(page, '.chat-transcript', -100000)
      await page.waitForFunction(() => document.querySelector('.chat-transcript').scrollTop < 3)
      await wheel(page, '.chat-transcript', 100000)
      await page.waitForFunction(() => { const el = document.querySelector('.chat-transcript'); return el.scrollHeight - el.clientHeight - el.scrollTop < 3 })
      if (await chat.evaluate(el => el.scrollHeight > el.clientHeight + 10)) {
        await chat.evaluate(el => { el.scrollTop = 0 })
        await wheel(page, '.chat-transcript', 100000)
        await page.waitForFunction(() => { const el = document.querySelector('.writing-chat'); return el.scrollTop > 0 && el.scrollHeight - el.clientHeight - el.scrollTop < 3 })
        await wheelToBottom(page, '.writing-chat', send)
      }

      // 长来源、长图示、竖向配图及上传后的底部操作。
      await page.getByRole('button',{name:'图示配图',exact:true}).click()
      await page.locator('.visual-panel > .source-preview summary').click()
      const kind = page.locator('.visual-fields .el-select').nth(1)
      await kind.click()
      await page.getByRole('option',{name:'流程图',exact:true}).click()
      await page.getByRole('button',{name:'生成预览',exact:true}).click()
      await page.locator('.diagram-preview .markdown-diagram > img').waitFor({state:'attached'})
      await page.waitForFunction(() => { const img = document.querySelector('.diagram-preview .markdown-diagram > img'); return img?.complete && img.naturalWidth > 0 })
      await wheelToBottom(page, '.visual-panel', page.getByRole('button',{name:'丢弃预览',exact:true}))
      await page.getByRole('button',{name:'丢弃预览',exact:true}).click()
      await kind.click()
      await page.getByRole('option',{name:'正文插图',exact:true}).click()
      await page.getByRole('button',{name:'生成预览',exact:true}).click()
      await page.waitForFunction(() => document.querySelector('.generated-image')?.naturalHeight === 900)
      const adopt = page.getByRole('button',{name:'采用并上传',exact:true})
      await wheelToBottom(page, '.visual-panel', page.getByRole('button',{name:'丢弃预览',exact:true}))
      await adopt.click()
      await page.getByRole('button',{name:'设为封面',exact:true}).waitFor()
      await wheelToBottom(page, '.visual-panel', page.getByRole('button',{name:'关闭结果（保留附件）',exact:true}))
      await page.getByRole('button',{name:'设为封面',exact:true}).click()
      assert.equal(await page.locator('.visual-panel').getByRole('button',{name:'已设为封面',exact:true}).count(), 1)

      // 二十章大纲和长篇章预览，底部采纳按钮仍可访问。
      await page.getByRole('button',{name:'文字写作',exact:true}).click()
      await page.getByRole('button',{name:'大纲到章节',exact:true}).click()
      await page.locator('#chapter-outline').fill(Array.from({length:20}, (_, i) => `## 第 ${i + 1} 章\n- 说明本章的内容与例子`).join('\n\n'))
      const confirm = page.getByRole('button',{name:'确认大纲，开始逐节写作',exact:true})
      await wheelToBottom(page, '.chapter-workflow', confirm)
      await confirm.click()
      await page.getByRole('button',{name:'生成本章',exact:true}).click()
      await page.locator('#chapter-draft').waitFor({state:'attached'})
      await page.locator('.draft-preview summary').click()
      const adoptChapter = page.getByRole('button',{name:'采纳本章到正文末尾',exact:true})
      await wheelToBottom(page, '.chapter-workflow', adoptChapter)
      await adoptChapter.click()
      assert((await page.evaluate(() => window.aiEditorTest.getContent())).includes('## 第 1 章'))
      assert.equal(await page.locator('.assistant-sections').evaluate(el => { const r = el.getBoundingClientRect(); return r.top >= 0 && r.bottom <= innerHeight }), true, '切换按钮始终可见')
      assert.deepEqual(errors, [])
      console.log(`PASS ${viewport.width}×${viewport.height}: 图示、竖图、上传操作、长回复、长大纲及章节滚动`)
      if (option('--screenshot', '') && viewport.height === 480) await page.screenshot({path:option('--screenshot','')})
      await context.close()
    }
  } finally { await browser.close() }
}
run().catch(error => { console.error(error); process.exitCode = 1 })
