<script setup lang="ts">
import type { ParamState } from '~/components/ParamPanel.vue'
import { gameMeta, themeMeta, buildSearchHaystack } from '~/utils/themeMeta'
import type { GameKey, ThemeKind } from '~/utils/themeMeta'

// The playground IS the landing page. /themes stays as a route alias so
// previously shared ?theme= links keep working.
definePageMeta({ alias: ['/themes'] })

const DEFAULT_THEME = 'lian-ren'

const { fetchThemes, fetchFThemes, fetchHotThemes, fetchConfig, buildCounterUrl, publicBase } = useApi()
const { t, locale } = useI18n()

// Client-side Nuxt data layer: setup starts all read-only requests during
// hydration instead of waiting for mounted. server:false keeps nuxt generate
// independent from the Go API while lazy:true avoids a second SSR payload.
type PlaygroundData = { themes: ThemeInfo[]; fthemes: string[]; hot: string[] }
const { data: homeData } = useAsyncData<PlaygroundData>(
  'playground-data',
  async () => {
    const [themeList, fthemeList, hotList] = await Promise.all([
      fetchThemes().catch(() => []),
      fetchFThemes().catch(() => []),
      fetchHotThemes().catch(() => []),
      fetchConfig(),
    ])
    return { themes: themeList, fthemes: fthemeList, hot: hotList }
  },
  { server: false, lazy: true, immediate: true },
)
const themes = computed(() => homeData.value?.themes ?? [])
const fthemes = computed(() => homeData.value?.fthemes ?? [])

// Thumbnail load state for the reference-style skeleton background:
// cards show the animated gray ramp until their image has decoded.
const loadedThumbs = ref<Record<string, boolean>>({})
const markThumbLoaded = (name: string): void => { loadedThumbs.value[name] = true }

// Join the API theme list with local metadata. Themes missing metadata
// (e.g. added before the meta table is updated) still show up under
// the "other" game bucket.
const themeMetaByName = new Map(themeMeta.map((meta) => [meta.name, meta]))
const merged = computed(() =>
  themes.value.map((tth) => {
    const meta = themeMetaByName.get(tth.name)
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

const hotNames = computed(() => homeData.value?.hot ?? [])
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
    // Princess Connect lists named characters first and unnamed NPC themes at
    // the end of the group, matching how the portraits are named.
    const items = filteredThemes.value
      .filter((tth) => !consumed.has(tth.name) && (tth.meta?.gameKey ?? 'other') === key)
      .sort((a, b) => {
        if (key !== 'princess-connect') return 0
        const unnamedOrder = Number(a.meta?.unnamed ?? false) - Number(b.meta?.unnamed ?? false)
        return unnamedOrder || a.name.localeCompare(b.name)
      })
    if (items.length > 0) {
      out.push({ key, label: gameLabel(key), hot: false, items })
    }
  }
  return out
})

// Group expand/render state: collapsed groups stay out of the DOM until
// first opened. Previously rendered groups remain mounted so closing does
// not discard scroll state or thumbnail skeletons. Filters still mount all
// matching groups so results remain immediately visible while searching.
const expandedGroups = reactive(new Set<string>())
const renderedGroups = reactive(new Set<string>())
const toggleGroup = (key: string) => {
  renderedGroups.add(key)
  if (expandedGroups.has(key)) expandedGroups.delete(key)
  else expandedGroups.add(key)
}
const isGroupOpen = (key: string) => filtersActive.value || expandedGroups.has(key)
const isGroupRendered = (key: string) => filtersActive.value || renderedGroups.has(key)

// Selected theme + preview with a cache-buster (same reload trick as the
// home page showcase: the back-end picks a random frame per request).
// Animated (emote) themes render through the shared WebGL preview instead
// of the static SVG image endpoint.
const selectedTheme = ref(DEFAULT_THEME)
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
  const lianRen = themes.value.find((tth) => tth.name === DEFAULT_THEME)
  return lianRen ? lianRen.name : (themes.value[0]?.name ?? '')
}

