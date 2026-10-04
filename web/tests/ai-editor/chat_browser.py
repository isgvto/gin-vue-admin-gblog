"""写作 Chat 真实组件回归；模型请求全部模拟，不调用真实供应商。"""
import argparse
import json
import base64
import sys
from playwright.sync_api import sync_playwright, expect
sys.stdout.reconfigure(encoding='utf-8')
parser=argparse.ArgumentParser()
parser.add_argument('--base-url',default='http://127.0.0.1:5191')
parser.add_argument('--channel',default='chrome')
parser.add_argument('--screenshot')
args=parser.parse_args()
with sync_playwright() as playwright:
    browser=playwright.chromium.launch(headless=True,channel=args.channel)
    page=browser.new_page(viewport={'width':1600,'height':1050})
    errors=[]
    page.on('pageerror',lambda error:errors.append(str(error)))
    output=[{'kind':'advice','content':'先明确核心观点，再补充例子。'}]
    requests=[]
    deferred=[False]
    pending=[]
    incomplete=[False]
    def chat(route):
        payload=route.request.post_data_json
        requests.append(payload)
        if payload['action']=='chapter':
            text='这是章节正文。'
        else:
            assert payload['action']=='conversation'
            text=json.dumps(output[0],ensure_ascii=False) if isinstance(output[0],dict) else output[0]
        body='event: message\ndata: '+json.dumps({'delta':text},ensure_ascii=False)+'\n\n'
        if not incomplete[0]:body+='event: done\ndata: {"finishReason":"stop"}\n\n'
        if deferred[0]:pending.append((route,body))
        else:route.fulfill(status=200,content_type='text/event-stream',body=body)
    page.route('**/test-api/blog/ai/chat',chat)
    page.route('**/test-api/blog/ai/status*',lambda route:route.fulfill(json={'code':0,'data':{'enabled':True,'model':'fixture'}}))
    previews={}
    png='iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aAlsAAAAASUVORK5CYII='
    page.route('https://images.example.test/**',lambda route:route.fulfill(content_type='image/png',body=base64.b64decode(png)))
    def visual(route):
        action=route.request.url.rsplit('/',1)[-1]
        data=route.request.post_data_json if route.request.method=='POST' else {}
        if action=='status':result={'imageEnabled':True,'storage':'local'}
        elif action=='plan':result={'kind':'flowchart','prompt':'描述写作到保存的流程','reason':'流程图适合表达步骤'}
        elif action=='generate':
            result={'id':len(previews)+1,'kind':data['kind']}
            if data['kind'] in ['flowchart','structure']:result['mermaid']='flowchart LR\n A[写作] --> B[保存]'
            else:result['preview']='data:image/png;base64,'+png
            previews[result['id']]=result
        elif action=='adopt':
            if data.get('png'):
                image=base64.b64decode(data['png'].split(',')[1]);width=int.from_bytes(image[16:20],'big');height=int.from_bytes(image[20:24],'big')
                assert 3200<=max(width,height)<=4096
            result={**previews[data['id']],'url':f'https://images.example.test/{data["id"]}.png','fileId':1}
        else:result={}
        route.fulfill(json={'code':0,'data':result})
    page.route('**/test-api/blog/ai/visual/*',visual)
    url=args.base_url+'/tests/ai-editor/index.html'
    def reset(text='前文\n\n选中内容\n\n后文',selection=False):
        page.goto(url+('?selection' if selection else ''))
        page.wait_for_load_state('networkidle')
        page.evaluate('localStorage.clear()')
        page.reload()
        page.wait_for_load_state('networkidle')
        page.locator('.markdown-textarea').fill(text)
        if selection:page.get_by_role('button',name='AI 助手',exact=True).click()
        expect(page.locator('.writing-chat')).to_be_visible()
    def select(text):
        page.locator('.markdown-textarea').evaluate('(el,text)=>{const start=el.value.indexOf(text);el.focus();el.setSelectionRange(start,start+text.length);el.dispatchEvent(new Event("select"));document.dispatchEvent(new Event("selectionchange"));}',text)
    def send(text='帮我看看这篇文章'):
        previous=page.locator('.chat-message.assistant').count()
        page.get_by_role('textbox',name='写作要求').fill(text)
        page.get_by_role('button',name='发送',exact=True).click()
        expect(page.locator('.chat-message.assistant')).to_have_count(previous+1)
        expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
    def content():return page.locator('.markdown-textarea').input_value()

    reset()
    page.get_by_role('textbox',name='写作要求').fill('保留专业术语，面向初学者')
    page.get_by_role('button',name='润色',exact=True).click()
    expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
    assert '保留专业术语，面向初学者' in requests[-1]['instruction']
    assert '润色' in requests[-1]['instruction']
    print('PASS 快捷操作保留补充要求')

    for label in ['起标题','摘要','推荐标签','审阅文章','生成大纲']:
        reset();select('选中内容')
        page.get_by_role('button',name=label,exact=True).click()
        expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
        assert requests[-1]['selection']=='' and requests[-1]['content']==content()
    reset();select('选中内容')
    page.locator('.scope-row .el-select').click()
    page.get_by_role('option',name='当前选区',exact=True).click()
    page.get_by_role('button',name='审阅文章',exact=True).click()
    expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
    assert requests[-1]['selection']=='选中内容'
    reset();select('选中内容')
    page.get_by_role('textbox',name='写作要求').fill('只为这段起两个标题')
    page.get_by_role('button',name='起标题',exact=True).click()
    expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
    assert requests[-1]['selection']=='选中内容'
    assert '只为这段起两个标题' in requests[-1]['instruction']
    print('PASS 全文类快捷操作的默认范围、明确选区及自定义补充范围')

    reset('开头\n\n原文第一段\n\n原文第二段\n\n结尾')
    output[0]={'kind':'edit','content':'新第一段\n\n不想要的新第二段'}
    select('原文第一段\n\n原文第二段');send('改写两段')
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    page.locator('.paragraph-diff').get_by_role('button',name='采用 AI 版',exact=True).first.click()
    page.locator('.paragraph-diff').get_by_role('button',name='保留原文',exact=True).first.click()
    expect(page.locator('.paragraph-diff')).to_have_count(0)
    page.get_by_role('button',name='引用',exact=True).click()
    expect(page.locator('.followup-context')).to_contain_text('实际保留')
    output[0]={'kind':'edit','content':'新第一段更简洁\n\n原文第二段'}
    send('再简洁一点')
    assert requests[-1]['selection']=='新第一段\n\n原文第二段'
    history=json.loads(requests[-1]['history'][-1]['content'])
    assert history['content']=='新第一段\n\n原文第二段'
    assert '不想要的新第二段' not in requests[-1]['history'][-1]['content']
    page.get_by_role('button',name='引用',exact=True).first.click()
    expect(page.locator('.followup-context')).to_contain_text('第 1 条')
    page.get_by_role('button',name='取消引用',exact=True).click()
    print('PASS 部分采用后引用真实正文、旧版本追问范围明确')

    reset('重复\n重复\n')
    page.locator('.markdown-textarea').evaluate('(el)=>{el.setSelectionRange(0,0);document.dispatchEvent(new Event("selectionchange"))}')
    output[0]={'kind':'insert','content':'重复'}
    send('从文首先写一段')
    page.get_by_role('button',name='插入原光标位置',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value('重复\n重复\n重复\n')
    page.get_by_role('button',name='引用',exact=True).click()
    output[0]={'kind':'edit','content':'首段新内容\n'}
    send('只改刚插入的这一段')
    assert requests[-1]['selection']=='重复\n'
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    page.get_by_role('button',name='采用剩余 AI 修改',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value('首段新内容\n重复\n重复\n')
    print('PASS 插入后引用回复精确位置，重复段落不会错改')

    reset()
    output[0]={'kind':'review','content':'建议先明确这段的主体。','issues':[{'quote':'选中内容','reason':'读者不清楚主体是谁','suggestion':'补充具体主体'}]}
    send('审阅文章')
    expect(page.locator('.review-issue')).to_have_count(1)
    page.get_by_role('button',name='定位原文',exact=True).click()
    page.wait_for_function('()=>{const el=document.querySelector(".markdown-textarea");return el.value.slice(el.selectionStart,el.selectionEnd)==="选中内容"}')
    assert content()=='前文\n\n选中内容\n\n后文'
    page.get_by_role('button',name='讨论这一处',exact=True).click()
    expect(page.locator('.followup-context')).to_contain_text('审阅')
    assert '读者不清楚' in page.get_by_role('textbox',name='写作要求').input_value()
    page.get_by_role('button',name='按建议修改',exact=True).click()
    assert '请按照建议修改这段' in page.get_by_role('textbox',name='写作要求').input_value()
    output[0]={'kind':'edit','content':'作者选中的内容'}
    send('按照这个建议修改')
    assert requests[-1]['selection']=='选中内容'
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    page.get_by_role('button',name='采用剩余 AI 修改',exact=True).click()
    assert content()=='前文\n\n作者选中的内容\n\n后文'
    reset('重复内容\n\n重复内容')
    output[0]={'kind':'review','content':'需要说明主体。','issues':[{'quote':'重复内容','reason':'主体不清楚','suggestion':'补充主体'}]}
    send('审阅')
    expect(page.get_by_role('button',name='定位原文',exact=True)).to_be_disabled()
    expect(page.get_by_role('button',name='讨论这一处',exact=True)).to_be_disabled()
    print('PASS 审阅定位、段落讨论修改、重复原文不猜测定位')

    reset(selection=True)
    page.evaluate("""() => {
      const original = window.fetch;
      const encoder = new TextEncoder();
      window.fetch = (url, options) => {
        if (String(url).includes('/blog/ai/chat')) return Promise.resolve(new Response(new ReadableStream({start(controller) {
          window.testChatStream = {push(delta) {controller.enqueue(encoder.encode('event: message\\ndata: '+JSON.stringify({delta})+'\\n\\n'))}, end() {controller.enqueue(encoder.encode('event: done\\ndata: {"finishReason":"stop"}\\n\\n'));controller.close()}};
        }}), {headers: {'Content-Type':'text/event-stream'}}));
        return original(url, options);
      };
    }""")
    page.get_by_role('textbox',name='写作要求').fill('给我详细建议')
    page.get_by_role('button',name='发送',exact=True).click()
    page.wait_for_function('()=>!!window.testChatStream')
    long_text='写作建议，保留自己的表达。\n\n'*100
    page.evaluate("""(text)=>window.testChatStream.push('{"content":"'+JSON.stringify(text).slice(1,-1))""",long_text)
    expect(page.locator('.message-content')).to_contain_text('写作建议')
    page.wait_for_timeout(150)
    page.locator('.chat-transcript').evaluate('(el)=>{el.scrollTop=0;el.dispatchEvent(new Event("scroll"))}')
    page.evaluate('()=>window.testChatStream.push("补充建议")')
    expect(page.get_by_role('button',name='查看最新回复 ↓',exact=True)).to_be_visible()
    assert page.locator('.chat-transcript').evaluate('(el)=>el.scrollTop')==0
    page.get_by_role('button',name='查看最新回复 ↓',exact=True).click()
    page.wait_for_function('()=>{const el=document.querySelector(".chat-transcript");return el.scrollTop>0}')
    page.evaluate("""()=>{window.testChatStream.push('","kind":"advice"}');window.testChatStream.end()}""")
    expect(page.get_by_role('button',name='停止生成',exact=True)).to_have_count(0)
    print('PASS 流式回复不打断历史阅读、手动返回最新回复')

    output[0]={'kind':'advice','content':'先明确核心观点，再补充例子。'}
    reset()
    send('先分析文章，不修改')
    expect(page.locator('.message-content')).to_contain_text('核心观点')
    assert content()=='前文\n\n选中内容\n\n后文'
    send('再具体一点')
    assert requests[-1]['history'][0]['content']=='先分析文章，不修改'
    assert requests[-1]['history'][1]['role']=='assistant'
    page.get_by_role('textbox',name='写作要求').fill('还没发送的想法')
    page.wait_for_function("()=>Object.keys(localStorage).some(k=>k.startsWith('gblog:writing-chat:')&&localStorage[k].includes('还没发送的想法'))")
    page.reload();page.wait_for_load_state('networkidle')
    expect(page.locator('.chat-message.assistant')).to_have_count(2)
    expect(page.get_by_role('textbox',name='写作要求')).to_have_value('还没发送的想法')
    print('PASS 多轮历史、刷新恢复、未发送输入')

    reset()
    output[0]={'kind':'edit','content':'更清楚的内容'}
    select('选中内容');send('润色这段')
    assert requests[-1]['selection']=='选中内容'
    page.get_by_role('textbox',name='写作要求').fill('再简洁一点')
    page.get_by_role('textbox',name='写作要求').press('Enter')
    expect(page.locator('.chat-message.assistant')).to_have_count(2)
    assert requests[-1]['selection']=='选中内容'
    page.get_by_role('button',name='查看修改对比',exact=True).last.click()
    expect(page.locator('.ai-diff-banner')).to_be_visible()
    page.get_by_role('button',name='采用 AI 版',exact=True).click()
    assert content()=='前文\n\n更清楚的内容\n\n后文'
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    assert content()=='前文\n\n选中内容\n\n后文'
    print('PASS 选区追问、对比采纳与撤销')

    original='前文\n\n甲\n\n乙😀\n\n丙\n\n后文'
    reset(original)
    output[0]={'kind':'edit','content':'更长的新甲\n\n新版乙😀\n\n新版丙'}
    select('甲\n\n乙😀\n\n丙');send('修改这三段')
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    expect(page.get_by_role('button',name='应用修改',exact=True)).to_have_count(0)
    expect(page.get_by_role('button',name='采用 AI 版',exact=True)).to_have_count(3)
    page.get_by_role('button',name='采用 AI 版',exact=True).first.click()
    expect(page.locator('.markdown-textarea')).to_have_value(original.replace('甲','更长的新甲'))
    expect(page.get_by_role('button',name='采用 AI 版',exact=True)).to_have_count(2)
    page.get_by_role('button',name='保留原文',exact=True).first.click()
    expect(page.locator('.markdown-textarea')).to_have_value(original.replace('甲','更长的新甲'))
    expect(page.get_by_role('button',name='采用 AI 版',exact=True)).to_have_count(1)
    page.get_by_role('button',name='采用 AI 版',exact=True).click()
    expect(page.locator('.ai-diff-banner')).to_have_count(0)
    expect(page.locator('.markdown-textarea')).to_have_value(original.replace('甲','更长的新甲').replace('丙','新版丙'))
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value(original.replace('甲','更长的新甲'))
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value(original)
    print('PASS 多处即时采用/保留，无全局应用，自动退出且逐步撤销')

    reset(original)
    select('甲\n\n乙😀\n\n丙');send('修改这三段')
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    page.get_by_role('button',name='采用 AI 版',exact=True).first.click()
    page.get_by_role('button',name='结束对比',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value(original.replace('甲','更长的新甲'))
    expect(page.locator('.ai-diff-banner')).to_have_count(0)
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    expect(page.locator('.markdown-textarea')).to_have_value(original)
    print('PASS 提前结束保留已采用修改，剩余部分保留原文')

    reset('前文\n\n选中内容\n\n后文')
    select('选中内容');send('改写这段')
    page.reload();page.wait_for_load_state('networkidle')
    page.locator('.markdown-textarea').fill('前文\n\n选中内容\n\n后文')
    expect(page.get_by_role('button',name='查看修改对比',exact=True)).to_be_enabled()
    page.locator('.markdown-textarea').fill('人工编辑后的正文')
    expect(page.get_by_role('button',name='查看修改对比',exact=True)).to_be_disabled()
    assert content()=='人工编辑后的正文'
    print('PASS 刷新恢复选区，正文变化拒绝旧结果')

    reset('甲😀乙后文')
    output[0]={'kind':'insert','content':'新内容'}
    page.locator('.markdown-textarea').evaluate('el=>{el.focus();el.setSelectionRange(3,3)}')
    send('从这里继续写')
    page.locator('.markdown-textarea').evaluate('el=>el.setSelectionRange(0,0)')
    page.get_by_role('button',name='插入原光标位置',exact=True).click()
    assert content()=='甲😀新内容\n乙后文'
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    assert content()=='甲😀乙后文'
    print('PASS 固定原光标位置插入及撤销')

    reset('标题参考正文')
    output[0]={'kind':'title','content':'两个候选标题','titles':['标题甲','标题乙']}
    send('起几个标题')
    page.get_by_role('button',name='采用标题',exact=True).nth(1).click()
    assert page.evaluate('window.aiEditorTest.getTitle()')=='标题乙'
    send('再给几个')
    page.evaluate("window.aiEditorTest.setTitle('人工修改的标题')")
    page.get_by_role('button',name='采用标题',exact=True).last.click()
    assert page.evaluate('window.aiEditorTest.getTitle()')=='人工修改的标题'
    output[0]={'kind':'summary','content':'这是生成摘要。'}
    send('生成摘要');page.get_by_role('button',name='采用摘要',exact=True).click()
    assert page.evaluate('window.aiEditorTest.getDescription()')=='这是生成摘要。'
    output[0]={'kind':'tags','content':'推荐技术分类','category':'技术','tags':['Vue'],'newTags':['新标签']}
    send('推荐标签')
    expect(page.get_by_role('checkbox',name='新标签',exact=True)).not_to_be_checked()
    page.get_by_role('button',name='采用分类与标签',exact=True).click()
    assert page.evaluate('window.aiEditorTest.getMetadataForm()')=={'cate':1,'tagList':[3,2]}
    print('PASS 标题冲突保护、摘要与标签回填，新标签默认不采用')

    reset()
    output[0]={'kind':'advice','content':'完整回复'}
    incomplete[0]=True;send('失败请求')
    expect(page.locator('.message-error')).to_be_visible()
    expect(page.get_by_role('button',name='重新尝试',exact=True)).to_be_visible()
    incomplete[0]=False
    deferred[0]=True
    page.get_by_role('textbox',name='写作要求').fill('慢速请求')
    page.get_by_role('button',name='发送',exact=True).click()
    expect(page.get_by_role('button',name='停止生成',exact=True)).to_be_visible()
    page.get_by_role('button',name='停止生成',exact=True).click()
    expect(page.locator('.message-error').last).to_contain_text('已停止')
    deferred[0]=False
    for route,body in pending:
        try:route.fulfill(status=200,content_type='text/event-stream',body=body)
        except Exception:pass
    pending.clear()
    print('PASS 失败和停止不会生成可采纳结果')

    reset()
    send('文章A的对话')
    page.evaluate("window.aiEditorTest.setDocumentId('article-b')")
    expect(page.locator('.chat-message')).to_have_count(0)
    send('文章B的对话')
    page.evaluate("window.aiEditorTest.setDocumentId('article-a')")
    expect(page.locator('.chat-message.user')).to_contain_text('文章A的对话')
    page.get_by_role('button',name='新对话',exact=True).click()
    expect(page.locator('.chat-message')).to_have_count(0)
    page.locator('.session-controls .el-select').click()
    page.get_by_role('option',name='文章A的对话',exact=True).click()
    expect(page.locator('.chat-message.user')).to_contain_text('文章A的对话')
    page.get_by_role('button',name='清空',exact=True).click()
    page.get_by_role('button',name='保留',exact=True).click()
    expect(page.locator('.chat-message.user')).to_contain_text('文章A的对话')
    print('PASS 文章隔离、多会话切换、清空确认')

    def delete_entry(title):
        button=page.get_by_role('button',name='删除会话：'+title,exact=True)
        expect(page.get_by_role('dialog',name='删除对话',exact=True)).to_be_hidden()
        page.keyboard.press('Escape')
        expect(button).to_be_hidden()
        page.locator('.session-controls .el-select').click()
        expect(button).to_be_visible()
        button.click()
    def stored_chat():
        return page.evaluate("()=>JSON.parse(localStorage[Object.keys(localStorage).find(k=>k.startsWith('gblog:writing-chat:')&&k.endsWith(':article-a'))])")
    delete_entry('新对话')
    dialog=page.get_by_role('dialog',name='删除对话',exact=True)
    dialog.get_by_role('button',name='取消',exact=True).click()
    assert len(stored_chat()['sessions'])==2
    expect(page.locator('.chat-message.user')).to_contain_text('文章A的对话')
    delete_entry('新对话')
    dialog.get_by_role('button',name='删除',exact=True).click()
    page.wait_for_function("()=>JSON.parse(localStorage[Object.keys(localStorage).find(k=>k.startsWith('gblog:writing-chat:')&&k.endsWith(':article-a'))]).sessions.length===1")
    expect(page.locator('.chat-message.user')).to_contain_text('文章A的对话')
    send('当前会话仍可继续')
    page.get_by_role('button',name='新对话',exact=True).click()
    send('要删除的当前会话')
    delete_entry('要删除的当前会话')
    dialog.get_by_role('button',name='删除',exact=True).click()
    expect(page.locator('.chat-message.user')).to_contain_text(['文章A的对话','当前会话仍可继续'])
    assert len(stored_chat()['sessions'])==1
    delete_entry('文章A的对话')
    dialog.get_by_role('button',name='删除',exact=True).click()
    expect(page.locator('.chat-message')).to_have_count(0)
    data=stored_chat()
    assert len(data['sessions'])==1 and data['sessions'][0]['messages']==[]
    assert data['activeId']==data['sessions'][0]['id']
    assert content()=='前文\n\n选中内容\n\n后文'
    page.reload();page.wait_for_load_state('networkidle')
    expect(page.locator('.chat-message')).to_have_count(0)
    assert len(stored_chat()['sessions'])==1
    page.evaluate("window.aiEditorTest.setDocumentId('article-b')")
    expect(page.locator('.chat-message.user')).to_contain_text('文章B的对话')
    print('PASS 删除取消、非当前会话、当前会话、最后会话、刷新和文章隔离')

    reset(selection=True)
    page.locator('.el-drawer__close-btn').click()
    page.wait_for_timeout(500)  # 等待抽屉退出动画结束，再测试重新打开。
    select('选中内容');output[0]={'kind':'edit','content':'浮动入口修改'}
    page.locator('.selection-ai-toolbar').get_by_role('button',name='AI 润色',exact=True).click()
    expect(page.get_by_role('button',name='查看修改对比',exact=True)).to_be_visible()
    page.get_by_role('button',name='查看修改对比',exact=True).click()
    expect(page.locator('.ai-diff-banner')).to_be_visible()
    expect(page.get_by_role('dialog',name='AI 助手')).not_to_be_visible()
    page.get_by_role('button',name='采用 AI 版',exact=True).click()
    assert '浮动入口修改' in content()
    print('PASS 真实抽屉选区入口，进入对比时关闭遮罩')

    reset()
    output[0]={'kind':'outline','content':'## 背景\n- 介绍问题\n\n## 实践\n- 举例说明'}
    send('给我大纲')
    page.get_by_role('button',name='按大纲逐节写作',exact=True).click()
    expect(page.locator('#chapter-outline')).to_have_value('## 背景\n- 介绍问题\n\n## 实践\n- 举例说明')
    page.get_by_role('button',name='确认大纲，开始逐节写作',exact=True).click()
    page.get_by_role('button',name='生成本章',exact=True).click()
    expect(page.get_by_role('button',name='采纳本章到正文末尾',exact=True)).to_be_enabled()
    page.get_by_role('button',name='采纳本章到正文末尾',exact=True).click()
    assert '这是章节正文' in content()
    print('PASS 大纲转章节，现有逐节生成与采纳保留')

    reset()
    page.get_by_role('button',name='图示配图',exact=True).click()
    page.get_by_role('button',name='AI 整理配图方案',exact=True).click()
    expect(page.get_by_role('button',name='生成预览',exact=True)).to_be_enabled()
    page.get_by_role('button',name='生成预览',exact=True).click()
    expect(page.get_by_role('button',name='插入可编辑图示',exact=True)).to_be_enabled(timeout=20000)
    page.get_by_text('编辑 Mermaid 源码',exact=True).click()
    original_source=page.locator('#visual-mermaid').input_value()
    source_before=content()
    long_preview='flowchart TD\n'+'\n'.join(f'S{i}["步骤{i}"] --> S{i+1}["步骤{i+1}"]' for i in range(14))
    page.locator('#visual-mermaid').fill(long_preview)
    expect(page.get_by_role('button',name='准备分阶段方案',exact=True)).to_be_visible(timeout=20000)
    count_before=len(previews)
    page.get_by_role('button',name='准备分阶段方案',exact=True).click()
    assert '阶段概览图' in page.locator('#visual-prompt').input_value()
    assert page.locator('#visual-mermaid').input_value()==long_preview
    assert content()==source_before and len(previews)==count_before
    page.locator('#visual-mermaid').fill(original_source)
    expect(page.get_by_role('button',name='准备分阶段方案',exact=True)).to_have_count(0)
    expect(page.get_by_role('button',name='插入可编辑图示',exact=True)).to_be_enabled(timeout=20000)
    print('PASS 分阶段要求需作者确认生成，不自动改动预览与正文')
    page.get_by_role('button',name='插入可编辑图示',exact=True).click()
    assert '```mermaid' in content()
    page.get_by_role('button',name='导出图片并上传',exact=True).click()
    expect(page.get_by_role('button',name='设为封面',exact=True)).to_be_enabled(timeout=20000)
    print('PASS 配图原流程：规划、生成、插入可编辑图示')

    reset()
    page.get_by_role('button',name='图示配图',exact=True).click()
    page.locator('.visual-fields .el-select').nth(1).click()
    page.get_by_role('option',name='封面图',exact=True).click()
    page.get_by_role('button',name='生成预览',exact=True).click()
    page.get_by_role('button',name='采用并上传',exact=True).click()
    page.get_by_role('button',name='设为封面',exact=True).click()
    page.get_by_role('button',name='插入正文',exact=True).click()
    assert 'https://images.example.test/' in content()
    page.get_by_role('button',name='文字写作',exact=True).click()
    page.get_by_role('button',name='撤销 AI 修改',exact=True).click()
    assert content()=='前文\n\n选中内容\n\n后文'
    print('PASS 图片预览、采用上传、设为封面、插入正文与撤销')

    reset()
    output[0]={'kind':'advice','content':'# 检查\n<img src="/missing" onerror="window.__aiXss=1"><iframe srcdoc="<script>parent.__aiXss=2</script>"></iframe>\n\n| A | B |\n|---|---|\n|1|2|\n\n```js\nconst ok = true;\n```'}
    page.route('**/missing',lambda route:route.fulfill(status=404,body=''))
    send('输出建议')
    assert page.locator('.message-content script,.message-content iframe,.message-content [onerror]').count()==0
    assert page.evaluate('window.__aiXss === undefined')
    expect(page.locator('.message-content table')).to_be_visible()
    print('PASS 聊天 Markdown 输出过滤危险 HTML')
    if args.screenshot:page.screenshot(path=args.screenshot,full_page=True)
    # 模型管理使用真实组件，验证旧 ID 响应兼容以及各操作发送的模型 ID。
    model_requests = []
    def model_route(route):
        request = route.request
        if request.method == 'GET':
            if request.url.endswith('/providers'):
                route.fulfill(json={'code': 0, 'data': [{'value': 'openai', 'label': 'OpenAI', 'needBaseUrl': True}]})
            elif request.url.rsplit('/',1)[-1] in ['image','errorAnalysis','workflow']:
                route.fulfill(json={'code':0,'data':{'config':{'enabled':False,'modelId':0,'timeoutSeconds':180,'provider':'openai','model':''},'models':[],'hasKey':False}})
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

    long_chart='flowchart TD\n'+'\n'.join(f'N{i}["步骤{i}"] --> N{i+1}["步骤{i+1}"]' for i in range(14))
    page.locator('.markdown-textarea').fill('```mermaid\n'+long_chart+'\n```')
    expect(page.locator('.markdown-diagram')).to_have_count(1)
    page.wait_for_function('()=>{const img=document.querySelector(".markdown-diagram > img");return img?.complete&&img.naturalHeight>0}')
    assert page.locator('.markdown-diagram > img').evaluate('(el)=>el.getBoundingClientRect().height')<=421
    expect(page.locator('.markdown-diagram')).to_have_attribute('data-diagram-readability','review')
    expect(page.locator('.diagram-readability')).to_contain_text('完整流程已保留')
    page.get_by_role('button',name='查看完整图示',exact=True).click()
    viewer=page.get_by_role('dialog',name='完整图示',exact=True)
    expect(viewer).to_be_visible()
    full=viewer.locator('img')
    initial=full.evaluate('(el)=>el.getBoundingClientRect().width')
    viewer.get_by_role('button',name='放大图示',exact=True).click()
    assert full.evaluate('(el)=>el.getBoundingClientRect().width')>initial
    viewer.get_by_role('button',name='适应窗口',exact=True).click()
    page.keyboard.press('Escape');expect(viewer).to_have_count(0)
    page.get_by_role('button',name='查看完整图示',exact=True).click()
    page.evaluate("window.aiEditorTest.setContent('替换后的正文')")
    expect(viewer).to_have_count(0)
    print('PASS 长图预览高度限制、完整查看、缩放、Esc关闭与卸载清理')
    # Root direction may change, but all content and editable source survive.
    balanced='flowchart TD\n'+'\n'.join(f'A["开始"] --> N{i}["步骤{i}"] --> Z["完成"]' for i in range(9))
    page.locator('.markdown-textarea').fill('```mermaid\n'+balanced+'\n```')
    expect(page.locator('.markdown-diagram')).to_have_count(1)
    expect(page.locator('.markdown-diagram')).to_have_attribute('data-layout-adjusted','true')
    page.locator('.markdown-diagram summary').click()
    expect(page.locator('.markdown-diagram details pre')).to_have_text(balanced)
    svg=page.locator('.markdown-diagram > img').evaluate('async el=>await (await fetch(el.src)).text()')
    assert svg.count('class="node default"') == 11
    print('PASS 渲染后自动选择布局，原始源码与十一个节点完整保留')


    assert not errors,errors
    browser.close()
    print('All writing chat browser regressions passed.')
