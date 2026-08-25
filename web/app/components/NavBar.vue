// NavBar: fixed top navigation bar. Left: icon + Lolicount title.
// Center: feature jump links. Right: GitHub star badge linking to repo.
<script setup lang="ts">
const { t } = useI18n()

const repoUrl = 'https://github.com/miaoledor/Lolicount'

const navLinks = [
  { href: '#howto', label: 'nav.howto' },
  { href: '#loli', label: 'nav.loli' },
  { href: '#themes', label: 'nav.themes' },
  { href: '#playground', label: 'nav.playground' },
  { href: '#about', label: 'nav.more' },
] as const
</script>

<template>
  <nav class="nav-bar">
    <a href="#top" class="nav-brand">
      <img src="/images/lolicount-icon.png" alt="Lolicount" class="nav-brand-icon" />
      <span class="nav-brand-text">Lolicount</span>
    </a>

    <div class="nav-links">
      <a
        v-for="link in navLinks"
        :key="link.href"
        :href="link.href"
        class="nav-link"
      >{{ t(link.label) }}</a>
    </div>

    <a :href="repoUrl" target="_blank" rel="noopener" class="nav-star">
      <span class="nav-star-icon">★</span>
      <span>Star</span>
    </a>
  </nav>
</template>

<style scoped>
.nav-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid #f0e3e8;
  height: 56px;
}
.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
}
.nav-brand-icon {
  width: 32px;
  height: 32px;
}
.nav-brand-text {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--loli-pink);
}
.nav-links {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
.nav-link {
  padding: 0.25rem 0.625rem;
  font-size: 0.85rem;
  color: #6b7280;
  text-decoration: none;
  border-radius: 0.375rem;
  transition: all 0.2s;
}
.nav-link:hover {
  color: var(--loli-pink);
  background: var(--loli-cream);
}
.nav-star {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.75rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #fff;
  background: var(--loli-pink);
  border-radius: 0.375rem;
  text-decoration: none;
  transition: opacity 0.2s;
}
.nav-star:hover {
  opacity: 0.9;
}
.nav-star-icon {
  font-size: 0.9rem;
}
</style>
