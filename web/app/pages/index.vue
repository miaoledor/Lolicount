<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, buildCounterUrl } = useApi()

const themes = ref<ThemeInfo[]>([])
const fthemes = ref<string[]>([])
const selectedTheme = ref('lian')

// Card-gallery refresh keys: clicking a card re-loads that card's image
// (M9) instead of selecting it as the playground theme. Each card holds
// its own cache-buster so only that image is re-fetched. The preview uses
// mode=random so a re-load actually shows a different frame (demo's
// FrameIndex is otherwise fixed at 0).
const cardKeys = reactive<Record<string, number>>({})

onMounted(async () => {
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
})

const themeByName = (name: string) => themes.value.find((t) => t.name === name)
const selectedKind = computed<'frame' | 'character'>(() => themeByName(state.theme)?.kind ?? 'frame')

const previewUrl = computed(() =>
  buildCounterUrl({ name: 'demo', theme: selectedTheme.value, number: 0 }),
)

const cardUrl = (name: string) => {
  const key = cardKeys[name] ?? 0
  const url = buildCounterUrl({ name: 'demo', theme: name, number: 0, unshowf: true, mode: 'random' })
  return key > 0 ? `${url}&_=${key}` : url
}

const reloadCard = (name: string) => {
  cardKeys[name] = (cardKeys[name] ?? 0) + 1
}

// Playground state (merged into the single page, M7.5).
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
  mode: 'seq',
})

const onUpdate = (patch: Partial<ParamState>) => Object.assign(state, patch)

// M9: Generate it! — the preview is only (re)generated on click, and the
// result + embed formats are shown below the button.
const generatedUrl = ref('')
const generatedName = ref('')

const generate = () => {
  // Character themes are always random; coerce mode so the URL stays
  // consistent with what the back-end will actually do.
  const params: ParamState = { ...state }
  if (selectedKind.value === 'character') params.mode = 'random'
  generatedUrl.value = buildCounterUrl(params)
  generatedName.value = state.name
}

const frameThemes = computed(() => themes.value.filter((t) => t.kind === 'frame'))
</script>

<template>
  <main class="max-w-6xl mx-auto px-4 py-8 font-sans">
    <!-- Hero -->
    <section id="top" class="mb-12 text-center">
      <h1 class="text-5xl font-bold text-loli-pink mb-3">Lolicount</h1>
      <p class="text-gray-600">萌系可换肤 SVG 访问计数器,往 README 贴一行链接即可计数。</p>
    </section>

    <!-- Random Loli character (M9) -->
    <section id="loli" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">立绘主题</h2>
      <p class="text-sm text-gray-500 mb-4">该主题由多个部件组成,如服装、腮红、表情等</p>
      <div class="flex justify-center rounded-xl bg-loli-cream py-8">
        <LoliCharacter />
      </div>
    </section>

    <!-- Theme gallery: clicking a card re-loads its image (M9). -->
    <section id="themes" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">卡片主题</h2>
      <p class="text-sm text-gray-500 mb-4">卡片主题每张展示只有一张图片构成,点击图片可重新加载</p>
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
        <button
          v-for="t in frameThemes"
          :key="t.name"
          class="border rounded-lg p-2 transition hover:shadow-md border-gray-200"
          :title="`重新加载 ${t.name}`"
          @click="reloadCard(t.name)"
        >
          <img :src="cardUrl(t.name)" :alt="t.name" class="w-full h-24 object-contain" />
          <p class="text-center text-sm mt-1">{{ t.name }}</p>
        </button>
      </div>
    </section>

    <!-- Playground -->
    <section id="playground" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">🎨 Playground</h2>
      <div class="grid md:grid-cols-2 gap-8">
        <div>
          <h3 class="text-lg font-medium mb-3">参数</h3>
          <ParamPanel
            :state="state"
            :themes="themes"
            :fthemes="fthemes"
            :kind="selectedKind"
            @update="onUpdate"
          />
          <button
            class="mt-4 w-full bg-loli-pink text-white py-2 rounded-lg font-medium hover:bg-loli-pink/90 transition"
            @click="generate"
          >
            Generate it!
          </button>
        </div>
        <div>
          <h3 class="text-lg font-medium mb-3">预览</h3>
          <div v-if="generatedUrl" class="rounded-xl bg-loli-cream p-4">
            <BgPreview :url="generatedUrl" :width="400" @drag="(x, y) => onUpdate({ x, y })" />
            <p class="text-xs text-gray-500 mt-2">在预览图上拖拽设置像素 x/y。</p>
          </div>
          <div v-else class="rounded-xl bg-gray-50 p-8 text-center text-sm text-gray-400">
            选择参数后点击 Generate it! 生成预览
          </div>
        </div>
      </div>
    </section>

    <!-- Embed formats: shown after generation (M9). -->
    <section v-if="generatedUrl" id="embed" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">📦 嵌入方式</h2>
      <LinkOutput :url="generatedUrl" :name="generatedName" />
    </section>

    <BackToTop />
  </main>
</template>
