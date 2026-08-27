<script setup lang="ts">
// NavBar: fixed top navigation bar, modeled after dicebear's VitePress
// nav. Left: brand icon + title (click cycles the color theme). Center:
// anchor links to page sections. Right: language picker, GitHub Star
// button with count. Collapses into a hamburger menu on narrow screens.

const { t, locale, localeLabels } = useI18n()
const { toggle: toggleTheme } = useTheme()
const { stars, repoUrl, fetchStars, formatStars } = useGitHub()

const navLinks = [
  { href: '#howto', label: 'nav.howto' },
  { href: '#themes', label: 'nav.themes' },
  { href: '#playground', label: 'nav.playground' },
  { href: '/editor', label: 'nav.editor', isRoute: true },
  { href: '#about', label: 'nav.more' },
] as const
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
      <!-- Brand: click cycles the color theme (pink <-> gray) -->
      <button type="button" class="nav-brand" :title="t('nav.theme')" @click="toggleTheme">
        <img src="/images/lolicount-icon.png" alt="Lolicount" class="nav-brand-icon" />
        <span class="nav-brand-text">Lolicount</span>
      </button>

      <!-- Desktop nav links -->
      <nav class="nav-links">
        <NuxtLink
          v-for="link in navLinks"
          :key="link.href"
          :to="link.isRoute ? link.href : undefined"
          :href="link.isRoute ? undefined : link.href"
          class="nav-link"
        >{{ t(link.label) }}</NuxtLink>
      </nav>

      <!-- Right actions -->
      <div class="nav-actions">
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
              d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"
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
        <NuxtLink
          v-for="link in navLinks"
          :key="link.href"
          :to="link.isRoute ? link.href : undefined"
          :href="link.isRoute ? undefined : link.href"
          class="nav-mobile-link"
          @click="closeMenu"
        >{{ t(link.label) }}</NuxtLink>
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
  flex-shrink: 0;
  padding: 0;
  background: transparent;
  border: none;
  cursor: pointer;
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
