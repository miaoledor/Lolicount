// Framework-free state machine for the gallery's collapsible source groups.
// A filter (kind or search) opens every matching group by default so results
// are immediately visible, but an explicit click must always be able to close
// one — and that choice must survive further keystrokes or kind switches.
// The state is a plain object so the page can wrap it in Vue's reactive()
// while plain Node tests exercise the same rules.
export type GroupCollapseState = {
  expanded: Set<string>
  collapsed: Set<string>
  rendered: Set<string>
}

export const createGroupCollapseState = (): GroupCollapseState => ({
  expanded: new Set<string>(),
  collapsed: new Set<string>(),
  rendered: new Set<string>(),
})

// An explicit collapse wins over the filter's open-by-default behavior.
export const isGroupOpen = (
  state: GroupCollapseState,
  key: string,
  filtersActive: boolean,
): boolean => !state.collapsed.has(key) && (filtersActive || state.expanded.has(key))

// Filters mount every matching group up front; otherwise a group stays out
// of the DOM until first opened. Once mounted it stays mounted so closing
// does not discard scroll state or thumbnail skeletons.
export const isGroupRendered = (
  state: GroupCollapseState,
  key: string,
  filtersActive: boolean,
): boolean => filtersActive || state.rendered.has(key)

export const toggleGroup = (
  state: GroupCollapseState,
  key: string,
  filtersActive: boolean,
): void => {
  state.rendered.add(key)
  if (isGroupOpen(state, key, filtersActive)) {
    state.collapsed.add(key)
    state.expanded.delete(key)
  } else {
    state.collapsed.delete(key)
    state.expanded.add(key)
  }
}

// Leaving the filtered view returns to the unfiltered default (all closed).
// Rendered groups are deliberately kept so reopening stays cheap.
export const resetGroupCollapse = (state: GroupCollapseState): void => {
  state.expanded.clear()
  state.collapsed.clear()
}
