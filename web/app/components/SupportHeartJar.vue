<script setup lang="ts">
import { cn } from '~/utils/cn'
import type { HeartJarScene } from '~/utils/support/heartJarScene'

const { t } = useI18n()
const emit = defineEmits<{ heart: [] }>()
const host = ref<HTMLElement | null>(null)
const status = ref<'loading' | 'ready' | 'unavailable'>('loading')
let scene: HeartJarScene | undefined
let disposed = false
let motion: MediaQueryList | undefined
let startX = 0
let previousX = 0
let dragging = false
let moved = false

const addHeart = () => {
  scene?.addHeart()
  emit('heart')
}
const clickJar = () => { if (!moved) addHeart() }
const pointerDown = (event: PointerEvent) => {
  if (event.button !== 0) return
  startX = previousX = event.clientX
  dragging = true
  moved = false
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
}
const pointerMove = (event: PointerEvent) => {
  if (!dragging) return
  if (Math.abs(event.clientX - startX) > 6) moved = true
  if (moved) scene?.rotate((event.clientX - previousX) * 0.009)
  previousX = event.clientX
}
const pointerEnd = () => { dragging = false }
const motionChange = () => scene?.setReducedMotion(motion?.matches ?? false)
const contextLost = (event: Event) => {
  event.preventDefault()
  scene?.dispose()
  scene = undefined
  status.value = 'unavailable'
}

onMounted(async () => {
  motion = window.matchMedia('(prefers-reduced-motion: reduce)')
  motion.addEventListener('change', motionChange)
  try {
    const { createHeartJarScene } = await import('~/utils/support/heartJarScene')
    if (disposed || !host.value) return
    scene = createHeartJarScene(host.value, motion.matches)
    host.value.querySelector('canvas')?.addEventListener('webglcontextlost', contextLost)
    status.value = 'ready'
  } catch {
    if (!disposed) status.value = 'unavailable'
  }
})

onBeforeUnmount(() => {
  disposed = true
  motion?.removeEventListener('change', motionChange)
  host.value?.querySelector('canvas')?.removeEventListener('webglcontextlost', contextLost)
  scene?.dispose()
})

defineExpose({ addHeart })
</script>

<template>
  <div class="heart-jar-wrap">
    <div
      :class="cn('heart-jar', status === 'ready' && 'is-ready')"
      role="button"
      tabindex="0"
      :aria-label="t('support.jarLabel')"
      aria-describedby="jar-hint"
      data-no-cursor-firework
      @click="clickJar"
      @pointerdown="pointerDown"
      @pointermove="pointerMove"
      @pointerup="pointerEnd"
      @pointercancel="pointerEnd"
      @keydown.enter.prevent="addHeart"
      @keydown.space.prevent="addHeart"
      @keydown.left.prevent="scene?.rotate(-0.2)"
      @keydown.right.prevent="scene?.rotate(0.2)"
    >
      <div ref="host" class="jar-canvas" />
      <div v-if="status !== 'ready'" class="jar-fallback" aria-hidden="true">
        <div class="jar-outline">
          <span v-for="i in 12" :key="i" :style="{ rotate: `${(i % 3 - 1) * 22}deg` }">♥</span>
        </div>
      </div>
    </div>
    <p id="jar-hint" class="jar-hint">
      <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="m8 9-3 3 3 3m8-6 3 3-3 3M5 12h14" /></svg>
      {{ t(status === 'unavailable' ? 'support.jarFallback' : 'support.jarHint') }}
    </p>
  </div>
</template>

<style scoped>
.heart-jar-wrap { position: relative; width: 100%; height: 100%; }
.heart-jar { position: relative; width: 100%; height: 100%; outline: none; touch-action: pan-y; cursor: grab; -webkit-tap-highlight-color: transparent; }
.heart-jar:active { cursor: grabbing; }
.heart-jar:focus-visible { outline: 2px solid var(--accent); outline-offset: -12px; border-radius: 1.5rem; }
.jar-canvas { position: absolute; inset: 0; }
.jar-canvas :deep(canvas) { display: block; width: 100%; height: 100%; }
.jar-hint { position: absolute; bottom: 0.5rem; left: 0; right: 0; display: flex; justify-content: center; align-items: center; gap: 0.4rem; margin: 0; color: var(--text-muted); font-size: 0.75rem; pointer-events: none; }
.jar-fallback { position: absolute; inset: 0; display: grid; place-items: center; }
.jar-outline { display: flex; flex-wrap: wrap-reverse; justify-content: center; align-content: end; width: 138px; height: 180px; padding: 12px; border: 3px solid var(--loli-border); border-top: 7px double var(--text-muted); border-radius: 12px 12px 24px 24px; box-shadow: inset 8px 0 12px #ffffff20; }
.jar-outline span { color: #f57e96; font-size: 38px; line-height: 0.95; }
</style>
