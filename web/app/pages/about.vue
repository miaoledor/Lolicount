<script setup lang="ts">
const { t } = useI18n()
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">
    <LangSwitch />

    <section id="about" class="mb-16 scroll-mt-20">
      <h2 class="text-2xl font-semibold mb-4 flex items-center gap-2">
        <img src="/images/lolicount-icon.png" alt="" class="h-7 w-7" />
        {{ t('about.title') }}
      </h2>
      <p class="text-sm text-gray-600 mb-6">{{ t('about.desc') }}</p>

      <h3 class="text-lg font-medium mb-2">{{ t('notice.title') }}</h3>
      <p class="text-sm text-gray-500 mb-4">{{ t('notice.desc') }}</p>
      <NoticeBoard />
    </section>

    <SiteFooter />
    <BackToTop />
  </main>
</template>
