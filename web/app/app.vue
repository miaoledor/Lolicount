<script setup lang="ts">
const { init: initLocale, locale, setLocale, t, locales } = useI18n()
const { init: initTheme, theme, setTheme } = useTheme()
onMounted(() => {
  initLocale()
  initTheme()
})

const shortLabels: Record<Locale, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}

// Theme swatches: pink and gray, shown as filled circles. The active
// theme gets a ring so it is distinguishable from the language circles.
const themeOptions: { value: ThemeName; swatch: string }[] = [
  { value: 'pink', swatch: '#e91e63' },
  { value: 'gray', swatch: '#6b7280' },
]
</script>

<template>
  <div>
    <NavBar />
    <div class="fixed top-16 right-4 z-50 flex flex-col items-end gap-2">
      <!-- Theme picker -->
      <div class="flex items-center gap-1.5" :title="t('nav.theme')">
        <button
          v-for="opt in themeOptions"
          :key="opt.value"
          type="button"
          class="h-6 w-6 rounded-full transition shadow-sm ring-offset-2"
          :class="opt.value === theme ? 'ring-2 ring-gray-400' : 'opacity-70 hover:opacity-100'"
          :style="{ backgroundColor: opt.swatch }"
          @click="setTheme(opt.value)"
        />
      </div>
      <!-- Language picker -->
      <div class="flex items-center gap-2" :title="t('nav.lang')">
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
    </div>
    <div class="pt-14">
      <NuxtPage />
    </div>
  </div>
</template>
