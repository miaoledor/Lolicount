<script setup lang="ts">
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
      'fixed bottom-6 right-6 z-50 w-12 h-12 rounded-full shadow-lg transition-all duration-300',
      'bg-loli-pink text-white hover:bg-loli-pink/90 flex items-center justify-center text-xl',
      visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4 pointer-events-none',
    )"
    aria-label="回到顶部"
    @click="scrollToTop"
  >
    ↑
  </button>
</template>
