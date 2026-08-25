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
  <img
    v-show="visible"
    src="/images/up.webp"
    alt="back to top"
    class="fixed bottom-6 right-6 z-50 w-20 h-20 cursor-pointer transition-all duration-300"
    :class="visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
    @click="scrollToTop"
  />
</template>
