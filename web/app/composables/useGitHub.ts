// useGitHub: fetches the repo star count from the public GitHub API.
// Cached for 10 minutes in localStorage so repeated page loads don't
// hammer the rate-limited unauthenticated endpoint.

const REPO = 'miaoledor/Lolicount'
const CACHE_KEY = 'lolicount-stars'
const CACHE_TTL = 10 * 60 * 1000

type StarCache = { count: number; ts: number }

const readCache = (): StarCache | null => {
  if (!import.meta.client) return null
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    return raw ? (JSON.parse(raw) as StarCache) : null
  } catch {
    return null
  }
}

const writeCache = (count: number) => {
  if (!import.meta.client) return
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({ count, ts: Date.now() }))
  } catch {
    // ignore quota errors
  }
}

const stars = ref<number | null>(null)

export const useGitHub = () => {
  const repoUrl = `https://github.com/${REPO}`

  const fetchStars = async () => {
    const cached = readCache()
    if (cached && Date.now() - cached.ts < CACHE_TTL) {
      stars.value = cached.count
      return
    }
    try {
      const data = await $fetch<{ stargazers_count?: number }>(
        `https://api.github.com/repos/${REPO}`,
        { headers: { Accept: 'application/vnd.github+json' } },
      )
      const count = data.stargazers_count ?? 0
      stars.value = count
      writeCache(count)
    } catch {
      // Rate-limited or offline: fall back to cached value if any.
      if (cached) stars.value = cached.count
    }
  }

  const formatStars = (n: number | null): string => {
    if (n == null) return ''
    if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k'
    return String(n)
  }

  return { stars, repoUrl, fetchStars, formatStars }
}
