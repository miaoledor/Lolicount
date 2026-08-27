<script setup lang="ts">
import type { EditorRequest, EditorLayer, EditorImage } from '~/composables/useEditorApi'
import EditorCanvas from '~/components/editor/EditorCanvas.vue'
import LayerPanel from '~/components/editor/LayerPanel.vue'
import TextLayerControls from '~/components/editor/TextLayerControls.vue'
import LayerImageControls from '~/components/editor/LayerImageControls.vue'
import { useEditorStorage } from '~/composables/useEditorStorage'

const { exportTheme, previewTheme } = useEditorApi()
const { saveDraft, saveDraftAs, loadDraft, listDrafts, deleteDraft } = useEditorStorage()
const { t, locale, locales, setLocale } = useI18n()

const themeName = ref('')
const canvasWidth = ref(500)
const canvasHeight = ref(800)
const displaySize = ref(400)
const layers = ref<EditorLayer[]>([])
const layerIdCounter = ref(0)

const counterText = ref('0123456789')
const fontSize = ref(50)
const scale = ref(0)
const unshowFont = ref(true)

const exportLoading = ref(false)
const exportImageLoading = ref(false)
const exportMenuOpen = ref(false)
const selectedLayerId = ref<number | null>(null)

// Per-layer selected image index for preview display. When a layer has
// multiple images, only the selected one is sent to the preview endpoint
// (export still sends all images as Candidates). Keyed by layer id.
const selectedImageIndex = ref<Record<number, number>>({})
const leftSidebarOpen = ref(true)
const rightSidebarOpen = ref(true)
const errorMsg = ref('')
const savedDrafts = ref<string[]>([])
const autoSaveEnabled = ref(true)

// Editor mode: 'workbench' = full layer editor, 'quick' = simplified
// single-layer card theme editor with auto canvas sizing.
const editorMode = ref<'workbench' | 'quick'>('quick')

const switchMode = (mode: 'workbench' | 'quick') => {
  if (editorMode.value === mode) return
  editorMode.value = mode
  if (mode === 'workbench') {
    // Entering workbench: show both sidebars (unless mobile).
    const isMobile = window.innerWidth <= 768
    leftSidebarOpen.value = !isMobile
    rightSidebarOpen.value = !isMobile
  } else {
    // Quick mode: hide sidebars to give canvas full space.
    leftSidebarOpen.value = false
    rightSidebarOpen.value = false
  }
}

const nonTextLayers = computed(() => layers.value.filter((l) => !l.fixed))
const isCardTheme = computed(() => nonTextLayers.value.length <= 1)
const selectedLayer = computed(() => layers.value.find((l) => l.id === selectedLayerId.value) || null)

const collectState = () => ({
  themeName: themeName.value,
  canvasWidth: canvasWidth.value,
  canvasHeight: canvasHeight.value,
  displaySize: displaySize.value,
  layers: layers.value.map((l) => {
    if (l.images.length <= 1) return l
    const idx = Math.min(getSelectedImageIndex(l.id), l.images.length - 1)
    return { ...l, images: [l.images[idx]] }
  }),
  layerIdCounter: layerIdCounter.value,
  counterText: counterText.value,
  fontSize: fontSize.value,
  scale: scale.value,
  unshowFont: unshowFont.value,
})

let saveTimer: ReturnType<typeof setTimeout> | null = null
const autoSave = () => {
  if (!autoSaveEnabled.value) return
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    saveDraft(collectState())
    savedDrafts.value = listDrafts()
  }, 500)
}

watch([themeName, canvasWidth, canvasHeight, displaySize, layers, counterText, fontSize, scale, unshowFont], autoSave, { deep: true })

const restoreDraft = (name: string) => {
  const state = loadDraft(name)
  if (!state) return
  autoSaveEnabled.value = false
  themeName.value = state.themeName
  canvasWidth.value = state.canvasWidth
  canvasHeight.value = state.canvasHeight
  displaySize.value = state.displaySize
  layers.value = state.layers
  layerIdCounter.value = state.layerIdCounter
  counterText.value = state.counterText
  fontSize.value = state.fontSize
  scale.value = state.scale
  unshowFont.value = state.unshowFont
  nextTick(() => { autoSaveEnabled.value = true })
}

