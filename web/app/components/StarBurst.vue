<script setup lang="ts">
// StarBurst renders a burst of star particles radiating from a point.
// Call trigger(x, y) from a click handler; GSAP animates each particle's
// translation + fade. The container is pointer-events-none so it never
// blocks clicks.
import gsap from 'gsap'

const STAR = '★'

const container = ref<HTMLElement | null>(null)
const particles = ref<HTMLElement[]>([])

const COLORS = ['#ff7eb3', '#ff758c', '#ffb86c', '#fff6a3', '#a3e6ff', '#c4b5fd']

const trigger = (x: number, y: number) => {
  const root = container.value
  if (!root) return
  const rect = root.getBoundingClientRect()
  const originX = x - rect.left
  const originY = y - rect.top

  const count = 18
  const fragment = document.createDocumentFragment()
  const created: HTMLElement[] = []

  for (let i = 0; i < count; i++) {
    const el = document.createElement('span')
    el.textContent = STAR
    el.style.position = 'absolute'
    el.style.left = `${originX}px`
    el.style.top = `${originY}px`
    el.style.fontSize = `${10 + Math.random() * 14}px`
    el.style.color = COLORS[i % COLORS.length] ?? '#ff758c'
    el.style.willChange = 'transform, opacity'
    el.style.userSelect = 'none'
    fragment.appendChild(el)
    created.push(el)
  }
  root.appendChild(fragment)

  created.forEach((el, i) => {
    const angle = (Math.PI * 2 * i) / count + Math.random() * 0.3
    const dist = 60 + Math.random() * 80
    const dx = Math.cos(angle) * dist
    const dy = Math.sin(angle) * dist
    const rot = (Math.random() - 0.5) * 360
    gsap.fromTo(
      el,
      { x: 0, y: 0, scale: 0, opacity: 1, rotation: 0 },
      {
        x: dx,
        y: dy,
        scale: 1 + Math.random(),
        rotation: rot,
        opacity: 0,
        duration: 0.7 + Math.random() * 0.4,
        ease: 'power2.out',
        onComplete: () => el.remove(),
      },
    )
  })
}

defineExpose({ trigger })
</script>

<template>
  <div ref="container" class="pointer-events-none absolute inset-0 overflow-hidden" />
</template>
