<script setup lang="ts">
import type { EditorRequest, EditorLayer } from '~/composables/useEditorApi'

const props = defineProps<{
  request: EditorRequest
  hasLayers: boolean
  canvasWidth: number
  canvasHeight: number
  layers: EditorLayer[]
  selectedLayerId: number | null
  selectedImageIndex: Record<number, number>
}>()

const emit = defineEmits<{
  updateImage: [layerId: number, index: number, patch: Partial<{ left: number; top: number }>]
  selectLayer: [id: number]
}>()

const { previewTheme } = useEditorApi()
const { t } = useI18n()

const svg = ref('')
const loading = ref(false)
const error = ref('')
const showCanvas = ref(true)
const zoom = ref(100)

let timer: ReturnType<typeof setTimeout> | null = null
// While dragging, suppress the debounced backend preview — the drag
// box (with its real <img>) already follows the cursor in real time.
// The actual SVG preview is refreshed on pointer up.
const dragging = ref(false)

watch(
  () => props.request,
  () => {
    if (!props.hasLayers) {
      svg.value = ''
      error.value = ''
      return
    }
    if (dragging.value) return
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
  let s = svg.value
  if (!showCanvas.value) return s
  if (s) {
    // Inject overflow="visible" so images dragged outside the canvas
    // bounds remain visible instead of being clipped by the viewBox.
    s = s.replace(/<svg\b(?=\s)/, '<svg overflow="visible"')
    return s
  }
  return gridSvg.value
})

const zoomStyle = computed(() => ({
  transform: `scale(${zoom.value / 100})`,
}))

const zoomIn = () => { zoom.value = Math.min(zoom.value + 25, 300) }
const zoomOut = () => { zoom.value = Math.max(zoom.value - 25, 25) }
const zoomReset = () => { zoom.value = 100 }

// --- Interactive image drag overlay ---
// The overlay sits on top of the rendered SVG. We measure the SVG's
// actual on-screen size and map canvas coordinates to screen pixels
// so each layer's selected image gets a draggable bounding box.

const svgContainerRef = ref<HTMLElement | null>(null)
const dragOverlayRef = ref<HTMLElement | null>(null)
const renderedSize = ref({ w: 0, h: 0 })

const measureSvg = () => {
  const el = svgContainerRef.value
  if (!el) return
  const svg = el.querySelector('svg')
  if (!svg) return
  // offsetWidth/Height give the layout size (unaffected by the zoom
  // transform), used for the overlay's base dimensions. The overlay
  // lives inside the zoomed wrapper so it is scaled by CSS transform.
  const baseW = svg.offsetWidth || svg.getBoundingClientRect().width
  const baseH = svg.offsetHeight || svg.getBoundingClientRect().height
  renderedSize.value = { w: baseW, h: baseH }
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  measureSvg()
  resizeObserver = new ResizeObserver(measureSvg)
  if (svgContainerRef.value) resizeObserver.observe(svgContainerRef.value)
})
onBeforeUnmount(() => {
  resizeObserver?.disconnect()
})

watch([svg, showCanvas], () => nextTick(measureSvg))

const scaleX = computed(() => {
  if (renderedSize.value.w <= 0 || props.canvasWidth <= 0) return 1
  return renderedSize.value.w / props.canvasWidth
})
const scaleY = computed(() => {
  if (renderedSize.value.h <= 0 || props.canvasHeight <= 0) return 1
  return renderedSize.value.h / props.canvasHeight
})

// One draggable box per non-empty layer (selected image only).
type DragBox = {
  layerId: number
  layerName: string
  imgIndex: number
  src: string
  left: number
  top: number
  width: number
  height: number
}

const dragBoxes = computed<DragBox[]>(() => {
  const boxes: DragBox[] = []
  for (const layer of props.layers) {
    if (layer.fixed || layer.images.length === 0) continue
    // The request computed already filters to the selected image per
    // layer, so reqLayer.images[0] is the currently displayed image.
    const reqLayer = props.request.layers.find((l) => l.id === layer.id)
    if (!reqLayer || reqLayer.images.length === 0) continue
    const img = reqLayer.images[0]
    // The index of the selected image in the original layer.images
    // array — needed so updateImage patches the correct image.
    const selectedIdx = Math.min(
      props.selectedImageIndex[layer.id] ?? 0,
      layer.images.length - 1,
    )
    boxes.push({
      layerId: layer.id,
      layerName: layer.name,
      imgIndex: selectedIdx,
      src: img.src,
      left: img.left,
      top: img.top,
      width: img.width,
      height: img.height,
    })
  }
  return boxes
})

