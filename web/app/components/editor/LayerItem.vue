<script setup lang="ts">
import type { EditorLayer } from '~/composables/useEditorApi'

const props = defineProps<{
  layer: EditorLayer
  index: number
  total: number
  selected: boolean
}>()

const emit = defineEmits<{
  remove: [id: number]
  move: [id: number, dir: -1 | 1]
  update: [id: number, patch: Partial<EditorLayer>]
  select: [id: number]
}>()

const { t } = useI18n()
const expanded = computed(() => props.selected)

const toggleExpand = () => {
  emit('select', props.layer.id)
}

const onNameChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  emit('update', props.layer.id, { name: target.value })
}
</script>

<template>
  <div class="layer-item" :class="{ 'layer-item-selected': selected }">
    <div class="layer-row" @click="toggleExpand">
      <span class="layer-toggle-icon" :class="{ 'layer-toggle-open': expanded }">
        <svg width="10" height="10" viewBox="0 0 10 10">
          <path d="M3 1 L7 5 L3 9" fill="none" stroke="currentColor" stroke-width="1.5" />
        </svg>
      </span>
      <input
        :value="layer.name"
        type="text"
        class="layer-name"
        @click.stop
        @input="onNameChange($event)"
      >
      <span class="layer-badge">{{ layer.images.length }}{{ t('editor.imgUnit') }}</span>
      <div class="layer-actions" @click.stop>
        <button class="layer-action-btn" :disabled="index === 0" @click="emit('move', layer.id, -1)">
          <svg width="10" height="10" viewBox="0 0 10 10"><path d="M5 1 L9 6 L6 6 L6 9 L4 9 L4 6 L1 6 Z" fill="currentColor"/></svg>
        </button>
        <button class="layer-action-btn" :disabled="index === total - 1" @click="emit('move', layer.id, 1)">
          <svg width="10" height="10" viewBox="0 0 10 10"><path d="M5 9 L1 4 L4 4 L4 1 L6 1 L6 4 L9 4 Z" fill="currentColor"/></svg>
        </button>
        <button v-if="!layer.fixed" class="layer-action-btn layer-action-del" @click="emit('remove', layer.id)">
          <svg width="10" height="10" viewBox="0 0 10 10"><path d="M2 2 L8 8 M8 2 L2 8" stroke="currentColor" stroke-width="1.5"/></svg>
        </button>
      </div>
    </div>
    <Transition name="layer-expand">
      <div v-if="expanded" class="layer-body">
        <slot />
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.layer-item {
  border: 1px solid var(--border-color, #333);
  border-radius: 6px;
  overflow: hidden;
  transition: border-color 0.15s, background 0.15s;
}

.layer-item-selected {
  border-color: var(--accent, #ec4899);
  background: rgba(236, 72, 153, 0.05);
}

.layer-row {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.5rem;
  cursor: pointer;
}

.layer-toggle-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--text-muted, #999);
  transition: transform 0.15s ease;
}

.layer-toggle-open {
  transform: rotate(90deg);
}

.layer-name {
  flex: 1;
  min-width: 0;
  padding: 0.2rem 0.375rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.75rem;
  font-weight: 600;
}

.layer-badge {
  font-size: 0.625rem;
  padding: 0.1rem 0.3rem;
  border-radius: 8px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-muted, #999);
  flex-shrink: 0;
}

.layer-actions {
  display: flex;
  gap: 0.125rem;
  flex-shrink: 0;
}

.layer-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  padding: 0;
}

.layer-action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.layer-action-btn:not(:disabled):hover {
  border-color: var(--accent, #ec4899);
  color: var(--accent, #ec4899);
}

.layer-action-del:not(:disabled):hover {
  border-color: #ef4444;
  color: #ef4444;
}

.layer-body {
  padding: 0.5rem;
  border-top: 1px solid var(--border-color, #333);
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
  max-height: 600px;
}
</style>
