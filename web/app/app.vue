<script setup lang="ts">
const { init: initLocale } = useI18n()
const { init: initTheme } = useTheme()
useFireworkFx()
useRippleFx()
useReveal()

const progressEl = ref<HTMLElement | null>(null)
let ticking = false

// Top-of-frame scroll progress line, styled after the reference site
// (2px accent bar pinned to the frame's top edge, scaled on scroll).
const updateProgress = () => {
  ticking = false
  const max = document.documentElement.scrollHeight - window.innerHeight
  const ratio = max > 0 ? Math.min(1, window.scrollY / max) : 0
  if (progressEl.value) progressEl.value.style.transform = `scaleX(${ratio})`
}

const scheduleProgress = () => {
  if (ticking) return
  ticking = true
  requestAnimationFrame(updateProgress)
}

onMounted(() => {
  initLocale()
  initTheme()
  updateProgress()
  window.addEventListener('scroll', scheduleProgress, { passive: true })
  window.addEventListener('resize', scheduleProgress, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', scheduleProgress)
  window.removeEventListener('resize', scheduleProgress)
})
</script>

<template>
  <div class="app-frame">
    <LoadingScreen />
    <div ref="progressEl" class="scroll-progress" aria-hidden="true" />
    <div class="page-frame" aria-hidden="true" />
    <NavBar />
    <NuxtPage />
    <noscript>
      <!-- <style>
        /* Without JS the entrance-hidden state must not persist. */
        .reveal,
        .hero-letter,
        .major-letter,
        .hero-avatar { opacity: 1 !important; transform: none !important; }
        [data-nojs-reveal] { opacity: 1 !important; transform: none !important; }
        .hero-subtitle { height: auto !important; overflow: visible !important; }
      </style> -->
    </noscript>
  </div>
</template>

<style scoped>
.app-frame {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: calc(100dvh - 2 * var(--frame-gutter));
  margin: var(--frame-gutter);
  background: var(--loli-surface);
  /* Clip children to the rounded rectangle so full-bleed backgrounds (nav
     frost band, page sections) cannot square off the frame corners; clip
     (unlike hidden) keeps the sticky nav pinned to the viewport. */
  overflow: clip;
  border-radius: var(--frame-radius);
}

/* Rounded-window mask, faithful to the reference site's #page-frame: a
   fixed rounded rectangle whose 100vmax spread shadow paints the crust
   color over everything outside it. Fixed-position children (loading
   screen, dialogs, physics overlays) escape .app-frame's clip, so
   the mask is what keeps their edges inside the rounded frame too. */
.page-frame {
  pointer-events: none;
  position: fixed;
  top: var(--frame-gutter);
  left: var(--frame-gutter);
  right: var(--frame-gutter);
  height: calc(100dvh - 2 * var(--frame-gutter));
  z-index: 110;
  border-radius: var(--frame-radius);
  box-shadow: 0 0 0 100vmax var(--loli-crust);
  transition: box-shadow 300ms;
}

.scroll-progress {
  position: fixed;
  top: var(--frame-gutter);
  left: var(--frame-gutter);
  right: var(--frame-gutter);
  height: 2px;
  background: var(--url-color);
  border-radius: var(--frame-radius) var(--frame-radius) 0 0;
  transform: scaleX(0);
  transform-origin: left center;
  pointer-events: none;
  z-index: 120;
}
</style>

<style>
/* .reveal is applied to elements inside child components, so the hidden
state must be global (not scoped). useReveal removes the class when the
reveal plays. */
.reveal {
  opacity: 0;
  transform: translateY(50px);
}

@media (prefers-reduced-motion: reduce) {
  .reveal {
    opacity: 1;
    transform: none;
  }
}
</style>
