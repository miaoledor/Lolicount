// useRippleFx appends an absolutely positioned span to any [data-ripple]
// host. If the host is not positioned, the circle is offset from the sticky
// nav bar instead of the button and is never clipped. These assertions keep
// both LangSwitch variants in sync with the other ripple hosts in NavBar.
import assert from 'node:assert/strict'
import test from 'node:test'
import { readFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const source = await readFile(join(root, 'web/app/components/LangSwitch.vue'), 'utf8')

test('the nav language button opts into the ripple effect', () => {
  const nav = source.slice(source.indexOf('<button'), source.indexOf('</button>'))
  assert.match(nav, /data-ripple/, 'nav button must carry data-ripple')
  assert.match(nav, /class="lang-cycle"/)
})

test('the .lang-cycle rule anchors and clips the ripple', () => {
  const rule = source.match(/\.lang-cycle\s*\{([\s\S]*?)\}/)?.[1] ?? ''
  assert.match(rule, /position:\s*relative/, 'ripple must be positioned against the button')
  assert.match(rule, /overflow:\s*hidden/, 'ripple must be clipped to the button')
})

test('inline locale buttons get the same ripple treatment', () => {
  const inline = source.slice(source.indexOf('v-else'), source.indexOf('</template>'))
  assert.match(inline, /data-ripple/, 'inline buttons must carry data-ripple')
  assert.match(inline, /relative overflow-hidden/, 'inline buttons must anchor and clip the ripple')
})
