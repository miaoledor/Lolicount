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
        pink: '#e91e63',
        cream: '#fff5f7',
      },
    },
  },
})
