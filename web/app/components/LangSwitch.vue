<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()

const shortLabels: Record<string, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}
</script>

<template>
  <div class="flex justify-end gap-1.5 mb-4">
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
