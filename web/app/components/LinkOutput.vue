<script setup lang="ts">
// LinkOutput: copyable embed formats. When widgetSnippet is provided,
// the animated renderer's widget or iframe replaces the URL-based formats.
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
  <div class="embed-list">
    <div v-for="f in formats" :key="f.label" class="embed-row">
      <div class="embed-head">
        <span class="embed-label">{{ f.label }}</span>
        <button
          type="button"
          :class="cn('embed-copy', copied === f.label && 'is-copied')"
          @click="copy(f.value, f.label)"
        >
          <svg v-if="copied === f.label" viewBox="0 0 16 16" fill="none" aria-hidden="true">
            <path d="M3.2 8.4 6.4 11.6 12.8 4.8" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <svg v-else viewBox="0 0 16 16" fill="none" aria-hidden="true">
            <rect x="5.4" y="5.4" width="8" height="8" rx="1.6" stroke="currentColor" stroke-width="1.5" />
            <path d="M10.6 3.2H4.2A1.6 1.6 0 0 0 2.6 4.8v6.4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
          {{ copied === f.label ? t('embed.copied') : t('embed.copy') }}
        </button>
      </div>
      <pre class="embed-code">{{ f.value }}</pre>
    </div>
  </div>
</template>

<style scoped>
.embed-list {
  display: grid;
  gap: 0.6rem;
  min-width: 0;
}

.embed-row {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
  padding: 0.6rem 0.7rem;
  border: 1px solid color-mix(in srgb, var(--loli-border) 60%, transparent);
  border-radius: 0.7rem;
  background: color-mix(in srgb, var(--loli-surface) 45%, var(--bg-card));
}

.embed-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.embed-label {
  font-size: 0.72rem;
  font-weight: 650;
  letter-spacing: 0.02em;
  color: var(--text-muted);
}

.embed-code {
  min-width: 0;
  margin: 0;
  overflow-x: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.72rem;
  line-height: 1.5;
  white-space: pre;
  color: var(--text-color);
}

.embed-copy {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  flex: none;
  padding: 0.22rem 0.5rem;
  border: 1px solid color-mix(in srgb, var(--loli-border) 70%, transparent);
  border-radius: 0.45rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  background: var(--bg-card);
  cursor: pointer;
  transition: color 0.2s ease, background 0.2s ease, border-color 0.2s ease;
}

.embed-copy:hover:not(.is-copied) {
  color: var(--loli-pink);
  border-color: var(--loli-pink);
}

.embed-copy.is-copied {
  color: var(--loli-pink);
  background: var(--accent-soft);
  border-color: transparent;
}

.embed-copy svg {
  width: 13px;
  height: 13px;
  flex: none;
}

@media (prefers-reduced-motion: reduce) {
  .embed-copy {
    transition: none;
  }
}
</style>
