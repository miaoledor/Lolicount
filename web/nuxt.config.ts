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
    // Default lang is zh; useI18n reconciles to the stored preference on
    // mount. Setting it here keeps SSR output consistent. A tiny inline
    // script (below) runs before hydration to apply the persisted theme
    // and locale to <html>, eliminating the first-paint flash under SSG
    // (SSR always renders the defaults zh + gray).
    head: {
      htmlAttrs: { lang: 'zh' },
      title: 'Lolicount',
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/lolicount-icon.png' },
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '萌系可换肤 SVG 访问计数器 / A cute themeable SVG visitor counter' },
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
