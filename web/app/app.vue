<script setup lang="ts">
const { init, locale, setLocale, t, localeLabels, locales } = useI18n()
onMounted(init)

// Dropdown open state. Closes on outside click or escape.
const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const closeOnOutside = (e: MouseEvent) => {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) open.value = false
}
const closeOnEscape = (e: KeyboardEvent) => {
  if (e.key === 'Escape') open.value = false
}
onMounted(() => {
  document.addEventListener('click', closeOnOutside)
  document.addEventListener('keydown', closeOnEscape)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', closeOnOutside)
  document.removeEventListener('keydown', closeOnEscape)
})

const pick = (l: Locale) => {
  setLocale(l)
  open.value = false
}
</script>

<template>
  <div>
    <div ref="rootRef" class="fixed top-4 right-4 z-50">
      <button
        type="button"
        class="flex items-center gap-1 rounded-full bg-white/80 px-3 py-1 text-sm shadow backdrop-blur transition hover:bg-white"
        :title="t('nav.lang')"
        @click="open = !open"
      >
        <span>{{ localeLabels[locale] }}</span>
        <span class="text-xs transition-transform" :class="open ? 'rotate-180' : ''">▾</span>
      </button>
      <ul
        v-if="open"
        class="absolute right-0 mt-1 min-w-28 overflow-hidden rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
      >
        <li v-for="l in locales" :key="l">
          <button
            type="button"
            class="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-sm transition"
            :class="l === locale ? 'bg-loli-cream text-loli-pink font-semibold' : 'text-gray-700 hover:bg-gray-50'"
            @click="pick(l)"
          >
            <span>{{ localeLabels[l] }}</span>
            <span v-if="l === locale">✓</span>
          </button>
        </li>
      </ul>
    </div>
    <NuxtPage />
  </div>
</template>
