<script setup lang="ts">
import type { EditorLayer } from '~/composables/useEditorApi'
import LayerItem from './LayerItem.vue'

const props = defineProps<{
  layers: EditorLayer[]
  selectedLayerId: number | null
}>()

const emit = defineEmits<{
  add: []
  remove: [id: number]
  move: [id: number, dir: -1 | 1]
  update: [id: number, patch: Partial<EditorLayer>]
  select: [id: number]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="layer-panel">
    <div class="layer-panel-header">
      <span class="layer-panel-title">{{ t('editor.layers') }} ({{ layers.length }})</span>
      <button class="layer-add-btn" @click="emit('add')">
        <span>+</span> {{ t('editor.addLayer') }}
      </button>
    </div>
    <div class="layer-list">
      <LayerItem
        v-for="(layer, i) in layers"
        :key="layer.id"
        :layer="layer"
        :index="i"
        :total="layers.length"
        :selected="layer.id === props.selectedLayerId"
        @remove="emit('remove', $event)"
        @move="(id, dir) => emit('move', id, dir)"
        @update="(id, patch) => emit('update', id, patch)"
        @select="emit('select', $event)"
      >
        <slot name="layerContent" :layer="layer" />
      </LayerItem>
      <div v-if="layers.length === 0" class="layer-empty">
        {{ t('editor.errNoLayers') }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.layer-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.layer-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.375rem;
  border-bottom: 1px solid var(--border-color, #333);
}

.layer-panel-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #999);
}

.layer-add-btn {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  border: 1px solid var(--loli-pink);
  border-radius: 4px;
  background: transparent;
  color: var(--loli-pink);
  cursor: pointer;
  font-size: 0.6875rem;
  font-weight: 600;
  transition: all 0.15s;
}

.layer-add-btn:hover {
  background: var(--loli-pink);
  color: #fff;
}

.layer-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.layer-empty {
  color: var(--text-muted, #999);
  font-size: 0.75rem;
  text-align: center;
  padding: 1rem;
}
</style>