const dragState = ref<{
  layerId: number
  imgIndex: number
  startX: number
  startY: number
  origLeft: number
  origTop: number
} | null>(null)

const onBoxPointerDown = (box: DragBox, e: PointerEvent) => {
  e.preventDefault()
  e.stopPropagation()
  emit('selectLayer', box.layerId)
  dragging.value = true
  dragState.value = {
    layerId: box.layerId,
    imgIndex: box.imgIndex,
    startX: e.clientX,
    startY: e.clientY,
    origLeft: box.left,
    origTop: box.top,
  }
  // Capture the pointer on the overlay element (not the drag-box)
  // so pointermove/up keep firing on the same element even as the
  // drag-box moves under the cursor. Without this, mobile touch
  // events get hijacked by scrolling or hit another element mid-drag.
  if (dragOverlayRef.value) {
    dragOverlayRef.value.setPointerCapture(e.pointerId)
  }
}

const onBoxPointerMove = (e: PointerEvent) => {
  if (!dragState.value) return
  // Screen pixels -> canvas coordinates. The overlay is inside the
  // zoomed wrapper, so screen pixels are scaled by both the base
  // layout scale (scaleX/Y) and the zoom factor.
  const zoomFactor = zoom.value / 100
  const dx = (e.clientX - dragState.value.startX) / (scaleX.value * zoomFactor)
  const dy = (e.clientY - dragState.value.startY) / (scaleY.value * zoomFactor)
  const newLeft = Math.round(dragState.value.origLeft + dx)
  const newTop = Math.round(dragState.value.origTop + dy)
  // Update the data model — the drag-box (with its <img>) follows in
  // real time via reactivity, no backend round-trip needed.
  emit('updateImage', dragState.value.layerId, dragState.value.imgIndex, {
    left: newLeft,
    top: newTop,
  })
}

const onBoxPointerUp = (e: PointerEvent) => {
  if (dragState.value && dragOverlayRef.value) {
    dragOverlayRef.value.releasePointerCapture?.(e.pointerId)
  }
  dragState.value = null
  // Drag ended: refresh the backend preview immediately so the image
  // snaps to its new position without waiting for the debounce.
  if (dragging.value) {
    dragging.value = false
    if (timer) clearTimeout(timer)
    doPreview()
  }
}

onMounted(() => {
  if (props.hasLayers) doPreview()
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <div class="canvas-wrapper">
    <!-- Canvas dimensions label -->
    <div class="canvas-toolbar">
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
          <div
            ref="svgContainerRef"
            class="canvas-svg-container"
            :class="{ 'canvas-svg-dragging': dragging }"
            v-html="displaySvg"
          />

          <!-- Interactive drag overlay: lives inside the zoomed wrapper
               so it scales together with the SVG at any zoom level. -->
          <div
            v-if="displaySvg && dragBoxes.length > 0"
            ref="dragOverlayRef"
          class="canvas-drag-overlay"
            :style="{
              width: renderedSize.w + 'px',
              height: renderedSize.h + 'px',
            }"
            @pointermove="onBoxPointerMove"
            @pointerup="onBoxPointerUp"
          >
            <div
              v-for="box in dragBoxes"
              :key="box.layerId"
              class="drag-box"
              :class="{
                'drag-box-selected': box.layerId === selectedLayerId,
                'drag-box-dragging': dragging && dragState?.layerId === box.layerId,
              }"
              :style="{
                left: box.left * scaleX + 'px',
                top: box.top * scaleY + 'px',
                width: box.width * scaleX + 'px',
                height: box.height * scaleY + 'px',
              }"
              @pointerdown="onBoxPointerDown(box, $event)"
            >
              <img
                :src="box.src"
                class="drag-box-img"
                alt=""
                draggable="false"
              >
              <span class="drag-box-label">{{ box.layerName }}</span>
            </div>
          </div>

          <!-- Out-of-canvas dimmer: a canvas-sized box with a huge
               semi-transparent box-shadow spread, so anything outside
               the canvas bounds appears dimmed. Sits above drag boxes
               but below the grid overlay. -->
          <div
            v-if="displaySvg && dragBoxes.length > 0"
            class="canvas-outer-dim"
            :style="{
              width: renderedSize.w + 'px',
              height: renderedSize.h + 'px',
            }"
          />

          <!-- Grid overlay: always on top so the canvas bounds and
               guide lines remain visible even while dragging. -->
          <div
            v-if="displaySvg"
            class="canvas-grid-overlay"
            :style="{
              width: renderedSize.w + 'px',
              height: renderedSize.h + 'px',
            }"
            v-html="gridSvg"
          />
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
        <button
          class="zoom-btn zoom-eye"
          :class="{ 'zoom-eye-off': !showCanvas }"
          :title="t('editor.showCanvas')"
          @click="showCanvas = !showCanvas"
        >
          <svg v-if="showCanvas" width="14" height="14" viewBox="0 0 14 14">
            <path d="M1 7 C3 4, 5 3, 7 3 C9 3, 11 4, 13 7 C11 10, 9 11, 7 11 C5 11, 3 10, 1 7 Z" fill="none" stroke="currentColor" stroke-width="1.2"/>
            <circle cx="7" cy="7" r="2" fill="none" stroke="currentColor" stroke-width="1.2"/>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 14 14">
            <path d="M1 7 C3 4, 5 3, 7 3 C9 3, 11 4, 13 7 C11 10, 9 11, 7 11 C5 11, 3 10, 1 7 Z" fill="none" stroke="currentColor" stroke-width="1.2"/>
            <line x1="2" y1="2" x2="12" y2="12" stroke="currentColor" stroke-width="1.2"/>
          </svg>
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
  background: #fff;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.15);
}

