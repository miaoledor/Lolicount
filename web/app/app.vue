<script setup lang="ts">
const { init, locale, setLocale, t, localeLabels, locales } = useI18n()
onMounted(init)

// Short single-token labels for the segmented control.
const shortLabels: Record<Locale, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}
</script>

<template>
  <div>
    <div
      class="fixed top-4 right-4 z-50 flex items-center overflow-hidden rounded-full bg-white/80 shadow backdrop-blur"
      :title="t('nav.lang')"
    >
      <button
        v-for="l in locales"
        :key="l"
        type="button"
        class="px-3 py-1 text-sm transition"
        :class="l === locale ? 'bg-loli-pink text-white font-semibold' : 'text-gray-600 hover:bg-gray-50'"
        @click="setLocale(l)"
      >
        {{ shortLabels[l] }}
      </button>
    </div>
    <NuxtPage />
  </div>
</template>
