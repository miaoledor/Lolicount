<script setup lang="ts">
const { fetchThemes, buildCounterUrl } = useApi()
const themes = ref<string[]>([])
onMounted(async () => {
  themes.value = await fetchThemes()
})
</script>

<template>
  <main class="max-w-5xl mx-auto px-4 py-8 font-sans">
    <h1 class="text-3xl font-bold text-loli-pink mb-6">🖼️ 主题画廊</h1>
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
      <div v-for="t in themes" :key="t" class="border rounded-lg p-2">
        <img :src="buildCounterUrl({ name: 'demo', theme: t, number: 0, unshowf: true })" :alt="t" class="w-full h-32 object-contain" />
        <p class="text-center text-sm mt-1">{{ t }}</p>
      </div>
    </div>
  </main>
</template>
