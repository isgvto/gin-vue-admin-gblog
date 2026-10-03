import test from 'node:test'
import assert from 'node:assert/strict'
import { diagramExportSize } from '../../src/utils/diagramExport.js'
test('vector export preserves aspect ratio at high resolution regardless of sidebar width', () => {
  const size = diagramExportSize(1800, 540)
  assert.equal(size.width, 4096)
  assert.ok(Math.abs(size.width / size.height - 1800 / 540) < .01)
  assert.equal(diagramExportSize(300, 100).width, 3200)
  assert.equal(diagramExportSize(100, 300).height, 3200)
  assert.throws(() => diagramExportSize(0, 20))
  assert.throws(() => diagramExportSize(NaN, 20))
})
