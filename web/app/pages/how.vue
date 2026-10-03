<script setup lang="ts">
// How-to page: embed instructions for the counter links. The sample URL
// uses the literal name "name" and the public domain so links users copy
// point at the real origin once publicBase resolves.
const { buildCounterUrl, publicBase } = useApi()
const { t } = useI18n()

const howToUrl = computed(() => buildCounterUrl({ name: 'name' }, publicBase.value))
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">
    <section class="mb-12">
      <h1 class="text-3xl font-bold text-loli-pink mb-4">{{ t('howto.title') }}</h1>
      <p class="text-sm text-gray-600 mb-4">
        {{ t('howto.introPre') }}<NuxtLink to="/" class="text-loli-pink underline">{{ t('howto.introLink') }}</NuxtLink>{{ t('howto.introPost') }}
      </p>
      <p class="text-sm text-gray-500 mb-2">{{ t('howto.mdHint') }} ![name]({{ howToUrl }})</p>
    </section>

    <StarPleaCard class="mb-16" />

    <BackToTop />
  </main>
</template>
