"""真实组件浏览器回归。先运行 npm run test:ai:serve；模型接口全部使用本地响应。"""
import argparse
import json
import sys
from playwright.sync_api import sync_playwright, expect

sys.stdout.reconfigure(encoding='utf-8')

parser = argparse.ArgumentParser()
parser.add_argument('--base-url', default='http://127.0.0.1:5191')
parser.add_argument('--channel', default='chrome')
parser.add_argument('--screenshot', help='可选：保存逐段对比的界面截图')
args = parser.parse_args()

with sync_playwright() as playwright:
    browser = playwright.chromium.launch(headless=True, channel=args.channel)
    page = browser.new_page(viewport={'width': 1600, 'height': 1050})
    errors = []
    page.on('pageerror', lambda error: errors.append(str(error)))
    response_text = ['AI 修改']
    defer_response = [False]
    pending = []

    def fulfill_chat(route, done=True):
        body = 'event: message\ndata: ' + json.dumps({'delta': response_text[0]}, ensure_ascii=False) + '\n\n'
        if done:
            body += 'event: done\ndata: {"finishReason":"stop"}\n\n'
        route.fulfill(status=200, content_type='text/event-stream', body=body)

    def chat(route):
        if defer_response[0]:
            pending.append(route)
        else:
            fulfill_chat(route)

    page.route('**/test-api/blog/ai/status', lambda route: route.fulfill(json={'code': 0, 'data': {'enabled': True}}))
    page.route('**/test-api/blog/ai/chat', chat)
    # 无效图片返回 404，避免 Vite 将图片路径回退到正式应用的 index.html。
    page.route('**/missing', lambda route: route.fulfill(status=404, body=''))

    def reset(content):
        page.goto(args.base_url + '/tests/ai-editor/index.html')
        page.wait_for_load_state('networkidle')
        page.wait_for_function('Boolean(window.aiEditorTest)')
        page.evaluate('text => window.aiEditorTest.setContent(text)', content)

    def select(text):
        page.locator('.markdown-textarea').evaluate('''(el, text) => {
          const start = el.value.indexOf(text);
          if (start < 0) throw new Error('selection not found');
          el.focus(); el.setSelectionRange(start, start + text.length);
          document.dispatchEvent(new Event('selectionchange'));
        }''', text.replace('\r\n', '\n').replace('\r', '\n'))

    def generate(text):
        select(text)
        page.get_by_role('button', name='润色', exact=True).click()
        expect(page.locator('.ai-diff-banner')).to_be_visible()

    def content():
        return page.evaluate('window.aiEditorTest.getContent()')

    original = '前文\r\n\r\n待修改😀\r\n\r\n后文\r\n'
    for selected in ['前文', '待修改😀', '后文', original]:
        reset(original)
        response_text[0] = 'AI 修改'
        generate(selected)
        assert content() == original, '生成不能直接改正文'
        if selected != original:
            expect(page.locator('.markdown-preview')).to_contain_text('AI 修改')
            outside = '后文' if selected != '后文' else '前文'
            expect(page.locator('.markdown-preview')).to_contain_text(outside)
        page.get_by_role('button', name='应用修改', exact=True).click()
        assert content() == original.replace(selected, 'AI 修改', 1)
        expect(page.get_by_text('已在正文编辑器中打开 diff 对比', exact=False)).to_have_count(0)
        page.get_by_role('button', name='撤销 AI 修改', exact=True).click()
        assert content() == original, '撤销必须精确恢复 CRLF 和选区外正文'
    print('PASS: 首部/中部/尾部/全文生成、全文预览、应用和精确撤销')

    reset('前文\n\n旧 A\n\n旧 B\n\n后文')
    response_text[0] = '新 A\n\n新 B'
    generate('旧 A\n\n旧 B')
    blocks = page.locator('.diff-block.is-modified')
    expect(blocks).to_have_count(2)
    blocks.nth(0).get_by_role('button', name='保留原文', exact=True).click()
    if args.screenshot:
        page.screenshot(path=args.screenshot, full_page=True)
    page.get_by_role('button', name='应用修改', exact=True).click()
    assert content() == '前文\n\n旧 A\n\n新 B\n\n后文'
    page.locator('.markdown-textarea').fill(content() + '人工新增')
    expect(page.get_by_role('button', name='撤销 AI 修改', exact=True)).to_be_disabled()
    assert content().endswith('人工新增')
    print('PASS: 连续多段部分采纳；后续人工编辑受到撤销冲突保护')

    reset(original)
    generate('待修改😀')
    page.get_by_role('button', name='全部保留原文', exact=True).click()
    page.get_by_role('button', name='应用修改', exact=True).click()
    assert content() == original
    expect(page.get_by_role('button', name='撤销 AI 修改', exact=True)).to_have_count(0)
    generate('待修改😀')
    page.get_by_role('button', name='取消', exact=True).click()
    assert content() == original
    print('PASS: 全部保留、取消对比均不改变原文')

    reset(original)
    generate('待修改😀')
    page.evaluate('text => window.aiEditorTest.setContent(text)', original + '更新')
    expect(page.get_by_role('button', name='应用修改', exact=True)).to_be_disabled()
    assert content() == original + '更新'
    print('PASS: 对比期间正文变化后禁止过期替换')

    for switch_document in [False, True]:
        reset(original)
        defer_response[0] = True
        select('待修改😀')
        page.get_by_role('button', name='润色', exact=True).click()
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
        if switch_document:
            page.evaluate('window.aiEditorTest.changeDocument()')
        else:
            page.evaluate('text => window.aiEditorTest.setContent(text)', original + '更新')
        assert pending
        fulfill_chat(pending.pop())
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_have_count(0)
        expect(page.locator('.ai-diff-banner')).to_have_count(0)
        assert content() == (original if switch_document else original + '更新')
        defer_response[0] = False
    print('PASS: 生成期间改文或切换文章，迟到结果无法进入替换流程')

    reset(original)
    defer_response[0] = True
    select('待修改😀')
    page.get_by_role('button', name='润色', exact=True).click()
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
    fulfill_chat(pending.pop(), done=False)
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_have_count(0)
    expect(page.locator('.ai-diff-banner')).to_have_count(0)
    assert content() == original
    defer_response[0] = False
    print('PASS: 不完整输出不允许进入对比替换流程')

    # 第二批：JSON 请求取消与重试，以及旧请求 finally 不得清除新请求状态。
    pending_single = []
    page.route('**/test-api/blog/ai/summary', lambda route: pending_single.append(route))
    page.route('**/test-api/blog/ai/suggest-tags', lambda route: pending_single.append(route))
    for button, payload, expected, apply_button in [
        ('生成摘要', {'summary': '新的摘要'}, '新的摘要', '回填摘要'),
        ('推荐标签', {'category': '技术', 'tags': ['Vue'], 'newTags': []}, 'Vue', '回填标签')
    ]:
        reset('测试正文')
        page.get_by_role('button', name=button, exact=True).click()
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
        page.wait_for_function('true')
        assert pending_single
        old = pending_single.pop()
        page.get_by_role('button', name='停止生成', exact=True).click()
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_have_count(0)
        page.get_by_role('button', name=button, exact=True).click()
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
        old.fulfill(json={'code': 0, 'data': payload})
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
        assert pending_single
        pending_single.pop().fulfill(json={'code': 0, 'data': payload})
        expect(page.locator('.result-body')).to_contain_text(expected)
        page.get_by_role('button', name=apply_button, exact=True).click()
        if button == '生成摘要':
            assert page.evaluate('window.aiEditorTest.getDescription()') == '新的摘要'
        else:
            assert page.evaluate('window.aiEditorTest.getSuggestion()').get('tags') == ['Vue']
        expect(page.locator('.el-message--error')).to_have_count(0)

        reset('文章 A')
        page.get_by_role('button', name=button, exact=True).click()
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
        assert pending_single
        old = pending_single.pop()
        page.evaluate('window.aiEditorTest.changeDocument()')
        old.fulfill(json={'code': 0, 'data': payload})
        expect(page.locator('.result-body')).to_have_count(0)
        expect(page.get_by_role('button', name=apply_button, exact=True)).to_have_count(0)
        assert page.evaluate('window.aiEditorTest.getDescription()') == '原摘要'
        assert page.evaluate('window.aiEditorTest.getSuggestion()') is None
    print('PASS: 摘要/标签取消、立即重试、旧响应隔离、成功回填、跨文章清理')

    reset('测试正文')
    page.get_by_role('button', name='生成摘要', exact=True).click()
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
    pending_single.pop().fulfill(json={'code': 7, 'msg': '今日 AI 调用次数已达上限'})
    expect(page.locator('.el-message--error')).to_have_count(1)
    expect(page.locator('.el-message--error')).to_contain_text('今日 AI 调用次数已达上限')
    expect(page.get_by_role('button', name='回填摘要', exact=True)).to_have_count(0)
    page.get_by_role('button', name='推荐标签', exact=True).click()
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
    pending_single.pop().fulfill(json={'code': 0, 'data': {'tags': 'wrong', 'newTags': []}})
    expect(page.get_by_role('button', name='回填标签', exact=True)).to_have_count(0)
    print('PASS: JSON 业务错误只提示一次；无效标签结果不可回填')

    reset('测试正文')
    response_text[0] = '文章 A 的结果'
    page.locator('.custom-input textarea').fill('文章 A 的指令')
    page.get_by_role('button', name='发送', exact=True).click()
    expect(page.locator('.history-bar')).to_contain_text('1 轮')
    page.evaluate('window.aiEditorTest.changeDocument()')
    expect(page.locator('.result-body')).to_have_count(0)
    expect(page.locator('.history-bar')).to_have_count(0)
    expect(page.locator('.custom-input textarea')).to_have_value('')
    with page.expect_request('**/test-api/blog/ai/chat') as request:
        page.locator('.custom-input textarea').fill('文章 B 的指令')
        page.get_by_role('button', name='发送', exact=True).click()
    assert request.value.post_data_json['history'] == []
    expect(page.locator('.history-bar')).to_contain_text('1 轮')
    page.evaluate('window.aiEditorTest.leaveEditor()')
    expect(page.locator('.history-bar')).to_have_count(0)
    expect(page.locator('.result-body')).to_have_count(0)
    print('PASS: 文章切换和离开页面清空结果、指令、历史；新请求不带旧历史')

    reset('原正文')
    response_text[0] = '生成的大纲'
    page.get_by_role('button', name='生成大纲', exact=True).click()
    expect(page.get_by_role('button', name='插入到光标处', exact=True)).to_be_enabled()
    page.get_by_role('button', name='插入到光标处', exact=True).click()
    assert '生成的大纲' in content()
    print('PASS: 完整结果仍可正常插入正文')

    # 第三批：结构化建议选择、取材范围提示、光标与选区传参。
    reset('长文章正文')
    page.get_by_role('button', name='推荐标签', exact=True).click()
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
    pending_single.pop().fulfill(json={'code': 0, 'data': {
        'category': '技术', 'categoryId': 1, 'tags': ['Vue'], 'tagIds': [2], 'newTags': ['新标签'],
        'context': {'notice': '本次选取分布于全文的片段'}, 'warnings': []}})
    expect(page.locator('.context-notice')).to_contain_text('分布于全文')
    expect(page.get_by_role('checkbox', name='新标签', exact=True)).not_to_be_checked()
    page.get_by_role('button', name='回填标签', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getMetadataForm()') == {'cate': 1, 'tagList': [3, 2]}
    assert page.evaluate('window.aiEditorTest.getSuggestion()')['newTags'] == []
    page.get_by_role('button', name='推荐标签', exact=True).click()
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_be_visible()
    pending_single.pop().fulfill(json={'code': 0, 'data': {
        'category': '技术', 'categoryId': 1, 'tags': ['Vue'], 'tagIds': [2], 'newTags': ['新标签']}})
    page.get_by_role('checkbox', name='新标签', exact=True).check()
    page.get_by_role('checkbox', name='分类：技术', exact=True).uncheck()
    page.get_by_role('button', name='回填标签', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getMetadataForm()') == {'cate': 1, 'tagList': [3, 2, '新标签']}
    print('PASS: 新标签默认不采纳，勾选后追加，已有分类和标签不被清空')

    reset('甲😀乙后文')
    page.locator('.markdown-textarea').evaluate("el => { el.focus(); el.setSelectionRange(0,0); }")
    with page.expect_request('**/test-api/blog/ai/chat') as request:
        page.get_by_role('button', name='续写', exact=True).click()
    assert request.value.post_data_json['cursorOffset'] == 0
    assert request.value.post_data_json['cursorContext'] == ''
    expect(page.get_by_role('button', name='停止生成', exact=True)).to_have_count(0)
    select('😀')
    with page.expect_request('**/test-api/blog/ai/chat') as request:
        page.locator('.custom-input textarea').fill('翻译选区')
        page.get_by_role('button', name='发送', exact=True).click()
    assert request.value.post_data_json['selection'] == '😀'
    print('PASS: 续写光标零位置传空前文，自定义指令保留选区')

    # 第四批：标题回填、重试、自定义对比。运行需要 Python Playwright。
    reset('用于起标题的正文')
    response_text[0] = '1. 标题甲\n2. 标题乙'
    page.get_by_role('button', name='起标题', exact=True).click()
    expect(page.locator('.title-candidate')).to_have_count(2)
    page.locator('.title-candidate').nth(1).get_by_role('button', name='采用标题', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getTitle()') == '标题乙'
    page.get_by_role('button', name='起标题', exact=True).click()
    expect(page.locator('.title-candidate')).to_have_count(2)
    page.evaluate("window.aiEditorTest.setTitle('人工编辑的标题')")
    page.locator('.title-candidate').first.get_by_role('button', name='采用标题', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getTitle()') == '人工编辑的标题'
    print('PASS: 标题候选回填和人工编辑冲突保护')

    reset('前文原选区后文')
    select('原选区')
    if page.locator('.writing-preferences').get_attribute('open') is None:
        page.locator('summary').filter(has_text='正文生成偏好').click()
    page.get_by_role('combobox', name='写作语气').click()
    page.get_by_role('option', name='正式严谨', exact=True).click()
    page.locator('.writing-preferences .el-select').nth(2).click()
    page.get_by_role('option', name='轻度修饰', exact=True).click()
    response_text[0] = '新选区'
    page.locator('.custom-input textarea').fill('重新表述选区')
    with page.expect_request('**/test-api/blog/ai/chat') as request:
        page.get_by_role('button', name='发送', exact=True).click()
    assert request.value.post_data_json['tone'] == 'formal'
    assert request.value.post_data_json['editStrength'] == 'light'
    expect(page.get_by_role('button', name='与原选区对比', exact=True)).to_be_enabled()
    with page.expect_request('**/test-api/blog/ai/chat') as retry:
        page.get_by_role('button', name='重新生成', exact=True).click()
    assert retry.value.post_data_json == request.value.post_data_json
    expect(page.locator('.history-bar')).to_contain_text('1 轮')
    page.get_by_role('button', name='与原选区对比', exact=True).click()
    expect(page.locator('.ai-diff-banner')).to_be_visible()
    page.get_by_role('radio', name='分栏', exact=True).check()
    expect(page.locator('.split-pane')).to_have_count(1)
    page.get_by_role('button', name='应用修改', exact=True).click()
    assert content() == '前文新选区后文'
    print('PASS: 偏好传参、重试固定条件且不叠加历史、自定义选区对比与分栏采纳')

    malicious = '''# 标题
<img src="/missing" onerror="window.__aiXss=1">
<a href="java&#x09;script:window.__aiXss=2">危险链接</a>
<iframe srcdoc="<script>parent.__aiXss=3</script>"></iframe>
<svg onload="window.__aiXss=4"></svg><math><mtext>x</mtext></math>
<style>body{display:none}</style><form id="app"><input name="x" autofocus></form>
<img src="data:image/svg+xml,test"><a href="data:text/html,test">data</a>

| A | B |
| --- | --- |
| 1 | 2 |

- [x] 完成

```js
const text = '<img onerror="bad">'
```
'''
    reset(malicious)

    def assert_safe(selector):
        result = page.locator(selector).evaluate('''el => ({
          forbidden: el.querySelectorAll('script,iframe,svg,math,style,form,object,embed').length,
          badAttributes: [...el.querySelectorAll('*')].flatMap(node => [...node.attributes])
            .filter(a => /^on/i.test(a.name) || ['style','id','name','srcdoc'].includes(a.name)).length,
          badUrls: [...el.querySelectorAll('[href],[src]')].filter(node =>
            /^(javascript|data|vbscript):/i.test(node.getAttribute('href') || node.getAttribute('src') || '')).length,
          table: !!el.querySelector('table'), code: !!el.querySelector('pre code.hljs'),
          checked: !!el.querySelector('input[type="checkbox"][disabled][checked]')
        })''')
        assert result == {'forbidden': 0, 'badAttributes': 0, 'badUrls': 0, 'table': True, 'code': True, 'checked': True}, result
        assert page.evaluate('window.__aiXss === undefined')

    assert_safe('.markdown-preview')
    response_text[0] = malicious
    page.locator('.custom-input textarea').fill('输出测试结果')
    page.get_by_role('button', name='发送', exact=True).click()
    expect(page.locator('.result-body')).to_contain_text('标题')
    assert_safe('.result-body')
    select('标题')
    page.get_by_role('button', name='润色', exact=True).click()
    expect(page.locator('.ai-diff-banner')).to_be_visible()
    assert_safe('.markdown-preview')
    print('PASS: 正文预览、助手结果、diff 预览过滤危险 HTML，保留表格、代码高亮与任务列表')
    # 模型管理使用真实组件，验证旧 ID 响应兼容以及各操作发送的模型 ID。
    model_requests = []
    def model_route(route):
        request = route.request
        if request.method == 'GET':
            if request.url.endswith('/providers'):
                route.fulfill(json={'code': 0, 'data': [{'value': 'openai', 'label': 'OpenAI', 'needBaseUrl': True}]})
            else:
                route.fulfill(json={'code': 0, 'data': {'list': [{'ID': 42, 'name': '测试模型', 'provider': 'openai', 'model': 'test', 'baseUrl': 'https://example.invalid/v1', 'temperature': 0.7, 'maxTokens': 4096, 'status': True, 'hasKey': True, 'keyTail': '1234'}], 'total': 1}})
        else:
            model_requests.append((request.method, request.url, request.post_data_json if request.post_data else None))
            route.fulfill(json={'code': 0, 'data': {}})
    page.route('**/test-api/ai/modelConfig**', model_route)
    page.goto(args.base_url + '/tests/ai-editor/index.html?models')
    page.wait_for_load_state('networkidle')
    page.get_by_role('button', name='编辑', exact=True).click()
    expect(page.get_by_role('dialog')).to_contain_text('编辑模型')
    page.get_by_role('button', name='保存', exact=True).click()
    expect(page.get_by_role('dialog')).not_to_be_visible()
    assert model_requests[-1][0] == 'PUT' and model_requests[-1][2]['id'] == 42
    assert model_requests[-1][2]['apiKey'] == ''
    page.get_by_role('button', name='设为默认', exact=True).click()
    expect(page.locator('.el-message').filter(has_text='已设为默认')).to_be_visible()
    assert model_requests[-1][1].endswith('/setDefault/42')
    page.get_by_role('button', name='删除', exact=True).click()
    page.locator('.el-popconfirm__action .el-button--primary').click()
    expect(page.locator('.el-message').filter(has_text='删除成功')).to_be_visible()
    assert model_requests[-1][0] == 'DELETE' and model_requests[-1][1].endswith('/42')
    print('PASS: 模型列表 ID 兼容，编辑、默认、删除均携带正确 ID')
    reset('```mermaid\nflowchart TD\n A[写作] --> B[保存]\n```\n\n```mermaid\nsequenceDiagram\n 用户->>服务端: 请求\n 服务端-->>用户: 返回\n```')
    expect(page.locator('.markdown-diagram')).to_have_count(2)
    page.wait_for_function('''() => [...document.querySelectorAll('.markdown-diagram > img')]
      .every(img => img.complete && img.naturalWidth > 0)''')
    expect(page.locator('.markdown-preview svg')).to_have_count(0)
    page.locator('.markdown-diagram summary').first.click()
    expect(page.locator('.markdown-diagram details pre').first).to_contain_text('A[写作]')
    page.locator('.markdown-textarea').fill('```mermaid\nflowchart TD\n A[未闭合\n```')
    expect(page.locator('.mermaid-error')).to_contain_text('源码已保留')
    expect(page.locator('.markdown-preview code')).to_contain_text('未闭合')
    page.locator('.markdown-textarea').fill('```mermaid\n%%{init: {"securityLevel": "loose"}}%%\nflowchart TD\n A --> B\n```')
    expect(page.locator('.mermaid-error')).to_contain_text('配置指令暂不支持')
    page.locator('.markdown-textarea').fill('```mermaid\nflowchart LR\n A[恢复] --> B[成功]\n```')
    expect(page.locator('.markdown-diagram')).to_have_count(1)
    expect(page.locator('.mermaid-error')).to_have_count(0)
    print('PASS: Mermaid 多图渲染、源码查看、语法错误回退、配置隔离与编辑恢复')
    assert not errors, errors
    browser.close()
    print('All browser regressions passed.')
