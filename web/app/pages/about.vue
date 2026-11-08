<script setup lang="ts">
import { cn } from '~/utils/cn'
import { summarizeThemes } from '~/utils/themeStatistics'

const { t, locale } = useI18n()
const { fetchThemes, fetchConfig, buildCounterUrl, publicBase } = useApi()
const { data: themes, status, refresh } = useAsyncData('about-themes', fetchThemes, {
  server: false,
  lazy: true,
})
const statistics = computed(() => themes.value ? summarizeThemes(themes.value) : null)
const numberFormat = computed(() => new Intl.NumberFormat({ en: 'en-US', zh: 'zh-CN', jp: 'ja-JP' }[locale.value]))
const formatNumber = (value: number | null | undefined) => value == null ? '—' : numberFormat.value.format(value)
const busy = computed(() => status.value === 'idle' || status.value === 'pending')
const siteOrigin = ref('')
const exampleUrl = computed(() => buildCounterUrl(
  { name: 'your-name', theme: 'lian-ren' },
  publicBase.value || siteOrigin.value,
))
const markdown = computed(() => `![Lolicount](${exampleUrl.value})`)
const copied = ref<'idle' | 'success' | 'failed'>('idle')
let copyTimer: ReturnType<typeof setTimeout> | undefined

const copyExample = async () => {
  clearTimeout(copyTimer)
  try {
    await navigator.clipboard.writeText(markdown.value)
    copied.value = 'success'
  } catch {
    copied.value = 'failed'
  }
  copyTimer = setTimeout(() => { copied.value = 'idle' }, 2500)
}

onMounted(() => {
  siteOrigin.value = window.location.origin
  void fetchConfig()
})
onBeforeUnmount(() => clearTimeout(copyTimer))

useSeoMeta({
  title: () => `${t('nav.howAbout')} · Lolicount`,
  description: () => t('aboutPage.intro'),
  ogTitle: () => `${t('nav.howAbout')} · Lolicount`,
  ogDescription: () => t('aboutPage.intro'),
})

const steps = ['choose', 'generate', 'embed'] as const
</script>

<template>
  <main class="about-page">
    <section class="about-overview" aria-labelledby="about-title">
      <p class="about-eyebrow">LOLICOUNT</p>
      <h1 id="about-title" class="about-title droog">How / About</h1>
      <p class="about-intro">{{ t('aboutPage.intro') }}</p>

      <dl class="about-statistics" :aria-busy="busy" aria-live="polite">
        <div class="about-stat">
          <dt>{{ t('aboutPage.totalThemes') }}</dt>
          <dd class="stat-value">
            <span class="stat-number" data-stat="total">{{ formatNumber(statistics?.total) }}</span>
            <p class="stat-note">{{ t('aboutPage.totalNote') }}</p>
          </dd>
        </div>
        <div class="about-stat">
          <dt>{{ t('aboutPage.maxVariants') }}</dt>
          <dd class="stat-value">
            <span class="stat-number" data-stat="maximum">{{ formatNumber(statistics?.maxVariants) }}</span>
            <p class="stat-note">
              <NuxtLink v-if="statistics?.mostVariedTheme" :to="`/?theme=${encodeURIComponent(statistics.mostVariedTheme)}`" class="stat-theme">
                {{ statistics.mostVariedTheme }} <span aria-hidden="true">↗</span>
              </NuxtLink>
              <span v-else>{{ t('aboutPage.maxNote') }}</span>
            </p>
          </dd>
        </div>
      </dl>
      <div class="statistics-status" role="status">
        <template v-if="status === 'error'">
          <span>{{ t('aboutPage.statsUnavailable') }}</span>
          <button type="button" class="text-button" @click="refresh()">{{ t('aboutPage.retry') }}</button>
        </template>
        <span v-else>{{ t(busy ? 'aboutPage.statsLoading' : 'aboutPage.statsNote') }}</span>
      </div>
    </section>

    <section id="howto" class="about-howto" aria-labelledby="howto-title">
      <div class="about-howto-inner">
        <div class="howto-heading-row">
          <div>
            <h2 id="howto-title" class="section-title droog">How to use</h2>
            <p class="section-description">{{ t('aboutPage.howtoIntro') }}</p>
          </div>
          <NuxtLink to="/" class="playground-link">{{ t('aboutPage.openPlayground') }} <span aria-hidden="true">↗</span></NuxtLink>
        </div>

        <ol class="howto-steps">
          <li v-for="(step, index) in steps" :key="step" class="howto-step">
            <span class="step-number" aria-hidden="true">0{{ index + 1 }}</span>
            <h3>{{ t(`aboutPage.${step}Title`) }}</h3>
            <p>{{ t(`aboutPage.${step}Description`) }}</p>
          </li>
        </ol>

        <div class="embed-example">
          <div class="example-heading">
            <h3>{{ t('aboutPage.exampleTitle') }}</h3>
            <button type="button" :class="cn('copy-button', copied === 'success' && 'is-copied')" @click="copyExample">
              {{ t(copied === 'success' ? 'embed.copied' : 'embed.copy') }}
            </button>
          </div>
          <pre><code>{{ markdown }}</code></pre>
          <p>{{ t('aboutPage.exampleNote') }}</p>
          <span v-if="copied === 'failed'" class="copy-error" role="status">{{ t('aboutPage.copyFailed') }}</span>
        </div>
        <p class="embed-kind-note">{{ t('aboutPage.animatedNote') }}</p>
      </div>
    </section>
    <BackToTop />
  </main>
