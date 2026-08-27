// NoticeBoard: a kungal-style auto-paging showcase that loops through
// images in /public/aboutImg. Pages slide horizontally on an interval;
// dots and arrows allow manual navigation. Pauses on hover so users can
// inspect an image. The image list is resolved via Nuxt's glob import
// (import.meta.glob) so any file dropped into aboutImg is picked up
// automatically at build time — no manual list maintenance needed.

<script setup lang="ts">
const { t } = useI18n()

const current = ref(0)
const timer = ref<ReturnType<typeof setInterval> | null>(null)
const paused = ref(false)

const modules = import.meta.glob('~/public/aboutImg/*.{png,jpg,jpeg,webp,gif,svg,avif}', {
  eager: true,
  query: '?url',
  import: 'default',
}) as Record<string, string>

const images = Object.keys(modules)
  .sort((a, b) => a.localeCompare(b, undefined, { numeric: true }))
  .map((path) => {
    const file = path.split('/').pop() ?? path
    return { src: modules[path], alt: file.replace(/\.[^.]+$/, '') }
  })

const imageCount = computed(() => images.length)

const next = () => {
  if (imageCount.value === 0) return
  current.value = (current.value + 1) % imageCount.value
}

const prev = () => {
  if (imageCount.value === 0) return
  current.value = (current.value - 1 + imageCount.value) % imageCount.value
}

const goTo = (idx: number) => {
  current.value = idx
}

const startAuto = () => {
  stopAuto()
  if (imageCount.value <= 1) return
  timer.value = setInterval(() => {
    if (!paused.value) next()
  }, 4000)
}

const stopAuto = () => {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
}

onMounted(() => {
  startAuto()
})

onBeforeUnmount(() => {
  stopAuto()
})
</script>

<template>
  <div
    class="notice-board"
    @mouseenter="paused = true"
    @mouseleave="paused = false"
  >
    <div class="notice-viewport">
      <div
        v-if="imageCount > 0"
        class="notice-track"
        :style="{ transform: `translateX(-${current * 100}%)` }"
      >
        <div
          v-for="img in images"
          :key="img.src"
          class="notice-page"
        >
          <img
            :src="img.src"
            :alt="img.alt"
            class="notice-image"
            loading="lazy"
          />
          <span class="notice-label">{{ img.alt }}</span>
        </div>
      </div>

      <div v-else class="notice-empty">
        {{ t('loli.loading') }}
      </div>
    </div>

    <button
      v-if="imageCount > 1"
      type="button"
      class="notice-arrow notice-arrow-left"
      :aria-label="t('notice.prev')"
      @click="prev"
    >
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path fill="currentColor" d="M15.41 7.41 14 6l-6 6 6 6 1.41-1.41L10.83 12z" />
      </svg>
    </button>
    <button
      v-if="imageCount > 1"
      type="button"
      class="notice-arrow notice-arrow-right"
      :aria-label="t('notice.next')"
      @click="next"
    >
      <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path fill="currentColor" d="M8.59 16.59 10 18l6-6-6-6-1.41 1.41L13.17 12z" />
      </svg>
    </button>

    <div v-if="imageCount > 1" class="notice-dots">
      <button
        v-for="(img, idx) in images"
        :key="img.src"
        type="button"
        class="notice-dot"
        :class="{ 'notice-dot-active': idx === current }"
        :aria-label="img.alt"
        @click="goTo(idx)"
      />
    </div>
  </div>
</template>

<style scoped>
.notice-board {
  position: relative;
  width: 100%;
  border-radius: 1rem;
  overflow: hidden;
  background: var(--loli-cream);
  border: 1px solid var(--border-color);
}

.notice-viewport {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  overflow: hidden;
}

.notice-track {
  display: flex;
  width: 100%;
  height: 100%;
  transition: transform 0.6s cubic-bezier(0.22, 1, 0.36, 1);
}

.notice-page {
  position: relative;
  flex: 0 0 100%;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--loli-cream);
}

.notice-image {
  max-width: 90%;
  max-height: 90%;
  object-fit: contain;
}

.notice-label {
  position: absolute;
  bottom: 0.75rem;
  right: 0.75rem;
  padding: 0.2rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-color);
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border-radius: 0.5rem;
}

.notice-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.notice-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  color: var(--loli-pink);
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  z-index: 2;
}
.notice-arrow:hover {
  background: var(--loli-pink);
  color: #fff;
}
.notice-arrow-left {
  left: 0.5rem;
}
.notice-arrow-right {
  right: 0.5rem;
}

.notice-dots {
  position: absolute;
  bottom: 0.6rem;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 0.4rem;
  z-index: 2;
}
.notice-dot {
  width: 0.5rem;
  height: 0.5rem;
  padding: 0;
  background: rgba(0, 0, 0, 0.2);
  border: none;
  border-radius: 50%;
  cursor: pointer;
  transition: background 0.2s, transform 0.2s;
}
.notice-dot-active {
  background: var(--loli-pink);
  transform: scale(1.3);
}

@media (max-width: 767px) {
  .notice-arrow {
    width: 1.75rem;
    height: 1.75rem;
  }
  .notice-arrow-left {
    left: 0.25rem;
  }
  .notice-arrow-right {
    right: 0.25rem;
  }
}
</style>
