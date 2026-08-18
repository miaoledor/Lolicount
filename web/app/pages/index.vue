<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, buildCounterUrl } = useApi()
const { t } = useI18n()

const themes = ref<ThemeInfo[]>([])
const fthemes = ref<string[]>([])

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

const cardUrl = (name: string) => {
  const key = cardKeys[name] ?? 0
  const url = buildCounterUrl({ name: 'demo', theme: name, number: 0, unshowf: true, mode: 'random' })
  return key > 0 ? `${url}&_=${key}` : url
}

const reloadCard = (name: string) => {
  cardKeys[name] = (cardKeys[name] ?? 0) + 1
}

// Playground state (merged into the single page, M7.5). kind is stored
// explicitly so the type picker works before the theme list loads.
const state = reactive<ParamState>({
  name: 'my-counter',
  kind: 'frame',
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
// generatedUrl is the clean URL handed to LinkOutput (no cache-buster, so
// the copied embed code stays clean). previewUrl adds a per-click cache
// buster so clicking Generate repeatedly re-fetches the SVG even when the
// params are unchanged (needed for random/character themes where the user
// expects a new image each click). M9.6.
const generatedUrl = ref('')
const generatedName = ref('')
const previewUrl = ref('')
const generateKey = ref(0)

const generate = () => {
  // Character themes are always random; coerce mode so the URL stays
  // consistent with what the back-end will actually do.
  const params: ParamState = { ...state }
  if (state.kind === 'character') params.mode = 'random'
  const clean = buildCounterUrl(params)
  generatedUrl.value = clean
  generatedName.value = state.name
  generateKey.value += 1
  const sep = clean.includes('?') ? '&' : '?'
  previewUrl.value = `${clean}${sep}_=_${generateKey.value}`
}

const frameThemes = computed(() => themes.value.filter((tth) => tth.kind === 'frame'))
</script>

<template>
  <main class="max-w-6xl mx-auto px-4 py-8 font-sans">
    <!-- Hero -->
    <section id="top" class="mb-12 text-center">
      <h1 class="text-5xl font-bold text-loli-pink mb-3">{{ t('hero.title') }}</h1>
      <p class="text-gray-600">{{ t('app.desc') }}</p>
    </section>

    <!-- Random Loli character (M9) -->
    <section id="loli" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('loli.title') }}</h2>
      <p class="text-sm text-gray-500 mb-4">{{ t('loli.desc') }}</p>
      <div class="flex justify-center rounded-xl bg-loli-cream py-8">
        <LoliCharacter />
      </div>
    </section>

    <!-- Theme gallery: clicking a card re-loads its image (M9). -->
    <section id="themes" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('themes.title') }}</h2>
      <p class="text-sm text-gray-500 mb-4">{{ t('themes.desc') }}</p>
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
        <button
          v-for="tth in frameThemes"
          :key="tth.name"
          class="border rounded-lg p-2 transition hover:shadow-md border-gray-200"
          :title="`${t('themes.reload')} ${tth.name}`"
          @click="reloadCard(tth.name)"
        >
          <img :src="cardUrl(tth.name)" :alt="tth.name" class="w-full h-24 object-contain" />
          <p class="text-center text-sm mt-1">{{ tth.name }}</p>
        </button>
      </div>
    </section>

    <!-- Playground -->
    <section id="playground" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('playground.title') }}</h2>
      <div class="grid md:grid-cols-2 gap-8">
        <div>
          <h3 class="text-lg font-medium mb-3">{{ t('playground.params') }}</h3>
          <ParamPanel
            :state="state"
            :themes="themes"
            :fthemes="fthemes"
            @update="onUpdate"
          />
          <button
            class="mt-4 w-full bg-loli-pink text-white py-2 rounded-lg font-medium hover:bg-loli-pink/90 transition"
            @click="generate"
          >
            {{ t('playground.generate') }}
          </button>
        </div>
        <div>
          <h3 class="text-lg font-medium mb-3">{{ t('playground.preview') }}</h3>
          <div v-if="generatedUrl" class="rounded-xl bg-loli-cream p-4">
            <BgPreview :url="previewUrl" :width="400" />
          </div>
          <div v-else class="rounded-xl bg-gray-50 p-8 text-center text-sm text-gray-400">
            {{ t('playground.previewPlaceholder') }}
          </div>
        </div>
      </div>
    </section>

    <!-- Embed formats: shown after generation (M9). -->
    <section v-if="generatedUrl" id="embed" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('embed.title') }}</h2>
      <LinkOutput :url="generatedUrl" :name="generatedName" />
    </section>

    <Site-footer />
    <BackToTop />
  </main>
</template>
