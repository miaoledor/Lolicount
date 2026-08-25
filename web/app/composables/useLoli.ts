// useLoli implements the random character assembly for the front-end
// preview (LoliCharacter.vue). Each character theme ships a ren.json
// layer manifest + a ren.config.json (canvas dims + part-category index
// ranges) + a ren/ directory of transparent layer images under
// web/public/<theme>/. Five part categories (brow / eye / mouth / face /
// lass) are each picked at random by index range, then overlaid by
// absolute coordinates to compose a full portrait.

export type RenLayer = {
  name: string
  left: number
  top: number
  width: number
  height: number
  visible: number
  layer_id: number
  group_layer_id: number
}

export type RenConfig = {
  canvasW: number
  canvasH: number
  ranges: Record<string, { first: number; last: number }>
}

export type LoliParts = {
  loliBodyLeft: string
  loliBodyTop: string
  loliEyeLeft: string
  loliEyeTop: string
  loliBrowLeft: string
  loliBrowTop: string
  loliMouthLeft: string
  loliMouthTop: string
  loliFaceLeft: string
  loliFaceTop: string
  body: string
  eye: string
  brow: string
  mouth: string
  face: string
  bbox: { left: number; top: number; width: number; height: number }
}

// layerUrl returns the public URL for a layer image, trying webp first
// then png (mirrors the back-end readLayerDataURI extension order).
const layerUrl = (base: string, layerId: number) => `${base}/ren/${layerId}.webp`

// Cache fetched ren.json + ren.config.json per theme so repeated rerolls
// don't re-fetch. Keyed by the public base path.
const manifestCache = new Map<string, { layers: RenLayer[]; config: RenConfig }>()

const loadManifest = async (base: string): Promise<{ layers: RenLayer[]; config: RenConfig }> => {
  const cached = manifestCache.get(base)
  if (cached) return cached
  const [renRes, cfgRes] = await Promise.all([
    $fetch<RenLayer[]>(`${base}/ren.json`),
    $fetch<RenConfig>(`${base}/ren.config.json`),
  ])
  const entry = { layers: renRes, config: cfgRes }
  manifestCache.set(base, entry)
  return entry
}

export const getLoli = async (theme: string): Promise<LoliParts> => {
  // lian lives at the public root (/ren.json, /ren/); other themes live
  // under /<theme>/ (e.g. /hinata/ren.json, /hinata/ren/).
  const base = theme === 'lian' ? '' : `/${theme}`

  const { layers, config } = await loadManifest(base)

  const pick = (cat: string) => {
    const r = config.ranges[cat]
    if (!r) throw new Error(`theme ${theme}: missing range for ${cat}`)
    return layers[randomNum(r.first, r.last)]!
  }

  const lass = pick('lass')
  const eye = pick('eye')
  const brow = pick('brow')
  const mouth = pick('mouth')
  const face = pick('face')

  const parts = [lass, eye, brow, mouth, face]
  const left = Math.min(...parts.map((p) => p.left))
  const top = Math.min(...parts.map((p) => p.top))
  const bbox = {
    left,
    top,
    width: Math.max(...parts.map((p) => p.left + p.width)) - left,
    height: Math.max(...parts.map((p) => p.top + p.height)) - top,
  }

  const blobUrls = await Promise.all(
    [lass, eye, brow, mouth, face].map(async (p) => {
      const blob = await $fetch<Blob>(layerUrl(base, p.layer_id), { responseType: 'blob' })
      return URL.createObjectURL(blob)
    }),
  )

  return {
    loliBodyLeft: `${lass.left}px`,
    loliBodyTop: `${lass.top}px`,
    loliEyeLeft: `${eye.left}px`,
    loliEyeTop: `${eye.top}px`,
    loliBrowLeft: `${brow.left}px`,
    loliBrowTop: `${brow.top}px`,
    loliMouthLeft: `${mouth.left}px`,
    loliMouthTop: `${mouth.top}px`,
    loliFaceLeft: `${face.left}px`,
    loliFaceTop: `${face.top}px`,
    body: blobUrls[0]!,
    eye: blobUrls[1]!,
    brow: blobUrls[2]!,
    mouth: blobUrls[3]!,
    face: blobUrls[4]!,
    bbox,
  }
}

export const useLoli = () => ({ getLoli })
