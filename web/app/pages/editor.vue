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
const unshowFont = ref(false)

const exportLoading = ref(false)
const errorMsg = ref('')
const savedDrafts = ref<string[]>([])
const autoSaveEnabled = ref(true)

const nonTextLayers = computed(() => layers.value.filter((l) => !l.fixed))
const isCardTheme = computed(() => nonTextLayers.value.length <= 1)

const collectState = () => ({
  themeName: themeName.value,
  canvasWidth: canvasWidth.value,
  canvasHeight: canvasHeight.value,
  displaySize: displaySize.value,
  layers: layers.value,
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

const buildRequest = (): EditorRequest => ({
  name: themeName.value || 'untitled',
  canvas: { width: canvasWidth.value, height: canvasHeight.value },
  display: displaySize.value > 0 ? { size: displaySize.value } : null,
  layers: layers.value,
  text: counterText.value,
  fsize: fontSize.value,
  scale: scale.value,
  unshowf: unshowFont.value,
})

const addLayer = (category = 'lass') => {
  layerIdCounter.value++
  layers.value.push({
    id: layerIdCounter.value,
    category,
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
  ;[layers.value[idx], layers.value[target]] = [layers.value[target], layers.value[idx]]
  layers.value.forEach((l, i) => (l.zIndex = i))
}

const updateLayer = (id: number, patch: Partial<EditorLayer>) => {
  const layer = layers.value.find((l) => l.id === id)
  if (layer) Object.assign(layer, patch)
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
    const blob = await exportTheme(buildRequest())
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
  <div class="editor-page">
    <div class="editor-layout">
      <div class="editor-left-panel">
        <h2 class="editor-title">{{ t('editor.title') }}</h2>

        <div class="editor-field">
          <label>{{ t('editor.themeName') }}</label>
          <input v-model="themeName" type="text" :placeholder="t('editor.namePlaceholder')" class="editor-input">
        </div>

        <div v-if="savedDrafts.length > 0" class="editor-drafts">
          <label>{{ t('editor.savedDrafts') }}</label>
          <div class="editor-draft-list">
            <div v-for="d in savedDrafts" :key="d" class="editor-draft-item">
              <button class="editor-draft-btn" @click="restoreDraft(d)">{{ d }}</button>
              <button class="btn-xs btn-danger" @click="removeDraft(d)">×</button>
            </div>
          </div>
        </div>

        <div class="editor-field-row">
          <div class="editor-field">
            <label>{{ t('editor.canvasW') }}</label>
            <input v-model.number="canvasWidth" type="number" class="editor-input">
          </div>
          <div class="editor-field">
            <label>{{ t('editor.canvasH') }}</label>
            <input v-model.number="canvasHeight" type="number" class="editor-input">
          </div>
        </div>

        <div class="editor-field">
          <label>{{ t('editor.displaySize') }}</label>
          <input v-model.number="displaySize" type="number" class="editor-input">
        </div>

        <LayerPanel
          :layers="layers"
          @add="addLayer()"
          @remove="removeLayer"
          @move="moveLayer"
          @update="updateLayer"
        >
          <template #layerContent="{ layer }">
            <LayerImageControls
              :images="layer.images"
              @add-image="(img) => addImage(layer.id, img)"
              @remove-image="(i) => removeImage(layer.id, i)"
              @update-image="(i, p) => updateImage(layer.id, i, p)"
            />
          </template>
        </LayerPanel>

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

        <div class="editor-actions">
          <button class="editor-btn-primary" :disabled="exportLoading" @click="doExport">
            {{ exportLoading ? '...' : t('editor.export') }}
          </button>
        </div>

        <p v-if="errorMsg" class="editor-error">{{ errorMsg }}</p>
        <p class="editor-info">{{ isCardTheme ? t('editor.cardTheme') : t('editor.characterTheme') }}</p>
      </div>

      <div class="editor-canvas-area">
        <EditorCanvas :request="buildRequest()" :has-layers="nonTextLayers.length > 0" :canvas-width="canvasWidth" :canvas-height="canvasHeight" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.editor-page {
  min-height: calc(100vh - 4rem);
  padding: 1rem;
}

.editor-layout {
  display: flex;
  gap: 1rem;
  max-width: 1400px;
  margin: 0 auto;
}

.editor-left-panel {
  width: 360px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.editor-canvas-area {
  flex: 1;
}

.editor-title {
  font-size: 1.25rem;
  margin: 0;
}

.editor-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.editor-field-row {
  display: flex;
  gap: 0.5rem;
}

.editor-field-row .editor-field {
  flex: 1;
}

.editor-input {
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-input, #1a1a1a);
  color: var(--text-color, #eee);
  font-size: 0.875rem;
}

.editor-actions {
  display: flex;
  gap: 0.5rem;
}

.editor-btn-primary {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  background: var(--accent, #ec4899);
  color: #fff;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 600;
  flex: 1;
}

.editor-btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-error {
  color: #ef4444;
  font-size: 0.8125rem;
  margin: 0;
}

.editor-info {
  color: var(--text-muted, #999);
  font-size: 0.8125rem;
  margin: 0;
}

.editor-drafts {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.editor-draft-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.editor-draft-item {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  overflow: hidden;
}

.editor-draft-btn {
  padding: 0.125rem 0.375rem;
  border: none;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.6875rem;
}

.editor-draft-btn:hover {
  background: var(--accent, #ec4899);
  color: #fff;
}

.btn-xs {
  padding: 0.125rem 0.375rem;
  border: none;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
}

.btn-danger:hover {
  color: #ef4444;
}

@media (max-width: 768px) {
  .editor-layout {
    flex-direction: column;
  }

  .editor-left-panel {
    width: 100%;
  }
}
</style>