.canvas-svg-container :deep(svg) {
  display: block;
  max-width: 60vw;
  max-height: 70vh;
  width: auto;
  height: auto;
}

/* While dragging, hide only the SVG content (images) — the white
   canvas background lives on the container so it stays visible. */
.canvas-svg-dragging :deep(svg) {
  opacity: 0;
}

/* Out-of-canvas dimmer: the box itself is canvas-sized and transparent;
   the huge box-shadow spread covers everything outside the canvas with
   a semi-transparent white veil, dimming image parts that exceed the
   canvas bounds. */
.canvas-outer-dim {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
  z-index: 15;
  box-shadow: 0 0 0 9999px rgba(255, 255, 255, 0.6);
}

/* Grid overlay: rendered above everything so canvas bounds and guide
   lines stay visible during drag. Does not capture pointer events. */
.canvas-grid-overlay {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
  z-index: 20;
}

.canvas-grid-overlay :deep(svg) {
  display: block;
  width: 100%;
  height: 100%;
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

.zoom-eye {
  border-left: 1px solid var(--border-color, #333);
  margin-left: 2px;
  padding-left: 2px;
}

.zoom-eye:hover {
  color: var(--loli-pink);
}

.zoom-eye-off {
  opacity: 0.5;
}

/* Interactive drag overlay */
.canvas-drag-overlay {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
  z-index: 5;
  overflow: visible;
}

.drag-box {
  position: absolute;
  border: 2px solid transparent;
  border-radius: 2px;
  cursor: move;
  pointer-events: auto;
  box-sizing: border-box;
  transition: border-color 0.12s;
  touch-action: none;
  overflow: visible;
}

.drag-box:hover {
  border-color: rgba(107, 114, 128, 0.6);
}

.drag-box-selected {
  border-color: var(--loli-pink) !important;
}

.drag-box-dragging {
  z-index: 10;
}

.drag-box-dragging .drag-box-img {
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.2));
}

.drag-box-img {
  width: 100%;
  height: 100%;
  display: block;
  pointer-events: none;
  user-select: none;
  -webkit-user-drag: none;
}

.drag-box-label {
  position: absolute;
  top: -16px;
  left: 0;
  font-size: 0.625rem;
  font-family: monospace;
  color: var(--loli-pink);
  background: rgba(255, 255, 255, 0.85);
  padding: 0 4px;
  border-radius: 2px;
  white-space: nowrap;
  pointer-events: none;
}

@media (max-width: 768px) {
  .canvas-viewport {
    padding: 0.25rem;
  }

  /* Ensure at least one full canvas is visible on mobile. Use nearly
     full viewport width/height since sidebars are hidden in quick mode. */
  .canvas-svg-container :deep(svg) {
    max-width: 95vw;
    max-height: 60vh;
  }
}
</style>
