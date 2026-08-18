// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  devServer: { port: 3721, strictPort: true },
  modules: ['@unocss/nuxt'],
  // Back-end API base. In dev the Go server runs on :8721; the Nuxt dev
  // server uses a different port so we point at the absolute origin.
  // At SSG build time the same origin is baked into the static output.
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://127.0.0.1:8721',
    },
  },
  app: {
    head: {
      title: 'Lolicount',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '萌系可换肤 SVG 访问计数器' },
      ],
    },
  },
})
