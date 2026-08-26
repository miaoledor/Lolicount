// useApi wraps the back-end read-only endpoints. All functions are arrow
// functions (AGENTS.md front-end rule). The SVG counter URL is built from
// the same apiBase so it works in both dev and SSG output.
//
// publicBase holds the configured BASE_URL (served by GET /api/config).
// When set, embed links use it so the URLs users paste into READMEs point
// at the public domain instead of 127.0.0.1 / a relative path. The live
// preview <img> always uses apiBase (same-origin) to avoid cross-origin /
// cache issues.

export type ThemeInfo = {
  name: string
  kind: 'frame' | 'character'  // runtime-inferred by backend, not a directory-based split
}

export type CounterParams = {
  name: string
  theme?: string
  ftheme?: string
  fsize?: number
  scale?: number
  number?: number
  unshowf?: boolean
  x?: number
  y?: number
  rx?: number
  ry?: number
}

export const useApi = () => {
  const config = useRuntimeConfig()
  const base = config.public.apiBase as string

  // Public origin for embed links. Initialized from runtimeConfig so it
  // works in SSG output (baked at build time via NUXT_PUBLIC_BASE_URL /
  // BASE_URL) without a runtime fetch. fetchConfig() refreshes it from the
  // back-end at runtime so a deployed SSG bundle picks up a BASE_URL
  // change without a rebuild.
  const publicBase = ref((config.public.baseUrl as string || '').replace(/\/+$/, ''))

  const fetchThemes = async (): Promise<ThemeInfo[]> => {
    const data = await $fetch<{ themes: ThemeInfo[] }>(`${base}/api/themes`)
    return data.themes ?? []
  }

  const fetchFThemes = async (): Promise<string[]> => {
    const data = await $fetch<{ fthemes: string[] }>(`${base}/api/fthemes`)
    return data.fthemes ?? []
  }

  const fetchConfig = async () => {
    try {
      const data = await $fetch<{ baseUrl: string }>(`${base}/api/config`)
      publicBase.value = (data.baseUrl ?? '').replace(/\/+$/, '')
    } catch {
      // Non-fatal: keep whatever was injected at build time (or empty).
      // Blankining publicBase here would discard a valid BASE_URL baked
      // into the SSG bundle just because the runtime fetch failed.
    }
  }

  const buildCounterUrl = (params: CounterParams, origin?: string): string => {
    const q = new URLSearchParams()
    if (params.theme) q.set('theme', params.theme)
    if (params.ftheme) q.set('ftheme', params.ftheme)
    if (params.fsize && params.fsize > 0) q.set('fsize', String(params.fsize))
    if (params.scale && params.scale > 0) q.set('scale', String(params.scale))
    if (params.number && params.number > 0) q.set('number', String(params.number))
    if (params.unshowf) q.set('unshowf', 'true')
    if (params.x !== undefined) q.set('x', String(params.x))
    if (params.y !== undefined) q.set('y', String(params.y))
    if (params.rx !== undefined) q.set('rx', String(params.rx))
    if (params.ry !== undefined) q.set('ry', String(params.ry))
      const qs = q.toString()
    // Prefer the explicit origin (public domain) when provided and non-empty;
    // otherwise use the same-origin apiBase for live preview.
    const root = origin && origin.length > 0 ? origin : base
    return `${root}/@${encodeURIComponent(params.name)}${qs ? `?${qs}` : ''}`
  }

  const buildEmbedFormats = (url: string, name: string) => ({
    svg: url,
    img: `<img src="${url}" alt="${name}" />`,
    markdown: `![${name}](${url})`,
  })

  return { fetchThemes, fetchFThemes, fetchConfig, buildCounterUrl, buildEmbedFormats, publicBase }
}