const removeDraft = (name: string) => {
  deleteDraft(name)
  savedDrafts.value = listDrafts()
}

onMounted(() => {
  savedDrafts.value = listDrafts()
  const drafts = listDrafts()
  if (drafts.length > 0) restoreDraft(drafts[drafts.length - 1])
})

const getSelectedImageIndex = (layerId: number): number => {
  const idx = selectedImageIndex.value[layerId]
  return idx === undefined ? 0 : idx
}

const setSelectedImageIndex = (layerId: number, index: number) => {
  selectedImageIndex.value = { ...selectedImageIndex.value, [layerId]: index }
}

const request = computed<EditorRequest>(() => ({
  name: themeName.value || 'untitled',
  canvas: { width: canvasWidth.value, height: canvasHeight.value },
  display: displaySize.value > 0 ? { size: displaySize.value } : null,
  layers: layers.value.map((l) => {
    if (l.images.length <= 1) return l
    const idx = Math.min(getSelectedImageIndex(l.id), l.images.length - 1)
    return { ...l, images: [l.images[idx]] }
  }),
  text: counterText.value,
  fsize: fontSize.value,
  scale: scale.value,
  unshowf: unshowFont.value,
}))

// Full state without image filtering — used for export so all images
// per layer are preserved as Candidates in the exported theme.
const exportRequest = computed<EditorRequest>(() => ({
  name: themeName.value || 'untitled',
  canvas: { width: canvasWidth.value, height: canvasHeight.value },
  display: displaySize.value > 0 ? { size: displaySize.value } : null,
  layers: layers.value,
  text: counterText.value,
  fsize: fontSize.value,
  scale: scale.value,
  unshowf: unshowFont.value,
}))

const addLayer = () => {
  layerIdCounter.value++
  layers.value.push({
    id: layerIdCounter.value,
    name: `${t('editor.layer')} ${layerIdCounter.value}`,
    zIndex: layers.value.length,
    fixed: false,
    images: [],
  })
}

const removeLayer = (id: number) => {
  const layer = layers.value.find((l) => l.id === id)
  if (layer?.fixed) return
  layers.value = layers.value.filter((l) => l.id !== id)
}

const moveLayer = (id: number, dir: -1 | 1) => {
  const idx = layers.value.findIndex((l) => l.id === id)
  if (idx < 0) return
  const target = idx + dir
  if (target < 0 || target >= layers.value.length) return
  const arr = [...layers.value]
  const tmp = arr[idx]
  arr[idx] = arr[target]
  arr[target] = tmp
  arr.forEach((l, i) => (l.zIndex = i))
  layers.value = arr
}

const updateLayer = (id: number, patch: Partial<EditorLayer>) => {
  const layer = layers.value.find((l) => l.id === id)
  if (layer) Object.assign(layer, patch)
}

const selectLayer = (id: number) => {
  selectedLayerId.value = selectedLayerId.value === id ? null : id
}

const addImage = (layerId: number, img: EditorImage) => {
  const layer = layers.value.find((l) => l.id === layerId)
  if (layer) layer.images.push(img)
}

const removeImage = (layerId: number, index: number) => {
  const layer = layers.value.find((l) => l.id === layerId)
  if (layer) layer.images.splice(index, 1)
}

const updateImage = (layerId: number, index: number, patch: Partial<EditorImage>) => {
  const layer = layers.value.find((l) => l.id === layerId)
  if (layer) Object.assign(layer.images[index], patch)
}

// --- Quick mode: single layer, auto canvas sizing ---
// In quick mode the user uploads a batch of images. The canvas size is
// auto-calculated from the first image's natural dimensions. All images
// go into one layer as random-frame candidates (card theme model).

const fileToDataURI = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

const getImageNaturalSize = (src: string): Promise<{ w: number; h: number }> => {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve({ w: img.naturalWidth, h: img.naturalHeight })
    img.onerror = () => resolve({ w: 0, h: 0 })
    img.src = src
  })
}

