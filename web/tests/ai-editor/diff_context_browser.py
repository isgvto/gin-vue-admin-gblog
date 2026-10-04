"""Keep the full document visible around a fixed selection during AI comparison."""
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
    errors = []
    page.on('pageerror', lambda error: errors.append(str(error)))
    result = [{'kind': 'edit', 'content': '新第一段，增加一些内容\r\n\r\n范围内保持原样\r\n\r\n新第二段'}]
    def chat(route):
        text = json.dumps(result[0], ensure_ascii=False)
        route.fulfill(content_type='text/event-stream', body='event: message\ndata: ' + json.dumps({'delta': text}, ensure_ascii=False) + '\n\nevent: done\ndata: {"finishReason":"stop"}\n\n')
    page.route('**/test-api/blog/ai/chat', chat)
    page.route('**/test-api/blog/ai/status*', lambda route: route.fulfill(json={'code': 0, 'data': {'enabled': True, 'model': 'fixture'}}))
    url = args.base_url + '/tests/ai-editor/index.html'

    def reset(text):
        page.goto(url)
        page.wait_for_load_state('networkidle')
        page.evaluate('localStorage.clear()')
        page.reload()
        page.wait_for_load_state('networkidle')
        page.evaluate('(text)=>window.aiEditorTest.setContent(text)', text)

    def compare(selection):
        page.locator('.markdown-textarea').evaluate('(el, text)=>{const start=el.value.indexOf(text);el.focus();el.setSelectionRange(start,start+text.length);document.dispatchEvent(new Event("selectionchange"))}', selection.replace('\r\n', '\n'))
        page.get_by_role('textbox', name='写作要求').fill('润色选中的内容')
        page.get_by_role('button', name='发送', exact=True).click()
        expect(page.get_by_role('button', name='查看修改对比', exact=True)).to_be_enabled()
        page.get_by_role('button', name='查看修改对比', exact=True).click()
        expect(page.locator('.paragraph-diff')).to_be_visible()

    before = '# 开头\r\n\r\n选区之前的原文😀\r\n\r\n'
    selected = '原第一段\r\n\r\n范围内保持原样\r\n\r\n原第二段'
    after = '\r\n\r\n选区之后的原文\r\n\r\n结尾'
    original = before + selected + after
    reset(original)
    compare(selected)
    assert page.locator('.before-context').text_content() == before
    assert page.locator('.after-context').text_content() == after
    expect(page.locator('.diff-block.is-equal')).to_contain_text('范围内保持原样')
    expect(page.get_by_role('button', name='采用 AI 版', exact=True)).to_have_count(2)
    assert page.evaluate('window.aiEditorTest.getContent()') == original
    page.locator('.diff-toolbar').get_by_text('分栏', exact=True).click()
    expect(page.locator('.split-pane')).to_have_count(2)
    assert page.locator('.before-context').text_content() == before
    assert page.locator('.after-context').text_content() == after
    if args.screenshot:
        page.screenshot(path=args.screenshot, full_page=True)
    page.get_by_role('button', name='采用 AI 版', exact=True).first.click()
    expect(page.locator('.paragraph-diff')).to_be_visible()
    assert page.locator('.before-context').text_content() == before
    assert page.locator('.after-context').text_content() == after
    assert page.evaluate('window.aiEditorTest.getContent()') == original.replace('原第一段', '新第一段，增加一些内容')
    page.get_by_role('button', name='保留原文', exact=True).click()
    expect(page.locator('.paragraph-diff')).to_have_count(0)
    assert page.evaluate('window.aiEditorTest.getContent()') == original.replace('原第一段', '新第一段，增加一些内容')
    page.get_by_role('button', name='撤销 AI 修改', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getContent()') == original
    print('PASS full context in inline/split comparison, correct hunk indexes, adopt/keep and undo')

    for text, selection, has_before, has_after in [
        ('第一\n\n后文', '第一', False, True),
        ('前文\n\n最后', '最后', True, False),
        ('全文', '全文', False, False),
        ('开头ABC结尾', 'ABC', True, True),
    ]:
        result[0] = {'kind': 'edit', 'content': '新内容'}
        reset(text)
        compare(selection)
        expect(page.locator('.before-context')).to_have_count(int(has_before))
        expect(page.locator('.after-context')).to_have_count(int(has_after))
        snapshot = page.evaluate('window.aiEditorTest.getContent()')
        page.get_by_role('button', name='结束对比', exact=True).click()
        assert page.evaluate('window.aiEditorTest.getContent()') == snapshot
    print('PASS document start/end, whole-document and mid-sentence selections; close preserves text')

    text = ('前文很长\n\n' * 100) + '要修改的段落\n\n尾文'
    result[0] = {'kind': 'edit', 'content': '修改后的段落'}
    reset(text)
    compare('要修改的段落')
    page.wait_for_function('()=>{const box=document.querySelector(".diff-block.is-modified").getBoundingClientRect();return box.top<innerHeight&&box.bottom>0}')
    assert page.locator('.before-context').text_content() == '前文很长\n\n' * 100
    page.evaluate('(text)=>window.aiEditorTest.setContent(text)', text + '\n\n人工新增')
    expect(page.get_by_role('button', name='采用 AI 版', exact=True)).to_be_disabled()
    page.get_by_role('button', name='结束对比', exact=True).click()
    assert page.evaluate('window.aiEditorTest.getContent()').endswith('人工新增')
    print('PASS long-document focus, full uncollapsed context and conflict protection')
    assert not errors, errors
    browser.close()
