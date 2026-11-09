<script setup lang="ts">
import type { Locale } from '~/i18n/locales'

const { t, locale, locales, localeLabels, setLocale } = useI18n()

defineProps<{
  // When true, renders in a compact inline style suited for the nav bar.
  nav?: boolean
}>()

const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)
const triggerEl = ref<HTMLElement | null>(null)

const toggleMenu = (): void => {
  open.value = !open.value
}

const close = (): void => {
  open.value = false
}

const choose = (next: Locale): void => {
  setLocale(next)
  close()
  triggerEl.value?.focus()
}

// Close on an outside press (capture phase so a child's stopPropagation
// cannot keep the menu open) and on Escape, which also returns focus.
const onPointerDown = (event: PointerEvent): void => {
  if (open.value && !rootEl.value?.contains(event.target as Node | null)) close()
}

const onKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape' && open.value) {
    close()
    triggerEl.value?.focus()
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', onPointerDown, true)
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onPointerDown, true)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <!-- Nav mode: the 文A trigger opens a real language picker. The menu is a
       sibling of the button because .lang-cycle clips the press ripple. -->
  <div v-if="nav" ref="rootEl" class="lang-switch">
    <button
      ref="triggerEl"
      type="button"
      class="lang-cycle"
      data-ripple
      :title="t('nav.language')"
      :aria-label="t('nav.language')"
      aria-haspopup="menu"
      :aria-expanded="open"
      @click="toggleMenu"
    >
      <svg class="lang-cycle-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
        <text x="1" y="10.5" font-size="10.5" font-weight="700" fill="currentColor">文</text>
        <text x="12" y="20" font-size="11.5" font-weight="700" fill="currentColor">A</text>
      </svg>
      <svg
        :class="cn('lang-chevron', open && 'is-open')"
        viewBox="0 0 12 12"
        width="10"
        height="10"
        aria-hidden="true"
      >
        <path d="M2.5 4.5 6 8l3.5-3.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <Transition name="lang-menu">
      <div v-if="open" class="lang-menu" role="menu" :aria-label="t('nav.language')">
        <button
          v-for="l in locales"
          :key="l"
          type="button"
          role="menuitemradio"
          :class="cn('lang-option', l === locale && 'is-active')"
          :aria-checked="l === locale"
          @click="choose(l)"
        >
          <span>{{ localeLabels[l] }}</span>
          <span v-if="l === locale" class="lang-option-dot" aria-hidden="true" />
        </button>
      </div>
    </Transition>
  </div>

  <!-- Legacy page-inline mode (kept for backward compatibility). -->
  <div
    v-else
    class="flex justify-end gap-1.5 mb-4"
  >
    <button
      v-for="l in locales"
      :key="l"
      type="button"
      class="flex items-center justify-center w-8 h-8 rounded-full text-xs font-semibold border transition-all relative overflow-hidden"
      :class="l === locale
        ? 'text-white bg-[var(--loli-pink)] border-[var(--loli-pink)]'
        : 'text-gray-500 bg-transparent border-[var(--loli-cream)] hover:text-[var(--loli-pink)] hover:border-[var(--loli-pink)]'"
      data-ripple
      @click="setLocale(l)"
    >{{ localeLabels[l] }}</button>
  </div>
</template>

<style scoped>
/* The menu needs a positioned anchor, but must not be clipped by the
   trigger's overflow:hidden (which exists only to contain the ripple). */
.lang-switch {
  position: relative;
  flex: 0 0 auto;
}

/* useRippleFx appends an absolutely positioned span, so every ripple host
   must be its own containing block and clip the circle to the button. */
.lang-cycle {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  background: transparent;
  border: 0;
  border-radius: 0.75rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.lang-cycle:hover {
  background: var(--loli-selection);
  color: var(--text-color);
}

.lang-cycle-icon {
  font-family: 'helvetica-lt-pro', sans-serif;
}

.lang-chevron {
  position: absolute;
  right: 0.15rem;
  bottom: 0.1rem;
  opacity: 0.65;
  transition: transform 0.2s ease;
}

.lang-chevron.is-open {
  transform: rotate(180deg);
}

.lang-menu {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  z-index: 60;
  display: grid;
  gap: 0.15rem;
  min-width: 9rem;
  padding: 0.4rem;
  border: 1px solid var(--loli-border);
  border-radius: 0.8rem;
  background: var(--bg-card);
  box-shadow: 0 14px 32px rgba(76, 79, 105, 0.18);
  transform-origin: top right;
}

.lang-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.65rem;
  border: 0;
  border-radius: 0.55rem;
  font-size: 0.85rem;
  color: var(--text-color);
  text-align: left;
  background: transparent;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.lang-option:hover {
  background: var(--loli-selection);
}

.lang-option.is-active {
  color: var(--loli-pink);
  font-weight: 650;
}

.lang-option-dot {
  width: 0.4rem;
  height: 0.4rem;
  flex: none;
  border-radius: 999px;
  background: var(--loli-pink);
}

.lang-menu-enter-active,
.lang-menu-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.lang-menu-enter-from,
.lang-menu-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.97);
}

@media (prefers-reduced-motion: reduce) {
  .lang-menu-enter-active,
  .lang-menu-leave-active,
  .lang-chevron,
  .lang-option {
    transition: none;
  }
}
</style>
