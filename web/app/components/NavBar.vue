<script setup lang="ts">
// NavBar: fixed top navigation bar, modeled after dicebear's VitePress
// nav. Left: brand icon + title. Center: anchor links to page sections.
// Right: theme toggle, language picker, GitHub Star button with count.
// Collapses into a hamburger menu on narrow screens.

const { t, locale, locales, setLocale, localeLabels } = useI18n()
const { theme, setTheme } = useTheme()
const { stars, repoUrl, fetchStars, formatStars } = useGitHub()

const navLinks = [
  { href: '#howto', label: 'nav.howto' },
  { href: '#loli', label: 'nav.loli' },
  { href: '#themes', label: 'nav.themes' },
  { href: '#playground', label: 'nav.playground' },
  { href: '#about', label: 'nav.more' },
] as const

const shortLabels: Record<string, string> = {
  zh: '中',
  en: 'EN',
  jp: 'JP',
}

const themeOptions = [
  { value: 'pink' as const, swatch: '#e91e63' },
  { value: 'gray' as const, swatch: '#6b7280' },
]

const menuOpen = ref(false)

const closeMenu = () => {
  menuOpen.value = false
}

onMounted(() => {
  fetchStars()
})
</script>

<template>
  <header class="nav-bar">
    <div class="nav-wrapper">
      <!-- Brand -->
      <a href="#top" class="nav-brand" @click="closeMenu">
        <img src="/images/lolicount-icon.png" alt="Lolicount" class="nav-brand-icon" />
        <span class="nav-brand-text">Lolicount</span>
      </a>

      <!-- Desktop nav links -->
      <nav class="nav-links">
        <a
          v-for="link in navLinks"
          :key="link.href"
          :href="link.href"
          class="nav-link"
        >{{ t(link.label) }}</a>
      </nav>

      <!-- Right actions -->
      <div class="nav-actions">
        <!-- Theme picker -->
        <div class="nav-action-group" :title="t('nav.theme')">
          <button
            v-for="opt in themeOptions"
            :key="opt.value"
            type="button"
            class="nav-swatch"
            :class="{ 'nav-swatch-active': opt.value === theme }"
            :style="{ backgroundColor: opt.swatch }"
            :aria-label="opt.value"
            @click="setTheme(opt.value)"
          />
        </div>

        <!-- Language picker -->
        <div class="nav-action-group" :title="t('nav.lang')">
          <button
            v-for="l in locales"
            :key="l"
            type="button"
            class="nav-lang"
            :class="{ 'nav-lang-active': l === locale }"
            @click="setLocale(l)"
          >{{ shortLabels[l] }}</button>
        </div>

        <!-- GitHub Star button -->
        <a
          :href="repoUrl"
          target="_blank"
          rel="noopener"
          class="nav-star"
          :title="localeLabels[locale]"
        >
          <svg class="nav-star-icon" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            <path
              fill="currentColor"
              d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25z"
            />
          </svg>
          <span class="nav-star-label">Star</span>
          <span v-if="stars != null" class="nav-star-count">{{ formatStars(stars) }}</span>
        </a>

        <!-- Mobile hamburger -->
        <button
          type="button"
          class="nav-hamburger"
          :class="{ 'nav-hamburger-active': menuOpen }"
          :aria-label="t('nav.features')"
          :aria-expanded="menuOpen"
          @click="menuOpen = !menuOpen"
        >
          <span class="nav-hamburger-line"></span>
          <span class="nav-hamburger-line"></span>
          <span class="nav-hamburger-line"></span>
        </button>
      </div>
    </div>

    <!-- Mobile dropdown menu -->
    <Transition name="nav-menu">
      <nav v-if="menuOpen" class="nav-mobile">
        <a
          v-for="link in navLinks"
          :key="link.href"
          :href="link.href"
          class="nav-mobile-link"
          @click="closeMenu"
        >{{ t(link.label) }}</a>
      </nav>
    </Transition>
  </header>
</template>

<style scoped>
.nav-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: var(--nav-bg, rgba(255, 255, 255, 0.85));
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--loli-cream);
  transition: background-color 0.25s, border-color 0.25s;
}

[data-theme='pink'] .nav-bar {
  --nav-bg: rgba(255, 255, 255, 0.85);
}

[data-theme='gray'] .nav-bar {
  --nav-bg: rgba(249, 250, 251, 0.85);
}

