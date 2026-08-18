<script setup lang="ts">
// Upload channel skeleton. The back-end POST /api/themes endpoint
// (server-side re-encoding, AGENTS.md Iron Rule 4) is a later milestone.
const placeholder = ref('上传通道即将上线,敬请期待~')
</script>

<template>
  <main class="max-w-3xl mx-auto px-4 py-8 font-sans">
    <h1 class="text-3xl font-bold text-loli-pink mb-6">⬆️ 上传主题</h1>
    <p class="text-gray-600">{{ placeholder }}</p>
  </main>
</template>
