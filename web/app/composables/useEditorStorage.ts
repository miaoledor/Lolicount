// useEditorStorage manages editor state persistence in localStorage.
// Editor content (layers, canvas, text params) is auto-saved under a
// per-draft key so the user can refresh without losing work. Multiple
// drafts are supported, keyed by theme name.

import type { EditorLayer } from './useEditorApi'

export type EditorState = {
  themeName: string
  canvasWidth: number
  canvasHeight: number
  displaySize: number
  layers: EditorLayer[]
  layerIdCounter: number
  counterText: string
  fontSize: number
  scale: number
  unshowFont: boolean
}

const STORAGE_PREFIX = 'lolicount-editor:'
const DRAFT_LIST_KEY = 'lolicount-editor:drafts'

export const useEditorStorage = () => {
  const saveDraft = (state: EditorState) => {
    const key = STORAGE_PREFIX + (state.themeName || 'untitled')
    localStorage.setItem(key, JSON.stringify(state))
    // Track draft name in the list
    const drafts = listDrafts()
    const name = state.themeName || 'untitled'
    if (!drafts.includes(name)) {
      drafts.push(name)
      localStorage.setItem(DRAFT_LIST_KEY, JSON.stringify(drafts))
    }
  }

  // saveDraftAs saves the current state under an explicit name as a
  // new manual draft, distinct from the auto-save draft. If a draft
  // with the same name already exists, it is overwritten.
  const saveDraftAs = (name: string, state: EditorState) => {
    const key = STORAGE_PREFIX + name
    const namedState = { ...state, themeName: name }
    localStorage.setItem(key, JSON.stringify(namedState))
    const drafts = listDrafts()
    if (!drafts.includes(name)) {
      drafts.push(name)
      localStorage.setItem(DRAFT_LIST_KEY, JSON.stringify(drafts))
    }
  }

  const loadDraft = (name: string): EditorState | null => {
    const key = STORAGE_PREFIX + name
    const raw = localStorage.getItem(key)
    if (!raw) return null
    try {
      return JSON.parse(raw) as EditorState
    } catch {
      return null
    }
  }

  const deleteDraft = (name: string) => {
    localStorage.removeItem(STORAGE_PREFIX + name)
    const drafts = listDrafts().filter((d) => d !== name)
    localStorage.setItem(DRAFT_LIST_KEY, JSON.stringify(drafts))
  }

  const listDrafts = (): string[] => {
    const raw = localStorage.getItem(DRAFT_LIST_KEY)
    if (!raw) return []
    try {
      return JSON.parse(raw) as string[]
    } catch {
      return []
    }
  }

  return { saveDraft, saveDraftAs, loadDraft, deleteDraft, listDrafts }
}
