<script setup lang="ts">
// EmotePreview: live WebGL preview of an E-mote (PSB) model playing a
// random motion, plus the visit count — used by the playground when an
// animated theme is selected (docs/emote-widget.md). All driver
// interaction happens client-side; only the model bytes and the count
// come from the Go API.
const props = defineProps<{
  model: string
  name: string
  text: string
}>()

const { t } = useI18n()
const config = useRuntimeConfig()
const apiBase = ((config.public.apiBase as string) || '').replace(/\/+$/, '')

// Preview canvas size (CSS px); physical pixels scale with devicePixelRatio.
const CSS_W = 360
const CSS_H = 520

const motionCount = ref(0)
const currentMotion = ref('')
const errorMsg = ref('')
const countText = ref('')

const container = ref<HTMLElement | null>(null)
let player: any = null
let previewCanvas: HTMLCanvasElement | null = null

// Ambient declaration for the driver's global lexical binding: the driver
// declares `class EmotePlayer` at the top level of a classic script, which
// is NOT a window property — only a bare identifier reference resolves it.
declare const EmotePlayer: any

const loadScript = (src: string) =>
  new Promise<void>((resolve, reject) => {
    const s = document.createElement('script')
    s.src = src
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(`failed to load ${src}`))
    document.head.appendChild(s)
  })

// The driver pair loads once per page and is shared by every preview
// instance. Order matters: emoteplayer.js declares the classes,
// FreeMoteDriver.js provides the emscripten runtime they call into.
let driverPromise: Promise<void> | null = null
const ensureDriver = () => {
  if (!driverPromise) {
    driverPromise = loadScript('/widget/emoteplayer.js')
      .then(() => loadScript('/widget/FreeMoteDriver.js'))
  }
  return driverPromise
}

// Contain-fit the model inside the canvas with 5% padding (same logic as
// the reference autoCenterPlayer).
const fitPlayer = () => {
  if (!player?.isCharaProfileAvailable) return
  const bounds = player.charaBounds
  if (!bounds || bounds.right === bounds.left) return
  const modelWidth = bounds.right - bounds.left
  const modelHeight = bounds.bottom - bounds.top
  if (modelWidth <= 0 || modelHeight <= 0) return
  const canvas = previewCanvas
  if (!canvas) return
  const scale = Math.min(canvas.width / modelWidth, canvas.height / modelHeight) * 0.95
  const centerX = (bounds.left + bounds.right) / 2
  const centerY = (bounds.top + bounds.bottom) / 2
  player.setScale(scale, 0)
  player.setCoord(-centerX * scale, -centerY * scale, 0)
}

// Labels that must never be picked at random (separator rows, the
// initialization timeline, and pointer-driven gaze-follow).
const excludedMotion = (label: string) =>
  label.startsWith('-') || label === '初期化' || label === '視線追従'

// Pick a random motion and play it. IsLoopTimeline is unreliable right
// after a model load (reports true for nearly everything), so liveliness
// is enforced by startMotionLoop below instead.
const playRandomMotion = () => {
  const labels = (player?.mainTimelineLabels || []).filter(
    (l: string) => !excludedMotion(l) && l !== currentMotion.value,
  )
  if (!labels.length || !player) return
  const pick = labels[Math.floor(Math.random() * labels.length)]
  player.mainTimelineLabel = pick
  currentMotion.value = pick
}

// Coarse strided sample of the preview canvas (luma+alpha per 64B).
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
// pixel noise is ignored via a per-sample threshold.
const startMotionLoop = () => {
  let prev: Uint8Array | null = null
  let lastChange = Date.now()
  setInterval(() => {
    if (!previewCanvas || document.visibilityState !== 'visible') return
    const cur = sampleCanvas(previewCanvas)
    if (!cur) return
    if (prev && cur.length === prev.length) {
      let changed = 0
      for (let i = 0; i < cur.length; i++) {
        if (Math.abs(cur[i] - prev[i]) > 12) changed++
      }
      // Real motion changes hundreds of samples; post-motion physics
      // settle and render noise stay below ~150.
      if (changed > 50) {
        lastChange = Date.now()
      } else if (Date.now() - lastChange > 3000) {
        playRandomMotion()
        lastChange = Date.now()
      }
    }
    prev = cur
  }, 900)
}

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
  const model = props.model
  if (!model) return
  errorMsg.value = ''
  try {
    await ensureDriver()
    if (typeof EmotePlayer === 'undefined') throw new Error(t('emote.loadFailed'))
    if (!player) {
      const host = container.value
      if (!host) return
      const dpr = Math.min(window.devicePixelRatio || 1, 2)
      const canvas = document.createElement('canvas')
      canvas.width = Math.round(CSS_W * dpr)
      canvas.height = Math.round(CSS_H * dpr)
      canvas.style.width = `${CSS_W}px`
      canvas.style.height = `${CSS_H}px`
      host.appendChild(canvas)
      previewCanvas = canvas
      EmotePlayer.createRenderCanvas(canvas.width, canvas.height)
      player = new EmotePlayer(canvas)
      startMotionLoop()
    }
    await player.promiseLoadDataFromURL(`${apiBase}/psb/${encodeURIComponent(model)}`)
    fitPlayer()
    const labels = (player.mainTimelineLabels || []).slice()
    motionCount.value = labels.length
    currentMotion.value = ''
    playRandomMotion()
    await refreshCount()
  } catch (e: any) {
    errorMsg.value = e?.message || String(e)
  }
}

onMounted(loadModel)
</script>

<template>
  <div class="flex flex-col items-center">
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
