<script setup lang="ts">
const props = defineProps<{
  kind: 'spine' | 'live2d' | 'unity'
  model: string
  name: string
  text: string
}>()

const { t } = useI18n()
const config = useRuntimeConfig()
const apiBase = ((config.public.apiBase as string) || '').replace(/\/+$/, '')
const src = computed(() => {
  if (!props.model) return ''
  const query = new URLSearchParams()
  query.set('model', props.model)
  query.set('name', props.name.trim() || 'demo')
  if (props.text) query.set('text', props.text)
  return `${apiBase}/${props.kind}-player.html?${query.toString()}`
})
</script>

<template>
  <div class="flex flex-col items-center">
    <iframe
      v-if="src"
      :src="src"
      width="360"
      height="520"
      :title="model"
      class="border-0 bg-transparent max-w-full"
      loading="lazy"
      allow="autoplay"
      referrerpolicy="no-referrer-when-downgrade"
    />
    <p class="text-xs text-gray-400 mt-2">{{ t('emote.hintClick') }}</p>
  </div>
</template>
