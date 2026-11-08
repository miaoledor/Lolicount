type ThemeSummary = { name: string; variants?: number }

/** Count all available themes; variation maxima only use reported positive integers. */
export const summarizeThemes = (themes: readonly ThemeSummary[]) => {
  let maxVariants: number | null = null
  let mostVariedTheme: string | null = null
  for (const theme of themes) {
    const variants = theme.variants
    if (variants === undefined || !Number.isSafeInteger(variants) || variants <= 0) continue
    if (maxVariants === null || variants > maxVariants) {
      maxVariants = variants
      mostVariedTheme = theme.name
    }
  }
  return { total: themes.length, maxVariants, mostVariedTheme }
}
