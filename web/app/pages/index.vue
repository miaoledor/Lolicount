<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'
import { gameMeta, themeMeta, buildSearchHaystack } from '~/utils/themeMeta'
import type { GameKey, ThemeKind } from '~/utils/themeMeta'

// The playground IS the landing page. /themes stays as a route alias so
// previously shared ?theme= links keep working.
definePageMeta({ alias: ['/themes'] })

const { fetchThemes, fetchFThemes, fetchHotThemes, fetchConfig, buildCounterUrl, publicBase } = useApi()
const { t, locale } = useI18n()

const themes = ref<ThemeInfo[]>([])
const fthemes = ref<string[]>([])

// Join the API theme list with local metadata. Themes missing metadata
// (e.g. added before the meta table is updated) still show up under
// the "other" game bucket.
const merged = computed(() =>
  themes.value.map((tth) => {
    const meta = themeMeta.find((m) => m.name === tth.name)
    return {
      ...tth,
      meta: meta ?? null,
      haystack: meta ? buildSearchHaystack(meta) : tth.name.toLowerCase(),
    }
  }),
)

// Filters + search state. The game filter dropdown was replaced by
// per-source collapsible groups below.
const kindFilter = ref<ThemeKind | 'all'>('all')
const searchQuery = ref('')

const filteredThemes = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return merged.value.filter((tth) => {
    if (kindFilter.value !== 'all' && tth.meta?.kind !== kindFilter.value) return false
    if (q && !tth.haystack.includes(q)) return false
    return true
  })
})

// Game filter options: only games that actually have themes, "other" last.
// Drives the group ordering of the collapsible source sections.
const gameOptions = computed(() => {
  const present = new Set(merged.value.map((tth) => tth.meta?.gameKey ?? 'other'))
  const keys = Object.keys(gameMeta) as GameKey[]
  return keys
    .filter((key) => present.has(key))
    .sort((a, b) => (a === 'other' ? 1 : b === 'other' ? -1 : keys.indexOf(a) - keys.indexOf(b)))
})

const gameLabel = (key: GameKey) => {
  const lang = (['zh', 'en', 'jp'] as const).includes(locale.value as 'zh' | 'en' | 'jp')
    ? (locale.value as 'zh' | 'en' | 'jp')
    : 'en'
  return gameMeta[key].label[lang]
}

const resultCount = computed(() => filteredThemes.value.length)

// Collapsible source groups: a hot group (server's usage-based top 10,
// refreshed at startup and hourly) followed by one group per game.
// Themes surfaced in the hot group are excluded from their game group.
type ThemeGroup = { key: string; label: string; hot: boolean; items: typeof merged.value }

const hotNames = ref<string[]>([])
const filtersActive = computed(() => kindFilter.value !== 'all' || searchQuery.value.trim().length > 0)

const groups = computed<ThemeGroup[]>(() => {
  const out: ThemeGroup[] = []
  const consumed = new Set<string>()
  const hot = hotNames.value
    .map((name) => filteredThemes.value.find((tth) => tth.name === name))
    .filter((x): x is NonNullable<typeof x> => !!x)
  if (hot.length > 0) {
    out.push({ key: '__hot', label: t('themesGallery.hotGroup'), hot: true, items: hot })
    hot.forEach((x) => consumed.add(x.name))
  }
  for (const key of gameOptions.value) {
    const items = filteredThemes.value.filter(
      (tth) => !consumed.has(tth.name) && (tth.meta?.gameKey ?? 'other') === key,
    )
    if (items.length > 0) {
      out.push({ key, label: gameLabel(key), hot: false, items })
    }
  }
  return out
})

// Group expand state: collapsed by default (spec). While search/type
// filters are active, matching groups auto-expand so results stay
// visible without extra clicks.
const expandedGroups = reactive(new Set<string>())
const toggleGroup = (key: string) => {
  if (expandedGroups.has(key)) expandedGroups.delete(key)
  else expandedGroups.add(key)
}
const isGroupOpen = (key: string) => filtersActive.value || expandedGroups.has(key)

