<script setup lang="ts">
// Emote playground: live WebGL preview of the embedded E-mote (PSB) models
// with random-motion playback, plus the copyable widget embed snippet
// (docs/emote-widget.md). All driver interaction happens client-side; the
// page only fetches the model list / bytes / count from the Go API.
const { t } = useI18n()
const { publicBase, fetchConfig } = useApi()
const config = useRuntimeConfig()
const apiBase = ((config.public.apiBase as string) || '').replace(/\/+$/, '')

// Preview canvas size (CSS px); physical pixels scale with devicePixelRatio.
const CSS_W = 360
const CSS_H = 520

const models = ref<string[]>([])
const selectedModel = ref('')
const motionLabels = ref<string[]>([])
const currentMotion = ref('')
const loading = ref(false)
const errorMsg = ref('')
const countText = ref('')

// Widget snippet inputs.
const counterName = ref('demo')
const textTpl = ref('{n}')

const container = ref<HTMLElement | null>(null)
let player: any = null
let previewCanvas: HTMLCanvasElement | null = null

// Labels that must never be picked at random (separator rows, the
// initialization timeline, and pointer-driven gaze-follow).
const excludedMotion = (label: string) =>
  label.startsWith('-') || label === '初期化' || label === '視線追従'

// Pick a random motion and play it. IsLoopTimeline is unreliable right
// after a model load (reports true for nearly everything), so liveliness
// is enforced by startMotionLoop below instead.
const playRandomMotion = () => {
  const labels = (motionLabels.value || []).filter(
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

// Ambient declaration for the driver's global lexical binding (see the
// comment in loadModel); it exists only after ensureDriver() resolves.
declare const EmotePlayer: any

const loadScript = (src: string) =>
  new Promise<void>((resolve, reject) => {
    const s = document.createElement('script')
    s.src = src
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(`failed to load ${src}`))
    document.head.appendChild(s)
  })

// The driver pair loads once per page: emoteplayer.js declares the
// classes, FreeMoteDriver.js provides the emscripten runtime they use.
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
  const canvas = container.value?.querySelector('canvas')
  if (!canvas) return
  const scale = Math.min(canvas.width / modelWidth, canvas.height / modelHeight) * 0.95
  const centerX = (bounds.left + bounds.right) / 2
  const centerY = (bounds.top + bounds.bottom) / 2
  player.setScale(scale, 0)
  player.setCoord(-centerX * scale, -centerY * scale, 0)
}

const refreshCount = async () => {
  const name = counterName.value.trim() || 'demo'
  try {
    const d = await $fetch<{ name: string; num: number }>(
      `${apiBase}/api/count/@${encodeURIComponent(name)}`,
    )
    countText.value = textTpl.value
      ? textTpl.value.split('{n}').join(String(d.num))
      : String(d.num)
  } catch {
    countText.value = ''
  }
}

const loadModel = async () => {
  const model = selectedModel.value
  if (!model || loading.value) return
  loading.value = true
  errorMsg.value = ''
  try {
    await ensureDriver()
    // The driver declares `class EmotePlayer` at the top level of a
    // classic script — a global LEXICAL binding, not a window property.
    // A bare identifier reference resolves it; window.EmotePlayer does not.
    if (typeof EmotePlayer === 'undefined') throw new Error('emote driver failed to initialize')
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
    motionLabels.value = (player.mainTimelineLabels || []).slice()
    playRandomMotion()
    await refreshCount()
  } catch (e: any) {
    errorMsg.value = e?.message || String(e)
  }
  loading.value = false
}

// The snippet points at the public origin (BASE_URL) like the SVG embed
// links, falling back to the same-origin apiBase in dev.
const snippetOrigin = computed(() => (publicBase.value || apiBase || '').replace(/\/+$/, ''))

const widgetSnippet = computed(() => {
  const name = counterName.value.trim() || 'my-counter'
  const model = selectedModel.value || 'model-name'
  const tpl = textTpl.value && textTpl.value !== '{n}'
    ? `\n  data-text="${textTpl.value}"`
    : ''
  return `<div data-lolicount="${name}"\n  data-model="${model}"${tpl}>\n</div>\n<script src="${snippetOrigin.value}/widget/widget.js" defer><\/script>`
})

