<script setup lang="ts">
import type { EditorImage } from '~/composables/useEditorApi'

const props = defineProps<{
  images: EditorImage[]
}>()

const emit = defineEmits<{
  addImage: [img: EditorImage]
  removeImage: [index: number]
  updateImage: [index: number, patch: Partial<EditorImage>]
}>()

const { t } = useI18n()

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

const updateNumber = (index: number, field: keyof EditorImage, e: Event) => {
  const v = Number((e.target as HTMLInputElement).value)
  emit('updateImage', index, { [field]: v } as Partial<EditorImage>)
}

// Drag-to-edit: hold and drag horizontally on a number input to
// increment/decrement the value, like vue-fabric-editor's InputNumber.
const dragState = ref<{ field: keyof EditorImage; startX: number; startVal: number; index: number } | null>(null)

const onDragStart = (index: number, field: keyof EditorImage, e: PointerEvent) => {
  const input = e.target as HTMLInputElement
  if (document.activeElement === input) return
  e.preventDefault()
  const currentVal = Number(input.value) || 0
  dragState.value = { field, startX: e.clientX, startVal: currentVal, index }
  input.setPointerCapture(e.pointerId)
}

const onDragMove = (e: PointerEvent) => {
  if (!dragState.value) return
  const dx = e.clientX - dragState.value.startX
  const newVal = Math.round(dragState.value.startVal + dx)
  emit('updateImage', dragState.value.index, { [dragState.value.field]: newVal } as Partial<EditorImage>)
}

const onDragEnd = () => {
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
        <img :src="img.src" class="img-thumb" alt="">
        <span class="img-index">#{{ i + 1 }}</span>
        <button class="img-del-btn" @click="emit('removeImage', i)">
          <svg width="10" height="10" viewBox="0 0 10 10"><path d="M2 2 L8 8 M8 2 L2 8" stroke="currentColor" stroke-width="1.5"/></svg>
        </button>
      </div>
      <div class="img-fields">
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

.img-thumb {
  width: 24px;
  height: 24px;
  object-fit: cover;
  border-radius: 3px;
  border: 1px solid var(--border-color, #444);
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
