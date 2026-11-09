// Framework-free gate for the preview reload cooldown. Keeping the timing
// rule out of the page component makes the 2s behavior testable in plain Node
// and gives the duration a single source of truth shared by the timer and the
// indicator.
export const RELOAD_COOLDOWN_MS = 2000

export type ReloadCooldown = {
  cooling: () => boolean
  trigger: () => boolean
  remaining: () => number
  reset: () => void
}

// remainingWholeSeconds converts the gate's millisecond remainder into the
// integer seconds shown inside the indicator. Ceiling keeps the first tick at
// the full duration (2 -> 1 while cooling, 0 once the window closes), so the
// countdown never shows a fraction and never shows zero while still visible.
export const remainingWholeSeconds = (remainingMs: number): number =>
  Math.max(0, Math.ceil(remainingMs / 1000))

// createReloadCooldown returns a gate accepting at most one trigger per
// cooldown window. The clock is injected so tests can advance virtual time
// without real timers; a custom duration is only useful for those tests.
export const createReloadCooldown = (
  now: () => number = () => Date.now(),
  durationMs: number = RELOAD_COOLDOWN_MS,
): ReloadCooldown => {
  let deadline = 0
  const remaining = (): number => Math.max(0, deadline - now())
  return {
    cooling: (): boolean => remaining() > 0,
    trigger: (): boolean => {
      if (remaining() > 0) return false
      deadline = now() + durationMs
      return true
    },
    remaining,
    reset: (): void => { deadline = 0 },
  }
}
