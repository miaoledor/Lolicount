import { clsx, type ClassValue } from 'clsx'

// cn merges class names (AGENTS.md: front-end uses cn to combine classes).
export const cn = (...inputs: ClassValue[]): string => clsx(inputs)
