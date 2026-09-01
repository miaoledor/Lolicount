<script setup lang="ts">
import type { EditorImage } from '~/composables/useEditorApi'

const props = defineProps<{
  images: EditorImage[]
  selectedIndex: number
}>()

const emit = defineEmits<{
  addImage: [img: EditorImage]
  removeImage: [index: number]
  updateImage: [index: number, patch: Partial<EditorImage>]
  selectImage: [index: number]
}>()

const { t } = useI18n()

const proportionalScale = ref(true)

const fileToDataURI = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

const onFileSelect = async (e: Event) => {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  for (const file of Array.from(input.files)) {
    if (!file.type.startsWith('image/')) continue
    const src = await fileToDataURI(file)
    emit('addImage', { src, left: 0, top: 0, width: 100, height: 100 })
  }
  input.value = ''
}

const buildDimensionPatch = (
  image: EditorImage,
  field: 'width' | 'height',
  value: number,
  baseWidth = image.width,
  baseHeight = image.height,
): Partial<EditorImage> => {
  if (!Number.isFinite(value)) return { [field]: value }
  if (!proportionalScale.value || baseWidth <= 0 || baseHeight <= 0 || value <= 0) {
    return { [field]: value }
  }
  if (field === 'width') {
    return { width: value, height: Math.round((value * baseHeight) / baseWidth) }
  }
  return { width: Math.round((value * baseWidth) / baseHeight), height: value }
}

const updateNumber = (index: number, field: keyof EditorImage, e: Event) => {
  const input = e.target as HTMLInputElement
  const value = Number(input.value)
  if (field !== 'width' && field !== 'height') {
    emit('updateImage', index, { [field]: value } as Partial<EditorImage>)
    return
  }
  emit('updateImage', index, buildDimensionPatch(props.images[index]!, field, value))
}

const selectImage = (index: number) => {
  emit('selectImage', index)
}

// Drag-to-edit: hold and drag horizontally on a number input to
// increment/decrement the value, like vue-fabric-editor's InputNumber.
const dragState = ref<{
  field: keyof EditorImage
  startX: number
  startVal: number
  index: number
  startWidth: number
  startHeight: number
  dragging: boolean
} | null>(null)

const onDragStart = (index: number, field: keyof EditorImage, e: PointerEvent) => {
  const input = e.target as HTMLInputElement
  const currentVal = Number(input.value) || 0
  const image = props.images[index]!
  dragState.value = {
    field,
    startX: e.clientX,
    startVal: currentVal,
    index,
    startWidth: image.width,
    startHeight: image.height,
    dragging: false,
  }
  input.setPointerCapture(e.pointerId)
}

const onDragMove = (e: PointerEvent) => {
  if (!dragState.value) return
  const dx = e.clientX - dragState.value.startX
  if (!dragState.value.dragging && Math.abs(dx) < 3) return
  dragState.value.dragging = true
  const newVal = Math.round(dragState.value.startVal + dx)
  const { index, field, startWidth, startHeight } = dragState.value
  if (field !== 'width' && field !== 'height') {
    emit('updateImage', index, { [field]: newVal } as Partial<EditorImage>)
    return
  }
  emit('updateImage', index, buildDimensionPatch(props.images[index]!, field, newVal, startWidth, startHeight))
}

const onDragEnd = (e: PointerEvent) => {
  if (dragState.value?.dragging) {
    (e.target as HTMLElement).releasePointerCapture?.(e.pointerId)
  }
  dragState.value = null
}
</script>