// Selected theme + preview with a cache-buster (same reload trick as the
// home page showcase: the back-end picks a random frame per request).
// Animated (emote) themes render through the shared WebGL preview instead
// of the static SVG image endpoint.
const selectedTheme = ref('')
const previewKey = ref(0)

// The selected theme lives in the URL as ?theme=<name> so a refresh or a
// shared link restores the exact selection. Written on every user-driven
// selection (replace, not push — no history spam), read after the theme
// list loads in onMounted and in the query watcher below (Nuxt reuses
// this component on query-only navigation, so onMounted does not rerun).
const route = useRoute()
const router = useRouter()

// Single writer for selection state: preview + playground stay in sync.
// Does not touch the URL — the plain default selection keeps a clean URL.
const applyThemeSelection = (name: string) => {
  selectedTheme.value = name
  state.theme = name
}

// Default selection: lian-ren when present, else the first theme.
const defaultThemeName = () => {
  const lianRen = themes.value.find((tth) => tth.name === 'lian-ren')
  return lianRen ? lianRen.name : (themes.value[0]?.name ?? '')
}

// Soft-navigation support: picking up ?theme= when the query changes
// while the page stays mounted. Skips unknown names; the writer path
// (selectTheme) is a no-op here because the name already matches. When
// the param is removed (e.g. the navbar links to plain /themes), the
// selection falls back to the default so the clean URL and the preview
// stay consistent with a fresh visit.
watch(
  () => route.query.theme,
  (q) => {
    if (themes.value.length === 0) return
    const name = typeof q === 'string' ? q : ''
    if (name === selectedTheme.value) return
    if (name && !themes.value.some((tth) => tth.name === name)) return
    applyThemeSelection(name || defaultThemeName())
  },
)

const selectedAnimated = computed(() =>
  themes.value.some((tth) => tth.name === selectedTheme.value && tth.animated),
)

// Which animated renderer the selected theme uses: "psb" (E-mote), "spine" or
// "live2d". Absent for static card/character themes. Drives the preview
// component choice (EmotePreview for psb/spine, Live2DPreview for live2d).
const selectedKind = computed(() =>
  themes.value.find((tth) => tth.name === selectedTheme.value)?.kind ?? '',
)

const selectTheme = (name: string) => {
  applyThemeSelection(name)
  router.replace({ query: { ...route.query, theme: name } })
}

// True while the static preview <img> is fetching a fresh SVG. The
// overlay blocks reload clicks and disables the frame download, so a
// slow no-store fetch can't hand the user the previous theme's frame.
const previewLoading = ref(true)
watch([selectedTheme, previewKey], () => {
  if (!selectedAnimated.value) previewLoading.value = true
})

const previewUrl = computed(() => {
  if (!selectedTheme.value || selectedAnimated.value) return ''
  const base = buildCounterUrl({
    name: 'demo',
    theme: selectedTheme.value,
    number: 0,
    unshowf: true,
  })
  const key = previewKey.value
  return key > 0 ? `${base}&_=${key}` : base
})

