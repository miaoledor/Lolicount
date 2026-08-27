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

// Debounce preview requests so rapid parameter changes don't flood
// the backend. The editor updates the request object reactively; this
// watch fires on every change but the timeout collapses bursts.
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

// gridSvg generates a standalone SVG showing only the coordinate grid
// and canvas border. Used when showCanvas is on but no preview SVG is
// loaded yet, so the user can see the PSD coordinate space.
const gridSvg = computed(() => {
  const w = props.canvasWidth
  const h = props.canvasHeight
  if (w <= 0 || h <= 0) return ''

  const major = 100
  const lines: string[] = []
  const labels: string[] = []

  for (let x = 0; x <= w; x += major) {
    lines.push(`<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.3" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="${x + 3}" y="14" fill="#ec4899" font-size="11" font-family="monospace">${x}</text>`)
  }
  for (let y = 0; y <= h; y += major) {
    lines.push(`<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.3" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="3" y="${y + 14}" fill="#ec4899" font-size="11" font-family="monospace">${y}</text>`)
  }

  const border = `<rect x="0" y="0" width="${w}" height="${h}" fill="none" stroke="#ec4899" stroke-width="2" stroke-opacity="0.5"/>`
  const body = border + lines.join('') + labels.join('')

  return `<?xml version="1.0" encoding="UTF-8"?>\n<svg viewBox="0 0 ${w} ${h}" width="${w}" height="${h}" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">\n${body}\n</svg>`
})

// gridOverlay is the grid fragment injected into the preview SVG when
// showCanvas is on and a preview is loaded.
const gridOverlay = computed(() => {
  const w = props.canvasWidth
  const h = props.canvasHeight
  if (w <= 0 || h <= 0) return ''

  const major = 100
  const lines: string[] = []
  const labels: string[] = []

  for (let x = 0; x <= w; x += major) {
    lines.push(`<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.3" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="${x + 3}" y="14" fill="#ec4899" font-size="11" font-family="monospace">${x}</text>`)
  }
  for (let y = 0; y <= h; y += major) {
    lines.push(`<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="#ec4899" stroke-width="1" stroke-opacity="0.3" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="3" y="${y + 14}" fill="#ec4899" font-size="11" font-family="monospace">${y}</text>`)
  }

  const border = `<rect x="0" y="0" width="${w}" height="${h}" fill="none" stroke="#ec4899" stroke-width="2" stroke-opacity="0.5"/>`
  return border + lines.join('') + labels.join('')
})

// displaySvg merges the preview SVG with the grid overlay when
// showCanvas is on. When no preview is loaded but showCanvas is on,
// gridSvg provides a standalone coordinate grid.
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
  <div class="editor-canvas">
    <div class="editor-canvas-dims">{{ canvasWidth }} × {{ canvasHeight }}</div>
    <div v-if="loading" class="editor-canvas-loading">...</div>
    <div v-else-if="error" class="editor-canvas-error">{{ error }}</div>
    <div v-else-if="displaySvg" class="editor-canvas-svg" v-html="displaySvg" />
    <div v-else class="editor-canvas-placeholder">
      <p>{{ t('editor.previewHint') }}</p>
    </div>
  </div>
</template>

<style scoped>
.editor-canvas {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
  min-height: 400px;
  padding: 1rem;
}

.editor-canvas-dims {
  font-size: 0.75rem;
  font-family: monospace;
  color: #ec4899;
  margin-bottom: 0.5rem;
}

.editor-canvas-svg {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.editor-canvas-svg :deep(svg) {
  max-width: 100%;
  height: auto;
}

.editor-canvas-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  width: 100%;
  border: 2px dashed var(--border-color, #333);
  border-radius: 8px;
  color: var(--text-muted, #999);
}

.editor-canvas-loading {
  color: var(--text-muted, #999);
  padding: 2rem;
}

.editor-canvas-error {
  color: #ef4444;
  padding: 2rem;
  font-size: 0.875rem;
}
</style>
