<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()

defineProps<{
  // When true, renders in a compact inline style suited for the nav bar.
  nav?: boolean
}>()

const fullLabels: Record<string, string> = {
  zh: '中文',
  en: 'English',
  jp: '日本語',
}

// Mobile header space is tight, so the dropdown collapses into a single
// 40px icon button that cycles the locale (the reference's 文A button).
const cycleLocale = (): void => {
  const idx = locales.indexOf(locale.value)
  setLocale(locales[(idx + 1) % locales.length])
}
</script>

<template>
  <!-- Nav mode: a single locale-cycling icon button at every width,
       mirroring the reference site's 文A header control. -->
  <button
    v-if="nav"
    type="button"
    class="lang-cycle"
    data-ripple
    :title="fullLabels[locale]"
    :aria-label="fullLabels[locale]"
    @click="cycleLocale"
  >
    <svg class="lang-cycle-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <text x="1" y="10.5" font-size="10.5" font-weight="700" fill="currentColor">文</text>
      <text x="12" y="20" font-size="11.5" font-weight="700" fill="currentColor">A</text>
    </svg>
  </button>

  <!-- Legacy page-inline mode (kept for backward compatibility). -->
  <div
    v-else
    class="flex justify-end gap-1.5 mb-4"
  >
    <button
      v-for="l in locales"
      :key="l"
      type="button"
      class="flex items-center justify-center w-8 h-8 rounded-full text-xs font-semibold border transition-all relative overflow-hidden"
      :class="l === locale
        ? 'text-white bg-[var(--loli-pink)] border-[var(--loli-pink)]'
        : 'text-gray-500 bg-transparent border-[var(--loli-cream)] hover:text-[var(--loli-pink)] hover:border-[var(--loli-pink)]'"
      data-ripple
      @click="setLocale(l)"
    >{{ fullLabels[l] }}</button>
  </div>
</template>

<style scoped>
/* useRippleFx appends an absolutely positioned span, so every ripple host
   must be its own containing block and clip the circle to the button. */
.lang-cycle {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  background: transparent;
  border: 0;
  border-radius: 0.75rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.lang-cycle:hover {
  background: var(--loli-selection);
  color: var(--text-color);
}

.lang-cycle-icon {
  font-family: 'helvetica-lt-pro', sans-serif;
}

</style>
