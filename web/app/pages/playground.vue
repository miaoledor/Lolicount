<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, buildCounterUrl, buildEmbedFormats } = useApi()

const themes = ref<string[]>([])
const fthemes = ref<string[]>([])
onMounted(async () => {
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
})

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
const embeds = computed(() => buildEmbedFormats(counterUrl.value, state.name))
</script>

<template>
  <main class="max-w-6xl mx-auto px-4 py-8 font-sans">
    <h1 class="text-3xl font-bold text-loli-pink mb-6">🎨 Playground</h1>
    <div class="grid md:grid-cols-2 gap-8">
      <section>
        <h2 class="text-xl font-semibold mb-3">参数</h2>
        <ParamPanel :state="state" :themes="themes" :fthemes="fthemes" @update="onUpdate" />
      </section>
      <section>
        <h2 class="text-xl font-semibold mb-3">预览</h2>
        <BgPreview :url="counterUrl" :width="400" @drag="(x, y) => onUpdate({ x, y })" />
        <p class="text-xs text-gray-500 mt-2">在预览图上拖拽设置像素 x/y。</p>
        <div class="mt-6">
          <h3 class="text-lg font-medium mb-2">嵌入链接</h3>
          <LinkOutput :url="counterUrl" :name="state.name" />
        </div>
      </section>
    </div>
  </main>
</template>
