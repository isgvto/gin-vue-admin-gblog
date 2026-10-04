"""Paragraph numbering, fixed-range chat and gutter layout; mocked model only."""
import argparse
import json
from playwright.sync_api import sync_playwright, expect

parser = argparse.ArgumentParser()
parser.add_argument('--channel', default='chrome')
parser.add_argument('--base-url', default='http://127.0.0.1:5191')
parser.add_argument('--screenshot')
args = parser.parse_args()

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True, channel=args.channel)
    page = browser.new_page(viewport={'width': 1600, 'height': 1050})
    errors, requests = [], []
    page.on('pageerror', lambda error: errors.append(str(error)))
    output = [{'kind': 'edit', 'content': '新的第三段'}]

    def chat(route):
        requests.append(route.request.post_data_json)
        body = 'event: message\ndata: ' + json.dumps({'delta': json.dumps(output[0], ensure_ascii=False)}, ensure_ascii=False)
        route.fulfill(content_type='text/event-stream', body=body + '\n\nevent: done\ndata: {"finishReason":"stop"}\n\n')

    page.route('**/test-api/blog/ai/chat', chat)
    page.route('**/test-api/blog/ai/status*', lambda route: route.fulfill(json={'code': 0, 'data': {'enabled': True, 'model': 'fixture'}}))
    page.goto(args.base_url + '/tests/ai-editor/index.html')
    page.wait_for_load_state('networkidle')
    textarea = page.locator('.markdown-textarea')
    original = '# 标题\r\n\r\n重复😀\r\n\r\n重复😀\r\n\r\n```js\r\nconst a=1\r\n\r\nconsole.log(a)\r\n```\r\n\r\n- 一\r\n- 二'
    page.evaluate('(text)=>window.aiEditorTest.setContent(text)', original)
    expect(page.locator('.paragraph-number')).to_have_count(5)
    assert page.locator('.paragraph-number').all_text_contents() == ['1', '2', '3', '4', '5']
    page.get_by_role('button', name='选中第3段', exact=True).click()
    assert textarea.evaluate('(el)=>el.value.slice(el.selectionStart,el.selectionEnd)') == '重复😀'
    # Deliberately move the caret elsewhere: the command must resolve by number.
    textarea.evaluate('(el)=>el.setSelectionRange(0,0)')

    def send(command):
        count = page.locator('.chat-message.assistant').count()
        page.get_by_role('textbox', name='写作要求').fill(command)
        page.get_by_role('button', name='发送', exact=True).click()
        expect(page.locator('.chat-message.assistant')).to_have_count(count + 1)
        expect(page.get_by_role('button', name='停止生成', exact=True)).to_have_count(0)

    send('润色第三段')
    assert requests[-1]['selection'] == '重复😀'
    start = original.index('重复😀', original.index('重复😀') + 1)
    assert requests[-1]['cursorOffset'] == len(original[:start].encode('utf-16-le')) // 2
    assert '正文的第3段' in requests[-1]['instruction']
    expect(page.locator('.chat-message').last).to_contain_text('第3段')
    page.get_by_role('button', name='查看修改对比', exact=True).click()
    page.get_by_role('button', name='采用 AI 版', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getContent()') == original.replace('重复😀\r\n\r\n```', '新的第三段\r\n\r\n```')
    page.get_by_role('button', name='撤销 AI 修改', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getContent()') == original
    print('PASS paragraph numbering, duplicate/CRLF-safe selection, adopt and undo')

    output[0] = {'kind': 'advice', 'content': '这两段可以这样衔接。'}
    send('解释第2段到第3段的关系')
    assert requests[-1]['selection'] == '重复😀\r\n\r\n重复😀'
    assert page.evaluate('window.aiEditorTest.getContent()') == original
    send('说明“第8段”这个词')
    assert '编辑器定位说明' not in requests[-1]['instruction']
    count = len(requests)
    page.get_by_role('textbox', name='写作要求').fill('润色第99段')
    page.get_by_role('button', name='发送', exact=True).click()
    expect(page.get_by_text('段落编号无效，当前正文共5段；请查看左侧编号后重试', exact=True)).to_be_visible()
    assert len(requests) == count
    assert page.get_by_role('textbox', name='写作要求').input_value() == '润色第99段'
    page.get_by_role('textbox', name='写作要求').fill('修改第1和第3段')
    page.get_by_role('button', name='发送', exact=True).click()
    expect(page.get_by_text('本次请指定一个段落或连续范围，例如“修改第2至4段”；多个不连续段落请分别处理', exact=True)).to_be_visible()
    assert len(requests) == count
    print('PASS contiguous ranges, discussion, quoted examples and invalid-number protection')

    long = ('这是一段需要在编辑器中自动换行的文字。' * 45) + '\n\n第二段\n\n第三段'
    textarea.fill(long)
    expect(page.locator('.paragraph-number')).to_have_count(3)
    page.wait_for_function('parseFloat(document.querySelectorAll(".paragraph-number")[1].style.top)>100')
    wide = page.locator('.paragraph-number').nth(1).evaluate('(el)=>parseFloat(el.style.top)')
    page.set_viewport_size({'width': 1100, 'height': 1050})
    page.wait_for_function('(before)=>parseFloat(document.querySelectorAll(".paragraph-number")[1].style.top)>before', arg=wide)
    textarea.evaluate('(el)=>{el.scrollTop=300;el.dispatchEvent(new Event("scroll"))}')
    page.wait_for_function('()=>{const el=document.querySelector(".markdown-textarea");return Math.abs(parseFloat(document.querySelector(".paragraph-number").style.top)-(18-el.scrollTop))<2}')
    if args.screenshot:
        page.screenshot(path=args.screenshot, full_page=True)
    textarea.evaluate('(el)=>{el.style.maxHeight="520px"}')
    page.wait_for_function('document.querySelector(".paragraph-gutter").clientHeight===document.querySelector(".markdown-textarea").clientHeight')
    textarea.evaluate('(el)=>{el.scrollTop=300;el.dispatchEvent(new Event("scroll"))}')
    assert textarea.evaluate('(el)=>el.scrollTop') > 0
    page.wait_for_function('()=>{const el=document.querySelector(".markdown-textarea");return Math.abs(parseFloat(document.querySelector(".paragraph-number").style.top)-(18-el.scrollTop))<2}')
    page.locator('.markdown-toolbar .toolbar-group').last.locator('.tool-button').nth(1).click()
    expect(page.locator('.editor-pane.has-paragraph-numbers')).to_have_class('editor-pane is-alone has-paragraph-numbers')
    print('PASS gutter alignment across wrapping, resizing, scrolling and preview toggle')
    output[0] = {'kind': 'edit', 'content': '第二段的新稿'}
    send('改写第2段')
    textarea.fill(long + '\n\n新增段落')
    expect(page.get_by_role('button', name='查看修改对比', exact=True).last).to_be_disabled()
    page.evaluate('window.aiEditorTest.leaveEditor()')
    count = len(requests)
    page.get_by_role('textbox', name='写作要求').fill('修改第1段')
    page.get_by_role('button', name='发送', exact=True).click()
    expect(page.get_by_text('请在文章编辑页使用段落编号，或直接粘贴要处理的文字', exact=True)).to_be_visible()
    assert len(requests) == count
    print('PASS stale-result and absent-editor protection')
    assert not errors, errors
    browser.close()
