<script setup lang="ts">
// LinkOutput: copyable embed formats. When widgetSnippet is provided
// (animated themes), the script-widget embed replaces the URL-based
// formats — those themes have no SVG image endpoint.
const props = defineProps<{ url: string; name: string; widgetSnippet?: string }>()
const { t } = useI18n()
const copied = ref('')

const formats = computed(() => {
  if (props.widgetSnippet) {
    return [{ label: t('embed.widget'), value: props.widgetSnippet }]
  }
  return [
    { label: t('embed.svg'), value: props.url },
    { label: t('embed.img'), value: `<img src="${props.url}" alt="${props.name}" />` },
    { label: t('embed.markdown'), value: `![${props.name}](${props.url})` },
  ]
})

const copy = async (text: string, label: string) => {
  await navigator.clipboard.writeText(text)
  copied.value = label
  setTimeout(() => (copied.value = ''), 1500)
}
</script>

<template>
  <div class="space-y-2">
    <div v-for="f in formats" :key="f.label" class="flex items-start gap-2">
      <pre class="flex-1 bg-gray-50 p-2 rounded text-xs overflow-x-auto">{{ f.value }}</pre>
      <button @click="copy(f.value, f.label)" :class="cn('px-2 py-1 text-xs rounded transition', copied === f.label ? 'bg-green-500 text-white' : 'bg-loli-pink text-white hover:bg-loli-pink/90')">
        {{ copied === f.label ? t('embed.copied') : t('embed.copy') }}
      </button>
    </div>
  </div>
</template>