.nav-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 1.5rem;
  max-width: 1440px;
  margin: 0 auto;
}

/* Brand */
.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  flex-shrink: 0;
}
.nav-brand-icon {
  width: 32px;
  height: 32px;
}
.nav-brand-text {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--loli-pink);
  white-space: nowrap;
}

/* Desktop nav links */
.nav-links {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
.nav-link {
  padding: 0 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
  text-decoration: none;
  line-height: 64px;
  transition: color 0.25s;
}
.nav-link:hover {
  color: var(--loli-pink);
}

/* Right actions */
.nav-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}
.nav-action-group {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

/* Theme swatches */
.nav-swatch {
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 9999px;
  border: none;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.2s, box-shadow 0.2s;
}
.nav-swatch:hover {
  opacity: 1;
}
.nav-swatch-active {
  opacity: 1;
  box-shadow: 0 0 0 2px #fff, 0 0 0 4px var(--loli-pink);
}

/* Language picker */
.nav-lang {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #6b7280;
  background: transparent;
  border: 1px solid var(--loli-cream);
  cursor: pointer;
  transition: all 0.2s;
}
.nav-lang:hover {
  color: var(--loli-pink);
  border-color: var(--loli-pink);
}
.nav-lang-active {
  color: #fff;
  background: var(--loli-pink);
  border-color: var(--loli-pink);
}

/* GitHub Star button */
.nav-star {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #4b5563;
  background: var(--loli-cream);
  border: 1px solid var(--loli-pink);
  border-radius: 0.5rem;
  text-decoration: none;
  white-space: nowrap;
  transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
}
.nav-star:hover {
  border-color: var(--loli-pink);
  background: var(--loli-pink);
  color: #fff;
  box-shadow: 0 0 0 3px var(--loli-cream);
}
.nav-star-icon {
  flex-shrink: 0;
}
.nav-star-count {
  display: inline-flex;
  align-items: center;
  padding: 0 0.375rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--loli-pink);
  background: #fff;
  border-radius: 0.375rem;
  line-height: 1.4;
}
.nav-star:hover .nav-star-count {
  color: var(--loli-pink);
  background: #fff;
}

/* Hamburger (mobile) */
.nav-hamburger {
  display: none;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  background: transparent;
  border: 1px solid var(--loli-cream);
  border-radius: 0.5rem;
  cursor: pointer;
}
.nav-hamburger-line {
  display: block;
  width: 1rem;
  height: 2px;
  margin: 0 auto;
  background: var(--loli-pink);
  border-radius: 1px;
  transition: transform 0.25s, opacity 0.25s;
}
.nav-hamburger-active .nav-hamburger-line:nth-child(1) {
  transform: translateY(6px) rotate(45deg);
}
.nav-hamburger-active .nav-hamburger-line:nth-child(2) {
  opacity: 0;
}
.nav-hamburger-active .nav-hamburger-line:nth-child(3) {
  transform: translateY(-6px) rotate(-45deg);
}

/* Mobile dropdown */
.nav-mobile {
  display: none;
  flex-direction: column;
  padding: 0.5rem 1.5rem 1rem;
  border-top: 1px solid var(--loli-cream);
  background: var(--nav-bg, rgba(255, 255, 255, 0.95));
}
.nav-mobile-link {
  padding: 0.75rem 0;
  font-size: 0.95rem;
  font-weight: 500;
  color: #374151;
  text-decoration: none;
  border-bottom: 1px solid var(--loli-cream);
}
.nav-mobile-link:last-child {
  border-bottom: none;
}
.nav-mobile-link:hover {
  color: var(--loli-pink);
}

/* Responsive: hide desktop links + show hamburger under 768px */
@media (max-width: 767px) {
  .nav-links {
    display: none;
  }
  .nav-hamburger {
    display: flex;
  }
  .nav-mobile {
    display: flex;
  }
  .nav-star-label {
    display: none;
  }
  .nav-wrapper {
    padding: 0 1rem;
  }
}

/* Tablet: condense nav links */
@media (max-width: 1024px) and (min-width: 768px) {
  .nav-link {
    padding: 0 0.5rem;
  }
}

/* Mobile menu transition */
.nav-menu-enter-active,
.nav-menu-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.nav-menu-enter-from,
.nav-menu-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