<template>
  <div class="img-controls" @pointermove="onDragMove" @pointerup="onDragEnd">
    <label class="img-upload-label">
      <input type="file" accept="image/*" multiple class="img-upload-input" @change="onFileSelect">
      <span class="img-upload-btn">{{ t('editor.uploadImage') }}</span>
    </label>

    <div v-if="images.length === 0" class="img-empty">
      {{ t('editor.noImages') }}
    </div>

    <div v-for="(img, i) in images" :key="i" class="img-item">
      <div class="img-item-header">
        <button
          type="button"
          class="img-thumb-wrap"
          :class="{ 'img-thumb-selected': i === selectedIndex }"
          :title="t('editor.selectImage')"
          @click="selectImage(i)"
        >
          <img :src="img.src" class="img-thumb" alt="">
          <span v-if="i === selectedIndex" class="img-thumb-badge">✓</span>
        </button>
        <span class="img-index">#{{ i + 1 }}</span>
        <button class="img-del-btn" @click="emit('removeImage', i)">
          <svg width="10" height="10" viewBox="0 0 10 10"><path d="M2 2 L8 8 M8 2 L2 8" stroke="currentColor" stroke-width="1.5"/></svg>
        </button>
      </div>
      <div class="img-fields">
        <label class="img-proportional">
          <input v-model="proportionalScale" type="checkbox">
          <span>{{ t('editor.proportionalScale') }}</span>
        </label>
        <div class="img-field">
          <label>X</label>
          <input
            :value="img.left"
            type="number"
            class="img-num-input"
            @input="updateNumber(i, 'left', $event)"
            @pointerdown="onDragStart(i, 'left', $event)"
          >
        </div>
        <div class="img-field">
          <label>Y</label>
          <input
            :value="img.top"
            type="number"
            class="img-num-input"
            @input="updateNumber(i, 'top', $event)"
            @pointerdown="onDragStart(i, 'top', $event)"
          >
        </div>
        <div class="img-field">
          <label>W</label>
          <input
            :value="img.width"
            type="number"
            class="img-num-input"
            @input="updateNumber(i, 'width', $event)"
            @pointerdown="onDragStart(i, 'width', $event)"
          >
        </div>
        <div class="img-field">
          <label>H</label>
          <input
            :value="img.height"
            type="number"
            class="img-num-input"
            @input="updateNumber(i, 'height', $event)"
            @pointerdown="onDragStart(i, 'height', $event)"
          >
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.img-controls {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.img-upload-label {
  cursor: pointer;
}

.img-upload-input {
  display: none;
}

.img-upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  border: 1px dashed var(--loli-pink);
  border-radius: 4px;
  background: transparent;
  color: var(--loli-pink);
  font-size: 0.6875rem;
  font-weight: 600;
  transition: all 0.15s;
}

.img-upload-btn:hover {
  background: rgba(107, 114, 128, 0.1);
}

.img-empty {
  color: var(--text-muted, #999);
  font-size: 0.6875rem;
  text-align: center;
  padding: 0.5rem;
}

.img-item {
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  padding: 0.375rem;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.img-item-header {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.img-thumb-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 2px solid transparent;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  padding: 0;
  transition: border-color 0.15s;
}

.img-thumb-wrap:hover {
  border-color: var(--text-muted, #999);
}

.img-thumb-selected {
  border-color: var(--loli-pink) !important;
}

.img-thumb {
  width: 24px;
  height: 24px;
  object-fit: cover;
  border-radius: 3px;
}

.img-thumb-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--loli-pink);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
}

.img-index {
  font-size: 0.6875rem;
  color: var(--text-muted, #999);
}

.img-del-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-left: auto;
  border: 1px solid var(--border-color, #444);
  border-radius: 3px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  padding: 0;
}

.img-del-btn:hover {
  border-color: #ef4444;
  color: #ef4444;
}

.img-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.25rem;
}

.img-proportional {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  color: var(--text-muted, #999);
  font-size: 0.625rem;
  user-select: none;
}

.img-proportional input {
  accent-color: var(--loli-pink);
}

.img-field {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.img-field label {
  font-size: 0.625rem;
  color: var(--text-muted, #999);
  min-width: 0.75rem;
  font-family: monospace;
}

.img-num-input {
  width: 100%;
  padding: 0.125rem 0.25rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 3px;
  background: #fff;
  color: #111;
  font-size: 0.6875rem;
  cursor: ew-resize;
  user-select: none;
}

.img-num-input:focus {
  cursor: text;
  user-select: text;
  border-color: var(--loli-pink);
  outline: none;
}
</style>
