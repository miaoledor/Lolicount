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
