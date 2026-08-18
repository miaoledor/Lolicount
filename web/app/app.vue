<script setup lang="ts">
const { init, locale, toggle, t } = useI18n()
onMounted(init)
</script>

<template>
  <div>
    <button
      class="fixed top-4 right-4 z-50 rounded-full bg-white/80 px-3 py-1 text-sm shadow backdrop-blur transition hover:bg-white"
      :title="t('nav.lang')"
      @click="toggle"
    >
      {{ locale === 'zh' ? 'EN' : '中' }}
    </button>
    <NuxtPage />
  </div>
</template>
