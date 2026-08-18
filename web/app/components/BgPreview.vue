<script setup lang="ts">
const props = defineProps<{ url: string; width: number }>()
const emit = defineEmits<{ drag: [x: number, y: number] }>()

const container = ref<HTMLElement | null>(null)
const { dragging, onPointerDown, onPointerMove, onPointerUp } = useDragPosition((x, y) => emit('drag', x, y))
</script>

<template>
  <div
    ref="container"
    :class="cn('relative inline-block cursor-move select-none border rounded', dragging && 'cursor-grabbing')"
    :style="{ maxWidth: width + 'px' }"
    @pointerdown="onPointerDown"
    @pointermove="onPointerMove($event, container!)"
    @pointerup="onPointerUp"
    @pointerleave="onPointerUp"
  >
    <img :src="url" alt="counter preview" class="w-full h-auto pointer-events-none" draggable="false" />
  </div>
</template>
