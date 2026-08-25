<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, fetchConfig, buildCounterUrl, publicBase } = useApi()
const { t } = useI18n()

const themes = ref<ThemeInfo[]>([])
const fthemes = ref<string[]>([])

// Card-gallery refresh keys: clicking a card re-loads that card's image
// (M9) instead of selecting it as the playground theme. Each card holds
// its own cache-buster so only that image is re-fetched. The preview uses
// mode=random so a re-load actually shows a different frame (demo's
// FrameIndex is otherwise fixed at 0).
const cardKeys = reactive<Record<string, number>>({})

// Card-gallery large-card view: a dropdown picks the active card theme,
// the big card preview reloads its image on click (like the character
// theme section). Defaults to the first frame theme once loaded.
const selectedCard = ref('')

onMounted(async () => {
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
  // Default the card-theme picker to "wenders" (the project default, same
  // as the playground's state.theme) when available; fall back to the
  // first frame theme otherwise.
  if (!selectedCard.value) {
    const wenders = frameThemes.value.find((tth) => tth.name === 'wenders')
    selectedCard.value = wenders ? wenders.name : (frameThemes.value[0]?.name ?? '')
  }
  await fetchConfig()
})

const cardUrl = (name: string) => {
  const key = cardKeys[name] ?? 0
  const url = buildCounterUrl({ name: 'demo', theme: name, number: 0, unshowf: true, mode: 'random' })
  return key > 0 ? `${url}&_=${key}` : url
}

const reloadCard = (name: string) => {
  cardKeys[name] = (cardKeys[name] ?? 0) + 1
}

const reloadSelectedCard = () => {
  if (selectedCard.value) reloadCard(selectedCard.value)
}

// Playground state (merged into the single page, M7.5). kind is stored
// explicitly so the type picker works before the theme list loads.
const state = reactive<ParamState>({
  name: '',
  kind: 'frame',
  theme: 'wenders',
  ftheme: '',
  fsize: 16,
  scale: 1,
  unshowf: true,
  x: undefined,
  y: undefined,
  rx: undefined,
  ry: undefined,
  mode: 'seq',
  number: 0,
})

const onUpdate = (patch: Partial<ParamState>) => Object.assign(state, patch)

const nameEmpty = computed(() => !state.name.trim())

// Collapsible sections: loli (character themes), themes (card themes),
// and about (more) are collapsed by default; click the header to toggle.
const loliExpanded = ref(false)
const themesExpanded = ref(false)
const aboutExpanded = ref(false)

// About gallery: static showcase images from /public/images/.
// Add images to web/public/images/ and list them here.
const aboutImages = [
  { src: '/images/nbg2.png', alt: 'nbg2' },
]

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

const starBurst = ref<{ trigger: (x: number, y: number) => void } | null>(null)

const generate = (e: MouseEvent) => {
  // Guard: require a non-empty counter name before generating.
  const trimmed = state.name.trim()
  if (!trimmed) {
    return
  }
  // Star burst from the click point.
  starBurst.value?.trigger(e.clientX, e.clientY)
  // Character themes are always random; coerce mode so the URL stays
  // consistent with what the back-end will actually do.
  const params: ParamState = { ...state }
  params.name = trimmed
  if (state.kind === 'character') params.mode = 'random'
  // Embed links use the configured public domain (BASE_URL) so users paste
  // the real origin into READMEs; the live preview stays same-origin to
  // avoid cross-origin issues.
  generatedUrl.value = buildCounterUrl(params, publicBase.value)
  generatedName.value = trimmed
  const preview = buildCounterUrl(params)
  generateKey.value += 1
  const sep = preview.includes('?') ? '&' : '?'
  previewUrl.value = `${preview}${sep}_=_${generateKey.value}`
}

const frameThemes = computed(() => themes.value.filter((tth) => tth.kind === 'frame'))

