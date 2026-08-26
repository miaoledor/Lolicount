import { defineConfig, presetUno } from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
  ],
  theme: {
    fontFamily: {
      sans: 'Nunito, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    },
    colors: {
      loli: {
        // Theme colors are driven by CSS variables so the whole palette
        // can switch (pink/gray) via [data-theme] on <html>. The values
        // are defined in app/assets/css/theme.css.
        pink: 'var(--loli-pink)',
        cream: 'var(--loli-cream)',
      },
    },
  },
})
