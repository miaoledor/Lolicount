import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveAppVersion } from './app-version.mjs'

const noTag = () => ''
const nearestTag = () => 'v0.4.4'

test('an explicit build-time version wins over every other source', () => {
  const env = {
    NUXT_PUBLIC_APP_VERSION: 'v9.9.9',
    APP_VERSION: 'v1.0.0',
    GITHUB_REF_NAME: 'v2.0.0',
  }
  assert.equal(resolveAppVersion(env, nearestTag), 'v9.9.9')
})

test('APP_VERSION is used when the Nuxt variable is absent', () => {
  assert.equal(resolveAppVersion({ APP_VERSION: 'v0.4.5' }, noTag), 'v0.4.5')
})

test('a GitHub tag ref supplies the version without a build arg', () => {
  assert.equal(resolveAppVersion({ GITHUB_REF_NAME: 'v0.5.0' }, noTag), 'v0.5.0')
})

test('local builds fall back to the nearest git tag', () => {
  assert.equal(resolveAppVersion({}, nearestTag), 'v0.4.4')
})

test('branch refs and missing tags produce no version', () => {
  assert.equal(resolveAppVersion({ GITHUB_REF_NAME: 'main' }, noTag), '')
  assert.equal(resolveAppVersion({}, noTag), '')
  assert.equal(resolveAppVersion({}, () => 'not-a-tag'), '')
})
