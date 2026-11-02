<script lang="ts">
// EmotePreview: live WebGL preview of an E-mote (PSB) model playing a
// random motion, plus the visit count — used by the playground when an
// animated theme is selected (docs/emote-widget.md).
//
// The WebGL player is a PAGE-LEVEL SINGLETON living in this module-scope
// block (shared across component instances): the reference implementation
// never destroys a player — reloading model data on the same instance is
// the supported path, while tearing instances down mid-render aborts the
// asm.js runtime (heap use-after-free) and instantiating a second player
// with an ~18MB model exhausts its fixed heap. The playground remounts
// this component on every regenerate, so each mount only re-parents the
// singleton canvas and re-picks a motion.

// Ambient declaration for the driver's global lexical binding: the driver
// declares `class EmotePlayer` at the top level of a classic script,
// which is NOT a window property — only a bare identifier resolves it.
declare const EmotePlayer: any

// Preview canvas size (CSS px); physical pixels scale with devicePixelRatio.
const CSS_W = 360
const CSS_H = 520

const loadScript = (src: string) =>
  new Promise<void>((resolve, reject) => {
    const s = document.createElement('script')
    s.src = src
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(`failed to load ${src}`))
    document.head.appendChild(s)
  })

// The driver pair loads once per page. Order matters: emoteplayer.js
// declares the classes, FreeMoteDriver.js provides the emscripten
// runtime they call into. The runtime's default asm.js heap (96MB)
// cannot hold several large models in sequence; EmoteModule is its
// documented config hook, so a larger initial heap is reserved before
// it loads.
let driverPromise: Promise<void> | null = null
const ensureDriver = () => {
  if (!driverPromise) {
    ;(window as any).EmoteModule = (window as any).EmoteModule || { TOTAL_MEMORY: 536870912 }
    driverPromise = loadScript('/widget/emoteplayer.js')
      .then(() => loadScript('/widget/FreeMoteDriver.js'))
  }
  return driverPromise
}

let sharedPlayer: any = null
let sharedCanvas: HTMLCanvasElement | null = null
let sharedModel = ''

// Model loads are serialized: rapid theme switching must never overlap
// two promiseLoadDataFromURL calls on the same player (its loadData
// unloads/reinitializes in place, and interleaved calls abort the
// asm.js runtime).
let modelChain: Promise<void> = Promise.resolve()
const loadModelData = (url: string) => {
  const task = modelChain.then(() => sharedPlayer!.promiseLoadDataFromURL(url))
  modelChain = task.catch(() => {})
  return task
}

// The active instance's motion re-pick callback; reassigned on mount so
// the loop always talks to the live component.
let repickMotion: (() => void) | null = null

// True while the loaded model has pickable diff timelines (body overlays
// such as idle loops or special gestures). Diff loops keep the canvas
// moving, so the loop's static check alone would freeze the expression.
let hasDiffOverlay = false

// Labels that must never be picked at random (separator rows, the
// initialization timeline, and pointer-driven gaze-follow).
const excludedMotion = (label: string) =>
  label.startsWith('-') || label === '初期化' || label === '視線追従'

// Coarse strided sample of the visible canvas (luma+alpha per 64B).
const sampleCanvas = (canvas: HTMLCanvasElement): Uint8Array | null => {
  try {
    const d = canvas.getContext('2d')!.getImageData(0, 0, canvas.width, canvas.height).data
    const out = new Uint8Array(Math.ceil(d.length / 64))
    for (let i = 0, j = 0; i < d.length; i += 64, j++) {
      out[j] = (d[i] + d[i + 3]) & 0xff
    }
    return out
  } catch {
    return null
  }
}

// One-shot motions freeze on their last frame; re-pick a random motion
// when the canvas stays visibly static for a few seconds. Sub-visual
// pixel noise is ignored via a per-sample threshold (real motion changes
// hundreds of samples; post-motion physics settle stays below ~150).
// When a diff overlay loops forever the canvas never goes static, so a
// timed re-pick keeps the expression rotating as well.
const startMotionLoop = () => {
  let prev: Uint8Array | null = null
  let lastChange = Date.now()
  let lastPick = Date.now()
  const repick = () => {
    repickMotion?.()
    lastChange = Date.now()
    lastPick = lastChange
  }
  setInterval(() => {
    if (!sharedCanvas || document.visibilityState !== 'visible') return
    const cur = sampleCanvas(sharedCanvas)
    if (!cur) return
    if (prev && cur.length === prev.length) {
      let changed = 0
      for (let i = 0; i < cur.length; i++) {
        if (Math.abs(cur[i] - prev[i]) > 12) changed++
      }
      if (changed > 50) {
        lastChange = Date.now()
      } else if (Date.now() - lastChange > 3000) {
        repick()
      }
      if (hasDiffOverlay && Date.now() - lastPick > 6000) {
        repick()
      }
    }
    prev = cur
  }, 900)
}
</script>

<script setup lang="ts">
const props = defineProps<{
  model: string
  name: string
  text: string
}>()

const { t } = useI18n()
const config = useRuntimeConfig()
const apiBase = ((config.public.apiBase as string) || '').replace(/\/+$/, '')

const motionCount = ref(0)
const currentMain = ref('')
const currentDiff = ref('')
const currentMotion = computed(() =>
  currentDiff.value ? `${currentMain.value} + ${currentDiff.value}` : currentMain.value,
)
const errorMsg = ref('')
const countText = ref('')

const container = ref<HTMLElement | null>(null)

