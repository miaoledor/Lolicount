<script setup lang="ts">
import type { EditorRequest } from '~/composables/useEditorApi'

const props = defineProps<{
  request: EditorRequest
  hasLayers: boolean
  canvasWidth: number
  canvasHeight: number
}>()

const { previewTheme } = useEditorApi()
const { t } = useI18n()

const svg = ref('')
const loading = ref(false)
const error = ref('')
const showCanvas = ref(true)

let timer: ReturnType<typeof setTimeout> | null = null

watch(
  () => props.request,
  () => {
    if (!props.hasLayers) {
      svg.value = ''
      error.value = ''
      return
    }
    if (timer) clearTimeout(timer)
    timer = setTimeout(doPreview, 300)
  },
  { deep: true },
)

const doPreview = async () => {
  if (!props.hasLayers) return
  loading.value = true
  error.value = ''
  try {
    svg.value = await previewTheme(props.request)
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || 'preview failed'
  } finally {
    loading.value = false
  }
}

const buildGrid = (w: number, h: number): string => {
  if (w <= 0 || h <= 0) return ''
  const major = 100
  const lines: string[] = []
  const labels: string[] = []
  for (let x = 0; x <= w; x += major) {
    lines.push(`<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.25" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="${x + 3}" y="14" fill="#ec4899" font-size="11" font-family="monospace">${x}</text>`)
  }
  for (let y = 0; y <= h; y += major) {
    lines.push(`<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.25" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="3" y="${y + 14}" fill="#ec4899" font-size="11" font-family="monospace">${y}</text>`)
  }
  const border = `<rect x="0" y="0" width="${w}" height="${h}" fill="none" stroke="#ec4899" stroke-width="2" stroke-opacity="0.4"/>`
  return border + lines.join('') + labels.join('')
}

const gridSvg = computed(() => {
  const w = props.canvasWidth
  const h = props.canvasHeight
  if (w <= 0 || h <= 0) return ''
  const body = buildGrid(w, h)
  return `<?xml version="1.0" encoding="UTF-8"?>\n<svg viewBox="0 0 ${w} ${h}" width="${w}" height="${h}" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">\n${body}\n</svg>`
})

const gridOverlay = computed(() => buildGrid(props.canvasWidth, props.canvasHeight))

const displaySvg = computed(() => {
  if (!showCanvas.value) return svg.value
  if (svg.value) {
    return svg.value.replace('</svg>', gridOverlay.value + '</svg>')
  }
  return gridSvg.value
})

onMounted(() => {
  if (props.hasLayers) doPreview()
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <div class="canvas-wrapper">
    <!-- Canvas toolbar -->
    <div class="canvas-toolbar">
      <label class="canvas-toggle">
        <input v-model="showCanvas" type="checkbox">
        <span>{{ t('editor.showCanvas') }}</span>
      </label>
      <span class="canvas-dims">{{ canvasWidth }} × {{ canvasHeight }} px</span>
    </div>

    <!-- Canvas viewport with checkered background -->
    <div class="canvas-viewport">
      <div v-if="loading" class="canvas-loading">
        <span class="canvas-spinner" />
      </div>
      <div v-else-if="error" class="canvas-error">{{ error }}</div>
      <div v-else-if="displaySvg" class="canvas-svg-container" v-html="displaySvg" />
      <div v-else class="canvas-empty">
        <p>{{ t('editor.previewHint') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.canvas-wrapper {
  display: flex;
  flex-direction: column;
  flex: 1;
  height: 100%;
}

/* Mini toolbar above canvas */
.canvas-toolbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.375rem 0.75rem;
  background: var(--bg-card, #1e1e1e);
  border-bottom: 1px solid var(--border-color, #333);
  flex-shrink: 0;
}

.canvas-toggle {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--text-muted, #999);
  user-select: none;
}

.canvas-toggle input {
  cursor: pointer;
}

.canvas-dims {
  font-size: 0.6875rem;
  font-family: monospace;
  color: #ec4899;
}

/* Checkered background viewport */
.canvas-viewport {
  flex: 1;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background-color: #2a2a2a;
  background-image:
    linear-gradient(45deg, #333 25%, transparent 25%),
    linear-gradient(-45deg, #333 25%, transparent 25%),
    linear-gradient(45deg, transparent 75%, #333 75%),
    linear-gradient(-45deg, transparent 75%, #333 75%);
  background-size: 20px 20px;
  background-position: 0 0, 0 10px, 10px -10px, -10px 0;
}

.canvas-svg-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.canvas-svg-container :deep(svg) {
  max-width: 100%;
  max-height: 100%;
  height: auto;
  background: #fff;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.4);
}

.canvas-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted, #999);
  font-size: 0.875rem;
}

.canvas-empty p {
  margin: 0;
  padding: 2rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.canvas-loading {
  display: flex;
  align-items: center;
  justify-content: center;
}

.canvas-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid var(--border-color, #444);
  border-top-color: var(--accent, #ec4899);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.canvas-error {
  color: #ef4444;
  font-size: 0.8125rem;
  padding: 1rem;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 6px;
}
</style>
