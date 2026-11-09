// The nav language control is now a real dropdown. These assertions pin the
// accessibility contract and, importantly, that the menu anchor is NOT
// clipped — .lang-cycle must keep overflow:hidden for its ripple, so the
// menu has to live outside the button.
import assert from 'node:assert/strict'
import test from 'node:test'
import { readFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const component = await readFile(join(root, 'web/app/components/LangSwitch.vue'), 'utf8')
const locales = await readFile(join(root, 'web/app/i18n/locales.ts'), 'utf8')

test('the nav trigger announces a popup menu', () => {
  assert.match(component, /aria-haspopup="menu"/)
  assert.match(component, /:aria-expanded="open"/)
  assert.match(component, /@click="toggleMenu"/)
})

test('the dropdown lists every locale as a selectable radio item', () => {
  assert.match(component, /role="menu"/)
  assert.match(component, /v-for="l in locales"/)
  assert.match(component, /role="menuitemradio"/)
  assert.match(component, /:aria-checked="l === locale"/)
  assert.match(component, /@click="choose\(l\)"/)
  assert.match(component, /setLocale\(next\)/, 'choosing must actually switch the locale')
})

test('outside press and Escape close the menu and return focus', () => {
  assert.match(component, /addEventListener\('pointerdown', onPointerDown, true\)/)
  assert.match(component, /removeEventListener\('pointerdown', onPointerDown, true\)/)
  assert.match(component, /event\.key === 'Escape'/)
  assert.match(component, /triggerEl\.value\?\.focus\(\)/)
})

test('the menu anchor is positioned but never clipped', () => {
  const anchor = component.match(/\.lang-switch\s*\{([\s\S]*?)\}/)?.[1] ?? ''
  assert.match(anchor, /position:\s*relative/)
  assert.ok(!/overflow:\s*hidden/.test(anchor), 'clipping the anchor would hide the dropdown')

  const trigger = component.match(/\.lang-cycle\s*\{([\s\S]*?)\}/)?.[1] ?? ''
  assert.match(trigger, /overflow:\s*hidden/, 'the trigger must still clip its press ripple')
})

test('the language label is translated in all three locales', () => {
  const values = [...locales.matchAll(/'nav\.language':\s*'([^']*)'/g)].map((m) => m[1])
  assert.equal(values.length, 3, 'expected one nav.language entry per locale')
  assert.equal(new Set(values).size, 3, 'labels must not fall back to the same text')
  assert.ok(values.every((value) => value.trim().length > 0))
})