const copied = ref(false)
const copySnippet = async () => {
  await navigator.clipboard.writeText(widgetSnippet.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}

onMounted(async () => {
  await fetchConfig()
  try {
    const d = await $fetch<{ models: string[] }>(`${apiBase}/api/psb/models`)
    models.value = d.models ?? []
    selectedModel.value = models.value[0] ?? ''
    if (selectedModel.value) await loadModel()
  } catch {
    errorMsg.value = t('emote.loadFailed')
  }
})

watch(selectedModel, () => {
  motionLabels.value = []
  currentMotion.value = ''
  loadModel()
})
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">
    <section class="mb-8 text-center">
      <h1 class="text-4xl font-bold text-loli-pink mb-3">{{ t('emote.title') }}</h1>
      <p class="text-gray-600 text-sm">{{ t('emote.desc') }}</p>
    </section>

    <!-- No models: show placement guidance instead of a broken preview. -->
    <div v-if="!loading && !models.length" class="rounded-xl bg-loli-cream p-6 text-center">
      <p class="font-medium text-gray-700 mb-2">{{ t('emote.noModels') }}</p>
      <p class="text-xs text-gray-500">{{ t('emote.noModelsHint') }}</p>
    </div>

    <template v-else>
      <div class="grid md:grid-cols-[220px_1fr] gap-8 items-start">
        <!-- Controls -->
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">{{ t('emote.model') }}</label>
            <select
              v-model="selectedModel"
              class="w-full border rounded-lg px-3 py-2 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-loli-pink/40 focus:border-loli-pink cursor-pointer transition"
            >
              <option v-for="m in models" :key="m" :value="m">{{ m }}</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">{{ t('emote.counterName') }}</label>
            <input
              v-model="counterName"
              type="text"
              :placeholder="t('param.namePlaceholder')"
              class="w-full border rounded-lg px-3 py-2 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-loli-pink/40 focus:border-loli-pink transition"
              @change="refreshCount"
            />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">{{ t('emote.textTemplate') }}</label>
            <input
              v-model="textTpl"
              type="text"
              :placeholder="t('param.textPlaceholder')"
              class="w-full border rounded-lg px-3 py-2 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-loli-pink/40 focus:border-loli-pink transition"
              @change="refreshCount"
            />
          </div>

          <button
            :disabled="!motionLabels.length"
            :class="cn(
              'w-full py-2 rounded-lg font-medium transition',
              motionLabels.length
                ? 'bg-loli-pink text-white hover:bg-loli-pink/90'
                : 'bg-gray-200 text-gray-400 cursor-not-allowed'
            )"
            @click="playRandomMotion"
          >
            🎲 {{ t('emote.randomMotion') }}
          </button>
        </div>

        <!-- Live preview -->
        <div class="rounded-xl bg-loli-cream p-4 flex flex-col items-center">
          <div ref="container" class="flex justify-center"></div>
          <p v-if="countText" class="text-lg font-semibold mt-2">{{ countText }}</p>
          <p v-if="currentMotion" class="text-xs text-gray-500 mt-1">
            {{ t('emote.currentMotion') }}: {{ currentMotion }}
          </p>
          <p v-if="motionLabels.length" class="text-xs text-gray-400 mt-0.5">
            {{ t('emote.motions') }}: {{ motionLabels.length }}
          </p>
          <p v-if="loading" class="text-sm text-gray-400 mt-4">{{ t('loli.loading') }}</p>
          <p v-if="errorMsg" class="text-sm text-red-600 mt-4">{{ errorMsg }}</p>
        </div>
      </div>

      <!-- Copyable widget snippet -->
      <section class="mt-10">
        <h2 class="text-2xl font-semibold mb-4">{{ t('emote.snippetTitle') }}</h2>
        <p class="text-sm text-gray-600 mb-3">{{ t('emote.snippetDesc') }}</p>
        <div class="flex items-start gap-2">
          <pre class="flex-1 bg-gray-50 p-3 rounded text-xs overflow-x-auto">{{ widgetSnippet }}</pre>
          <button
            :class="cn(
              'px-3 py-1.5 text-xs rounded transition shrink-0',
              copied ? 'bg-green-500 text-white' : 'bg-loli-pink text-white hover:bg-loli-pink/90'
            )"
            @click="copySnippet"
          >
            {{ copied ? t('embed.copied') : t('embed.copy') }}
          </button>
        </div>
      </section>
    </template>

    <SiteFooter />
    <BackToTop />
  </main>
</template>
