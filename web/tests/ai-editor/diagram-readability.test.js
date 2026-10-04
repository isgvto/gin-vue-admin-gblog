import { test } from 'node:test'
import assert from 'node:assert/strict'
import { assessDiagram, diagramLayoutCandidates, preferDiagramLayout } from '../../src/utils/diagramReadability.js'

test('fits article preview while preserving readable labels', () => {
  assert.equal(assessDiagram({ width: 600, height: 350 }).needsReview, false)
  assert.equal(assessDiagram({ width: 300, height: 1800 }).needsReview, true)
  assert.equal(assessDiagram({ width: 2200, height: 200 }).needsReview, true)
  assert.equal(assessDiagram({ width: 600, height: 350, fontSize: 10 }).needsReview, true)
  assert.equal(assessDiagram({ width: NaN, height: 350 }).needsReview, true)
})

test('layout attempts change only root direction, preserving groups, labels and edges', () => {
  const source = 'flowchart TD\nsubgraph phase[创作]\ndirection TD\nA["准备"] --> B{"达标？"}\nB -->|否| A\nend'
  const choices = diagramLayoutCandidates(source)
  assert.equal(choices.length, 2)
  assert.equal(choices[1], source.replace('flowchart TD', 'flowchart LR'))
  assert.deepEqual(diagramLayoutCandidates('sequenceDiagram\nA->>B: 消息'), ['sequenceDiagram\nA->>B: 消息'])
})

test('retain original layout unless the candidate is meaningfully better', () => {
  const original = assessDiagram({ width: 300, height: 1800 })
  assert.equal(preferDiagramLayout(original, assessDiagram({ width: 650, height: 350 })), true)
  assert.equal(preferDiagramLayout(original, assessDiagram({ width: 300, height: 1800 })), false)
})
