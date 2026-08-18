<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, buildCounterUrl, buildEmbedFormats } = useApi()

const themes = ref<string[]>([])
const fthemes = ref<string[]>([])
const selectedTheme = ref('lian')

onMounted(async () => {
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
})

const previewUrl = computed(() =>
  buildCounterUrl({ name: 'demo', theme: selectedTheme.value, number: 0 }),
)

// Playground state (merged into the single page).
const state = reactive<ParamState>({
  name: 'my-counter',
  theme: 'lian',
  ftheme: '',
  fsize: 16,
  scale: 1,
  unshowf: false,
  x: undefined,
  y: undefined,
  rx: undefined,
  ry: undefined,
})

const onUpdate = (patch: Partial<ParamState>) => Object.assign(state, patch)
const counterUrl = computed(() => buildCounterUrl(state))
</script>

<template>
  <main class="max-w-6xl mx-auto px-4 py-8 font-sans">
    <!-- Hero -->
    <section id="top" class="mb-12 text-center">
      <h1 class="text-5xl font-bold text-loli-pink mb-3">Lolicount</h1>
      <p class="text-gray-600">萌系可换肤 SVG 访问计数器,往 README 贴一行链接即可计数。</p>
    </section>


    <!-- Theme gallery -->
    <section id="themes" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">主题选择</h2>
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

    <!-- Playground -->
    <section id="playground" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">🎨 Playground</h2>
      <div class="grid md:grid-cols-2 gap-8">
        <div>
          <h3 class="text-lg font-medium mb-3">参数</h3>
          <ParamPanel :state="state" :themes="themes" :fthemes="fthemes" @update="onUpdate" />
        </div>
        <div>
          <h3 class="text-lg font-medium mb-3">预览</h3>
          <BgPreview :url="counterUrl" :width="400" @drag="(x, y) => onUpdate({ x, y })" />
          <p class="text-xs text-gray-500 mt-2">在预览图上拖拽设置像素 x/y。</p>
        </div>
      </div>
    </section>

    <!-- Embed formats -->
    <section id="embed" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">📦 嵌入方式</h2>
      <LinkOutput :url="counterUrl" :name="state.name" />
    </section>

    <BackToTop />
  </main>
</template>
