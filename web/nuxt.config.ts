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
    },
  },
  app: {
    // Default lang is en; useI18n reconciles to the stored preference on
    // mount. Setting it here keeps SSR output consistent. A tiny inline
    // script (below) runs before hydration to apply the persisted theme
    // and locale to <html>, eliminating the first-paint flash under SSG
    // (SSR always renders the defaults en + gray).
    head: {
      htmlAttrs: { lang: 'en' },
      title: 'Lolicount',
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/lolicount-icon.png' },
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'A cute themeable SVG visitor counter / 萌系可换肤 SVG 访问计数器' },
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
          // No-flash bootstrap: read persisted theme/locale from
          // localStorage and set <html data-theme> / <html lang> before
          // the page paints, so SSG users with a saved preference do
          // not see the default (zh/gray) flash on first frame.
          innerHTML: [
            '(function(){try{',
            'var t=localStorage.getItem("lolicount-theme");',
            'if(t==="pink"||t==="gray")document.documentElement.setAttribute("data-theme",t);',
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
