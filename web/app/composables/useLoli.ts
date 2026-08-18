// useLoli implements the "莲" random character assembly, ported from
// kungal-forum's setting-panel Loli. A PSD is split into layered transparent
// webp files described by ren.json; five part categories (brow / eye / mouth /
// face / lass) are each picked at random by index range, then overlaid by
// absolute coordinates to compose a full character portrait.
import renData from '~~/public/ren.json'

type RenLayer = {
  name: string
  left: number
  top: number
  width: number
  height: number
  visible: number
  layer_id: number
  group_layer_id: number
}

const layers = renData as RenLayer[]

// Index ranges (1-based, closed) for each part category in ren.json.
// Index 0 (汗/sweat) is skipped; 71-79 are PSD group labels, unused.
const RANGE = {
  brow: [1, 18],
  eye: [19, 36],
  mouth: [37, 56],
  face: [57, 62],
  lass: [63, 70],
} as const

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

export const getLoli = async (): Promise<LoliParts> => {
  const assetUrl = (layerId: number) => `/ren/${layerId}.webp`

  const pick = (range: readonly [number, number]) =>
    layers[randomNum(range[0], range[1])]!

  const lass = pick(RANGE.lass)
  const eye = pick(RANGE.eye)
  const brow = pick(RANGE.brow)
  const mouth = pick(RANGE.mouth)
  const face = pick(RANGE.face)

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
      const blob = await $fetch<Blob>(assetUrl(p.layer_id), { responseType: 'blob' })
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