// Contain-fit the model inside the canvas with 5% padding (same logic as
// the reference autoCenterPlayer).
const fitPlayer = () => {
  if (!sharedPlayer?.isCharaProfileAvailable) return
  const bounds = sharedPlayer.charaBounds
  if (!bounds || bounds.right === bounds.left) return
  const modelWidth = bounds.right - bounds.left
  const modelHeight = bounds.bottom - bounds.top
  if (modelWidth <= 0 || modelHeight <= 0 || !sharedCanvas) return
  const scale = Math.min(sharedCanvas.width / modelWidth, sharedCanvas.height / modelHeight) * 0.95
  const centerX = (bounds.left + bounds.right) / 2
  const centerY = (bounds.top + bounds.bottom) / 2
  sharedPlayer.setScale(scale, 0)
  sharedPlayer.setCoord(-centerX * scale, -centerY * scale, 0)
}

// Pick a random main timeline (expression) and a random diff timeline
// (body overlay). Many E-mote models ship their body animation only as
// diff timelines — the main list is expressions only — so playing one of
// each is what makes the preview move like the game. IsLoopTimeline is
// unreliable right after a model load (reports true for nearly
// everything), so liveliness is enforced by the module-scope motion loop.
const pickableLabels = (labels: string[] | undefined) =>
  (labels || []).filter((l: string) => !excludedMotion(l))

const playRandomMotion = () => {
  if (!sharedPlayer) return
  const mainLabels = pickableLabels(sharedPlayer.mainTimelineLabels).filter(
    (l: string) => l !== currentMain.value,
  )
  if (!mainLabels.length) return
  currentMain.value = mainLabels[Math.floor(Math.random() * mainLabels.length)]
  sharedPlayer.mainTimelineLabel = currentMain.value

  const diffLabels = pickableLabels(sharedPlayer.diffTimelineLabels)
  hasDiffOverlay = diffLabels.length > 0
  if (!hasDiffOverlay) {
    currentDiff.value = ''
    sharedPlayer.setDiffTimelineLabel(0, '', sharedPlayer.diffTimelineFadeoutTime)
    return
  }
  const nextDiff = diffLabels.filter((l: string) => l !== currentDiff.value)
  if (nextDiff.length) {
    currentDiff.value = nextDiff[Math.floor(Math.random() * nextDiff.length)]
    sharedPlayer.setDiffTimelineLabel(0, currentDiff.value, sharedPlayer.diffTimelineFadeoutTime)
  }
}
repickMotion = playRandomMotion

const refreshCount = async () => {
  const name = props.name.trim() || 'demo'
  try {
    const d = await $fetch<{ name: string; num: number }>(
      `${apiBase}/api/count/@${encodeURIComponent(name)}`,
    )
    countText.value = props.text
      ? props.text.split('{n}').join(String(d.num))
      : String(d.num)
  } catch {
    countText.value = ''
  }
}

const loadModel = async () => {
  const host = container.value
  if (!host || !props.model) return
  errorMsg.value = ''
  try {
    await ensureDriver()
    if (typeof EmotePlayer === 'undefined') throw new Error(t('emote.loadFailed'))
    if (!sharedPlayer) {
      const dpr = Math.min(window.devicePixelRatio || 1, 2)
      const canvas = document.createElement('canvas')
      canvas.width = Math.round(CSS_W * dpr)
      canvas.height = Math.round(CSS_H * dpr)
      canvas.style.width = '100%'
      canvas.style.maxWidth = `${CSS_W}px`
      canvas.style.height = 'auto'
      canvas.style.aspectRatio = `${CSS_W} / ${CSS_H}`
      sharedCanvas = canvas
      EmotePlayer.createRenderCanvas(canvas.width, canvas.height)
      sharedPlayer = new EmotePlayer(canvas)
      startMotionLoop()
    }
    // Re-parent the singleton canvas into this instance's container.
    host.appendChild(sharedCanvas)
    // Reload only when the model actually changed; the player handles
    // unload/reload internally on the same instance.
    if (sharedModel !== props.model) {
      await loadModelData(`${apiBase}/psb/${encodeURIComponent(props.model)}`)
      sharedModel = props.model
    }
    fitPlayer()
    motionCount.value =
      pickableLabels(sharedPlayer.mainTimelineLabels).length +
      pickableLabels(sharedPlayer.diffTimelineLabels).length
    playRandomMotion()
    await refreshCount()
  } catch (e: any) {
    errorMsg.value = e?.message || String(e)
  }
}

// Re-load when the model prop changes: the component is kept alive across
// card selections, and the shared singleton player would otherwise keep
// showing the previously loaded model.
watch(() => props.model, loadModel)
onMounted(loadModel)
</script>

<template>
  <div class="preview-body">
    <div ref="container" class="flex justify-center"></div>
    <p v-if="countText" class="text-lg font-semibold mt-2">{{ countText }}</p>
    <p v-if="currentMotion" class="text-xs text-gray-500 mt-1">
      {{ t('emote.currentMotion') }}: {{ currentMotion }}
    </p>
    <p v-if="motionCount" class="text-xs text-gray-400 mt-0.5">
      {{ t('emote.motions') }}: {{ motionCount }}
    </p>
    <p v-if="errorMsg" class="text-sm text-red-600 mt-4">{{ errorMsg }}</p>
  </div>
</template>

<style scoped>
.preview-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  min-width: 0;
}

.preview-body > div {
  width: 100%;
}

.preview-body canvas {
  display: block;
  width: 100%;
  height: auto;
  max-width: 360px;
  aspect-ratio: 360 / 520;
  margin: 0 auto;
}
</style>
