<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()

defineProps<{
  // When true, renders in a compact inline style suited for the nav bar
  // (no bottom margin, no right alignment).
  nav?: boolean
}>()

const shortLabels: Record<string, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}
</script>

<template>
  <!-- Nav mode: a single segmented button styled like the GitHub Star
       button next to it. Three equal segments; the active one is filled. -->
  <div
    v-if="nav"
    class="lang-segmented"
  >
    <button
      v-for="l in locales"
      :key="l"
      type="button"
      class="lang-segment"
      :class="{ 'lang-segment-active': l === locale }"
      @click="setLocale(l)"
    >{{ shortLabels[l] }}</button>
  </div>

  <!-- Legacy page-inline mode (kept for backward compatibility). -->
  <div
    v-else
    class="flex justify-end gap-1.5 mb-4"
  >
    <button
      v-for="l in locales"
      :key="l"
      type="button"
      class="flex items-center justify-center w-8 h-8 rounded-full text-xs font-semibold border transition-all"
      :class="l === locale
        ? 'text-white bg-[var(--loli-pink)] border-[var(--loli-pink)]'
        : 'text-gray-500 bg-transparent border-[var(--loli-cream)] hover:text-[var(--loli-pink)] hover:border-[var(--loli-pink)]'"
      @click="setLocale(l)"
    >{{ shortLabels[l] }}</button>
  </div>
</template>

<style scoped>
.lang-segmented {
  display: inline-flex;
  align-items: stretch;
  height: 2rem;
  padding: 0;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #4b5563;
  background: var(--loli-cream);
  border: 1px solid var(--loli-pink);
  border-radius: 0.5rem;
  overflow: hidden;
  white-space: nowrap;
}
.lang-segment {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.25rem;
  padding: 0 0.625rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #6b7280;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 0.2s, background 0.2s;
}
.lang-segment + .lang-segment {
  border-left: 1px solid var(--loli-pink);
}
.lang-segment:hover:not(.lang-segment-active) {
  color: var(--loli-pink);
}
.lang-segment-active {
  color: #fff;
  background: var(--loli-pink);
}
</style>
