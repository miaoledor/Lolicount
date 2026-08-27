<script setup lang="ts">
import type { EditorLayer } from '~/composables/useEditorApi'
import LayerItem from './LayerItem.vue'

const props = defineProps<{
  layers: EditorLayer[]
}>()

const emit = defineEmits<{
  add: []
  remove: [id: number]
  move: [id: number, dir: -1 | 1]
  update: [id: number, patch: Partial<EditorLayer>]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="layer-panel">
    <div class="layer-panel-header">
      <span>{{ t('editor.layers') }} ({{ layers.length }})</span>
      <button class="btn-sm" @click="emit('add')">{{ t('editor.addLayer') }}</button>
    </div>
    <div class="layer-list">
      <LayerItem
        v-for="(layer, i) in layers"
        :key="layer.id"
        :layer="layer"
        :index="i"
        :total="layers.length"
        @remove="emit('remove', $event)"
        @move="emit('move', $event[0], $event[1])"
        @update="emit('update', $event[0], $event[1])"
      >
        <slot name="layerContent" :layer="layer" />
      </LayerItem>
    </div>
  </div>
</template>

<style scoped>
.layer-panel {
  border: 1px solid var(--border-color, #333);
  border-radius: 6px;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.layer-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
  font-weight: 600;
}

.layer-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.btn-sm {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: var(--bg-btn, #2a2a2a);
  color: var(--text-color, #eee);
  cursor: pointer;
  font-size: 0.75rem;
}
</style>
