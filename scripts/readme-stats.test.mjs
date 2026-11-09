import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

// The three README cards must stay structurally identical: GitHub keeps the
// table markup but not Markdown inside HTML blocks, so a missing column or a
// stale number would only show up on the repository front page.
const cards = [
  { file: 'README.md', label: 'Fastest request speed' },
  { file: 'README.zh.md', label: '最快请求速度' },
  { file: 'README.ja.md', label: '最速リクエスト速度' },
]

const readCard = (file) => {
  const html = readFileSync(new URL(`../${file}`, import.meta.url), 'utf8')
  const start = html.indexOf('<table>')
  const end = html.indexOf('</table>')
  assert.ok(start >= 0 && end > start, `${file} has no stats table`)
  return html.slice(start, end)
}

test('all three README cards show the same three statistics', () => {
  for (const { file, label } of cards) {
    const card = readCard(file)
    assert.equal((card.match(/width="33%"/g) || []).length, 3, `${file} should have three columns`)
    assert.ok(card.includes('<h2>1,574</h2>'), `${file} missing theme count`)
    assert.ok(card.includes('<h2>311,040</h2>'), `${file} missing variation count`)
    assert.ok(card.includes('<h2>14,000+</h2>'), `${file} missing fastest request speed`)
    assert.ok(card.includes(`<b>${label}</b>`), `${file} missing localized label`)
    assert.ok(card.includes('colspan="3"'), `${file} badge row should span three columns`)
  }
})