const onQuickUpload = async (e: Event) => {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  const files = Array.from(input.files).filter((f) => f.type.startsWith('image/'))
  if (files.length === 0) return

  const dataUris = await Promise.all(files.map(fileToDataURI))

  // Auto-calculate canvas size from the first image.
  const { w, h } = await getImageNaturalSize(dataUris[0])
  if (w > 0 && h > 0) {
    canvasWidth.value = w
    canvasHeight.value = h
  }

  // Reset to a single layer with all uploaded images, centered at 0,0
  // with full canvas dimensions.
  layers.value = [{
    id: 1,
    name: t('editor.quickLayer'),
    zIndex: 0,
    fixed: false,
    images: dataUris.map((src) => ({
      src,
      left: 0,
      top: 0,
      width: canvasWidth.value,
      height: canvasHeight.value,
    })),
  }]
  layerIdCounter.value = 1
  selectedLayerId.value = 1
  selectedImageIndex.value = { 1: 0 }
  input.value = ''
}

// doSaveDraft prompts for a draft name and saves the current editor
// state as a new manual draft, distinct from the auto-save draft.
const doSaveDraft = () => {
  const name = window.prompt(t('editor.draftNamePrompt'), themeName.value || 'untitled')
  if (!name?.trim()) return
  saveDraftAs(name.trim(), collectState())
  savedDrafts.value = listDrafts()
  themeName.value = name.trim()
}

// doExportImage renders the current preview SVG to a PNG via a canvas
// element and triggers a download. No backend rasterizer is needed —
// the browser does the SVG-to-PNG conversion client-side.
const doExportImage = async () => {
  errorMsg.value = ''
  if (nonTextLayers.value.length === 0) {
    errorMsg.value = t('editor.errNoLayers')
    return
  }
  exportImageLoading.value = true
  try {
    const svgStr = await previewTheme(request.value)
    const blob = new Blob([svgStr], { type: 'image/svg+xml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const img = new Image()
    img.crossOrigin = 'anonymous'
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve()
      img.onerror = () => reject(new Error('SVG render failed'))
      img.src = url
    })
    const canvas = document.createElement('canvas')
    canvas.width = canvasWidth.value
    canvas.height = canvasHeight.value
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('canvas 2d context unavailable')
    ctx.fillStyle = '#fff'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    URL.revokeObjectURL(url)
    canvas.toBlob((pngBlob) => {
      if (!pngBlob) return
      const pngUrl = URL.createObjectURL(pngBlob)
      const a = document.createElement('a')
      a.href = pngUrl
      a.download = `${themeName.value || 'untitled'}.png`
      a.click()
      URL.revokeObjectURL(pngUrl)
    }, 'image/png')
  } catch (e: any) {
    errorMsg.value = e?.data?.message || e?.message || 'image export failed'
  } finally {
    exportImageLoading.value = false
  }
}