// Soft-navigation support: picking up ?theme= when the query changes
// while the page stays mounted. Skips unknown names; the writer path
// (selectTheme) is a no-op here because the name already matches. When
// the param is removed (e.g. the navbar links to plain /themes), the
// selection falls back to the default so the clean URL and the preview
// stay consistent with a fresh visit.
watch(
  [() => homeData.value, () => route.query.theme],
  ([data, q]) => {
    if (!data) return
    const name = typeof q === 'string' ? q : ''
    if (name === selectedTheme.value) return
    if (name && !themes.value.some((tth) => tth.name === name)) return
    applyThemeSelection(name || defaultThemeName())
  },
  { immediate: true },
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
  themeMetaByName.get(selectedTheme.value) ?? null,
)

// Playground state, same shape as the home page so the generation flow
// (ParamPanel + generate + LinkOutput) stays identical.
const state = reactive<ParamState>({
  name: '',
  theme: DEFAULT_THEME,
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
  // Only E-mote (PSB) themes embed through widget.js. Spine, Live2D and Unity
  // themes are iframe players with their own player pages.
  if (generatedKind.value !== 'psb') {
    const query = new URLSearchParams({
      model: state.theme,
      name: generatedName.value,
      text: state.text || '{n}',
    })
    return `<iframe src="${origin}/${generatedKind.value}-player.html?${query.toString()}" width="360" height="520" frameborder="0"></iframe>`
  }
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

onMounted(() => {
  // SSG may finish the preview fetch before hydration attaches @load —
  // clear the spinner for an already-complete image.
  if (previewImgEl.value?.complete) previewLoading.value = false
})

// Critical first-paint assets are preloaded only on the playground route;
// Nuxt injects these after its global head entries without duplicating them.
useHead({
  link: [
    { rel: 'preload', href: '/images/avatar.webp', as: 'image', fetchpriority: 'high' },
    { rel: 'preload', href: '/fonts/droog-900.woff2', as: 'font', type: 'font/woff2', crossorigin: true },
  ],
})
</script>

<template>
  <main class="home-page">
    <HomeHero />

    <section class="home-workspace">
      <div class="workspace-grid">
        <section class="tool-panel">
          <header class="panel-header reveal" data-reveal>
            <h2>{{ t('themesGallery.generateTitle') }}</h2>
          </header>

          <LazyParamPanel
            :state="state"
            :themes="themes"
            :fthemes="fthemes"
            @update="onUpdate"
          />

          <div class="generate-row">
            <StarBurst ref="starBurst" />
            <button
              type="button"
              class="generate-button"
              :disabled="nameEmpty"
              @click="generate($event)"
            >{{ nameEmpty ? t('param.nameEmpty') : t('playground.generate') }}</button>
          </div>

          <div v-if="generatedUrl || generatedAnimated" class="result-panel">
            <div class="result-preview">
              <EmotePreview
                v-if="generatedAnimated && generatedKind === 'psb'"
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
              <ModelIframePreview
                v-else-if="generatedAnimated && ['spine', 'unity'].includes(generatedKind)"
                :key="generateKey"
                :kind="generatedKind as 'spine' | 'unity'"
                :model="state.theme"
                :name="generatedName"
                :text="state.text || '{n}'"
              />
              <BgPreview v-else :url="generatedPreviewUrl" :width="400" />
            </div>
            <LazyLinkOutput
              :url="generatedAnimated ? '' : generatedUrl"
              :name="generatedName"
              :widget-snippet="generatedAnimated ? widgetSnippet : undefined"
            />
          </div>

          <div v-else class="empty-panel">
            <p>{{ t('playground.emptyHint1') }}</p>
            <p>{{ t('playground.emptyHint2') }}</p>
          </div>
        </section>
        <div class="preview-slot">

        <aside class="preview-panel">
          <header class="panel-header reveal" data-reveal>
            <h2>{{ t('themesGallery.previewTitle') }}</h2>
          </header>

          <div v-if="selectedTheme" class="preview-stage" @click="reloadPreview">
            <EmotePreview
              v-if="selectedAnimated && selectedKind === 'psb'"
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
            <ModelIframePreview
              v-else-if="selectedAnimated && ['spine', 'unity'].includes(selectedKind)"
              :key="previewKey"
              :kind="selectedKind as 'spine' | 'unity'"
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
              class="preview-image"
              @load="previewLoading = false"
              @error="previewLoading = false"
            >
            <div v-if="previewLoading && !selectedAnimated" class="preview-loading">
              <span class="spinner"></span>
              <span>{{ t('themesGallery.previewLoading') }}</span>
            </div>
          </div>

          <div v-else class="preview-empty">{{ t('loli.loading') }}</div>

          <button
            v-if="selectedTheme && !selectedAnimated"
            type="button"
            class="download-button"
            :disabled="previewLoading"
            @click="downloadFrame"
          >{{ t('themesGallery.downloadFrame') }}</button>

          <dl v-if="selectedMeta" class="preview-meta">
            <div>
              <dt>{{ t('themesGallery.gameLabel') }}</dt>
              <dd>{{ gameLabel(selectedMeta.gameKey) }}</dd>
            </div>
            <div>
              <dt>{{ t('themesGallery.characterLabel') }}</dt>
              <dd>{{ selectedMeta.character }}</dd>
            </div>
          </dl>
        </aside>
        </div>

      <section class="theme-library">
        <header class="library-header reveal" data-reveal>
          <div>
            <h2>{{ t('themesGallery.title') }}</h2>
            <p>{{ t('themesGallery.desc') }}</p>
          </div>
          <input
            v-model="searchQuery"
            type="search"
            class="library-search"
            :placeholder="t('themesGallery.searchPlaceholder')"
          >
        </header>

        <div class="library-filters reveal" data-reveal>
          <button
            v-for="opt in [
              { value: 'all', label: t('themesGallery.allKinds') },
              { value: 'card', label: t('themesGallery.kindCard') },
              { value: 'character', label: t('themesGallery.kindCharacter') },
              { value: 'live2d', label: t('themesGallery.kindLive2d') },
              { value: 'unity', label: t('themesGallery.kindUnity') },
            ]"
            :key="opt.value"
            type="button"
            class="library-filter"
            :class="{ active: kindFilter === opt.value }"
            @click="kindFilter = opt.value"
          >{{ opt.label }}</button>
          <span class="library-count">{{ t('themesGallery.resultCount', { n: resultCount }) }}</span>
        </div>

        <div v-if="groups.length" class="theme-groups">
          <section v-for="g in groups" :key="g.key" class="theme-group">
            <button
              type="button"
              class="theme-group-header"
              :aria-expanded="isGroupOpen(g.key)"
              @click="toggleGroup(g.key)"
            >
              <span>{{ g.label }}</span>
              <span v-if="g.hot" class="hot-badge">HOT</span>
              <span class="group-count">{{ g.items.length }}</span>
            </button>

            <div v-if="isGroupRendered(g.key)" v-show="isGroupOpen(g.key)" class="theme-grid">
              <button
                v-for="tth in g.items"
                :key="tth.name"
                type="button"
                class="theme-card"
                :class="{ selected: selectedTheme === tth.name }"
                @click="selectTheme(tth.name)"
              >
                <span class="thumb-box">
                  <span v-if="!loadedThumbs[tth.name]" class="thumb-skeleton" aria-hidden="true"></span>
                  <img
                    :src="tth.animated
                      ? `/images/emote-thumbs/${tth.name}.webp`
                      : `/images/theme-thumbs/${tth.name}.webp`"
                    :alt="tth.name"
                    loading="lazy"
                    @load="markThumbLoaded(tth.name)"
                    @error="markThumbLoaded(tth.name)"
                  >
                </span>
                <strong>{{ tth.name }}</strong>
                <span>{{ tth.meta ? tth.meta.character : t('themesGallery.unknownMeta') }}</span>
                <small v-if="tth.variants">{{ t('themes.variants', { n: tth.variants.toLocaleString() }) }}</small>
              </button>
            </div>
          </section>
        </div>

        <div v-else class="library-empty">{{ t('themesGallery.noResults') }}</div>
      </section>
      </div>
    </section>

    <LazyBackToTop />
    <LazyHomeFooter />
  </main>
</template>

<style scoped>
.home-page {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
}

.home-workspace {
  display: flex;
  flex-direction: column;
  gap: 4rem;
  padding: 4rem 1.5rem;
  background: var(--loli-surface);
}

.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 2rem;
  width: min(100%, 64rem);
  margin: 0 auto;
}

