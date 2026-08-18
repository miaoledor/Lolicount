// useI18n: minimal locale state shared app-wide. The chosen locale is
// persisted to localStorage and reflected on <html lang> so the value
// survives navigation and is correct for SSG hydration. SSR-safe: the
// initial value is always the default (zh) and is reconciled on mount.

import { locales, dictionaries, type Locale } from '~/i18n/locales'

const STORAGE_KEY = 'lolicount-locale'
const DEFAULT_LOCALE: Locale = 'zh'

const current = ref<Locale>(DEFAULT_LOCALE)

const isValid = (v: string | null): v is Locale =>
  !!v && (locales as string[]).includes(v)

const applyHtmlLang = (locale: Locale) => {
  if (import.meta.client) document.documentElement.lang = locale
}

const load = (): Locale => {
  if (!import.meta.client) return DEFAULT_LOCALE
  const stored = localStorage.getItem(STORAGE_KEY)
  return isValid(stored) ? stored : DEFAULT_LOCALE
}

const persist = (locale: Locale) => {
  if (!import.meta.client) return
  localStorage.setItem(STORAGE_KEY, locale)
  applyHtmlLang(locale)
}

export const useI18n = () => {
  const t = (key: string): string =>
    dictionaries[current.value]?.[key] ?? dictionaries[DEFAULT_LOCALE]?.[key] ?? key

  const setLocale = (locale: Locale) => {
    current.value = locale
    persist(locale)
  }

  const toggle = () =>
    setLocale(current.value === 'zh' ? 'en' : 'zh')

  // Reconcile with persisted preference once on the client. Called from
  // app.vue onMounted so it runs after hydration rather than clobbering
  // the SSR-rendered default.
  const init = () => {
    const stored = load()
    if (stored !== current.value) current.value = stored
    applyHtmlLang(current.value)
  }

  return { t, setLocale, toggle, init, locale: current }
}
