<script setup lang="ts">
import { cn } from '~/utils/cn'

const { t } = useI18n()
const { repoUrl } = useGitHub()
const jar = ref<{ addHeart: () => void } | null>(null)
const hearts = ref(0)
const thanked = ref(false)
let thanksTimer: ReturnType<typeof setTimeout> | undefined

useSeoMeta({
  title: () => `${t('nav.support')} · Lolicount`,
  description: () => t('support.intro'),
  ogTitle: () => `${t('nav.support')} · Lolicount`,
  ogDescription: () => t('support.intro'),
})

const thank = () => {
  hearts.value++
  thanked.value = true
  clearTimeout(thanksTimer)
  thanksTimer = setTimeout(() => { thanked.value = false }, 600)
}
onBeforeUnmount(() => clearTimeout(thanksTimer))

const ways = [
  { key: 'star', href: repoUrl, path: 'm12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-3-5.6 3 1.1-6.2L3 9.6l6.2-.9Z' },
  { key: 'theme', href: `${repoUrl}/blob/main/docs/contributing-themes.md`, path: 'M4 4h16v16H4z M4 16l5-5 4 4 3-3 4 4 M14 8h.01' },
  { key: 'code', href: `${repoUrl}/blob/main/docs/contributing-code.md`, path: 'm8 6-6 6 6 6m8-12 6 6-6 6m-3-14-2 16' },
]
</script>

<template>
  <main class="support-page">
    <section class="support-hero" aria-labelledby="support-title">
      <div class="support-hero-inner">
        <div class="support-title-wrap">
          <p class="support-eyebrow">{{ t('support.eyebrow') }}</p>
          <h1 id="support-title" class="support-title droog">Support</h1>
        </div>
        <div class="support-model">
          <SupportHeartJar ref="jar" @heart="thank" />
        </div>
      </div>
    </section>

    <div class="support-body">
      <div class="support-inner">
        <section class="support-intro" :aria-label="t('nav.support')">
          <div class="heart-action-row">
            <button
              type="button"
              :class="cn('give-heart', thanked && 'is-thanked')"
              :aria-label="t('support.addHeart')"
              data-no-cursor-firework
              @click="jar?.addHeart()"
            >
              <svg viewBox="0 0 64 64" width="52" height="52" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M32 54S7 40 7 22a13 13 0 0 1 25-5 13 13 0 0 1 25 5c0 18-25 32-25 32Z" /></svg>
              <span class="heart-plus" aria-hidden="true">+</span>
            </button>
            <span class="heart-status" role="status">{{ hearts ? t('support.thanks', { n: hearts }) : t('support.heartNote') }}</span>
          </div>
          <p class="support-intro-copy">{{ t('support.intro') }}</p>
        </section>

        <section class="sponsor-section" aria-labelledby="sponsor-title">
          <div class="support-section-head">
            <h2 id="sponsor-title" class="support-heading droog">{{ t('support.sponsorHeading') }}</h2>
            <span class="section-aside">{{ t('support.optional') }}</span>
          </div>
          <a class="sponsor-card" href="https://github.com/sponsors/miaoledor" target="_blank" rel="noopener noreferrer">
            <span class="sponsor-cup" aria-hidden="true">
              <svg viewBox="0 0 96 96" width="88" height="88" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M25 30h42l-5 51H30L25 30Zm-3 0h48M35 19h25M53 8l-6 22" /><path d="M36 48c4-8 10-5 12 0 3-5 10-8 13 0 3 7-13 17-13 17S33 55 36 48Z" fill="currentColor" stroke="none" /><circle cx="38" cy="72" r="2" fill="currentColor" stroke="none" /><circle cx="48" cy="75" r="2" fill="currentColor" stroke="none" /><circle cx="57" cy="71" r="2" fill="currentColor" stroke="none" /></svg>
            </span>
            <span class="sponsor-copy">
              <span class="sponsor-platform">GitHub Sponsors</span>
              <span class="sponsor-title">{{ t('support.sponsorTitle') }}</span>
              <span class="sponsor-description">{{ t('support.sponsorDescription') }}</span>
            </span>
            <svg class="link-arrow" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="M6 18 18 6M6 6h12v12" /></svg>
          </a>
        </section>

        <section aria-labelledby="other-ways-title">
          <div class="support-section-head">
            <h2 id="other-ways-title" class="support-heading droog">{{ t('support.waysHeading') }}</h2>
            <span class="section-aside">{{ t('support.waysAside') }}</span>
          </div>
          <div class="support-ways">
            <a v-for="way in ways" :key="way.key" class="support-way" :href="way.href" target="_blank" rel="noopener noreferrer">
              <svg class="way-icon" viewBox="0 0 24 24" width="30" height="30" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path :d="way.path" /></svg>
              <svg class="way-arrow" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="M6 18 18 6M6 6h12v12" /></svg>
              <h3>{{ t(`support.${way.key}Title`) }}</h3>
              <p>{{ t(`support.${way.key}Description`) }}</p>
            </a>
          </div>
        </section>

        <p class="support-signoff">{{ t('support.signoff') }} <span aria-hidden="true">♡</span></p>
      </div>
    </div>
    <HomeFooter />
    <BackToTop />
  </main>
