<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'

const { fetchThemes, fetchFThemes, fetchConfig, buildCounterUrl, publicBase } = useApi()
const { t } = useI18n()

const themes = ref<ThemeInfo[]>([])
const fthemes = ref<string[]>([])

// Unified theme showcase: a single picker lists all themes (both
// single-layer and multi-layer). The preview re-loads on click with a
// cache-buster. The back-end always uses random frame selection so each
// shows a fresh frame/combination.
const showcaseKey = ref(0)
const selectedShowcase = ref('')

onMounted(async () => {
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
  // Default the showcase picker to "lian-ren" when available; fall back
  // to the first theme otherwise.
  if (!selectedShowcase.value) {
    const lianRen = themes.value.find((tth) => tth.name === 'lian-ren')
    selectedShowcase.value = lianRen ? lianRen.name : (themes.value[0]?.name ?? '')
  }
  await fetchConfig()
})

const showcaseUrl = computed(() => {
  if (!selectedShowcase.value) return ''
  const base = buildCounterUrl({
    name: 'demo',
    theme: selectedShowcase.value,
    number: 0,
    unshowf: true,
  })
  const key = showcaseKey.value
  return key > 0 ? `${base}&_=${key}` : base
})

const reloadShowcase = () => {
  showcaseKey.value++
}

// Playground state (merged into the single page, M7.5).
const state = reactive<ParamState>({
  name: '',
  theme: 'wenders',
  ftheme: '',
  fsize: 16,
  scale: 1,
  unshowf: true,
  x: undefined,
  y: undefined,
  rx: undefined,
  ry: undefined,
  number: 0,
  text: '{n}',
})

const onUpdate = (patch: Partial<ParamState>) => Object.assign(state, patch)

const showcaseVariants = computed(() => {
  const found = themes.value.find((tth) => tth.name === selectedShowcase.value)
  return found?.variants ?? 0
})

const nameEmpty = computed(() => !state.name.trim())

// Collapsible sections: themes (all unified), about (more),
// and about (more) are collapsed by default; click the header to toggle.
const themesExpanded = ref(true)

// Animated (emote) models render via the WebGL widget, not an SVG image —
// keep them out of the image showcase but list them in the playground
// theme picker with their marker.
const showcaseThemes = computed(() => themes.value.filter((tth) => !tth.animated))

const isAnimatedTheme = (name: string) =>
  themes.value.some((tth) => tth.name === name && tth.animated)

// M9: Generate it! — the preview is only (re)generated on click, and the
// result + embed formats are shown below the button.
// generatedUrl is the clean URL handed to LinkOutput (no cache-buster, so
// the copied embed code stays clean). previewUrl adds a per-click cache
// buster so clicking Generate repeatedly re-fetches the SVG even when the
// params are unchanged (needed for themes where the user
// expects a new image each click). M9.6.
const generatedUrl = ref('')
const generatedName = ref('')
const previewUrl = ref('')
const generateKey = ref(0)
const generatedAnimated = ref(false)

const starBurst = ref<{ trigger: (x: number, y: number) => void } | null>(null)

// Widget snippet origin: the public domain when configured, otherwise the
// current origin so the copied script src is always absolute.
const snippetOrigin = computed(() =>
  publicBase.value || (import.meta.client ? window.location.origin : ''),
)

// Embed snippet for animated themes: a <div> + <script> widget pair
// instead of an image URL (those themes have no SVG endpoint).
const widgetSnippet = computed(() => {
  if (!generatedAnimated.value || !generatedName.value) return ''
  const tpl = state.text && state.text !== '{n}'
    ? `\n  data-text="${state.text}"`
    : ''
  return `<div data-lolicount="${generatedName.value}"\n  data-model="${state.theme}"${tpl}>\n</div>\n<script src="${snippetOrigin.value}/widget/widget.js" defer><\/script>`
})

const generate = (e: MouseEvent) => {
  // Guard: require a non-empty counter name before generating.
  const trimmed = state.name.trim()
  if (!trimmed) {
    return
  }
  // Star burst from the click point.
  starBurst.value?.trigger(e.clientX, e.clientY)
  generateKey.value += 1
  generatedName.value = trimmed
  // Animated (emote) themes: live WebGL preview + widget snippet embed.
  // There is no SVG URL for them, so the URL-based formats are skipped.
  if (isAnimatedTheme(state.theme)) {
    generatedAnimated.value = true
    generatedUrl.value = ''
    previewUrl.value = ''
    return
  }
  generatedAnimated.value = false
  // All theme types share the same compose path; themes always use
  // random frame selection.
  const params: ParamState = { ...state }
  params.name = trimmed
  // Embed links use the configured public domain (BASE_URL) so users paste
  // the real origin into READMEs; the live preview stays same-origin to
  // avoid cross-origin issues.
  generatedUrl.value = buildCounterUrl(params, publicBase.value)
  const preview = buildCounterUrl(params)
  const sep = preview.includes('?') ? '&' : '?'
  previewUrl.value = `${preview}${sep}_=_${generateKey.value}`
}

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

    <!-- Unified theme showcase: all themes in one section -->
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
            v-model="selectedShowcase"
            class="w-full border rounded-lg px-3 py-2 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-loli-pink/40 focus:border-loli-pink cursor-pointer transition"
          >
            <option v-for="tth in showcaseThemes" :key="tth.name" :value="tth.name">
              {{ tth.name }}{{ tth.variants ? ` (${tth.variants.toLocaleString()})` : '' }}
            </option>
          </select>
          <p class="text-xs text-gray-500 mt-2">{{ t('themes.reloadHint') }}</p>
        </div>
        <div class="relative flex h-[27rem] items-center justify-center rounded-xl bg-loli-cream p-4">
          <div
            v-if="selectedShowcase"
            class="flex h-full w-full cursor-pointer items-center justify-center"
            :title="t('themes.reload')"
            @click="reloadShowcase"
          >
            <img
              :src="showcaseUrl"
              :alt="selectedShowcase"
              class="h-full w-full object-contain"
            />
          </div>
          <span
            v-if="showcaseVariants > 0"
            class="absolute bottom-2 right-2 rounded-full bg-black/60 text-white text-xs px-2 py-0.5"
          >{{ t('themes.variants', { n: showcaseVariants.toLocaleString() }) }}</span>
          <div v-if="!selectedShowcase" class="h-full w-full flex items-center justify-center text-sm text-gray-400">
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
              ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
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
      <!-- Animated (emote) result: live WebGL preview + widget snippet. -->
      <div v-else-if="generatedAnimated" class="mt-6 space-y-4">
        <div class="rounded-xl bg-loli-cream p-4 flex justify-center">
          <EmotePreview
            :key="generateKey"
            :model="state.theme"
            :name="generatedName"
            :text="state.text || '{n}'"
          />
        </div>
        <h3 class="text-lg font-medium flex items-center gap-2">
          <img src="/images/lolicount-icon.png" alt="" class="h-5 w-5" />
          {{ t('embed.title') }}
        </h3>
        <LinkOutput url="" :name="generatedName" :widget-snippet="widgetSnippet" />
        <p class="text-xs text-gray-500">{{ t('playground.animatedHint') }}</p>
      </div>
      <div v-else class="mt-6 rounded-xl bg-loli-cream p-4">
        <div class="h-40 flex flex-col items-center justify-center text-center text-sm text-gray-400">
          <p>{{ t('playground.emptyHint1') }}</p>
          <p>{{ t('playground.emptyHint2') }}</p>
        </div>
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