.theme-library {
  width: 100%;
  min-width: 0;
  margin: 0;
}

.preview-slot {
  width: 100%;
  min-width: 0;
}

.tool-panel,
.preview-panel {
  min-width: 0;
  padding: 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--loli-border);
  border-radius: 1.25rem;
  box-shadow: 0 18px 45px rgba(76, 79, 105, 0.08);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.panel-header h2 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-color);
}

.generate-row {
  position: relative;
  margin-top: 1.25rem;
}

.generate-button {
  width: 100%;
  padding: 0.8rem 1rem;
  font-size: 0.95rem;
  font-weight: 650;
  color: #fff;
  background: var(--loli-pink);
  border: 0;
  border-radius: 0.75rem;
  cursor: pointer;
  transition: filter 0.2s ease, transform 0.2s ease;
}

.generate-button:hover:not(:disabled) {
  filter: brightness(1.05);
  transform: translateY(-1px);
}

.generate-button:disabled {
  color: var(--text-muted);
  background: var(--loli-cream);
  cursor: not-allowed;
}

.result-panel {
  display: grid;
  /* The implicit auto track would grow to the long URL's max-content and
     overflow the card; minmax(0, 1fr) lets the track shrink instead. */
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  margin-top: 1.25rem;
}

.result-preview {
  display: grid;
  place-items: center;
  min-height: 12rem;
  padding: 1rem;
  background: var(--loli-surface);
  border-radius: 0.85rem;
}

