import { resolveAppVersion } from '../scripts/app-version.mjs'

// ogImage: absolute URL for Open Graph / Twitter Card link previews.
// Telegram, Discord, etc. require an absolute URL — relative paths do
// not work. Uses BASE_URL (or NUXT_PUBLIC_BASE_URL) at build time so
// the SSG output bakes in the correct domain. Falls back to empty if
// not set (previews won't show an image, but the page still works).
const ogBase = (process.env.NUXT_PUBLIC_BASE_URL || process.env.BASE_URL || '').replace(/\/+$/, '')
const ogImage = ogBase ? `${ogBase}/images/lolicount-icon.png` : '/images/lolicount-icon.png'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  devServer: { port: 3721, strictPort: true },
  modules: ['@unocss/nuxt'],
  css: ['~/assets/css/theme.css'],
  // /themes is a route alias of the playground landing page kept for
  // previously shared ?theme= links; nothing links to it anymore, so it
  // must be listed explicitly for the static generator to emit it.
  nitro: {
    prerender: {
      routes: ['/themes', '/support'],
    },
  },
  // Back-end API base. The SSG frontend is served from the same origin
  // as the Go binary (embedded dist), so a relative base keeps every
  // request same-origin regardless of the deployed port. In dev the
  // Nuxt server (3721) talks to the Go server (9721) on a different
  // origin, so NUXT_PUBLIC_API_BASE must be set (see dev:web script).
  runtimeConfig: {
    public: {
      // apiBase: same-origin back-end origin (empty in SSG → relative).
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '',
      // baseUrl: public domain for embed links (BASE_URL on the back-end).
      // Injected at build/dev time so SSG output bakes it in without a
      // runtime fetch; also available via GET /api/config as a fallback.
      baseUrl: process.env.NUXT_PUBLIC_BASE_URL || process.env.BASE_URL || '',
      // appVersion: release tag shown in the footer. CD passes it as a
      // Docker build arg, tag builds expose GITHUB_REF_NAME, local builds
      // read the nearest git tag. Empty = omit the version, so a shallow
      // CI checkout never shows a misleading number.
      appVersion: resolveAppVersion(),
    },
  },
  app: {
    // Default lang is en; useI18n reconciles to the stored preference on
    // mount. Setting it here keeps SSR output consistent. A tiny inline
    // script (below) runs before hydration to apply the persisted theme,
    // day/night mode and locale to <html>, eliminating the first-paint
    // flash under SSG (SSR always renders the defaults en + gray + light).
    head: {
      htmlAttrs: { lang: 'en' },
      title: 'Lolicount',
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/lolicount-icon.png' },
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'A cute themeable SVG visitor counter / 萌系可换肤 SVG 访问计数器' },
        // Mobile browser chrome color, mirroring the reference's static
        // OS-scheme media query (the in-app toggle does not restyle it).
        { name: 'theme-color', content: '#11111b', media: '(prefers-color-scheme: dark)' },
        // Open Graph + Twitter Card: link previews in Telegram, Discord,
        // Twitter, etc. og:image must be an absolute URL; we use the
        // build-time BASE_URL (or NUXT_PUBLIC_BASE_URL) so the SSG
        // output bakes in the correct domain.
        { property: 'og:title', content: 'Lolicount' },
        { property: 'og:description', content: 'A cute themeable SVG visitor counter / 萌系可换肤 SVG 访问计数器' },
        { property: 'og:type', content: 'website' },
        { property: 'og:image', content: ogImage },
        { name: 'twitter:card', content: 'summary' },
        { name: 'twitter:title', content: 'Lolicount' },
        { name: 'twitter:description', content: 'A cute themeable SVG visitor counter / 萌系可换肤 SVG 访问计数器' },
        { name: 'twitter:image', content: ogImage },
      ],
      script: [
        {
          // No-flash bootstrap: read persisted mode/locale from
          // localStorage and set <html data-mode> / <html lang> before
          // the page paints, so SSG users with a saved preference do
          // not see the default (zh/light) flash on first frame.
          innerHTML: [
            '(function(){try{',
            'var m=localStorage.getItem("lolicount-mode");',
            'if(m!=="light"&&m!=="dark")m=window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light";',
            'if(m==="dark")document.documentElement.setAttribute("data-mode","dark");',
            'var l=localStorage.getItem("lolicount-locale");',
            'if(l==="zh"||l==="en"||l==="jp")document.documentElement.lang=l;',
            '}catch(e){}})();',
          ].join(''),
          tagPosition: 'head',
        },
      ],
    },
  },
})
