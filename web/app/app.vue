<template>
  <div>
    <nav class="bg-loli-cream border-b px-4 py-2 flex gap-4">
      <NuxtLink to="/" class="text-loli-pink font-bold">Lolicount</NuxtLink>
      <NuxtLink to="/themes" class="text-gray-700 hover:text-loli-pink">主题</NuxtLink>
      <NuxtLink to="/playground" class="text-gray-700 hover:text-loli-pink">Playground</NuxtLink>
      <NuxtLink to="/upload" class="text-gray-700 hover:text-loli-pink">上传</NuxtLink>
    </nav>
    <NuxtPage />
  </div>
</template>