.empty-panel {
  display: grid;
  gap: 0.35rem;
  place-items: center;
  min-height: 9rem;
  margin-top: 1.25rem;
  padding: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  text-align: center;
  background: var(--loli-surface);
  border-radius: 0.85rem;
}

.preview-stage {
  position: relative;
  display: grid;
  place-items: center;
  min-height: 16rem;
  cursor: pointer;
  background: var(--loli-surface);
  border-radius: 0.85rem;
  /* Very wide portraits fill the frame and fade out at both edges instead of
     spilling over the panel. Centred images narrower than the frame keep their
     own edges, because the transparent mask stops live outside them. The mask
     belongs to the portrait itself: masking the stage would fade its own
     background and rounded corners too. */
  overflow: hidden;
}

.preview-image {
  width: 100%;
  max-height: 18rem;
  object-fit: contain;
  object-position: center;
  -webkit-mask-image: linear-gradient(90deg, transparent 0, #000 8%, #000 92%, transparent 100%);
  mask-image: linear-gradient(90deg, transparent 0, #000 8%, #000 92%, transparent 100%);
}

.preview-loading {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  gap: 0.5rem;
  background: color-mix(in srgb, var(--loli-surface) 75%, transparent);
  border-radius: 0.85rem;
  color: var(--text-muted);
  font-size: 0.8rem;
}

.spinner {
  width: 1.5rem;
  height: 1.5rem;
  border: 2px solid var(--loli-pink);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.preview-empty {
  display: grid;
  place-items: center;
  min-height: 16rem;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.download-button {
  width: 100%;
  margin-top: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 650;
  color: var(--loli-pink);
  background: transparent;
  border: 1px solid var(--loli-pink);
  border-radius: 0.75rem;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.download-button:hover:not(:disabled) {
  color: #fff;
  background: var(--loli-pink);
}

.download-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.preview-meta {
  display: grid;
  gap: 0.5rem;
  margin: 1rem 0 0;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.preview-meta div {
  display: grid;
  grid-template-columns: 5rem 1fr;
  gap: 0.5rem;
}

.preview-meta dt {
  font-weight: 650;
}

.preview-meta dd {
  margin: 0;
}

.library-header {
  display: grid;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.library-header h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-color);
}

.library-header p {
  margin-top: 0.35rem;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.library-search {
  width: 100%;
  padding: 0.7rem 0.9rem;
  font-size: 0.9rem;
  color: var(--text-color);
  background: var(--bg-card);
  border: 1px solid var(--loli-border);
  border-radius: 0.75rem;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.library-search:focus {
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.library-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.library-filter {
  padding: 0.45rem 0.85rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: var(--bg-card);
  border: 1px solid var(--loli-border);
  border-radius: 999px;
  cursor: pointer;
  transition: color 0.2s ease, background-color 0.2s ease, border-color 0.2s ease;
}

.library-filter:hover,
.library-filter.active {
  color: #fff;
  background: var(--loli-pink);
  border-color: var(--loli-pink);
}

.library-count {
  margin-left: auto;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.theme-groups {
  display: grid;
  gap: 1rem;
}

.theme-group {
  min-width: 0;
}

.theme-group-header {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 0.9rem;
  font-weight: 650;
  color: var(--text-color);
  text-align: left;
  background: var(--bg-card);
  border: 1px solid var(--loli-border);
  border-radius: 0.85rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.theme-group-header:hover {
  background: var(--loli-selection);
}

.hot-badge {
  padding: 0.15rem 0.4rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: #fff;
  background: var(--loli-pink);
  border-radius: 999px;
}

.group-count {
  margin-left: auto;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted);
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1rem;
  content-visibility: auto;
  contain-intrinsic-size: auto 500px;
}

.theme-card {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
  padding: 0.85rem;
  text-align: left;
  background: var(--bg-card);
  border: 1px solid var(--loli-border);
  border-radius: 0.85rem;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.theme-card:hover,
.theme-card.selected {
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 3px var(--accent-soft);
  transform: translateY(-2px);
}

.thumb-box {
  position: relative;
  width: 100%;
  height: 8.5rem;
  margin-bottom: 0.35rem;
  /* Wide portraits stay centred and fade out where they leave the card; the
     mask sits on the image so the skeleton ramp keeps its full width. */
  overflow: hidden;
}

/* Reference's skeleton-cycle ramp behind each card image until it loads. */
.thumb-skeleton {
  position: absolute;
  inset: 0;
  background-color: var(--skeleton-0);
  animation: skeleton-cycle 6s ease-in-out infinite;
}

.theme-card img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  -webkit-mask-image: linear-gradient(90deg, transparent 0, #000 8%, #000 92%, transparent 100%);
  mask-image: linear-gradient(90deg, transparent 0, #000 8%, #000 92%, transparent 100%);
}

.theme-card strong {
  overflow: hidden;
  font-size: 0.875rem;
  color: var(--text-color);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.theme-card span,
.theme-card small {
  overflow: hidden;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.library-empty {
  padding: 3rem 1rem;
  font-size: 0.875rem;
  color: var(--text-muted);
  text-align: center;
  background: var(--bg-card);
  border-radius: 0.85rem;
}

@media (min-width: 768px) {
  .theme-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .library-header {
    grid-template-columns: minmax(0, 1fr) 20rem;
    align-items: end;
  }
}

@media (min-width: 1024px) {
  .workspace-grid {
    grid-template-columns: minmax(0, 1fr) 22.5rem;
    align-items: start;
  }

  .preview-slot {
    grid-column: 2;
    grid-row: 1 / span 2;
    align-self: stretch;
  }

  .preview-panel {
    position: sticky;
    top: 5rem;
    max-height: calc(100vh - 6rem);
    overflow-y: auto;
  }

  .theme-library {
    grid-column: 1;
  }
}
</style>
