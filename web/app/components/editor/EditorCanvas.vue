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
const zoom = ref(100)

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
    lines.push(`<line x1="${x}" y1="0" x2="${x}" y2="${h}" stroke="var(--loli-pink)" stroke-width="1" stroke-opacity="0.25" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="${x + 3}" y="14" fill="var(--loli-pink)" font-size="11" font-family="monospace">${x}</text>`)
  }
  for (let y = 0; y <= h; y += major) {
    lines.push(`<line x1="0" y1="${y}" x2="${w}" y2="${y}" stroke="var(--loli-pink)" stroke-width="1" stroke-opacity="0.25" stroke-dasharray="4 4"/>`)
    labels.push(`<text x="3" y="${y + 14}" fill="var(--loli-pink)" font-size="11" font-family="monospace">${y}</text>`)
  }
  const border = `<rect x="0" y="0" width="${w}" height="${h}" fill="none" stroke="var(--loli-pink)" stroke-width="2" stroke-opacity="0.4"/>`
  return border + lines.join('') + labels.join('')
}

const gridSvg = computed(() => {
  const w = props.canvasWidth
  const h = props.canvasHeight
  if (w <= 0 || h <= 0) return ''
  const body = buildGrid(w, h)
  return `<?xml version="1.0" encoding="UTF-8"?>\n<svg viewBox="0 0 ${w} ${h}" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">\n${body}\n</svg>`
})

const gridOverlay = computed(() => buildGrid(props.canvasWidth, props.canvasHeight))

const displaySvg = computed(() => {
  if (!showCanvas.value) return svg.value
  if (svg.value) {
    return svg.value.replace('</svg>', gridOverlay.value + '</svg>')
  }
  return gridSvg.value
})

const zoomStyle = computed(() => ({
  transform: `scale(${zoom.value / 100})`,
}))

const zoomIn = () => { zoom.value = Math.min(zoom.value + 25, 300) }
const zoomOut = () => { zoom.value = Math.max(zoom.value - 25, 25) }
const zoomReset = () => { zoom.value = 100 }

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
      <div class="canvas-stage">
        <div v-if="loading" class="canvas-loading">
          <span class="canvas-spinner" />
        </div>
        <div v-else-if="error" class="canvas-error">{{ error }}</div>
        <div
          v-else-if="displaySvg"
          class="canvas-svg-wrapper"
          :style="zoomStyle"
        >
          <div class="canvas-inside-shadow" />
          <div class="canvas-svg-container" v-html="displaySvg" />
        </div>
        <div v-else class="canvas-empty">
          <p>{{ t('editor.previewHint') }}</p>
        </div>
      </div>

      <!-- Zoom controls (bottom-right, like vue-fabric-editor) -->
      <div class="canvas-zoom">
        <button class="zoom-btn" title="-" @click="zoomOut">
          <svg width="14" height="14" viewBox="0 0 14 14"><path d="M3 7 L11 7" stroke="currentColor" stroke-width="2"/></svg>
        </button>
        <span class="zoom-label">{{ zoom }}%</span>
        <button class="zoom-btn" title="+" @click="zoomIn">
          <svg width="14" height="14" viewBox="0 0 14 14"><path d="M3 7 L11 7 M7 3 L7 11" stroke="currentColor" stroke-width="2"/></svg>
        </button>
        <button class="zoom-btn zoom-reset" title="1:1" @click="zoomReset">
          <svg width="14" height="14" viewBox="0 0 14 14"><path d="M3 3 L11 3 L11 11 L3 11 Z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
        </button>
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

.canvas-toggle input { cursor: pointer; }

.canvas-dims {
  font-size: 0.6875rem;
  font-family: monospace;
  color: var(--loli-pink);
}

.canvas-viewport {
  flex: 1;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  position: relative;
  background-color: #e8e8e8;
  background-image:
    linear-gradient(45deg, #d0d0d0 25%, transparent 25%),
    linear-gradient(-45deg, #d0d0d0 25%, transparent 25%),
    linear-gradient(45deg, transparent 75%, #d0d0d0 75%),
    linear-gradient(-45deg, transparent 75%, #d0d0d0 75%);
  background-size: 20px 20px;
  background-position: 0 0, 0 10px, 10px -10px, -10px 0;
}

.canvas-stage {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.canvas-svg-wrapper {
  position: relative;
  display: inline-flex;
  transform-origin: center center;
}

.canvas-inside-shadow {
  position: absolute;
  inset: 0;
  box-shadow: inset 0 0 12px 3px rgba(0, 0, 0, 0.12);
  z-index: 2;
  pointer-events: none;
}

.canvas-svg-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.canvas-svg-container :deep(svg) {
  display: block;
  max-width: 60vw;
  max-height: 70vh;
  width: auto;
  height: auto;
  background: #fff;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.15);
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
  border: 3px solid var(--border-color);
  border-top-color: var(--loli-pink);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.canvas-error {
  color: #ef4444;
  font-size: 0.8125rem;
  padding: 1rem;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 6px;
}

/* Zoom controls */
.canvas-zoom {
  position: absolute;
  right: 12px;
  bottom: 12px;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px;
  background: var(--bg-card, #1e1e1e);
  border: 1px solid var(--border-color, #333);
  border-radius: 6px;
  z-index: 10;
}

.zoom-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-color, #eee);
  cursor: pointer;
  padding: 0;
}

.zoom-btn:hover {
  background: var(--bg-btn, #2a2a2a);
  color: var(--loli-pink);
}

.zoom-label {
  font-size: 0.6875rem;
  font-family: monospace;
  color: var(--text-muted, #999);
  min-width: 36px;
  text-align: center;
}

.zoom-reset {
  border-left: 1px solid var(--border-color, #333);
  margin-left: 2px;
  padding-left: 2px;
}
</style>
