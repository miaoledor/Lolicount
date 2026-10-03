<script setup lang="ts">
// About page: the how-to-embed instructions on top, then the project
// details card, closed by the shared StarPleaCard. All copy comes from
// the i18n dictionary.
const { t } = useI18n()
const { buildCounterUrl, publicBase } = useApi()
const { repoUrl } = useGitHub()

// How-to-embed example URL. Uses the literal name "name" and the public
// domain so the sample links users copy point at the real origin once
// publicBase resolves. Same builder as the playground, so the format
// stays consistent.
const howToUrl = computed(() => buildCounterUrl({ name: 'name' }, publicBase.value))

const issuesUrl = `${repoUrl}/issues`
const contributeUrl = `${repoUrl}/blob/main/docs/contributing-themes.md`

// Static facts for the details grid. Links (feedback/contributing) are
// rendered separately as anchor cards.
const details = computed(() => [
  { label: t('about.techStack'), value: t('about.stackValue') },
  { label: t('about.deployment'), value: t('about.deploymentValue') },
  { label: t('about.license'), value: 'AGPL-3.0' },
])
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">

    <!-- How to use: embed instructions for the counter links. -->
    <section id="howto" class="mb-12 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4">{{ t('howto.title') }}</h2>
      <p class="text-sm text-gray-600 mb-4">
        {{ t('howto.introPre') }}<NuxtLink to="/" class="text-loli-pink underline">{{ t('howto.introLink') }}</NuxtLink>{{ t('howto.introPost') }}
      </p>
      <p class="text-sm text-gray-500 mb-2">{{ t('howto.mdHint') }} ![name]({{ howToUrl }})</p>
    </section>

    <!-- About header -->
    <section id="about" class="mb-10 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4 flex items-center gap-2">
        <img src="/images/lolicount-icon.png" alt="" class="h-7 w-7" />
        {{ t('about.title') }}
      </h2>
      <p class="text-sm text-gray-600">{{ t('about.desc') }}</p>
    </section>

    <!-- Project details -->
    <section class="mb-10">
      <h2 class="text-xl font-semibold mb-3">{{ t('about.detailsTitle') }}</h2>
      <div class="rounded-xl bg-loli-cream p-4 grid sm:grid-cols-2 gap-3">
        <div
          v-for="d in details"
          :key="d.label"
          class="rounded-lg bg-white p-3"
        >
          <p class="text-xs text-gray-400 mb-1">{{ d.label }}</p>
          <p class="text-sm font-medium text-gray-700">{{ d.value }}</p>
        </div>
        <a
          :href="issuesUrl"
          target="_blank"
          rel="noopener"
          class="rounded-lg bg-white p-3 group"
        >
          <p class="text-xs text-gray-400 mb-1">{{ t('about.feedback') }}</p>
          <p class="text-sm font-medium text-loli-pink group-hover:underline">
            GitHub Issues →
          </p>
        </a>
        <a
          :href="contributeUrl"
          target="_blank"
          rel="noopener"
          class="rounded-lg bg-white p-3 group"
        >
          <p class="text-xs text-gray-400 mb-1">{{ t('about.contribute') }}</p>
          <p class="text-sm font-medium text-loli-pink group-hover:underline">
            {{ t('about.contributeLink') }} →
          </p>
        </a>
      </div>
    </section>

    <!-- Star plea: begging mascot + live star count. Page closer — the
         old SiteFooter was folded into this card per design. -->
    <StarPleaCard class="mb-4" />

    <BackToTop />
  </main>
</template>
