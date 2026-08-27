<script setup lang="ts">
import type { EditorRequest, EditorLayer, EditorImage } from '~/composables/useEditorApi'
import EditorCanvas from '~/components/editor/EditorCanvas.vue'
import LayerPanel from '~/components/editor/LayerPanel.vue'
import TextLayerControls from '~/components/editor/TextLayerControls.vue'
import LayerImageControls from '~/components/editor/LayerImageControls.vue'
import { useEditorStorage } from '~/composables/useEditorStorage'

const { exportTheme } = useEditorApi()
const { saveDraft, loadDraft, listDrafts, deleteDraft } = useEditorStorage()
const { t } = useI18n()

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
      </div>
      <div class="editor-toolbar-right">
        <input v-model="themeName" type="text" :placeholder="t('editor.namePlaceholder')" class="editor-toolbar-input">
        <button class="editor-btn-export" :disabled="exportLoading" @click="doExport">
          {{ exportLoading ? '...' : t('editor.export') }}
        </button>
      </div>
    </div>

    <!-- Main three-column layout -->
    <div class="editor-main">
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
          <h3 class="sidebar-title">{{ t('editor.textLayer') }}</h3>
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

.editor-btn-export:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  .editor-main {
    flex-direction: column;
  }

  .editor-sidebar {
    width: 100%;
    max-height: 300px;
    border-right: none;
    border-bottom: 1px solid var(--border-color, #333);
  }

  .editor-sidebar-right {
    border-left: none;
    border-top: 1px solid var(--border-color, #333);
  }

  .editor-toolbar-input {
    width: 120px;
  }
}
</style>
