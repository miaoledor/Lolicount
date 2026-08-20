<script setup lang="ts">
const { init, locale, toggle, t, localeLabels } = useI18n()
onMounted(init)

// Label shown on the language button: the next locale's short label,
// so the button always tells the user what they will switch to.
const nextLabel = computed(() => {
  const order = ['zh', 'en', 'jp'] as const
  const idx = order.indexOf(locale.value as typeof order[number])
  const next = order[(idx + 1) % order.length]!
  return localeLabels[next]
})
</script>

<template>
  <div>
    <button
      class="fixed top-4 right-4 z-50 rounded-full bg-white/80 px-3 py-1 text-sm shadow backdrop-blur transition hover:bg-white"
      :title="t('nav.lang')"
      @click="toggle"
    >
      {{ nextLabel }}
    </button>
    <NuxtPage />
  </div>
</template>
