<script setup lang="ts">
const { init: initLocale } = useI18n()
const { init: initTheme } = useTheme()
onMounted(() => {
  initLocale()
  initTheme()
})
</script>

<template>
  <div>
    <NavBar />
    <div class="pt-16">
      <NuxtPage />
    </div>
  </div>
</template>
