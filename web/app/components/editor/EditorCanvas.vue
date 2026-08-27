<script setup lang="ts">
import type { EditorRequest } from '~/composables/useEditorApi'

const props = defineProps<{
  request: EditorRequest
  hasLayers: boolean
}>()

const { previewTheme } = useEditorApi()
const { t } = useI18n()

const svg = ref('')
const loading = ref(false)
const error = ref('')

// Debounce preview requests so rapid parameter changes don't flood
// the backend. The editor updates the request object reactively; this
// watch fires on every change but the timeout collapses bursts.
let timer: ReturnType<typeof setTimeout> | null = null

watch(
  () => props.request,
  () => {
    if (!props.hasLayers) {
      svg.value = ''
      error.value = ''
      return
    }
    if (timer) clearTimeout(timer)
    timer = setTimeout(doPreview, 300)
  },
  { deep: true },
)

const doPreview = async () => {
  if (!props.hasLayers) return
  loading.value = true
  error.value = ''
  try {
    svg.value = await previewTheme(props.request)
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || 'preview failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (props.hasLayers) doPreview()
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <div class="editor-canvas">
    <div v-if="loading" class="editor-canvas-loading">...</div>
    <div v-else-if="error" class="editor-canvas-error">{{ error }}</div>
    <div v-else-if="svg" class="editor-canvas-svg" v-html="svg" />
    <div v-else class="editor-canvas-placeholder">
      <p>{{ t('editor.previewHint') }}</p>
    </div>
  </div>
</template>

<style scoped>
.editor-canvas {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  width: 100%;
  min-height: 400px;
  padding: 1rem;
}

.editor-canvas-svg :deep(svg) {
  max-width: 100%;
  height: auto;
}

.editor-canvas-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  width: 100%;
  border: 2px dashed var(--border-color, #333);
  border-radius: 8px;
  color: var(--text-muted, #999);
}

.editor-canvas-loading {
  color: var(--text-muted, #999);
  padding: 2rem;
}

.editor-canvas-error {
  color: #ef4444;
  padding: 2rem;
  font-size: 0.875rem;
}
</style>
