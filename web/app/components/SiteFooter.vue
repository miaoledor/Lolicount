<script setup lang="ts">
// SiteFooter: bottom of every page. Holds the author/repo links,
// contributors credits, acknowledged projects, and donation channel.
// All copy is localized via useI18n; links are stable URLs.
const { t } = useI18n()

const authorUrl = 'https://github.com/miaoledor'
const repoUrl = 'https://github.com/miaoledor/Lolicount'

// Acknowledged upstream projects (M10).
const thanks = [
  { name: 'kun-galgame-forum', url: 'https://github.com/KunMoe/kun-galgame-forum' },
  { name: 'Moe-Counter', url: 'https://github.com/journey-ad/Moe-Counter' },
]

// Contributors list. Single source of truth: the GitHub contributors
// page. Kept as a link rather than a hardcoded list so it never goes
// stale.
const contributorsUrl = `${repoUrl}/graphs/contributors`

// Donation channel placeholder. Replace with a real sponsor link
// (e.g. 爱发电 / Ko-fi / GitHub Sponsors) when ready.
const donateUrl = 'https://github.com/sponsors/miaoledor'
</script>

<template>
  <footer class="mt-16 border-t border-gray-200 bg-loli-cream/50">
    <div class="max-w-4xl mx-auto px-4 py-10 grid gap-8 sm:grid-cols-3">
      <!-- Project + links -->
      <div>
        <h3 class="font-bold text-loli-pink mb-2">Lolicount</h3>
        <p class="text-sm text-gray-600 mb-3">{{ t('footer.tagline') }}</p>
        <p class="text-sm font-medium text-loli-pink mb-3">{{ t('footer.starHint') }}</p>
        <div class="flex flex-wrap gap-2">
          <a
            :href="repoUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-1.5 bg-loli-pink text-white text-sm font-semibold px-4 py-2 rounded-lg shadow-sm hover:bg-loli-pink/90 hover:shadow-md transition"
          >
            <span>★</span> {{ t('footer.repo') }}
          </a>
          <a
            :href="authorUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-1.5 border-2 border-loli-pink text-loli-pink text-sm font-semibold px-4 py-2 rounded-lg hover:bg-loli-pink hover:text-white transition"
          >
            {{ t('footer.author') }}: miaoledor
          </a>
        </div>
      </div>

      <!-- Contributors -->
      <div>
        <h3 class="font-semibold mb-2">{{ t('footer.contributors') }}</h3>
        <p class="text-sm text-gray-600 mb-2">{{ t('footer.credits') }}</p>
        <a :href="contributorsUrl" target="_blank" rel="noopener" class="text-sm text-loli-pink hover:underline">
          GitHub Contributors →
        </a>
      </div>

      <!-- Thanks + donate -->
      <div>
        <h3 class="font-semibold mb-2">{{ t('footer.thanks') }}</h3>
        <div class="flex flex-col gap-1 text-sm mb-4">
          <a
            v-for="item in thanks"
            :key="item.name"
            :href="item.url"
            target="_blank"
            rel="noopener"
            class="text-gray-600 hover:text-loli-pink transition"
          >
            {{ item.name }}
          </a>
        </div>
        <h3 class="font-semibold mb-1">{{ t('footer.donate') }}</h3>
        <p class="text-xs text-gray-500 mb-2">{{ t('footer.donateHint') }}</p>
        <a :href="donateUrl" target="_blank" rel="noopener" class="inline-block bg-loli-pink text-white text-sm px-3 py-1 rounded hover:bg-loli-pink/90 transition">
          ❤ {{ t('footer.donate') }}
        </a>
      </div>
    </div>

    <div class="text-center text-xs text-gray-400 pb-6">
      {{ t('footer.poweredBy') }}
    </div>
  </footer>
</template>
