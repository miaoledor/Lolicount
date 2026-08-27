// useEditorApi wraps the editor backend endpoints (preview + export).
// All functions are arrow functions (AGENTS.md front-end rule).

export type EditorImage = {
  src: string
  left: number
  top: number
  width: number
  height: number
}

export type EditorLayer = {
  id: number
  category: string
  zIndex: number
  fixed: boolean
  images: EditorImage[]
}

export type EditorCanvas = {
  width: number
  height: number
}

export type EditorDisplay = {
  size: number
  crop?: { left: number; top: number; width: number; height: number } | null
}

export type EditorRequest = {
  name: string
  canvas: EditorCanvas
  display?: EditorDisplay | null
  layers: EditorLayer[]
  text: string
  fsize: number
  scale: number
  unshowf: boolean
}

export const useEditorApi = () => {
  const config = useRuntimeConfig()
  const base = config.public.apiBase as string

  const previewTheme = async (req: EditorRequest): Promise<string> => {
    const resp = await $fetch<string>(`${base}/api/editor/preview`, {
      method: 'POST',
      body: req,
      headers: { 'Content-Type': 'application/json' },
      responseType: 'text',
    })
    return resp
  }

  const exportTheme = async (req: EditorRequest): Promise<Blob> => {
    const resp = await $fetch<Blob>(`${base}/api/editor/export`, {
      method: 'POST',
      body: req,
      headers: { 'Content-Type': 'application/json' },
      responseType: 'blob',
    })
    return resp
  }

  return { previewTheme, exportTheme }
}
