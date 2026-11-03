<script setup lang="ts">
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

// About page laid out on the reference site's about template: the hero
// component on top, then a mantle band holding the About Me / Skills
// columns, two card carousels and the socials grid. Copy comes from the
// i18n dictionary and the scroll reveals reuse the reference's
// cardReveal / overshoot easings.
useEases()

const { t } = useI18n()
const { fetchThemes, fetchHotThemes, buildCounterUrl, publicBase } = useApi()
const { repoUrl } = useGitHub()

const detailsEl = ref<HTMLElement | null>(null)
const detailsLeftEl = ref<HTMLElement | null>(null)
const detailsRightEl = ref<HTMLElement | null>(null)

const howToUrl = computed(() => buildCounterUrl({ name: 'name' }, publicBase.value))
const issuesUrl = `${repoUrl}/issues`
const contributeUrl = `${repoUrl}/blob/main/docs/contributing-themes.md`
const sponsorUrl = 'https://github.com/sponsors/miaoledor'

const skills = computed(() => [
  t('about.stackValue'),
  t('about.deploymentValue'),
  `${t('about.license')}: AGPL-3.0`,
])

// Hot themes back the Themes row, falling back to no cards when the API is down.
const hotCards = ref<{ name: string, animated?: boolean }[]>([])
const thumb = (name: string, animated?: boolean): string =>
  `/images/${animated ? 'emote-thumbs' : 'theme-thumbs'}/${name}.webp`

const themeItems = computed(() =>
  hotCards.value.slice(0, 6).map((card) => ({
    key: card.name,
    title: card.name,
    image: thumb(card.name, card.animated),
    to: `/?theme=${card.name}`,
  })),
)

const socials = computed(() => [
  { key: 'github', label: 'GitHub', url: repoUrl, priority: 3 },
  { key: 'sponsor', label: t('about.sponsor'), url: sponsorUrl, priority: 2 },
  { key: 'issues', label: t('about.feedback'), url: issuesUrl, priority: 2 },
  { key: 'docs', label: t('about.contribute'), url: contributeUrl, priority: 2 },
  { key: 'license', label: t('about.license'), url: `${repoUrl}/blob/main/LICENSE`, priority: 1 },
  { key: 'readme', label: 'README', url: `${repoUrl}#readme`, priority: 1 },
])

const prefersReducedMotion = (): boolean =>
  typeof window !== 'undefined'
  && window.matchMedia('(prefers-reduced-motion: reduce)').matches

let detailsContext: gsap.Context | null = null

// Details reveal: the reference staggers every heading, paragraph and list
// item in from y:60, with one shared trigger on desktop and one per column
// on mobile.
const setupDetails = (): void => {
  const root = detailsEl.value
  const left = detailsLeftEl.value
  const right = detailsRightEl.value
  if (!root || !left || !right) return
  const leftItems = left.querySelectorAll('.droog, p')
  const rightItems = right.querySelectorAll('.droog, .reveal-subheading, li')
  if (!leftItems.length || !rightItems.length) return
  const reveal = { opacity: 1, y: 0, duration: 0.8, stagger: 0.05, ease: 'power3.out' }
  if (prefersReducedMotion()) {
    gsap.set([leftItems, rightItems], { opacity: 1, y: 0 })
    return
  }
  detailsContext = gsap.context(() => {
    gsap.matchMedia().add({
      '(min-width: 768px)': () => {
        gsap.timeline({ scrollTrigger: { trigger: root, start: 'top 85%', once: true } })
          .fromTo(leftItems, { opacity: 0, y: 60 }, reveal, 0)
          .fromTo(rightItems, { opacity: 0, y: 60 }, reveal, 0)
      },
      '(max-width: 767px)': () => {
        gsap.fromTo(leftItems, { opacity: 0, y: 60 }, { ...reveal, scrollTrigger: { trigger: left, start: 'top 85%', once: true } })
        gsap.fromTo(rightItems, { opacity: 0, y: 60 }, { ...reveal, scrollTrigger: { trigger: right, start: 'top 85%', once: true } })
      },
    })
  }, root)
}

// Hot themes drive the Themes row; fall back to the full list when the
// hot endpoint is empty, and to no cards when the API is unreachable.
const loadCards = async (): Promise<void> => {
  try {
    const [all, hot] = await Promise.all([fetchThemes(), fetchHotThemes()])
    const names = hot.length > 0 ? hot : all.map((theme) => theme.name)
    hotCards.value = names.slice(0, 6).map((name) => all.find((theme) => theme.name === name) ?? { name })
  } catch {
    hotCards.value = []
  }
}

onMounted(() => {
  setupDetails()
  void loadCards()
})

onBeforeUnmount(() => {
  detailsContext?.revert()
  ScrollTrigger.getAll().forEach((trigger) => trigger.kill())
})
</script>

