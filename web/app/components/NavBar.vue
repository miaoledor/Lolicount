<script setup lang="ts">
import gsap from 'gsap'
const { t } = useI18n()
const { mode, toggleMode } = useTheme()
const { muted, toggleMute } = useSoundFx()
const { stars, repoUrl, fetchStars, formatStars } = useGitHub()
const route = useRoute()

useEases()

// Day/night icon entrance, ported from the reference's next-themes
// toggle: the glyph spins in from -180deg (light) or -360deg (dark)
// over 1s with the cardReveal curve, on mount and every mode change.
const modeIconEl = ref<HTMLElement | null>(null)

const playModeIconSpin = (): void => {
  if (!modeIconEl.value) return
  gsap.fromTo(
    modeIconEl.value,
    { rotation: mode.value === 'dark' ? -360 : -180 },
    { rotation: 0, duration: 1, ease: 'cardReveal', overwrite: true },
  )
}

onMounted(playModeIconSpin)
watch(mode, playModeIconSpin, { flush: 'post' })

const isLinkActive = (href: string) => route.path === href || route.path.startsWith(href + '/')
const navLinks = [
  { href: '/', label: 'nav.playground', isRoute: true },
  { href: '/about', label: 'nav.howAbout', isRoute: true },
] as const
const menuOpen = ref(false)
const startupDone = ref(false)
const headerEl = ref<HTMLElement | null>(null)
const menuPanelEl = ref<HTMLElement | null>(null)
const hamburgerIconEl = ref<HTMLElement | null>(null)
const isTouch = useIsTouch()

// Faded stroke glyphs for the takeover-menu tiles, echoing the reference
// site's icon cards (currentColor so they follow the theme).
const linkIcons: Record<string, string> = {
  '/': '<svg viewBox="0 0 24 24" width="100%" height="100%" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M7.5 5h9A4.5 4.5 0 0 1 21 9.5v4A4.5 4.5 0 0 1 16.5 18h-9A4.5 4.5 0 0 1 3 13.5v-4A4.5 4.5 0 0 1 7.5 5Z"/><path d="M6 10.8h3.6M7.8 9v3.6"/><path d="M15 10h.01M17.8 12.4h.01"/></svg>',
  '/about': '<svg viewBox="0 0 24 24" width="100%" height="100%" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="8.5"/><path d="M12 11.2V16"/><path d="M12 7.9h.01"/></svg>',
}

// Takeover-menu choreography (panel fade, tile rise, ghost slide, action
// fly-out, hamburger spin, pancake easter egg) lives in the composable;
// NavBar only supplies the element refs and the menuOpen state.
const { pancake, toggleMenu, closeMenu } = useTakeoverMenu({
  menuOpen,
  panelEl: menuPanelEl,
  headerEl,
  hamburgerIconEl,
  isTouch,
  startupDone,
})

// Header entrance, ported from the reference's startup timeline: the
// hamburger and right-side cluster drop in from y:-32 over 0.8s with a
// .05 stagger from the end. startupDone gates the menu's action fly-out
// so the two choreographies never overlap.
onMounted(() => {
  gsap.fromTo('.nav-hamburger', { opacity: 0, y: -32 }, { opacity: 1, y: 0, duration: 0.8, ease: 'power3.out' })
  gsap.fromTo(
    '.nav-link, .nav-actions > *',
    { opacity: 0, y: -32 },
    {
      opacity: 1,
      y: 0,
      duration: 0.8,
      ease: 'power3.out',
      stagger: { each: 0.05, from: 'end' },
      onComplete: () => {
        gsap.set('.nav-hamburger, .nav-link, .nav-actions > *', { clearProps: 'transform' })
        startupDone.value = true
      },
    },
  )
})

// Star count is non-critical and rate-limited by GitHub; start it only
// after Nuxt finishes hydration so it cannot compete with first paint.
onNuxtReady(() => { void fetchStars() })
</script>

