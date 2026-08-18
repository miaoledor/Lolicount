<script setup lang="ts">
defineProps<{ url: string; width: number }>()
</script>

<template>
  <div class="inline-block border rounded" :style="{ maxWidth: width + 'px' }">
    <img :src="url" alt="counter preview" class="w-full h-auto" draggable="false" />
  </div>
</template>
