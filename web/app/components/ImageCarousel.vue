// ImageCarousel: a cyclic image gallery with left/right navigation.
// Images are counter SVGs rendered with different themes, showcasing
// what Lolicount can produce. Clicking the left/right arrow buttons
// (or the image edges) cycles to the previous/next slide. The gallery
// wraps around at both ends.
<script setup lang="ts">
const props = defineProps<{
  images: { src: string; alt: string; caption?: string }[]
}>()

const current = ref(0)

const next = () => {
  current.value = (current.value + 1) % props.images.length
}

const prev = () => {
  current.value = (current.value - 1 + props.images.length) % props.images.length
}

const goTo = (index: number) => {
  current.value = index
}

// Keyboard navigation: arrow left/right.
const onKey = (e: KeyboardEvent) => {
  if (e.key === 'ArrowLeft') prev()
  if (e.key === 'ArrowRight') next()
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="carousel">
    <button
      class="carousel-arrow carousel-arrow-left"
      :aria-label="'Previous'"
      @click="prev"
    >
      <span>‹</span>
    </button>

    <div class="carousel-stage">
      <div
        class="carousel-track"
        :style="{ transform: `translateX(-${current * 100}%)` }"
      >
        <div
          v-for="(img, i) in images"
          :key="i"
          class="carousel-slide"
        >
          <img :src="img.src" :alt="img.alt" class="carousel-img" />
          <p v-if="img.caption" class="carousel-caption">{{ img.caption }}</p>
        </div>
      </div>
    </div>

    <button
      class="carousel-arrow carousel-arrow-right"
      :aria-label="'Next'"
      @click="next"
    >
      <span>›</span>
    </button>

    <div class="carousel-dots">
      <button
        v-for="(img, i) in images"
        :key="i"
        :class="cn('carousel-dot', i === current && 'is-active')"
        @click="goTo(i)"
        :aria-label="`Slide ${i + 1}`"
      />
    </div>
  </div>
</template>

<style scoped>
.carousel {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.carousel-stage {
  flex: 1;
  overflow: hidden;
  border-radius: 0.75rem;
  background: var(--loli-cream);
}
.carousel-track {
  display: flex;
  transition: transform 0.35s ease;
}
.carousel-slide {
  flex: 0 0 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  min-height: 200px;
}
.carousel-img {
  max-height: 240px;
  max-width: 100%;
  object-fit: contain;
}
.carousel-caption {
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: #6b7280;
}
.carousel-arrow {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  border: 1px solid #e5d4dc;
  background: #fff;
  color: var(--loli-pink);
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}
.carousel-arrow:hover {
  background: var(--loli-pink);
  color: #fff;
  border-color: var(--loli-pink);
}
.carousel-dots {
  position: absolute;
  bottom: 0.5rem;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 0.375rem;
}
.carousel-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  border: none;
  background: #d1c0c8;
  cursor: pointer;
  transition: background 0.2s;
  padding: 0;
}
.carousel-dot.is-active {
  background: var(--loli-pink);
}
</style>
