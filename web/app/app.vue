<script setup lang="ts">
const { init, locale, setLocale, t, locales } = useI18n()
onMounted(init)

const shortLabels: Record<Locale, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}
</script>

<template>
  <div>
    <div
      class="fixed top-4 right-4 z-50 flex items-center gap-2"
      :title="t('nav.lang')"
    >
      <button
        v-for="l in locales"
        :key="l"
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-full text-xs transition shadow-sm"
        :class="l === locale ? 'bg-loli-pink text-white font-semibold' : 'bg-white/80 text-gray-600 hover:bg-loli-cream'"
        @click="setLocale(l)"
      >
        {{ shortLabels[l] }}
      </button>
    </div>
    <NuxtPage />
  </div>
</template>