// How-to-embed example URL. Uses the literal name "name" and the public
// domain so the sample links users copy point at the real origin once
// publicBase resolves (via fetchConfig on mount). Same builder as the
// playground, so the format stays consistent.
const howToUrl = computed(() =>
  buildCounterUrl({ name: 'name' }, publicBase.value),
)
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">
    <!-- Hero -->
    <section id="top" class="mb-12 text-center">
      <h1 class="text-5xl font-bold text-loli-pink mb-3 flex items-center justify-center gap-3">
        <img src="/images/lolicount-icon.png" alt="Lolicount" class="h-12 w-12" />
        {{ t('hero.title') }}
      </h1>
      <p class="text-gray-600">{{ t('app.desc') }}</p>
    </section>

    <!-- How to use -->
    <section id="howto" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('howto.title') }}</h2>
      <p class="text-sm text-gray-600 mb-4">
        {{ t('howto.introPre') }}<a href="#playground" class="text-loli-pink underline">{{ t('howto.introLink') }}</a>{{ t('howto.introPost') }}
      </p>
      <p class="text-sm text-gray-500 mb-2">{{ t('howto.mdHint') }} ![name]({{ howToUrl }})</p>
      <pre class="text-sm text-gray-500 mb-2"></pre>
    </section>

    <!-- Random Loli character (M9) -->
    <section id="loli" class="mb-16 scroll-mt-20">
      <h2
        class="text-2xl font-semibold mb-4 cursor-pointer select-none flex items-center gap-2"
        @click="loliExpanded = !loliExpanded"
      >
        <span class="loli-toggle-icon">{{ loliExpanded ? '▼' : '▶' }}</span>
        {{ t('loli.title') }}
      </h2>
      <div v-show="loliExpanded">
        <p class="text-sm text-gray-500 mb-4">{{ t('loli.desc') }}</p>
        <div class="flex justify-center rounded-xl bg-loli-cream py-8">
          <LoliCharacter />
        </div>
      </div>
    </section>

    <!-- Card theme: dropdown to pick a theme, big card preview with
         click-to-reload (mirrors the character theme section). -->
    <section id="themes" class="mb-16 scroll-mt-20">
      <h2
        class="text-2xl font-semibold mb-4 cursor-pointer select-none flex items-center gap-2"
        @click="themesExpanded = !themesExpanded"
      >
        <span class="loli-toggle-icon">{{ themesExpanded ? '▼' : '▶' }}</span>
        {{ t('themes.title') }}
      </h2>
      <div v-show="themesExpanded">
        <p class="text-sm text-gray-500 mb-4">{{ t('themes.desc') }}</p>
        <div class="grid md:grid-cols-[200px_1fr] gap-8 items-start">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('themes.select') }}</label>
          <select
            v-model="selectedCard"
            class="w-full border rounded px-2 py-1"
          >
            <option v-for="tth in frameThemes" :key="tth.name" :value="tth.name">{{ tth.name }}</option>
          </select>
          <p class="text-xs text-gray-500 mt-2">{{ t('themes.reloadHint') }}</p>
        </div>
        <div class="flex justify-center rounded-xl bg-loli-cream py-8">
          <div
            v-if="selectedCard"
            class="cursor-pointer"
            :title="t('themes.reload')"
            @click="reloadSelectedCard"
          >
            <img
              :src="cardUrl(selectedCard)"
              :alt="selectedCard"
              class="max-h-80 object-contain"
            />
          </div>
          <div v-else class="h-80 flex items-center justify-center text-sm text-gray-400">
            {{ t('loli.loading') }}
          </div>
        </div>
      </div>
      </div>
    </section>

    <!-- Playground -->
    <section id="playground" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4 flex items-center gap-2">
        <img src="/images/lolicount-icon.png" alt="" class="h-7 w-7" />
        {{ t('playground.title') }}
      </h2>
      <ParamPanel
        :state="state"
        :themes="themes"
        :fthemes="fthemes"
        @update="onUpdate"
      />
      <div class="relative mt-4">
        <StarBurst ref="starBurst" />
        <button
          :disabled="nameEmpty"
          :class="cn(
            'relative w-full py-2 rounded-lg font-medium transition',
            nameEmpty
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-loli-pink text-white hover:bg-loli-pink/90'
          )"
          @click="generate($event)"
        >
          {{ nameEmpty ? t('param.nameEmpty') : t('playground.generate') }}
        </button>
      </div>
      <!-- Result: preview image + embed formats, shown after generation. -->
      <div v-if="generatedUrl" class="mt-6 space-y-4">
        <div class="rounded-xl bg-loli-cream p-4 flex justify-center">
          <BgPreview :url="previewUrl" :width="400" />
        </div>
        <h3 class="text-lg font-medium flex items-center gap-2">
          <img src="/images/lolicount-icon.png" alt="" class="h-5 w-5" />
          {{ t('embed.title') }}
        </h3>
        <LinkOutput :url="generatedUrl" :name="generatedName" />
      </div>
      <div v-else class="mt-6 rounded-xl bg-loli-cream p-4">
        <div class="h-40 flex flex-col items-center justify-center text-center text-sm text-gray-400">
          <p>{{ t('playground.emptyHint1') }}</p>
          <p>{{ t('playground.emptyHint2') }}</p>
        </div>
      </div>
    </section>

    <!-- More -->
    <section id="about" class="mb-16 scroll-mt-20">
      <h2
        class="text-2xl font-semibold mb-4 cursor-pointer select-none flex items-center gap-2"
        @click="aboutExpanded = !aboutExpanded"
      >
        <span class="loli-toggle-icon">{{ aboutExpanded ? '▼' : '▶' }}</span>
        <img src="/images/lolicount-icon.png" alt="" class="h-7 w-7" />
        {{ t('about.title') }}
      </h2>
      <div v-show="aboutExpanded">
        <p class="text-sm text-gray-600 mb-6">{{ t('about.desc') }}</p>
        <ImageCarousel :images="aboutImages" />
      </div>
    </section>

    <Site-footer />
    <BackToTop />
  </main>
</template>

<style scoped>
.loli-toggle-icon {
  font-size: 0.9rem;
  color: var(--loli-pink);
  transition: transform 0.2s;
}
</style>
