import assert from 'node:assert/strict'
import test from 'node:test'
import { summarizeThemes } from '../web/app/utils/themeStatistics.ts'
import { aboutMessages } from '../web/app/i18n/about.ts'

test('an empty collection has zero themes and no known variation maximum', () => {
  assert.deepEqual(summarizeThemes([]), { total: 0, maxVariants: null, mostVariedTheme: null })
})

test('the total includes static and every animated theme kind', () => {
  const themes = [
    { name: 'card', variants: 8 },
    { name: 'character', variants: 450 },
    ...['psb', 'spine', 'live2d', 'unity'].map((kind) => ({ name: kind, kind, animated: true })),
  ]
  assert.deepEqual(summarizeThemes(themes), { total: 6, maxVariants: 450, mostVariedTheme: 'character' })
})

test('the highlighted maximum is one theme, not a sum across themes', () => {
  const themes = [{ name: 'first', variants: 12 }, { name: 'largest', variants: 84 }, { name: 'last', variants: 40 }]
  assert.deepEqual(summarizeThemes(themes), { total: 3, maxVariants: 84, mostVariedTheme: 'largest' })
})

test('animated-only collections do not invent a variation count', () => {
  const result = summarizeThemes([{ name: 'motion-a' }, { name: 'motion-b' }])
  assert.equal(result.total, 2)
  assert.equal(result.maxVariants, null)
  assert.equal(result.mostVariedTheme, null)
})

test('invalid variation values are ignored without removing themes from the total', () => {
  const values = [undefined, 0, -2, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]
  const themes = values.map((variants, index) => ({ name: `invalid-${index}`, variants }))
  themes.push({ name: 'valid', variants: 1 })
  assert.deepEqual(summarizeThemes(themes), { total: values.length + 1, maxVariants: 1, mostVariedTheme: 'valid' })
})

test('equal maxima keep the first theme as a stable Playground destination', () => {
  assert.deepEqual(summarizeThemes([{ name: 'a', variants: 20 }, { name: 'b', variants: 20 }]), {
    total: 2, maxVariants: 20, mostVariedTheme: 'a',
  })
})

test('large collections and large safe integers preserve exact totals and maxima', () => {
  const themes = Array.from({ length: 25000 }, (_, index) => ({ name: `theme-${index}`, variants: index + 1 }))
  themes.push({ name: 'large-combination', variants: Number.MAX_SAFE_INTEGER })
  assert.deepEqual(summarizeThemes(themes), {
    total: 25001, maxVariants: Number.MAX_SAFE_INTEGER, mostVariedTheme: 'large-combination',
  })
})

test('statistics never reorder or mutate the API collection', () => {
  const themes = Object.freeze([Object.freeze({ name: 'b', variants: 5 }), Object.freeze({ name: 'a', variants: 9 })])
  summarizeThemes(themes)
  assert.deepEqual(themes.map((theme) => theme.name), ['b', 'a'])
})

test('all three locales contain complete introduction, statistics and usage messages', () => {
  const keys = Object.keys(aboutMessages.en).sort()
  for (const messages of Object.values(aboutMessages)) {
    assert.deepEqual(Object.keys(messages).sort(), keys)
    assert.ok(Object.values(messages).every((message) => message.trim().length > 0))
    for (const step of ['choose', 'generate', 'embed']) {
      assert.ok(messages[`aboutPage.${step}Title`])
      assert.ok(messages[`aboutPage.${step}Description`])
    }
  }
})
