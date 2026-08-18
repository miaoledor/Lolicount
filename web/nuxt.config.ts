// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  devServer: { port: 3721, strictPort: true },
  modules: ['@unocss/nuxt'],
  // Back-end API base. The SSG frontend is served from the same origin
  // as the Go binary (embedded dist), so a relative base keeps every
  // request same-origin regardless of the deployed port. In dev the
  // Nuxt server (3721) talks to the Go server (9721) on a different
  // origin, so NUXT_PUBLIC_API_BASE must be set (see dev:web script).
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '',
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
