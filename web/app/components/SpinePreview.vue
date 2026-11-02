<script setup lang="ts">
// SpinePreview reuses the self-contained Spine player page in an iframe so
// the gallery, generated preview and third-party embed share one renderer.
const props = defineProps<{
  model: string
  name: string
  text: string
}>()

const config = useRuntimeConfig()
const apiBase = ((config.public.apiBase as string) || '').replace(/\/+$/, '')

const src = computed(() => {
  if (!props.model) return ''
  const query = new URLSearchParams({
    model: props.model,
    name: props.name.trim() || 'demo',
    text: props.text || '{n}',
  })
  return `${apiBase}/spine-player.html?${query.toString()}`
})
</script>

<template>
  <div class="preview-body">
    <div class="flex justify-center">
      <iframe
        v-if="src"
        :src="src"
        width="360"
        height="520"
        :title="model"
        class="preview-frame border-0 bg-transparent"
        loading="lazy"
        allow="autoplay"
        referrerpolicy="no-referrer-when-downgrade"
      />
    </div>
  </div>
</template>

<style scoped>
.preview-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  min-width: 0;
}

.preview-body > div {
  width: 100%;
}

.preview-frame {
  display: block;
  width: 100%;
  height: auto;
  max-width: 360px;
  aspect-ratio: 4 / 5;
  margin: 0 auto;
}
</style>
