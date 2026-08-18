<script setup lang="ts">
const props = defineProps<{ url: string; name: string }>()
const copied = ref('')

const formats = computed(() => [
  { label: 'SVG 地址', value: props.url },
  { label: 'Img 标签', value: `<img src="${props.url}" alt="${props.name}" />` },
  { label: 'Markdown', value: `![${props.name}](${props.url})` },
])

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
      <button @click="copy(f.value, f.label)" :class="cn('px-2 py-1 text-xs rounded', copied === f.label ? 'bg-green-500 text-white' : 'bg-loli-pink text-white')">
        {{ copied === f.label ? '已复制' : '复制' }}
      </button>
    </div>
  </div>
</template>
