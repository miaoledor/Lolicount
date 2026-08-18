<script setup lang="ts">
const { fetchThemes, buildCounterUrl } = useApi()

const themes = ref<string[]>([])
const selectedTheme = ref('lian')

onMounted(async () => {
  themes.value = await fetchThemes()
})

const previewUrl = computed(() =>
  buildCounterUrl({
    name: 'demo',
    theme: selectedTheme.value,
    number: 0,
  }),
)
</script>

<template>
  <main class="max-w-5xl mx-auto px-4 py-8 font-sans">
    <h1 class="text-4xl font-bold text-loli-pink mb-2">🎀 Lolicount</h1>
    <p class="text-gray-600 mb-8">萌系可换肤 SVG 访问计数器,往 README 贴一行链接即可计数。</p>

    <section class="mb-12">
      <h2 class="text-2xl font-semibold mb-4">主题市场</h2>
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
        <button
          v-for="t in themes"
          :key="t"
          :class="cn(
            'border rounded-lg p-2 transition hover:shadow-md',
            selectedTheme === t ? 'border-loli-pink bg-loli-cream' : 'border-gray-200',
          )"
          @click="selectedTheme = t"
        >
          <img :src="buildCounterUrl({ name: 'demo', theme: t, number: 0, unshowf: true })" :alt="t" class="w-full h-24 object-contain" />
          <p class="text-center text-sm mt-1">{{ t }}</p>
        </button>
      </div>
    </section>

    <section class="mb-12">
      <h2 class="text-2xl font-semibold mb-4">预览</h2>
      <img :src="previewUrl" alt="preview" class="mx-auto" />
    </section>

    <section class="mb-12">
      <h2 class="text-2xl font-semibold mb-4">三种嵌入方式</h2>
      <pre class="bg-gray-50 p-4 rounded text-sm overflow-x-auto">{{ previewUrl }}
&lt;img src="{{ previewUrl }}" alt="my-counter" />
![my-counter]({{ previewUrl }})</pre>
      <p class="text-sm text-gray-500 mt-2">前往 <NuxtLink to="/playground" class="text-loli-pink underline">Playground</NuxtLink> 调整参数。</p>
    </section>
  </main>
</template>
