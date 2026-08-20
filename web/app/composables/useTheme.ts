// useTheme: pink/gray color theme shared app-wide. Persisted to
// localStorage and applied to <html data-theme="..."> so the CSS
// variables in theme.css recolor the whole site. SSR-safe: the initial
// value is always the default (pink); init() reconciles on the client.

export type ThemeName = 'pink' | 'gray'

const STORAGE_KEY = 'lolicount-theme'
const DEFAULT_THEME: ThemeName = 'gray'

const current = ref<ThemeName>(DEFAULT_THEME)

const isValid = (v: string | null): v is ThemeName => v === 'pink' || v === 'gray'

const applyHtmlTheme = (theme: ThemeName) => {
  if (import.meta.client) document.documentElement.setAttribute('data-theme', theme)
}

const load = (): ThemeName => {
  if (!import.meta.client) return DEFAULT_THEME
  const stored = localStorage.getItem(STORAGE_KEY)
  return isValid(stored) ? stored : DEFAULT_THEME
}

export const useTheme = () => {
  const setTheme = (theme: ThemeName) => {
    current.value = theme
    if (import.meta.client) localStorage.setItem(STORAGE_KEY, theme)
    applyHtmlTheme(theme)
  }

  const toggle = () => setTheme(current.value === 'pink' ? 'gray' : 'pink')

  // Reconcile with persisted preference once on the client, after
  // hydration, to avoid clobbering SSR output.
  const init = () => {
    const stored = load()
    if (stored !== current.value) current.value = stored
    applyHtmlTheme(current.value)
  }

  return { theme: current, setTheme, toggle, init }
}
