<template>
  <span class="m3-loading-indicator" role="status" :aria-label="label">
    <span class="m3-loading-indicator__rotor">
      <span class="m3-loading-indicator__shape"></span>
    </span>
    <span v-if="seconds > 0" class="m3-loading-indicator__seconds" aria-hidden="true">{{ seconds }}</span>
  </span>
</template>

<script setup lang="ts">
// Contained Material 3 Expressive loading indicator for short waits
// (200ms-5s). M3 uses a fully rounded container plus an active indicator that
// morphs through seven rounded shapes while rotating, rather than a plain
// ring spinner. The whole-second countdown sits outside the rotor so it never
// rotates with the shape. The component stays presentational: the preview
// page owns visibility, the 2s window and the seconds value.
withDefaults(defineProps<{ label?: string; seconds?: number }>(), {
  label: 'Loading',
  seconds: 0,
})
</script>

<style scoped>
/* 36px container with a 28px active indicator keeps M3's 48/38dp ratio.
   The container is opaque because the indicator sits over preview artwork. */
.m3-loading-indicator {
  position: absolute;
  right: 0.75rem;
  bottom: 0.75rem;
  z-index: 2;
  display: grid;
  place-items: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  background: color-mix(in srgb, var(--loli-pink) 20%, var(--loli-surface));
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--loli-pink) 32%, transparent),
    0 2px 6px color-mix(in srgb, var(--text-color) 18%, transparent);
  pointer-events: none;
}

.m3-loading-indicator__rotor {
  display: grid;
  place-items: center;
  width: 79%;
  height: 79%;
  animation: m3-rotate 2s linear infinite;
}

.m3-loading-indicator__shape {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: var(--loli-pink);
  animation: m3-morph 2s ease-in-out infinite;
}

/* The number stays upright above the morphing shape. A dark shadow keeps it
   legible against both the light and dark theme indicator colors. */
.m3-loading-indicator__seconds {
  position: absolute;
  z-index: 1;
  display: grid;
  place-items: center;
  width: 79%;
  height: 79%;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.55);
}

/* Seven rounded shapes in one loop, mirroring the M3 morph sequence
   (soft burst, cookie, pentagon, pill, sunny, cookie, oval) with the
   rotation supplied by the parent rotor. */
@keyframes m3-rotate {
  to { transform: rotate(360deg); }
}

@keyframes m3-morph {
  0%, 100% {
    border-radius: 50%;
    transform: scale(1);
  }
  14.28% {
    border-radius: 42% 58% 38% 62% / 58% 40% 60% 42%;
    transform: scale(0.88);
  }
  28.57% {
    border-radius: 32% 32% 32% 32% / 32% 32% 32% 32%;
    transform: scale(0.94);
  }
  42.86% {
    border-radius: 50% 18% 50% 18% / 18% 50% 18% 50%;
    transform: scale(0.84);
  }
  57.14% {
    border-radius: 38% 62% 42% 58% / 62% 38% 62% 38%;
    transform: scale(0.96);
  }
  71.43% {
    border-radius: 24% 76% 24% 76% / 76% 24% 76% 24%;
    transform: scale(0.86);
  }
  85.71% {
    border-radius: 50% 50% 50% 50% / 62% 38% 62% 38%;
    transform: scale(0.98);
  }
}

/* Reduced motion keeps the status visible as a static dot instead of
   removing the affordance entirely. */
@media (prefers-reduced-motion: reduce) {
  .m3-loading-indicator__rotor,
  .m3-loading-indicator__shape {
    animation: none;
  }

  .m3-loading-indicator__shape {
    transform: scale(0.88);
  }
}
</style>
