// ImageGallery: displays showcase images stacked vertically, each shown
// at maximum available width. No navigation controls, backgrounds, or
// indicators — just the images.
<script setup lang="ts">
defineProps<{
  images: { src: string; alt: string }[]
}>()
</script>

<template>
  <div class="image-gallery">
    <img
      v-for="(img, i) in images"
      :key="i"
      :src="img.src"
      :alt="img.alt"
      class="image-gallery-item"
    />
  </div>
</template>

<style scoped>
.image-gallery {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}
.image-gallery-item {
  width: 100%;
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
}
</style>