<template>
  <main class="about-page">
    <AboutHero />

    <section class="details-band">
      <div class="details-inner">
        <div ref="detailsEl" class="details-grid">
          <div ref="detailsLeftEl" class="details-col">
            <div class="droog">
              <p class="section-title">About Me</p>
            </div>
            <div class="details-copy">
              <p>{{ t('about.desc') }}</p>
              <p>{{ t('about.stackValue') }}</p>
              <p>{{ t('about.deploymentValue') }}</p>
            </div>
            <div id="howto" class="details-howto scroll-mt-20">
              <p class="howto-text">
                {{ t('howto.introPre') }}<NuxtLink to="/" class="howto-link">{{ t('howto.introLink') }}</NuxtLink>{{ t('howto.introPost') }}
              </p>
              <p class="howto-hint">{{ t('howto.mdHint') }}</p>
              <code class="howto-code">![name]({{ howToUrl }})</code>
            </div>
          </div>

          <div ref="detailsRightEl" class="details-col">
            <div class="droog">
              <p class="section-title">Skills</p>
            </div>
            <div class="details-group">
              <div class="reveal-subheading">
                <p class="details-subheading">{{ t('about.techStack') }}</p>
              </div>
              <ul class="details-list">
                <li v-for="skill in skills" :key="skill">{{ skill }}</li>
              </ul>
            </div>
            <div class="details-group">
              <div class="reveal-subheading">
                <p class="details-subheading">{{ t('about.contribute') }}</p>
              </div>
              <ul class="details-list">
                <li><a :href="issuesUrl" target="_blank" rel="noopener">{{ t('about.feedback') }}</a></li>
                <li><a :href="contributeUrl" target="_blank" rel="noopener">{{ t('about.contributeLink') }}</a></li>
                <li><a :href="repoUrl" target="_blank" rel="noopener">GitHub</a></li>
              </ul>
            </div>
          </div>
        </div>

        <!-- <AboutCarousel
          v-if="themeItems.length"
          title="Themes"
          :more-label="t('about.goThemes')"
          more-to="/"
          :view-label="t('about.view')"
          :items="themeItems"
        />
        -->

        <AboutSocials :title="t('about.socials')" :links="socials" />
      </div>
    </section>

    <StarPleaCard class="about-plea" />
    <BackToTop />
    <HomeFooter />
  </main>
</template>

<style scoped>
.about-page {
  display: flex;
  width: 100%;
  flex-direction: column;
  margin: 0;
  padding: 0;
}

.details-band {
  position: relative;
  z-index: 3;
  display: flex;
  width: 100%;
  height: fit-content;
  flex-direction: column;
  padding: 3rem 0;
  background: var(--loli-cream);
}

.details-inner {
  display: flex;
  width: 100%;
  max-width: 60rem;
  flex-direction: column;
  gap: 4.5rem;
  margin: 0 auto;
  padding: 0 1.5rem;
}

.details-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 3rem;
  align-items: start;
  justify-content: space-between;
}

.details-col {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-title {
  margin: 0;
  font-size: 1.875rem;
  font-weight: 700;
  line-height: 1.1;
  color: var(--text-color);
  user-select: none;
}

.details-copy {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 1.5rem;
  color: var(--text-color);
  font-size: 1rem;
  line-height: 1.75;
  overflow-wrap: break-word;
}

.details-copy p {
  margin: 0;
}

.details-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.details-subheading {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-muted);
}

.details-list {
  margin: 0;
  padding-left: 1.5rem;
  list-style: disc;
  color: var(--text-color);
}

.details-list li + li {
  margin-top: 0.75rem;
}

.details-list a {
  color: var(--text-color);
  text-decoration: underline;
  text-decoration-color: transparent;
  text-underline-offset: 3px;
  transition: text-decoration-color 0.2s ease;
}

.details-list a:hover {
  text-decoration-color: currentColor;
}

.details-howto {
  padding: 1rem 1.25rem;
  background: var(--loli-surface);
  border-radius: 1rem;
}

.howto-text {
  margin: 0 0 0.75rem;
  font-size: 0.9375rem;
  line-height: 1.7;
  color: var(--text-color);
}

.howto-link {
  font-weight: 650;
  color: var(--loli-pink);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.howto-hint {
  margin: 0 0 0.5rem;
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.howto-code {
  display: block;
  padding: 0.625rem 0.875rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.8125rem;
  color: var(--text-color);
  overflow-wrap: anywhere;
  background: var(--loli-cream);
  border-radius: 0.5rem;
}

.about-plea {
  width: 100%;
  max-width: 60rem;
  margin: 3rem auto 0;
  padding: 0 1.5rem;
}

@media (min-width: 640px) {
  .details-inner {
    padding: 0;
  }

  .about-plea {
    padding: 0;
  }
}

@media (min-width: 768px) {
  .details-grid {
    grid-template-columns: 3fr 2fr;
    gap: 1.5rem;
  }
}

@media (min-width: 1024px) {
  .section-title {
    font-size: 2.25rem;
  }

  .details-subheading {
    font-size: 1.5rem;
  }
}
</style>
