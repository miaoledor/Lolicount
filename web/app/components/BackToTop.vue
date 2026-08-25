<script setup lang="ts">
const { t } = useI18n()
const visible = ref(false)

const onScroll = () => {
  visible.value = window.scrollY > 300
}

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
})
onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<template>
  <button
    :class="cn(
      'fixed bottom-6 right-6 z-50 rounded-full shadow-lg transition-all duration-300',
      'hover:scale-110',
      visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4 pointer-events-none',
    )"
    :aria-label="t('nav.top')"
    @click="scrollToTop"
  >
    <img src="/images/up.webp" alt="back to top" class="w-12 h-12" />
  </button>
</template>
