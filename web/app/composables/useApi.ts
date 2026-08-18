// useApi wraps the back-end read-only endpoints. All functions are arrow
// functions (AGENTS.md front-end rule). The SVG counter URL is built from
// the same apiBase so it works in both dev and SSG output.

export type ThemeInfo = {
  name: string
  kind: 'frame' | 'character'
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
  mode?: 'seq' | 'random'
}

export const useApi = () => {
  const config = useRuntimeConfig()
  const base = config.public.apiBase as string

  const fetchThemes = async (): Promise<ThemeInfo[]> => {
    const data = await $fetch<{ themes: ThemeInfo[] }>(`${base}/api/themes`)
    return data.themes ?? []
  }

  const fetchFThemes = async (): Promise<string[]> => {
    const data = await $fetch<{ fthemes: string[] }>(`${base}/api/fthemes`)
    return data.fthemes ?? []
  }

  const buildCounterUrl = (params: CounterParams): string => {
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
  if (params.mode) q.set('mode', params.mode)
    const qs = q.toString()
    return `${base}/@${encodeURIComponent(params.name)}${qs ? `?${qs}` : ''}`
  }

  const buildEmbedFormats = (url: string, name: string) => ({
    svg: url,
    img: `<img src="${url}" alt="${name}" />`,
    markdown: `![${name}](${url})`,
  })

  return { fetchThemes, fetchFThemes, buildCounterUrl, buildEmbedFormats }
}
