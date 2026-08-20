import { defineConfig, presetUno, presetWebFonts } from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetWebFonts({
      provider: 'google',
      fonts: {
        sans: 'Nunito',
      },
    }),
  ],
  theme: {
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
