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
</script>

<template>
  <!-- Nav mode: a compact dropdown selector. -->
  <select
    v-if="nav"
    :value="locale"
    class="lang-select"
    @change="setLocale(($event.target as HTMLSelectElement).value as any)"
  >
    <option v-for="l in locales" :key="l" :value="l">{{ fullLabels[l] }}</option>
  </select>

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
    >{{ fullLabels[l] }}</button>
  </div>
</template>

<style scoped>
.lang-select {
  height: 2rem;
  padding: 0 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #4b5563;
  background: var(--loli-cream);
  border: 1px solid var(--loli-pink);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.lang-select:focus {
  outline: none;
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--loli-pink) 40%, transparent);
}
</style>