</template>

<style scoped>
.support-page { width: 100%; }
.support-hero { background: var(--loli-surface); }
.support-hero-inner { position: relative; display: grid; grid-template-columns: 1fr 1fr; align-items: end; max-width: 60rem; min-height: 23rem; margin: 0 auto; padding: 0 2rem; box-sizing: border-box; }
.support-title-wrap { position: relative; z-index: 1; padding-bottom: 2.5rem; pointer-events: none; }
.support-eyebrow { margin: 0 0 1rem; color: var(--text-muted); font-size: 0.875rem; }
.support-title { margin: 0; font-size: clamp(3.4rem, 7.5vw, 6rem); line-height: 1; letter-spacing: -0.035em; }
.support-model { height: 24rem; margin: -1.5rem -2.5rem 0 -1rem; }
.support-body { background: var(--loli-cream); }
.support-inner { box-sizing: border-box; max-width: 60rem; margin: 0 auto; padding: 3rem 2rem 2.5rem; }
.support-intro { margin-bottom: 3.5rem; }
.heart-action-row { display: flex; align-items: center; gap: 1.2rem; }
.give-heart { position: relative; display: grid; place-items: center; width: 58px; height: 58px; padding: 0; border: 0; border-radius: 1rem; background: transparent; color: var(--text-color); transition: color 200ms, transform 200ms; }
.give-heart:hover { color: #f07890; transform: rotate(-8deg) scale(1.08); }
.give-heart.is-thanked { color: #f07890; animation: heart-pop 500ms ease; }
.heart-plus { position: absolute; top: -2px; right: -3px; font-size: 1.25rem; color: var(--text-muted); }
.heart-status { color: var(--text-muted); font-size: 0.8125rem; }
.support-intro-copy { max-width: 48rem; margin: 1.5rem 0 0; font-size: 1.25rem; line-height: 1.85; text-wrap: pretty; }
.support-section-head { display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: 0.75rem 1.5rem; margin-bottom: 1.5rem; }
.support-heading { margin: 0; font-size: 1.75rem; line-height: 1.2; }
.section-aside { color: var(--text-muted); font-size: 0.8125rem; }
.sponsor-section { margin-bottom: 3.5rem; }
.sponsor-card { display: flex; position: relative; align-items: center; gap: 2rem; padding: 2rem; border-radius: 1.25rem; background: var(--loli-surface); color: var(--text-color); text-decoration: none; transition: background 200ms, transform 200ms; }
.sponsor-card:hover { background: color-mix(in srgb, var(--loli-surface) 85%, #f07890); transform: translateY(-3px); }
.sponsor-cup { display: grid; place-items: center; width: 7rem; height: 7rem; flex-shrink: 0; border-radius: 50%; color: #dc7893; background: color-mix(in srgb, #f07890 10%, transparent); transform: rotate(-8deg); }
.sponsor-copy { display: flex; flex-direction: column; align-items: start; gap: 0.65rem; padding-right: 1rem; }
.sponsor-platform { color: var(--text-muted); font-size: 0.75rem; }
.sponsor-title { font-size: 1.5rem; font-weight: 700; }
.sponsor-description { color: var(--text-muted); font-size: 0.9375rem; line-height: 1.7; }
.link-arrow { flex-shrink: 0; margin-left: auto; }
.support-ways { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1rem; }
.support-way { position: relative; display: block; padding: 1.5rem; border-radius: 1rem; color: var(--text-color); background: var(--loli-surface); text-decoration: none; transition: background 200ms, transform 200ms; }
.support-way:hover { background: var(--loli-border); transform: translateY(-3px); }
.way-icon { margin-bottom: 1.5rem; }
.way-arrow { position: absolute; top: 1.5rem; right: 1.5rem; color: var(--text-muted); transition: transform 200ms; }
.support-way:hover .way-arrow { transform: translate(2px, -2px); }
.support-way h3 { margin: 0 0 0.65rem; font-size: 1rem; }
.support-way p { margin: 0; color: var(--text-muted); font-size: 0.875rem; line-height: 1.7; }
.support-signoff { display: flex; justify-content: center; align-items: center; gap: 0.5rem; margin: 3rem 0 0; color: var(--text-muted); font-size: 0.8125rem; line-height: 1.8; text-align: center; }
.support-signoff span { color: #dc7893; font-size: 1.5rem; }
.support-page a:focus-visible, .give-heart:focus-visible { outline: 2px solid var(--accent); outline-offset: 5px; }
@keyframes heart-pop { 0%, 100% { transform: scale(1); } 35% { transform: scale(1.25) rotate(-8deg); } 65% { transform: scale(0.95); } }
@media (max-width: 767px) {
  .support-hero-inner { min-height: 21rem; padding: 0 1.5rem; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .support-title-wrap { padding-bottom: 2rem; }
  .support-title { font-size: clamp(2.7rem, 8.5vw, 4rem); }
  .support-eyebrow { max-width: 10rem; line-height: 1.7; font-size: 0.75rem; }
  .support-model { height: 21rem; margin: 0 -3rem 0 -3.5rem; }
  .support-inner { padding: 2.5rem 1.5rem; }
  .support-intro-copy { font-size: 1.125rem; }
  .support-heading { font-size: 1.5rem; }
  .sponsor-card { gap: 1rem; padding: 1.5rem; }
  .sponsor-cup { width: 4rem; height: 4rem; }
  .sponsor-cup svg { width: 58px; height: 58px; }
  .sponsor-title { font-size: 1.25rem; }
  .sponsor-description { font-size: 0.875rem; }
  .link-arrow { position: absolute; top: 1.5rem; right: 1.5rem; width: 18px; }
  .support-ways { grid-template-columns: 1fr; }
  .support-way { padding: 1.5rem 1.5rem 1.5rem 4.75rem; }
  .way-icon { position: absolute; top: 1.5rem; left: 1.5rem; margin: 0; }
  .support-way h3 { padding-right: 1.5rem; }
}
@media (max-width: 479px) {
  .support-hero-inner { display: flex; flex-direction: column-reverse; align-items: stretch; }
  .support-model { height: 18rem; margin: -1.5rem -2rem -0.5rem; }
  .support-title-wrap { padding-bottom: 1.75rem; }
  .support-title { font-size: clamp(2.7rem, 12vw, 4rem); }
  .support-eyebrow { max-width: none; margin-bottom: 0.65rem; }
  .sponsor-card { align-items: start; }
  .sponsor-cup { width: 3rem; height: 3rem; margin-top: 1rem; }
  .sponsor-cup svg { width: 48px; height: 48px; }
  .sponsor-copy { padding-right: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .give-heart, .sponsor-card, .support-way, .way-arrow { transition: none; animation: none; }
}
</style>
