import assert from 'node:assert/strict'
import test from 'node:test'
import {
  createGroupCollapseState,
  isGroupOpen,
  isGroupRendered,
  toggleGroup,
  resetGroupCollapse,
} from '../web/app/utils/groupCollapse.ts'

const openKeys = (state, keys, filtersActive) => keys.filter((key) => isGroupOpen(state, key, filtersActive))

test('unfiltered groups start closed and toggle on click', () => {
  const state = createGroupCollapseState()
  assert.deepEqual(openKeys(state, ['a', 'b'], false), [])
  toggleGroup(state, 'a', false)
  assert.equal(isGroupOpen(state, 'a', false), true)
  assert.equal(isGroupOpen(state, 'b', false), false)
  toggleGroup(state, 'a', false)
  assert.equal(isGroupOpen(state, 'a', false), false)
})

test('a filtered group opens by default but an explicit click collapses it', () => {
  const state = createGroupCollapseState()
  assert.deepEqual(openKeys(state, ['a', 'b'], true), ['a', 'b'])
  toggleGroup(state, 'a', true)
  assert.deepEqual(openKeys(state, ['a', 'b'], true), ['b'])
  toggleGroup(state, 'a', true)
  assert.deepEqual(openKeys(state, ['a', 'b'], true), ['a', 'b'])
})

test('a collapsed group stays collapsed while the user keeps typing or switches kind', () => {
  const state = createGroupCollapseState()
  toggleGroup(state, 'a', true)
  // More keystrokes and a different kind filter keep filtersActive true.
  for (const filtersActive of [true, true, true]) {
    assert.equal(isGroupOpen(state, 'a', filtersActive), false, 'explicit collapse must win')
  }
  // A newly matching group still opens by default.
  assert.equal(isGroupOpen(state, 'c', true), true)
})

test('leaving the filtered view returns every group to the closed default', () => {
  const state = createGroupCollapseState()
  toggleGroup(state, 'a', true)
  toggleGroup(state, 'b', true)
  resetGroupCollapse(state)
  assert.deepEqual(openKeys(state, ['a', 'b', 'c'], false), [])
  toggleGroup(state, 'c', false)
  assert.deepEqual(openKeys(state, ['a', 'b', 'c'], false), ['c'])
})

test('filters mount all groups; closing keeps a rendered group mounted', () => {
  const state = createGroupCollapseState()
  assert.equal(isGroupRendered(state, 'a', true), true)
  assert.equal(isGroupRendered(state, 'a', false), false)
  toggleGroup(state, 'a', false)
  assert.equal(isGroupRendered(state, 'a', false), true, 'first open mounts the group')
  toggleGroup(state, 'a', false)
  assert.equal(isGroupOpen(state, 'a', false), false)
  assert.equal(isGroupRendered(state, 'a', false), true, 'closing must not unmount it')
})

test('collapse survives a filtered-to-unfiltered-to-filtered round trip only after a reset', () => {
  const state = createGroupCollapseState()
  toggleGroup(state, 'a', true)
  // Without a reset the explicit choice is kept (used while filters stay on).
  assert.equal(isGroupOpen(state, 'a', true), false)
  resetGroupCollapse(state)
  assert.equal(isGroupOpen(state, 'a', true), true, 'a fresh filter opens results again')
})