<template>
  <header ref="headerEl" class="nav-bar">
    <div class="nav-wrapper">
      <button
        type="button"
        class="nav-hamburger"
        data-ripple
        :aria-label="menuOpen ? t('nav.closeMenu') : t('nav.features')"
        :aria-expanded="menuOpen"
        @click="toggleMenu"
      >
        <span ref="hamburgerIconEl" class="nav-hamburger-icon" aria-hidden="true">
          <span v-if="pancake && !menuOpen" class="nav-pancake">🥞</span>
          <svg v-else-if="menuOpen" class="nav-hamburger-svg" viewBox="0 0 256 256" width="24" height="24">
            <path
              fill="currentColor"
              d="M232.49,80.49l-128,128a12,12,0,0,1-17,0l-56-56a12,12,0,1,1,17-17L96,183,215.51,63.51a12,12,0,0,1,17,17Z"
            />
          </svg>
          <svg v-else class="nav-hamburger-svg" viewBox="0 0 256 256" width="24" height="24">
            <path
              fill="currentColor"
              d="M228,128a12,12,0,0,1-12,12H40a12,12,0,0,1,0-24H216A12,12,0,0,1,228,128ZM40,76H216a12,12,0,0,0,0-24H40a12,12,0,0,0,0,24ZM216,180H40a12,12,0,0,0,0,24H216a12,12,0,0,0,0-24Z"
            />
          </svg>
        </span>
      </button>

      <nav class="nav-links" aria-label="Lolicount">
        <NuxtLink
          v-for="link in navLinks"
          :key="link.href"
          :to="link.href"
          class="nav-link"
          data-ripple
          :class="{ 'nav-link-active': isLinkActive(link.href) }"
        >{{ t(link.label) }}</NuxtLink>
      </nav>

      <div class="nav-actions">
        <button type="button" class="nav-icon-button nav-sound" data-ripple :title="t('nav.sound')" :aria-label="t('nav.sound')" :aria-pressed="String(!muted)" @click="toggleMute">
          <svg v-if="muted" class="nav-sound-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
            <path fill="currentColor" d="M11 4.5 6.5 8H3.5a.5.5 0 0 0-.5.5v7a.5.5 0 0 0 .5.5h3L11 20.5v-16Z" opacity=".55" />
            <path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" d="m16 9.5 5 5m0-5-5 5" />
          </svg>
          <svg v-else class="nav-sound-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
            <path fill="currentColor" d="M11 4.5 6.5 8H3.5a.5.5 0 0 0-.5.5v7a.5.5 0 0 0 .5.5h3L11 20.5v-16Z" />
            <path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" d="M15.5 9.2a4 4 0 0 1 0 5.6m2.6-8.2a7.4 7.4 0 0 1 0 10.4" />
          </svg>
        </button>
        <LangSwitch nav />

        <a
          :href="repoUrl"
          target="_blank"
          rel="noopener"
          class="nav-icon-button nav-star"
          data-ripple
          :title="t('nav.star')"
          :aria-label="t('nav.star')"
        >
          <svg class="nav-star-icon" viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            <path
              fill="currentColor"
              d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"
            />
          </svg>
          <span v-if="stars != null" class="nav-star-count">{{ formatStars(stars) }}</span>
        </a>

        <!-- Sun fill in light, moon fill in dark: the reference's
             next-themes toggle (tabler fill glyphs on the 256 grid). -->
        <button type="button" class="nav-icon-button nav-mode" data-ripple :title="t('nav.mode')" :aria-label="t('nav.mode')" @click="toggleMode">
          <span ref="modeIconEl" class="nav-theme-icon">
            <svg v-if="mode === 'dark'" viewBox="0 0 256 256" width="18" height="18" aria-hidden="true">
              <path fill="currentColor" d="M235.54,150.21a104.84,104.84,0,0,1-37,52.91A104,104,0,0,1,32,120,103.09,103.09,0,0,1,52.88,57.48a104.84,104.84,0,0,1,52.91-37,8,8,0,0,1,10,10,88.08,88.08,0,0,0,109.8,109.8,8,8,0,0,1,10,10Z" />
            </svg>
            <svg v-else viewBox="0 0 256 256" width="18" height="18" aria-hidden="true">
              <path fill="currentColor" d="M120,40V16a8,8,0,0,1,16,0V40a8,8,0,0,1-16,0Zm8,24a64,64,0,1,0,64,64A64.07,64.07,0,0,0,128,64ZM58.34,69.66A8,8,0,0,0,69.66,58.34l-16-16A8,8,0,0,0,42.34,53.66Zm0,116.68-16,16a8,8,0,0,0,11.32,11.32l16-16a8,8,0,0,0-11.32-11.32ZM192,72a8,8,0,0,0,5.66-2.34l16-16a8,8,0,0,0-11.32-11.32l-16,16A8,8,0,0,0,192,72Zm5.66,114.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32-11.32ZM48,128a8,8,0,0,0-8-8H16a8,8,0,0,0,0,16H40A8,8,0,0,0,48,128Zm80,80a8,8,0,0,0-8,8v24a8,8,0,0,0,16,0V216A8,8,0,0,0,128,208Zm112-88H216a8,8,0,0,0,0,16h24a8,8,0,0,0,0-16Z" />
            </svg>
          </span>
        </button>
      </div>
    </div>

  </header>

  <!-- Always-mounted takeover panel. Teleported to <body> (like the
       reference's portal) so it stacks under the header in the root
       context: z-45 < header z-50, instead of covering the header's
       own content from inside its stacking context. -->
  <Teleport to="body">
    <div ref="menuPanelEl" class="nav-menu-panel" :aria-hidden="menuOpen ? undefined : 'true'" aria-label="Menu" @click="closeMenu">
      <div class="nav-menu-inner">
        <div class="nav-menu-grid-wrap">
          <nav class="nav-menu-grid" :aria-label="t('nav.features')">
            <NuxtLink
              v-for="(link, i) in navLinks"
              :key="link.href"
              :to="link.href"
              class="nav-tile"
              :class="{ 'nav-tile-large': i === 0 }"
            >
              <span class="nav-tile-ghost" aria-hidden="true">
                <span v-html="linkIcons[link.href]" />
              </span>
              <span class="nav-tile-fade" aria-hidden="true" />
              <span class="nav-tile-icon" aria-hidden="true">
                <span v-html="linkIcons[link.href]" />
              </span>
              <span class="nav-tile-label">{{ t(link.label) }}</span>
            </NuxtLink>
          </nav>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.nav-bar {
  position: sticky;
  top: var(--frame-gutter);
  z-index: 50;
  height: 64px;
  flex: 0 0 auto;
}

/* Frosted band: blurs page content sliding under the header and fades out
   below it, matching the reference site's header::before treatment. */
.nav-bar::before {
  content: '';
  position: absolute;
  inset: 0 0 -125%;
  z-index: -1;
  pointer-events: none;
  background: var(--loli-surface);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  -webkit-mask-image: linear-gradient(#000, #000000f2 10%, #000000de 20%, #000000bf 30%, #00000094 40%, #0006 50%, #00000040 60%, #00000024 70%, #00000012 80%, #00000005 90%, #0000);
  mask-image: linear-gradient(#000, #000000f2 10%, #000000de 20%, #000000bf 30%, #00000094 40%, #0006 50%, #00000040 60%, #00000024 70%, #00000012 80%, #00000005 90%, #0000);
}

.nav-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  max-width: 96rem;
  padding: 0 1.5rem;
  margin: 0 auto;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.nav-link {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  height: 2.5rem;
  padding: 0 1rem;
  font-size: 0.875rem;
  color: var(--text-muted);
  text-decoration: none;
  border-radius: 0.75rem;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.nav-link:hover,
.nav-link-active {
  color: var(--text-color);
  background: var(--loli-selection);
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

/* The right cluster mirrors the reference header: a uniform run of
   40px icon buttons (sound, language, star, theme color, day/night). */
.nav-icon-button {
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
  text-decoration: none;
  /* 0.3s matches the reference's --fade-bg hover language. */
  transition: background-color 0.3s ease, color 0.3s ease;
}

.nav-icon-button:hover {
  background: var(--loli-selection);
  color: var(--text-color);
}

/* The star button widens to show the live count while there is room;
   narrow screens keep the uniform 40px icon cluster. */
.nav-star-count {
  display: none;
}

@media (min-width: 768px) {
  /* Only widens while a count is actually rendered, so a failed fetch
     keeps the uniform 40px icon cluster. */
  .nav-icon-button.nav-star:has(.nav-star-count) {
    width: auto;
    gap: 0.4rem;
    padding: 0 0.625rem;
  }

  .nav-star-count {
    display: inline-flex;
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--loli-pink);
    white-space: nowrap;
  }
}

.nav-hamburger {
  position: relative;
  overflow: hidden;
  display: none;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  width: 2.5rem;
  height: 2.5rem;
  background: transparent;
  border: 0;
  border-radius: 0.75rem;
  color: var(--text-color);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.nav-theme-icon {
  display: inline-flex;
}

.nav-sound[aria-pressed='true'] {
  color: var(--accent);
}

.nav-hamburger:hover {
  background: var(--loli-selection);
}

/* The icon container is the spin target: it rotates in from -90deg on
   every toggle while its child swaps List <-> X (or the pancake). */
.nav-hamburger-icon {
  display: flex;
  width: 1.5rem;
  height: 1.5rem;
  align-items: center;
  justify-content: center;
}

.nav-hamburger-svg {
  display: block;
}

.nav-pancake {
  font-size: 1.125rem;
  line-height: 1;
  user-select: none;
}

/* Takeover menu (the reference site's mobile pattern): an always-mounted
   fixed panel filling the frame below the header (z-45 < header z-50),
   its opacity/visibility driven by useTakeoverMenu. The tile grid is a
   square, centered in the remaining space, sized from the container
   query width. */
.nav-menu-panel {
  position: fixed;
  top: var(--frame-gutter);
  left: var(--frame-gutter);
  right: var(--frame-gutter);
  bottom: var(--frame-gutter);
  z-index: 45;
  visibility: hidden;
  opacity: 0;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  background: var(--loli-surface);
  border-radius: var(--frame-radius);
  /* pt-16: reserve the sticky header; the bottom padding keeps the grid
     reachable when the mobile browser UI shrinks the visual viewport. */
  padding-top: 4rem;
  padding-bottom: max(4rem, calc(4rem + 100lvh - var(--menu-visible-h, 100svh)));
}

.nav-menu-inner {
  display: flex;
  min-height: 100%;
  width: 100%;
  align-items: safe center;
  justify-content: center;
  padding: 0.75rem;
}

@media (min-width: 40rem) {
  .nav-menu-inner {
    padding: 1.5rem;
  }
}

.nav-menu-grid-wrap {
  container-type: inline-size;
  width: 100%;
  max-width: calc((var(--menu-visible-h, 100svh) - 9.875rem) / 1.5);
}

@media (min-width: 40rem) {
  .nav-menu-grid-wrap {
    max-width: min(36rem, calc(var(--menu-visible-h, 100svh) - 11rem));
  }
}

.nav-menu-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  grid-auto-rows: calc((100cqw - 0.75rem) / 2);
  gap: 0.75rem;
}

/* The first tile spans both rows, like the reference's large tiles. */
.nav-tile-large {
  grid-row: span 2;
}

.nav-tile {
  --nav-tile-bg: var(--loli-tile);
  position: relative;
  display: block;
  overflow: hidden;
  border-radius: 0.75rem;
  padding: 0.75rem;
  background: var(--nav-tile-bg);
  text-decoration: none;
  transition: --nav-tile-bg 0.3s ease-out;
}

@media (min-width: 40rem) {
  .nav-tile {
    border-radius: 1rem;
    padding: 1rem;
  }
}

.nav-tile:hover {
  --nav-tile-bg: var(--loli-tile-hover);
}

.nav-tile:focus-visible {
  outline: 2px solid var(--url-color);
  outline-offset: 2px;
}

.nav-tile-ghost {
  position: absolute;
  display: flex;
  height: 100%;
  right: -20%;
  bottom: -8%;
  transform-origin: bottom center;
  opacity: 0;
  pointer-events: none;
  color: var(--loli-ghost);
  transition: color 0.4s ease;
}

@media (hover: none) and (pointer: coarse) {
  .nav-tile-ghost {
    color: var(--loli-ghost-touch);
  }
}

.nav-tile-ghost svg {
  height: 100%;
  width: auto;
}

.nav-tile-fade {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 4rem;
  pointer-events: none;
  background: linear-gradient(
    to top,
    color-mix(in srgb, var(--nav-tile-bg) 80%, transparent),
    color-mix(in srgb, var(--nav-tile-bg) 35%, transparent) 45%,
    transparent
  );
}

@media (min-width: 40rem) {
  .nav-tile-fade {
    height: 5rem;
  }
}

.nav-tile-icon {
  position: absolute;
  top: 0.75rem;
  left: 0.75rem;
  z-index: 1;
  height: 1.75rem;
  color: var(--text-color);
  pointer-events: none;
}

.nav-tile-icon svg {
  height: 100%;
  width: auto;
}

@media (min-width: 40rem) {
  .nav-tile-icon {
    top: 1rem;
    left: 1rem;
  }

  .nav-tile-large .nav-tile-icon {
    height: 2.25rem;
  }
}

.nav-tile-label {
  position: absolute;
  bottom: 0.75rem;
  left: 0.75rem;
  z-index: 1;
  font-family: 'droog', sans-serif;
  font-size: 1.125rem;
  font-weight: 900;
  line-height: 1;
  text-align: left;
  color: var(--text-color);
}

@media (min-width: 40rem) {
  .nav-tile-label {
    bottom: 1rem;
    left: 1rem;
    font-size: 1.25rem;
  }
}


@media (max-width: 767px) {
  /* Reference mobile header: hamburger on the left, the same uniform
     icon cluster on the right, links hidden. */
  .nav-links {
    display: none;
  }

  .nav-hamburger {
    display: flex;
  }
}
</style>
