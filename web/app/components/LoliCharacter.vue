<script setup lang="ts">
// LoliCharacter shows a randomly assembled character portrait by
// requesting the back-end SVG directly (@demo?theme=...).
// Clicking the portrait re-loads the image with a cache-buster so the
// back-end assembles a fresh random combination. The displayed size is
// controlled by the theme's display.json (size + crop) on the back-end,
// so all character themes render at a consistent height when configured
// with the same size.
const props = defineProps<{
  theme: string
}>()

const { buildCounterUrl } = useApi()
const { t } = useI18n()

const reloadKey = ref(0)
const loading = ref(true)

const src = computed(() => {
  const base = buildCounterUrl({
    name: 'demo',
    theme: props.theme,
    number: 0,
    unshowf: true,
  })
  const key = reloadKey.value
  return key > 0 ? `${base}&_=${key}` : base
})

const onImgLoad = () => {
  loading.value = false
}

const reroll = () => {
  loading.value = true
  reloadKey.value++
}

watch(() => props.theme, () => {
  loading.value = true
  reloadKey.value = 0
})
</script>

<template>
  <div
    class="relative cursor-pointer overflow-hidden rounded-lg"
    :title="loading ? t('loli.rolling') : t('loli.reroll')"
    @click="reroll"
  >
    <img
      v-if="src"
      :src="src"
      :alt="props.theme"
      class="max-w-full object-contain"
      :class="cn(loading && 'opacity-40')"
      @load="onImgLoad"
      @error="onImgLoad"
    />
    <div
      v-if="loading"
      class="absolute inset-0 flex items-center justify-center text-sm text-gray-400"
    >
      {{ t('loli.loading') }}
    </div>
  </div>
</template>
