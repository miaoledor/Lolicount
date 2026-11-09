// The dictionary is a TypeScript module with extensionless relative imports,
// so Node's type stripping cannot load it directly. Parse the three locale
// blocks as text instead: this keeps the assertions dependency-free and
// independent of how the i18n module is bundled.
import assert from 'node:assert/strict'
import test from 'node:test'
import { readFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const source = await readFile(join(root, 'web/app/i18n/locales.ts'), 'utf8')

const boundaries = [
  ['zh', 'const zh: Dict = {', 'const en: Dict = {'],
  ['en', 'const en: Dict = {', 'const jp: Dict = {'],
  ['jp', 'const jp: Dict = {', 'export const dictionaries'],
]

const blocks = Object.fromEntries(boundaries.map(([locale, start, end]) => {
  const from = source.indexOf(start)
  const to = source.indexOf(end, from + 1)
  assert.ok(from >= 0 && to > from, `${locale} dictionary block not found`)
  return [locale, source.slice(from, to)]
}))

const valueOf = (block, key) => {
  const pattern = new RegExp(`'${key.replaceAll('.', '\\.')}':\\s*'([^']*)'`)
  return block.match(pattern)?.[1]
}

test('the generator heading is localized in all three languages', () => {
  assert.equal(valueOf(blocks.zh, 'themesGallery.generateTitle'), '生成计数器')
  assert.equal(valueOf(blocks.en, 'themesGallery.generateTitle'), 'Create a counter')
  assert.equal(valueOf(blocks.jp, 'themesGallery.generateTitle'), 'カウンターを作成')
  assert.ok(!/[\u3400-\u9fff]/.test(valueOf(blocks.en, 'themesGallery.generateTitle')),
    'the English heading must not fall back to Chinese')
  assert.notEqual(valueOf(blocks.jp, 'themesGallery.generateTitle'), '生成计数器',
    'the Japanese heading must not fall back to Chinese')
})

test('every locale labels the name field and shows an example placeholder', () => {
  for (const [locale, block] of Object.entries(blocks)) {
    assert.ok(valueOf(block, 'param.nameLabel')?.trim(), `${locale} missing param.nameLabel`)
    const placeholder = valueOf(block, 'param.namePlaceholder')
    assert.ok(placeholder?.trim(), `${locale} missing param.namePlaceholder`)
    assert.ok(placeholder.includes('my-counter'), `${locale} placeholder should show an example`)
  }
})

test('copy/embed labels stay available in every locale', () => {
  for (const [locale, block] of Object.entries(blocks)) {
    for (const key of ['embed.svg', 'embed.img', 'embed.markdown', 'embed.widget', 'embed.copy', 'embed.copied']) {
      assert.ok(valueOf(block, key)?.trim(), `${locale} missing ${key}`)
    }
  }
})