</template>

<style scoped>
.about-page { width: 100%; }
.about-overview, .about-howto-inner { box-sizing: border-box; max-width: 60rem; margin: 0 auto; padding: 3.5rem 2rem; }
.about-eyebrow { margin: 0 0 1rem; font-size: 0.75rem; letter-spacing: 0.18em; color: var(--text-muted); }
.about-title { margin: 0; font-size: clamp(2.5rem, 6vw, 4.5rem); line-height: 1.1; letter-spacing: -0.025em; }
.about-intro { max-width: 43rem; margin: 1.5rem 0 0; font-size: 1.125rem; line-height: 1.85; text-wrap: pretty; }
.about-statistics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 3rem; margin: 2.75rem 0 0; }
.about-stat { min-width: 0; padding-top: 1.5rem; border-top: 1px solid var(--loli-border); }
.about-stat dt { color: var(--text-muted); font-size: 0.9375rem; }
.stat-value { margin: 0; }
.stat-number { display: block; margin: 0.75rem 0 0; font-size: clamp(2.75rem, 6vw, 4.75rem); font-weight: 700; line-height: 1.1; letter-spacing: -0.05em; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.stat-note { margin: 0.75rem 0 0; color: var(--text-muted); font-size: 0.8125rem; line-height: 1.7; }
.stat-theme { color: var(--text-color); text-underline-offset: 4px; }
.stat-theme span { margin-left: 0.3rem; }
.statistics-status { display: flex; align-items: center; flex-wrap: wrap; gap: 0.5rem 1rem; min-height: 1.5rem; margin-top: 1.5rem; font-size: 0.75rem; line-height: 1.7; color: var(--text-muted); }
.text-button { padding: 0; border: 0; background: none; color: var(--text-color); text-decoration: underline; text-underline-offset: 3px; font: inherit; }
.about-howto { background: var(--loli-cream); scroll-margin-top: 5rem; }
.howto-heading-row { display: flex; align-items: start; justify-content: space-between; gap: 1.5rem; }
.section-title { margin: 0; font-size: 2rem; line-height: 1.2; }
.section-description { margin: 1rem 0 0; color: var(--text-muted); font-size: 0.9375rem; line-height: 1.7; }
.playground-link { display: flex; align-items: center; gap: 0.75rem; flex-shrink: 0; padding: 0.75rem 1rem; border-radius: 0.75rem; background: var(--loli-surface); color: var(--text-color); text-decoration: none; font-size: 0.875rem; transition: background 200ms; }
.playground-link:hover { background: var(--loli-border); }
.howto-steps { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 2rem; margin: 2.5rem 0; padding: 0; list-style: none; }
.step-number { display: block; margin-bottom: 1rem; color: var(--text-muted); font-size: 0.8125rem; font-variant-numeric: tabular-nums; }
.howto-step h3 { margin: 0 0 0.75rem; font-size: 1rem; }
.howto-step p { margin: 0; color: var(--text-muted); font-size: 0.9375rem; line-height: 1.8; }
.embed-example { padding: 1.5rem; border-radius: 1rem; background: var(--loli-surface); }
.example-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.example-heading h3 { margin: 0; font-size: 0.875rem; }
.copy-button { flex-shrink: 0; padding: 0.5rem 0.75rem; border: 0; border-radius: 0.5rem; color: var(--text-color); background: var(--loli-cream); font: inherit; font-size: 0.8125rem; }
.copy-button:hover, .copy-button.is-copied { background: var(--loli-border); }
.embed-example pre { margin: 1rem 0; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.8125rem; line-height: 1.8; white-space: pre-wrap; overflow-wrap: anywhere; user-select: all; }
.embed-example p, .embed-kind-note { margin: 0; font-size: 0.8125rem; line-height: 1.8; color: var(--text-muted); }
.embed-kind-note { margin-top: 1rem; }
.copy-error { display: block; margin-top: 0.5rem; font-size: 0.8125rem; }
.about-page a:focus-visible, .about-page button:focus-visible { outline: 2px solid var(--accent); outline-offset: 5px; }
@media (max-width: 767px) {
  .about-overview, .about-howto-inner { padding: 2.5rem 1.5rem; }
  .about-intro { font-size: 1rem; }
  .about-statistics { gap: 1.5rem; margin-top: 2rem; }
  .stat-number { font-size: clamp(2rem, 7vw, 3rem); }
  .about-stat dt { min-height: 3.2em; font-size: 0.8125rem; line-height: 1.6; }
  .stat-note { font-size: 0.75rem; }
  .howto-heading-row { flex-direction: column; }
  .section-title { font-size: 1.75rem; }
  .howto-steps { grid-template-columns: 1fr; gap: 1.75rem; }
  .howto-step { position: relative; padding-left: 2.5rem; }
  .step-number { position: absolute; top: 0.2rem; left: 0; margin: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .playground-link { transition: none; }
}
</style>
