// useDragPosition converts pointer events over a preview area into x/y
// pixel coordinates. Debounced callback so rapid drags don't flood state.
export const useDragPosition = (onUpdate: (x: number, y: number) => void) => {
  const dragging = ref(false)

  const onPointerDown = () => {
    dragging.value = true
  }

  const onPointerMove = (e: PointerEvent, target: HTMLElement) => {
    if (!dragging.value) return
    const rect = target.getBoundingClientRect()
    const x = Math.round(e.clientX - rect.left)
    const y = Math.round(e.clientY - rect.top)
    onUpdate(x, y)
  }

  const onPointerUp = () => {
    dragging.value = false
  }

  return { dragging, onPointerDown, onPointerMove, onPointerUp }
}