// Download the EXACT frame currently displayed. The preview URL renders a
// fresh random frame on every fetch (no-store + per-request PRNG seed), so
// a plain <a download> would save a different frame — instead the loaded
// <img> is rasterized through a canvas and exported as a PNG. Static themes
// only: animated previews are WebGL canvases with no per-frame image
// endpoint.
//
// Supersampling: the SVG's intrinsic size is the ~400px-class display
// size, but it embeds the theme art at full source resolution (up to
// ~3500px). Drawing the SVG to a larger canvas re-rasterizes the layout
// at the target size and samples the embedded raster at its native
// detail, so the export lands near the source resolution instead of the
// display resolution. Factor targets a ~1600px long edge, clamped 2x..5x
// to bound canvas memory; small-source frame themes upscale noisily-free
// (vector text/shapes stay crisp, rasters go no further than they have).
const previewImgEl = ref<HTMLImageElement | null>(null)
const downloadFrame = () => {
  const img = previewImgEl.value
  if (!img || !img.naturalWidth || !img.naturalHeight) return
  if (previewLoading.value) return
  try {
    const long = Math.max(img.naturalWidth, img.naturalHeight)
    const factor = Math.min(5, Math.max(2, Math.ceil(1600 / long)))
    const canvas = document.createElement('canvas')
    canvas.width = img.naturalWidth * factor
    canvas.height = img.naturalHeight * factor
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
    // Synchronous toDataURL keeps the download inside the user-gesture
    // task — Safari blocks downloads from async callbacks (toBlob) after
    // the gesture window closes.
    const url = canvas.toDataURL('image/png')
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedTheme.value}-frame-${canvas.height}px.png`
    // Safari ignores clicks on detached anchors — mount before clicking.
    document.body.appendChild(a)
    a.click()
    a.remove()
  } catch {
    // Tainted-canvas SecurityError (cross-origin preview without CORS):
    // fall back to opening the render URL so the frame can be saved
    // manually.
    window.open(previewUrl.value, '_blank', 'noopener')
  }
}

const reloadPreview = () => {
  // Blocked while a fetch is in flight: the click would only queue
  // another random frame behind the pending one.
  if (previewLoading.value) return
  previewKey.value++
}

const selectedMeta = computed(() =>
  themeMeta.find((m) => m.name === selectedTheme.value) ?? null,
)

// Playground state, same shape as the home page so the generation flow
// (ParamPanel + generate + LinkOutput) stays identical.
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

const onUpdate = (patch: Partial<ParamState>) => {
  Object.assign(state, patch)
  if (patch.theme && patch.theme !== selectedTheme.value) {
    selectTheme(patch.theme)
    previewKey.value++
  }
}

const nameEmpty = computed(() => !state.name.trim())

const generatedUrl = ref('')
const generatedName = ref('')
const generatedPreviewUrl = ref('')
const generateKey = ref(0)

const starBurst = ref<{ trigger: (x: number, y: number) => void } | null>(null)

// Animated themes have no SVG endpoint: the result shows the live WebGL
// preview and a widget snippet instead of image URL formats.
const generatedAnimated = ref(false)
const generatedKind = ref('')
const widgetSnippet = computed(() => {
  if (!generatedAnimated.value || !generatedName.value) return ''
  const origin = publicBase.value || (import.meta.client ? window.location.origin : '')
  const tpl = state.text && state.text !== '{n}'
    ? `
  data-text="${state.text}"`
    : ''
  return `<div data-lolicount="${generatedName.value}"
  data-model="${state.theme}"${tpl}>
</div>
<script src="${origin}/widget/widget.js" defer></` + 'script>'
})

const generate = (e: MouseEvent) => {
  const trimmed = state.name.trim()
  if (!trimmed) return
  // Fire the fireworks from the button center so keyboard activation
  // (clientX/Y of 0) still bursts in the right place.
  const btnRect = (e.currentTarget as HTMLElement | null)?.getBoundingClientRect()
  starBurst.value?.trigger(
    btnRect ? btnRect.left + btnRect.width / 2 : e.clientX,
    btnRect ? btnRect.top + btnRect.height / 2 : e.clientY,
  )
  const params: ParamState = { ...state }
  params.name = trimmed
  const generatedTheme = themes.value.find((tth) => tth.name === params.theme)
  generatedAnimated.value = !!generatedTheme?.animated
  generatedKind.value = generatedTheme?.kind ?? ''
  generatedUrl.value = generatedAnimated.value ? '' : buildCounterUrl(params, publicBase.value)
  generatedName.value = trimmed
  const preview = buildCounterUrl(params)
  generateKey.value += 1
  const sep = preview.includes('?') ? '&' : '?'
  generatedPreviewUrl.value = `${preview}${sep}_=_${generateKey.value}`
}

onMounted(async () => {
  // SSG may finish the preview fetch before hydration attaches @load —
  // clear the spinner for an already-complete image.
  if (previewImgEl.value?.complete) previewLoading.value = false
  themes.value = await fetchThemes()
  fthemes.value = await fetchFThemes()
  // Hot list is non-critical: an API hiccup just hides the hot group.
  try {
    hotNames.value = await fetchHotThemes()
  } catch {
    hotNames.value = []
  }
  // Restore ?theme= from the URL; fall back to the default selection for
  // plain visits or unknown names (URL stays clean in the fallback).
  const fromQuery = typeof route.query.theme === 'string' ? route.query.theme : ''
  if (fromQuery && themes.value.some((tth) => tth.name === fromQuery)) {
    applyThemeSelection(fromQuery)
  } else {
    applyThemeSelection(defaultThemeName())
  }
  await fetchConfig()
})
</script>

<template>
  <main class="max-w-5xl mx-auto px-4 py-8 font-sans">
    <!-- Header -->
    <section class="mb-8">
      <h1 class="text-3xl font-bold text-loli-pink mb-2">{{ t('themesGallery.title') }}</h1>
      <p class="text-sm text-gray-600">{{ t('themesGallery.desc') }}</p>
    </section>
    <!-- min-w-0 on every child: grid items default to min-width auto, so
         unbreakable output content (embed URLs) would stretch the whole
         page on phones. -->
    <div class="grid lg:grid-cols-[minmax(0,1fr)_360px] gap-8 [&>*]:min-w-0">
      <!-- Generator: first in DOM so mobile stacks it above the preview. -->
      <div>
        <!-- Playground -->
        <section class="rounded-xl bg-loli-cream p-4">
          <h2 class="text-lg font-semibold mb-3 flex items-center gap-2">
            <img src="/images/lolicount-icon.png" alt="" class="h-5 w-5" />
            {{ t('themesGallery.generateTitle') }}
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
          <div v-if="generatedUrl || generatedAnimated" class="mt-4 space-y-3">
            <div class="rounded-xl bg-white p-3 flex justify-center">
              <EmotePreview
                v-if="generatedAnimated && generatedKind !== 'live2d'"
                :key="generateKey"
                :model="state.theme"
                :name="generatedName"
                :text="state.text || '{n}'"
              />
              <Live2DPreview
                v-else-if="generatedAnimated && generatedKind === 'live2d'"
                :key="generateKey"
                :model="state.theme"
                :name="generatedName"
                :text="state.text || '{n}'"
              />
              <BgPreview v-else :url="generatedPreviewUrl" :width="400" />
            </div>
            <LinkOutput
              :url="generatedAnimated ? '' : generatedUrl"
              :name="generatedName"
              :widget-snippet="generatedAnimated ? widgetSnippet : undefined"
            />
          </div>
          <div v-else class="mt-4 rounded-xl bg-white p-3">
            <div class="h-24 flex flex-col items-center justify-center text-center text-xs text-gray-400">
              <p>{{ t('playground.emptyHint1') }}</p>
              <p>{{ t('playground.emptyHint2') }}</p>
            </div>
          </div>
        </section>
      </div>

      <!-- Preview panel: second in DOM so mobile shows it right below
           the generator. On desktop it spans both grid rows of the
           right column and the inner rail is sticky, so scrolling moves
           only the left column. -->
      <div class="lg:row-span-2 lg:self-stretch">
        <div class="preview-rail">
        <!-- Preview panel -->
        <section class="rounded-xl bg-loli-cream p-4">
          <h2 class="text-lg font-semibold mb-3">{{ t('themesGallery.previewTitle') }}</h2>
          <div
            v-if="selectedTheme"
            class="cursor-pointer relative flex justify-center"
            :title="t('themes.reload')"
            @click="reloadPreview"
          >
            <EmotePreview
              v-if="selectedAnimated && selectedKind !== 'live2d'"
              :key="previewKey"
              :model="selectedTheme"
              name="demo"
              text="{n}"
            />
            <Live2DPreview
              v-else-if="selectedAnimated && selectedKind === 'live2d'"
              :key="previewKey"
              :model="selectedTheme"
              name="demo"
              text="{n}"
            />
            <img
              v-else
              ref="previewImgEl"
              :src="previewUrl"
              :alt="selectedTheme"
              crossorigin="anonymous"
              class="max-h-72 object-contain"
              @load="previewLoading = false"
              @error="previewLoading = false"
            />
            <!-- Loading overlay: blocks the click-to-reload hit area and
                 signals that a fresh no-store render is on its way. -->
            <div
              v-if="previewLoading && !selectedAnimated"
              class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-2 bg-white/70 rounded-lg"
              aria-live="polite"
            >
              <span class="h-6 w-6 animate-spin rounded-full border-2 border-loli-pink border-t-transparent" aria-hidden="true" />
              <span class="text-xs text-gray-500">{{ t('themesGallery.previewLoading') }}</span>
            </div>
          </div>
          <div v-else class="h-40 flex items-center justify-center text-sm text-gray-400">
            {{ t('loli.loading') }}
          </div>
          <div v-if="selectedTheme && !selectedAnimated" class="mt-3 flex justify-center">
            <button
              type="button"
              :disabled="previewLoading"
              :class="cn(
                'inline-flex items-center gap-1.5 border-2 border-loli-pink text-loli-pink text-sm font-semibold px-4 py-1.5 rounded-lg hover:bg-loli-pink hover:text-white transition disabled:opacity-50 disabled:cursor-not-allowed'
              )"
              @click="downloadFrame"
            >↓ {{ t('themesGallery.downloadFrame') }}</button>
          </div>
          <div v-if="selectedMeta" class="mt-3 text-xs text-gray-500 space-y-1">
            <p>{{ t('themesGallery.gameLabel') }}: {{ gameLabel(selectedMeta.gameKey) }}</p>
            <p>{{ t('themesGallery.characterLabel') }}: {{ selectedMeta.character }}</p>
          </div>
          <p class="mt-2 text-xs text-gray-500">{{ t('themesGallery.previewHint') }}</p>
        </section>
        </div>
      </div>

      <!-- Filters + theme grid: last on mobile; on desktop grid
           auto-placement puts it in the second row, left column. -->
      <div>
        <!-- Filter bar: search + type filter. The game dropdown was
             replaced by the collapsible per-source groups below. -->
        <section class="mb-6 rounded-xl bg-loli-cream p-4 space-y-3">
          <!-- box-border: no global CSS reset is loaded, so <input> keeps
               the UA content-box and w-full + px-3 + border would overflow
               the panel by 26px on the right. text-base on mobile: 16px
               control text prevents iOS focus auto-zoom. -->
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('themesGallery.searchPlaceholder')"
            class="w-full box-border border rounded-lg px-3 py-2 bg-white text-base sm:text-sm focus:outline-none focus:ring-2 focus:ring-loli-pink/40 focus:border-loli-pink transition"
          />
          <div class="flex items-center gap-2 flex-wrap">
            <button
              v-for="opt in [
                { value: 'all', label: t('themesGallery.allKinds') },
                { value: 'card', label: t('themesGallery.kindCard') },
                { value: 'character', label: t('themesGallery.kindCharacter') },
                { value: 'live2d', label: t('themesGallery.kindLive2d') },
              ]"
              :key="opt.value"
              type="button"
              :class="cn(
                'px-3 py-1.5 rounded-full text-sm transition border',
                kindFilter === opt.value
                  ? 'bg-loli-pink text-white border-loli-pink'
                  : 'bg-white text-gray-600 border-gray-200 hover:border-loli-pink'
              )"
              @click="kindFilter = opt.value"
            >{{ opt.label }}</button>
            <span class="text-xs text-gray-500 ml-auto">{{ t('themesGallery.resultCount', { n: resultCount }) }}</span>
          </div>
        </section>

        <!-- Theme groups: hot first, then one collapsible section per
             source game. Collapsed by default; auto-expanded while
             search/type filters are active. -->
        <section>
          <div v-if="groups.length" class="space-y-3">
            <div v-for="g in groups" :key="g.key">
              <button
                type="button"
                class="w-full flex items-center gap-2 rounded-lg bg-loli-cream px-4 py-2.5 text-left transition hover:bg-loli-pink/10"
                :aria-expanded="isGroupOpen(g.key)"
                @click="toggleGroup(g.key)"
              >
                <span class="text-xs text-loli-pink w-4">{{ isGroupOpen(g.key) ? '▼' : '▶' }}</span>
                <span class="font-medium text-sm">{{ g.label }}</span>
                <span
                  v-if="g.hot"
                  class="text-[10px] leading-none px-1.5 py-1 rounded-full bg-loli-pink text-white font-semibold"
                  :title="t('themesGallery.hotHint')"
                >HOT</span>
                <span class="ml-auto text-xs text-gray-500">{{ g.items.length }}</span>
              </button>
              <div v-show="isGroupOpen(g.key)" class="grid grid-cols-2 md:grid-cols-3 gap-4 mt-3">
                <button
                  v-for="tth in g.items"
                  :key="tth.name"
                  type="button"
                  :class="cn(
                    'rounded-xl border-2 p-3 text-left transition bg-white',
                    selectedTheme === tth.name
                      ? 'border-loli-pink shadow-sm'
                      : 'border-transparent hover:border-loli-pink/40'
                  )"
                  @click="selectTheme(tth.name)"
                >
                  <div class="rounded-lg bg-loli-cream flex items-center justify-center overflow-hidden mb-2 h-28">
                    <!-- Static pre-rendered thumbs (cmd/gen-theme-thumbs, compressed
                         to card-sized webp by scripts/gen-theme-thumbs-webp.mjs): the live
                         /@demo URL fires one render request per card, and a 500-card grid
                         bursts past the IP rate limit (429) and breaks every image. -->
                    <img
                      :src="tth.animated
                        ? `/images/emote-thumbs/${tth.name}.webp`
                        : `/images/theme-thumbs/${tth.name}.webp`"
                      :alt="tth.name"
                      class="max-h-24 object-contain"
                      loading="lazy"
                    />
                  </div>
                  <p class="text-sm font-medium truncate">{{ tth.name }}</p>
                  <p class="text-xs text-gray-500 truncate">
                    {{ tth.meta ? tth.meta.character : t('themesGallery.unknownMeta') }}
                  </p>
                  <!-- Variation count from /api/themes (product of random layer
                       candidates; absent for animated themes). -->
                  <p v-if="tth.variants" class="text-xs text-gray-400 truncate">
                    {{ t('themes.variants', { n: tth.variants.toLocaleString() }) }}
                  </p>
                </button>
              </div>
            </div>
          </div>
          <div v-else class="rounded-xl bg-loli-cream p-10 text-center text-sm text-gray-400">
            {{ t('themesGallery.noResults') }}
          </div>
        </section>
      </div>
    </div>

    <BackToTop />
  </main>
</template>

<style scoped>
/* Desktop only: pin the preview rail to the viewport so scrolling moves
 * just the left column (generator + theme picker). Below lg the rail is
 * a plain block in the normal flow. */
@media (min-width: 1024px) {
  .preview-rail {
    position: sticky;
    top: 5rem;
    max-height: calc(100vh - 6rem);
    overflow-y: auto;
  }
}
</style>
