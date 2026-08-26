<script setup lang="ts">
// LangSwitch: fixed language picker pinned to the top-right corner of
// the page. Decoupled from NavBar so it stays accessible while scrolling.
const { locale, locales, setLocale } = useI18n()

const shortLabels: Record<string, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}
</script>

<template>
  <div class="lang-switch">
    <button
      v-for="l in locales"
      :key="l"
      type="button"
      class="lang-btn"
      :class="{ 'lang-btn-active': l === locale }"
      @click="setLocale(l)"
    >{{ shortLabels[l] }}</button>
  </div>
</template>

<style scoped>
.lang-switch {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 60;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}
.lang-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #6b7280;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border: 1px solid var(--loli-cream);
  cursor: pointer;
  transition: all 0.2s;
}
.lang-btn:hover {
  color: var(--loli-pink);
  border-color: var(--loli-pink);
}
.lang-btn-active {
  color: #fff;
  background: var(--loli-pink);
  border-color: var(--loli-pink);
}
</style>