const doExport = async () => {
  errorMsg.value = ''
  if (!themeName.value.trim()) {
    errorMsg.value = t('editor.errNoName')
    return
  }
  if (nonTextLayers.value.length === 0) {
    errorMsg.value = t('editor.errNoLayers')
    return
  }
  exportLoading.value = true
  try {
    const blob = await exportTheme(exportRequest.value)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${themeName.value}.zip`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) {
    errorMsg.value = e?.data?.message || e?.message || 'export failed'
  } finally {
    exportLoading.value = false
  }
}
</script>

<template>
  <div class="editor-root">
    <!-- Top toolbar -->
    <div class="editor-toolbar">
      <div class="editor-toolbar-left">
        <h2 class="editor-toolbar-title">{{ t('editor.title') }}</h2>
        <span class="editor-toolbar-badge">{{ isCardTheme ? t('editor.cardTheme') : t('editor.characterTheme') }}</span>
        <div class="editor-mode-switch">
          <button
            class="editor-mode-btn"
            :class="{ 'editor-mode-btn-active': editorMode === 'workbench' }"
            @click="switchMode('workbench')"
          >{{ t('editor.modeWorkbench') }}</button>
          <button
            class="editor-mode-btn"
            :class="{ 'editor-mode-btn-active': editorMode === 'quick' }"
            @click="switchMode('quick')"
          >{{ t('editor.modeQuick') }}</button>
        </div>
      </div>
      <div class="editor-toolbar-right">
        <div class="editor-lang-switch">
          <button
            v-for="l in locales"
            :key="l"
            type="button"
            class="editor-lang-btn"
            :class="{ 'editor-lang-btn-active': l === locale }"
            @click="setLocale(l as string)"
          >{{ l === 'zh' ? '中' : l === 'en' ? 'EN' : 'JP' }}</button>
        </div>
        <input v-model="themeName" type="text" :placeholder="t('editor.namePlaceholder')" class="editor-toolbar-input">
        <button class="editor-btn-upload" disabled :title="t('editor.uploadThemeTodo')">
          {{ t('editor.uploadTheme') }}
        </button>
        <button class="editor-btn-draft" @click="doSaveDraft">
          {{ t('editor.saveDraft') }}
        </button>
        <div v-if="exportMenuOpen" class="editor-export-backdrop" @click="exportMenuOpen = false" />
        <div class="editor-export-dropdown">
          <button
            class="editor-btn-export"
            :disabled="exportLoading || exportImageLoading"
            @click="doExportImage"
          >
            {{ (exportLoading || exportImageLoading) ? '...' : t('editor.exportImage') }}
          </button>
          <button
            class="editor-export-arrow"
            :disabled="exportLoading || exportImageLoading"
            @click="exportMenuOpen = !exportMenuOpen"
          >
            <svg width="10" height="10" viewBox="0 0 10 10"><path d="M2 3 L5 7 L8 3" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
          </button>
          <div v-if="exportMenuOpen" class="editor-export-menu">
            <button class="editor-export-item" @click="exportMenuOpen = false; doExportImage()">
              {{ t('editor.exportImage') }}
            </button>
            <button class="editor-export-item" @click="exportMenuOpen = false; doExport()">
              {{ t('editor.export') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Main layout: transitions between workbench and quick mode -->
    <div class="editor-main">
      <Transition name="mode-fade" mode="out-in">
        <!-- Quick mode: simplified single-panel layout -->
        <div v-if="editorMode === 'quick'" key="quick" class="quick-mode-layout">
          <aside class="quick-panel">
            <h3 class="sidebar-title">{{ t('editor.quickUpload') }}</h3>
            <label class="quick-upload-label">
              <input type="file" accept="image/*" multiple class="img-upload-input" @change="onQuickUpload">
              <span class="quick-upload-btn">{{ t('editor.uploadImage') }}</span>
            </label>
            <p v-if="layers.length > 0 && layers[0].images.length > 0" class="quick-info">
              {{ layers[0].images.length }} {{ t('editor.imgUnit') }} · {{ canvasWidth }} × {{ canvasHeight }} px
            </p>
            <div v-if="layers.length > 0 && layers[0].images.length > 0" class="quick-thumbs">
              <img
                v-for="(img, i) in layers[0].images"
                :key="i"
                :src="img.src"
                class="quick-thumb"
                :class="{ 'quick-thumb-selected': (selectedImageIndex[1] ?? 0) === i }"
                @click="setSelectedImageIndex(1, i)"
              >
            </div>
          </aside>
          <main class="editor-canvas-area">
            <EditorCanvas
              :request="request"
              :has-layers="nonTextLayers.length > 0"
              :canvas-width="canvasWidth"
              :canvas-height="canvasHeight"
              :layers="layers"
              :selected-layer-id="selectedLayerId"
              :selected-image-index="selectedImageIndex"
              @update-image="updateImage"
              @select-layer="selectLayer"
            />
          </main>
        </div>

        <!-- Workbench mode: full three-column layout -->
        <div v-else key="workbench" class="workbench-layout">
      <!-- Left: settings panel -->
      <button class="sidebar-toggle sidebar-toggle-left" @click="leftSidebarOpen = !leftSidebarOpen">
        <svg width="12" height="12" viewBox="0 0 12 12"><path d="M3 1 L9 6 L3 11" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
      </button>
      <aside class="editor-sidebar editor-sidebar-left" :class="{ 'sidebar-collapsed': !leftSidebarOpen }">
        <div class="sidebar-section">
          <h3 class="sidebar-title">{{ t('editor.canvasSettings') }}</h3>
          <div class="sidebar-row">
            <label class="sidebar-label">{{ t('editor.canvasW') }}</label>
            <input v-model.number="canvasWidth" type="number" class="sidebar-input">
          </div>
          <div class="sidebar-row">
            <label class="sidebar-label">{{ t('editor.canvasH') }}</label>
            <input v-model.number="canvasHeight" type="number" class="sidebar-input">
          </div>
          <div class="sidebar-row">
            <label class="sidebar-label">{{ t('editor.displaySize') }}</label>
            <input v-model.number="displaySize" type="number" class="sidebar-input">
          </div>
        </div>

        <div class="sidebar-section">
          <TextLayerControls
            :text="counterText"
            :font-size="fontSize"
            :scale="scale"
            :unshow-font="unshowFont"
            @update:text="counterText = $event"
            @update:font-size="fontSize = $event"
            @update:scale="scale = $event"
            @update:unshow-font="unshowFont = $event"
          />
        </div>

        <div v-if="savedDrafts.length > 0" class="sidebar-section">
          <h3 class="sidebar-title">{{ t('editor.savedDrafts') }}</h3>
          <div class="sidebar-drafts">
            <div v-for="d in savedDrafts" :key="d" class="sidebar-draft-item">
              <button class="sidebar-draft-btn" @click="restoreDraft(d)">{{ d }}</button>
              <button class="sidebar-draft-del" @click="removeDraft(d)">×</button>
            </div>
          </div>
        </div>

        <p v-if="errorMsg" class="sidebar-error">{{ errorMsg }}</p>
      </aside>

      <!-- Center: canvas -->
      <main class="editor-canvas-area">
        <EditorCanvas
          :request="request"
          :has-layers="nonTextLayers.length > 0"
          :canvas-width="canvasWidth"
          :canvas-height="canvasHeight"
          :layers="layers"
          :selected-layer-id="selectedLayerId"
          :selected-image-index="selectedImageIndex"
          @update-image="updateImage"
          @select-layer="selectLayer"
        />
      </main>

      <!-- Right: layers panel -->
      <button class="sidebar-toggle sidebar-toggle-right" @click="rightSidebarOpen = !rightSidebarOpen">
        <svg width="12" height="12" viewBox="0 0 12 12"><path d="M9 1 L3 6 L9 11" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
      </button>
      <aside class="editor-sidebar editor-sidebar-right" :class="{ 'sidebar-collapsed': !rightSidebarOpen }">
        <LayerPanel
          :layers="layers"
          :selected-layer-id="selectedLayerId"
          @add="addLayer()"
          @remove="removeLayer"
          @move="moveLayer"
          @update="updateLayer"
          @select="selectLayer"
        >
          <template #layerContent="{ layer }">
            <LayerImageControls
              :images="layer.images"
              :selected-index="getSelectedImageIndex(layer.id)"
              @select-image="(i) => setSelectedImageIndex(layer.id, i)"
              @add-image="(img) => addImage(layer.id, img)"
              @remove-image="(i) => removeImage(layer.id, i)"
              @update-image="(i, p) => updateImage(layer.id, i, p)"
            />
          </template>
        </LayerPanel>
      </aside>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.editor-root {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 4rem);
  overflow: hidden;
}

/* Top toolbar */
.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  background: var(--bg-card, #1e1e1e);
  border-bottom: 1px solid var(--border-color, #333);
  flex-shrink: 0;
}

.editor-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.editor-toolbar-title {
  font-size: 1rem;
  margin: 0;
  font-weight: 700;
}

.editor-toolbar-badge {
  font-size: 0.6875rem;
  padding: 0.125rem 0.5rem;
  border-radius: 10px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-muted, #999);
}

.editor-toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.editor-toolbar-input {
  width: 180px;
  padding: 0.3rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.8125rem;
}

.editor-btn-export {
  padding: 0.3rem 1rem;
  border: none;
  border-radius: 4px;
  background: var(--loli-pink);
  color: #fff;
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 600;
  white-space: nowrap;
}

.editor-lang-switch {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.editor-lang-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted, #999);
  cursor: pointer;
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0;
  transition: all 0.15s;
}

.editor-lang-btn:hover {
  border-color: var(--loli-pink);
  color: var(--loli-pink);
}

.editor-lang-btn-active {
  background: var(--loli-pink);
  border-color: var(--loli-pink);
  color: #fff;
}

.editor-btn-upload {
  padding: 0.3rem 0.75rem;
  border: 1px dashed var(--border-color, #444);
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted, #999);
  cursor: not-allowed;
  font-size: 0.8125rem;
  font-weight: 600;
  white-space: nowrap;
  opacity: 0.6;
}

.editor-btn-draft {
  padding: 0.3rem 0.75rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: transparent;
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 600;
  white-space: nowrap;
}

.editor-btn-draft:hover {
  border-color: var(--loli-pink);
  color: var(--loli-pink);
}

.editor-btn-export:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-export-dropdown {
  position: relative;
  display: flex;
  align-items: stretch;
}

.editor-export-dropdown .editor-btn-export {
  border-radius: 4px 0 0 4px;
  padding-right: 0.5rem;
}

.editor-export-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  border: none;
  border-radius: 0 4px 4px 0;
  background: var(--loli-pink);
  color: #fff;
  cursor: pointer;
  padding: 0;
}

.editor-export-arrow:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-export-arrow:hover {
  filter: brightness(1.1);
}

.editor-export-backdrop {
  position: fixed;
  inset: 0;
  z-index: 40;
}

.editor-export-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 2px;
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #1e1e1e);
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  overflow: hidden;
  z-index: 50;
  min-width: 120px;
}

.editor-export-item {
  padding: 0.4rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
  text-align: left;
  white-space: nowrap;
}

.editor-export-item:hover {
  background: var(--loli-pink);
  color: #fff;
}

/* Main layout */
.editor-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* Sidebars */
.editor-sidebar {
  width: 280px;
  flex-shrink: 0;
  overflow-y: auto;
  background: var(--bg-card, #1e1e1e);
  border-right: 1px solid var(--border-color, #333);
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.editor-sidebar-right {
  border-right: none;
  border-left: 1px solid var(--border-color, #333);
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sidebar-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #999);
  margin: 0;
  padding-bottom: 0.375rem;
  border-bottom: 1px solid var(--border-color, #333);
}

.sidebar-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sidebar-label {
  font-size: 0.75rem;
  color: var(--text-muted, #999);
  min-width: 4rem;
}

.sidebar-input {
  flex: 1;
  width: 100%;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.8125rem;
}

.sidebar-drafts {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-draft-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  overflow: hidden;
}

.sidebar-draft-btn {
  flex: 1;
  padding: 0.25rem 0.5rem;
  border: none;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
  text-align: left;
}

.sidebar-draft-btn:hover {
  background: var(--loli-pink);
  color: #fff;
}

.sidebar-draft-del {
  padding: 0.25rem 0.5rem;
  border: none;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
}

.sidebar-draft-del:hover {
  color: #ef4444;
}

.sidebar-error {
  color: #ef4444;
  font-size: 0.75rem;
  margin: 0;
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 4px;
}

/* Canvas area */
.editor-canvas-area {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
}

/* Mode switch */
.editor-mode-switch {
  display: flex;
  align-items: center;
  gap: 2px;
  background: var(--bg-btn, #2a2a2a);
  border-radius: 6px;
  padding: 2px;
}

.editor-mode-btn {
  padding: 0.2rem 0.6rem;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted, #999);
  cursor: pointer;
  font-size: 0.6875rem;
  font-weight: 600;
  transition: all 0.2s ease;
}

.editor-mode-btn:hover {
  color: var(--text-color, #eee);
}

.editor-mode-btn-active {
  background: var(--loli-pink);
  color: #fff;
}

/* Mode transition animation */
.mode-fade-enter-active,
.mode-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.mode-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.mode-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Quick mode layout */
.quick-mode-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  width: 100%;
}

.quick-panel {
  width: 220px;
  flex-shrink: 0;
  overflow-y: auto;
  background: var(--bg-card, #1e1e1e);
  border-right: 1px solid var(--border-color, #333);
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.quick-upload-label {
  cursor: pointer;
}

.quick-upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: 1px dashed var(--loli-pink);
  border-radius: 4px;
  background: transparent;
  color: var(--loli-pink);
  font-size: 0.75rem;
  font-weight: 600;
  transition: all 0.15s;
}

.quick-upload-btn:hover {
  background: rgba(107, 114, 128, 0.1);
}

.quick-info {
  font-size: 0.6875rem;
  color: var(--text-muted, #999);
  font-family: monospace;
  margin: 0;
}

.quick-thumbs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.25rem;
}

.quick-thumb {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  border-radius: 3px;
  border: 2px solid transparent;
  cursor: pointer;
  transition: border-color 0.15s;
}

.quick-thumb:hover {
  border-color: var(--text-muted, #999);
}

.quick-thumb-selected {
  border-color: var(--loli-pink);
}

.workbench-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  width: 100%;
}

/* Responsive */
/* Sidebar toggle buttons */
.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  background: var(--bg-card, #1e1e1e);
  border: none;
  border-right: 1px solid var(--border-color, #333);
  color: var(--text-muted, #999);
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.15s;
}

.sidebar-toggle:hover {
  color: var(--loli-pink);
}

.sidebar-toggle-right {
  border-right: none;
  border-left: 1px solid var(--border-color, #333);
}

/* Collapsed sidebar */
.sidebar-collapsed {
  width: 0 !important;
  padding: 0 !important;
  overflow: hidden !important;
  border: none !important;
}

@media (max-width: 1024px) {
  .editor-sidebar {
    width: 240px;
  }
}

@media (max-width: 768px) {
  .editor-toolbar {
    flex-wrap: wrap;
    gap: 0.375rem;
    padding: 0.375rem 0.5rem;
  }

  .editor-toolbar-title {
    font-size: 0.875rem;
  }

  .editor-toolbar-badge {
    display: none;
  }

  .editor-toolbar-input {
    width: 100%;
    order: 3;
  }

  .editor-main {
    flex-direction: column;
  }

  /* On mobile, sidebars become full-width bottom drawers */
  .editor-sidebar {
    width: 100%;
    max-height: 35vh;
    border-right: none;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .sidebar-collapsed {
    max-height: 0;
  }

  .editor-sidebar-right {
    border-left: none;
    border-top: 1px solid var(--border-color, #333);
    order: 3;
  }

  /* Toggle buttons become horizontal bars in column layout */
  .sidebar-toggle {
    width: 100%;
    height: 20px;
    border-right: none;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .sidebar-toggle-right {
    border-left: none;
    border-top: 1px solid var(--border-color, #333);
  }

  .sidebar-toggle svg {
    transform: rotate(-90deg);
  }

  .quick-panel {
    width: 100%;
    max-height: 30vh;
    border-right: none;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .quick-mode-layout {
    flex-direction: column;
  }

  /* Compact toolbar buttons on mobile */
  .editor-toolbar-right {
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .editor-lang-btn {
    width: 24px;
    height: 24px;
    font-size: 0.625rem;
  }

  .editor-btn-draft,
  .editor-btn-upload,
  .editor-export-dropdown .editor-btn-export {
    font-size: 0.6875rem;
    padding: 0.25rem 0.5rem;
  }

  /* Give canvas maximum space: reduce sidebar max-height */
  .editor-sidebar {
    max-height: 28vh;
  }

  .canvas-viewport {
    padding: 0.25rem;
  }
}
</style>
