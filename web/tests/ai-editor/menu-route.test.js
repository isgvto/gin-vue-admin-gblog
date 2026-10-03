import test from 'node:test'
import assert from 'node:assert/strict'
import { joinMenuPath, firstMenuPage } from '../../src/utils/menuRoute.js'

test('移动分类后完整页面地址保持原样，普通嵌套路由继续拼接', () => {
  assert.equal(joinMenuPath(['admin', 'navNavigation'], '/layout/admin/menu'), '/layout/admin/menu')
  assert.equal(joinMenuPath(['gblog', 'navArticles'], '/layout/gblog/edit/:id?'), '/layout/gblog/edit/:id?')
  assert.equal(joinMenuPath(['systemTools'], 'autoCodeEdit/:id'), 'systemTools/autoCodeEdit/:id')
})

test('分类跳转到可见页面，跳过隐藏页面、外链和带参数的编辑页', () => {
  const target = firstMenuPage([
    { path: 'hidden', hidden: true },
    { path: 'https://example.com' },
    { path: 'group', children: [{ path: 'edit/:id' }, { path: '/layout/gblog/list' }] }
  ], ['gblog'])
  assert.equal(target, '/layout/gblog/list')
  assert.equal(firstMenuPage([{path:'state'}]), '/layout/state')
  assert.equal(firstMenuPage([{path:'hidden',hidden:true}]), null)
})
