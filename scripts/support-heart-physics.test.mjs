import assert from 'node:assert/strict'
import test from 'node:test'
import { createHeartBody, HEART_RADIUS, JAR_FLOOR, JAR_RADIUS, MAX_HEARTS, stepHearts } from '../web/app/utils/support/heartPhysics.ts'
import { supportMessages } from '../web/app/i18n/support.ts'

const assertContained = (bodies) => {
  for (const body of bodies) {
    assert.ok(Object.values(body).every(Number.isFinite), 'simulation must remain finite')
    assert.ok(body.y >= JAR_FLOOR + HEART_RADIUS - 1e-8, 'hearts must stay above the floor')
    assert.ok(Math.hypot(body.x, body.z) <= JAR_RADIUS - HEART_RADIUS + 1e-8, 'hearts must stay inside the glass')
  }
}

test('new hearts start above the opening and within the jar radius', () => {
  for (let index = 0; index < 200; index++) {
    const body = createHeartBody(index, true)
    assert.ok(body.y > 2.8)
    assert.ok(Math.hypot(body.x, body.z) < JAR_RADIUS - HEART_RADIUS)
  }
})

test('gravity moves a newly added heart down', () => {
  const body = createHeartBody(0, true)
  const initialHeight = body.y
  stepHearts([body], 1 / 60)
  assert.ok(body.y < initialHeight)
  assert.ok(body.vy < 0)
})

test('zero and negative time leave the simulation untouched', () => {
  const bodies = [createHeartBody(0, true), createHeartBody(1, true)]
  const snapshot = structuredClone(bodies)
  stepHearts(bodies, 0)
  stepHearts(bodies, -2)
  assert.deepEqual(bodies, snapshot)
})

test('a suspended tab cannot produce an unbounded physics step', () => {
  const longPause = [createHeartBody(1, true)]
  const bounded = structuredClone(longPause)
  stepHearts(longPause, 3600)
  stepHearts(bounded, 1 / 30)
  assert.deepEqual(longPause, bounded)
})

test('floor contact bounces a falling heart back into the jar', () => {
  const body = { x: 0, y: 0.1, z: 0, vx: 0.4, vy: -5, vz: 0.2 }
  stepHearts([body], 1 / 60)
  assertContained([body])
  assert.ok(body.vy > 0)
  assert.ok(body.vx < 0.4)
})

test('wall contact redirects outward velocity', () => {
  for (const angle of [0, 0.5, 1.5, 3, 5]) {
    const nx = Math.cos(angle)
    const nz = Math.sin(angle)
    const body = { x: nx * 1.1, y: 1, z: nz * 1.1, vx: nx * 5, vy: 0, vz: nz * 5 }
    stepHearts([body], 1 / 60)
    assertContained([body])
    assert.ok(body.vx * nx + body.vz * nz < 0)
  }
})

test('coincident hearts separate without division by zero', () => {
  const bodies = Array.from({ length: 8 }, () => ({ x: 0, y: 1, z: 0, vx: 0, vy: 0, vz: 0 }))
  for (let frame = 0; frame < 600; frame++) stepHearts(bodies, 1 / 60)
  assertContained(bodies)
  assert.ok(new Set(bodies.map((body) => body.x)).size > 1)
})

for (const count of [1, 20, MAX_HEARTS]) {
  test(`${count} hearts stay stable and contained over 40 simulated seconds`, () => {
    const bodies = Array.from({ length: count }, (_, index) => createHeartBody(index))
    for (let frame = 0; frame < 2400; frame++) {
      stepHearts(bodies, frame % 11 === 0 ? 1 / 30 : 1 / 60)
      assertContained(bodies)
    }
    assert.ok(Math.max(...bodies.map((body) => body.y)) < 2.8, 'a full jar must fit below its opening')
  })
}

test('an added heart joins an existing pile', () => {
  const bodies = Array.from({ length: 20 }, (_, index) => createHeartBody(index))
  for (let frame = 0; frame < 240; frame++) stepHearts(bodies, 1 / 60)
  const added = createHeartBody(20, true)
  bodies.push(added)
  for (let frame = 0; frame < 600; frame++) stepHearts(bodies, 1 / 60)
  assertContained(bodies)
  assert.ok(added.y < 2.8)
})

test('every support locale has complete messages and the heart count placeholder', () => {
  const expected = Object.keys(supportMessages.en).sort()
  for (const messages of Object.values(supportMessages)) {
    assert.deepEqual(Object.keys(messages).sort(), expected)
    assert.ok(Object.values(messages).every((message) => message.trim().length > 0))
    assert.ok(messages['support.thanks'].includes('{n}'))
  }
})
