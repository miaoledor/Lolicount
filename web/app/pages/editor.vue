<script setup lang="ts">
import type { EditorRequest, EditorLayer, EditorImage } from '~/composables/useEditorApi'

const { previewTheme, exportTheme } = useEditorApi()
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

const previewSvg = ref('')
const previewLoading = ref(false)
const exportLoading = ref(false)
const errorMsg = ref('')

const nonTextLayers = computed(() => layers.value.filter((l) => l.category !== 'text'))

const isCardTheme = computed(() => nonTextLayers.value.length <= 1)

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

const doPreview = async () => {
  errorMsg.value = ''
  if (nonTextLayers.value.length === 0) {
    errorMsg.value = t('editor.errNoLayers')
    return
  }
  previewLoading.value = true
  try {
    previewSvg.value = await previewTheme(buildRequest())
  } catch (e: any) {
    errorMsg.value = e?.data?.message || e?.message || 'preview failed'
  } finally {
    previewLoading.value = false
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

const addLayer = (category: string = 'lass') => {
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
</script>

<template>
  <div class="editor-page">
    <div class="editor-layout">
      <div class="editor-left-panel">
        <h2 class="editor-title">{{ t('editor.title') }}</h2>

        <div class="editor-field">
          <label>{{ t('editor.themeName') }}</label>
          <input v-model="themeName" type="text" :placeholder="t('editor.namePlaceholder')" class="editor-input" />
        </div>

        <div class="editor-field-row">
          <div class="editor-field">
            <label>{{ t('editor.canvasW') }}</label>
            <input v-model.number="canvasWidth" type="number" class="editor-input" />
          </div>
          <div class="editor-field">
            <label>{{ t('editor.canvasH') }}</label>
            <input v-model.number="canvasHeight" type="number" class="editor-input" />
          </div>
        </div>

        <div class="editor-field">
          <label>{{ t('editor.displaySize') }}</label>
          <input v-model.number="displaySize" type="number" class="editor-input" />
        </div>

        <div class="editor-layers-section">
          <div class="editor-layers-header">
            <span>{{ t('editor.layers') }} ({{ nonTextLayers.length }})</span>
            <button class="editor-btn-sm" @click="addLayer()">{{ t('editor.addLayer') }}</button>
          </div>

          <div v-for="layer in layers" :key="layer.id" class="editor-layer-item">
            <div class="editor-layer-header">
              <span class="editor-layer-cat">{{ layer.category }}</span>
              <span class="editor-layer-z">Z:{{ layer.zIndex }}</span>
              <span class="editor-layer-count">{{ layer.images.length }}img</span>
              <button v-if="!layer.fixed" class="editor-btn-xs" @click="removeLayer(layer.id)">×</button>
            </div>
            <div class="editor-layer-controls">
              <select v-model="layer.category" class="editor-select">
                <option value="lass">lass</option>
                <option value="brow">brow</option>
                <option value="eye">eye</option>
                <option value="mouth">mouth</option>
                <option value="face">face</option>
              </select>
              <button class="editor-btn-xs" @click="moveLayer(layer.id, -1)">↑</button>
              <button class="editor-btn-xs" @click="moveLayer(layer.id, 1)">↓</button>
            </div>
          </div>
        </div>

        <div class="editor-text-section">
          <h3>{{ t('editor.textLayer') }}</h3>
          <div class="editor-field">
            <label>{{ t('editor.counterText') }}</label>
            <input v-model="counterText" type="text" class="editor-input" />
          </div>
          <div class="editor-field-row">
            <div class="editor-field">
              <label>{{ t('editor.fontSize') }}</label>
              <input v-model.number="fontSize" type="number" class="editor-input" />
            </div>
            <div class="editor-field">
              <label>{{ t('editor.scale') }}</label>
              <input v-model.number="scale" type="number" step="0.1" class="editor-input" />
            </div>
          </div>
          <label class="editor-checkbox">
            <input v-model="unshowFont" type="checkbox" />
            {{ t('editor.hideText') }}
          </label>
        </div>

        <div class="editor-actions">
          <button class="editor-btn-primary" :disabled="previewLoading" @click="doPreview">
            {{ previewLoading ? '...' : t('editor.preview') }}
          </button>
          <button class="editor-btn-primary" :disabled="exportLoading" @click="doExport">
            {{ exportLoading ? '...' : t('editor.export') }}
          </button>
        </div>

        <p v-if="errorMsg" class="editor-error">{{ errorMsg }}</p>
        <p v-if="isCardTheme" class="editor-info">{{ t('editor.cardTheme') }}</p>
        <p v-else class="editor-info">{{ t('editor.characterTheme') }}</p>
      </div>

      <div class="editor-canvas-area">
        <div v-if="previewSvg" class="editor-preview" v-html="previewSvg" />
        <div v-else class="editor-placeholder">
          <p>{{ t('editor.previewHint') }}</p>
        </div>
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
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 1rem;
}

.editor-preview {
  max-width: 100%;
}

.editor-preview :deep(svg) {
  max-width: 100%;
  height: auto;
}

.editor-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: var(--text-muted, #999);
  border: 2px dashed var(--border-color, #333);
  border-radius: 8px;
  width: 100%;
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

.editor-input,
.editor-select {
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-input, #1a1a1a);
  color: var(--text-color, #eee);
  font-size: 0.875rem;
}

.editor-layers-section,
.editor-text-section {
  border: 1px solid var(--border-color, #333);
  border-radius: 6px;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.editor-layers-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
  font-weight: 600;
}

.editor-layer-item {
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.editor-layer-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
}

.editor-layer-cat {
  font-weight: 600;
}

.editor-layer-z,
.editor-layer-count {
  color: var(--text-muted, #999);
}

.editor-layer-controls {
  display: flex;
  gap: 0.25rem;
  align-items: center;
}

.editor-layer-controls .editor-select {
  flex: 1;
}

.editor-btn-sm,
.editor-btn-xs {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
}

.editor-btn-xs {
  padding: 0.125rem 0.375rem;
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
}

.editor-btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-actions {
  display: flex;
  gap: 0.5rem;
}

.editor-checkbox {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  cursor: pointer;
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

@media (max-width: 768px) {
  .editor-layout {
    flex-direction: column;
  }

  .editor-left-panel {
    width: 100%;
  }
}
</style>
