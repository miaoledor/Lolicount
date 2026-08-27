<script setup lang="ts">
import type { EditorLayer } from '~/composables/useEditorApi'

const props = defineProps<{
  layer: EditorLayer
  index: number
  total: number
}>()

const emit = defineEmits<{
  remove: [id: number]
  move: [id: number, dir: -1 | 1]
  update: [id: number, patch: Partial<EditorLayer>]
}>()

const { t } = useI18n()
const expanded = ref(false)

const toggleExpand = () => {
  expanded.value = !expanded.value
}

const onCategoryChange = (e: Event) => {
  const target = e.target as HTMLSelectElement
  emit('update', props.layer.id, { category: target.value })
}

const onNameChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  emit('update', props.layer.id, { name: target.value })
}
</script>

<template>
  <div class="layer-item" :class="{ 'layer-item-expanded': expanded }">
    <div class="layer-header">
      <button class="layer-expand-btn" @click="toggleExpand">
        <span class="layer-expand-icon" :class="{ 'layer-expand-icon-open': expanded }">▶</span>
      </button>
      <input
        :value="layer.name"
        type="text"
        class="layer-name-input"
        @input="onNameChange($event)"
      >
      <span class="layer-z">{{ t('editor.layerZ') }}:{{ index }}</span>
      <span class="layer-count">{{ layer.images.length }}{{ t('editor.imgUnit') }}</span>
      <div class="layer-actions">
        <button class="btn-xs" :disabled="index === 0" @click="emit('move', layer.id, -1)">↑</button>
        <button class="btn-xs" :disabled="index === total - 1" @click="emit('move', layer.id, 1)">↓</button>
        <button v-if="!layer.fixed" class="btn-xs btn-danger" @click="emit('remove', layer.id)">×</button>
      </div>
    </div>
    <Transition name="layer-expand">
      <div v-if="expanded" class="layer-body">
        <div class="layer-control-row">
          <label class="layer-label">{{ t('editor.category') }}</label>
          <select :value="layer.category" class="layer-select" @change="onCategoryChange">
            <option value="lass">lass</option>
            <option value="brow">brow</option>
            <option value="eye">eye</option>
            <option value="mouth">mouth</option>
            <option value="face">face</option>
          </select>
        </div>
        <slot />
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.layer-item {
  border: 1px solid var(--border-color, #333);
  border-radius: 4px;
  overflow: hidden;
}

.layer-item-expanded {
  border-color: var(--accent, #ec4899);
}

.layer-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  font-size: 0.75rem;
}

.layer-expand-btn {
  background: none;
  border: none;
  color: var(--text-muted, #999);
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.layer-expand-icon {
  display: inline-block;
  transition: transform 0.2s ease;
  font-size: 0.625rem;
}

.layer-expand-icon-open {
  transform: rotate(90deg);
}

.layer-name-input {
  width: 80px;
  padding: 0.125rem 0.25rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 3px;
  background: #fff;
  color: #111;
  font-size: 0.75rem;
  font-weight: 600;
}

.layer-z,
.layer-count {
  color: var(--text-muted, #999);
}

.layer-actions {
  margin-left: auto;
  display: flex;
  gap: 0.25rem;
}

.btn-xs {
  padding: 0.125rem 0.375rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
}

.btn-xs:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-danger:hover {
  color: #ef4444;
}

.layer-body {
  padding: 0.5rem;
  border-top: 1px solid var(--border-color, #333);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.layer-control-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.layer-label {
  font-size: 0.75rem;
  color: var(--text-muted, #999);
  min-width: 3rem;
}

.layer-select {
  flex: 1;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.75rem;
}

.layer-expand-enter-active,
.layer-expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.layer-expand-enter-from,
.layer-expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.layer-expand-enter-to,
.layer-expand-leave-from {
  max-height: 500px;
}
</style>
