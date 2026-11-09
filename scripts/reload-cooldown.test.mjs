import assert from 'node:assert/strict'
import test from 'node:test'
import { RELOAD_COOLDOWN_MS, createReloadCooldown, remainingWholeSeconds } from '../web/app/utils/reloadCooldown.ts'

test('the preview cooldown lasts two seconds', () => {
  assert.equal(RELOAD_COOLDOWN_MS, 2000)
})

test('the countdown shows only whole seconds', () => {
  assert.equal(remainingWholeSeconds(2000), 2)
  assert.equal(remainingWholeSeconds(1001), 2)
  assert.equal(remainingWholeSeconds(1000), 1)
  assert.equal(remainingWholeSeconds(1), 1)
  assert.equal(remainingWholeSeconds(0), 0)
  assert.equal(remainingWholeSeconds(-10), 0)
})

test('a second trigger inside the window is rejected', () => {
  let now = 1000
  const gate = createReloadCooldown(() => now)
  assert.equal(gate.trigger(), true)
  now += 500
  assert.equal(gate.cooling(), true)
  assert.equal(gate.trigger(), false)
  assert.equal(gate.remaining(), 1500)
})

test('the gate reopens exactly when the window elapses', () => {
  let now = 0
  const gate = createReloadCooldown(() => now)
  assert.equal(gate.trigger(), true)
  now = RELOAD_COOLDOWN_MS - 1
  assert.equal(gate.trigger(), false)
  now = RELOAD_COOLDOWN_MS
  assert.equal(gate.cooling(), false)
  assert.equal(gate.trigger(), true)
})

test('a custom window is honored without changing the default', () => {
  let now = 0
  const gate = createReloadCooldown(() => now, 300)
  gate.trigger()
  now = 299
  assert.equal(gate.trigger(), false)
  now = 300
  assert.equal(gate.trigger(), true)
})

test('reset cancels an active window', () => {
  let now = 0
  const gate = createReloadCooldown(() => now)
  gate.trigger()
  gate.reset()
  assert.equal(gate.cooling(), false)
  assert.equal(gate.remaining(), 0)
  assert.equal(gate.trigger(), true)
})
