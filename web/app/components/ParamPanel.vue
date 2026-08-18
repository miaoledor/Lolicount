<script setup lang="ts">
export type ParamState = {
  name: string
  theme: string
  kind: 'frame' | 'character'
  ftheme: string
  fsize: number
  scale: number
  unshowf: boolean
  x: number | undefined
  y: number | undefined
  rx: number | undefined
  ry: number | undefined
  mode: 'seq' | 'random'
}

const props = defineProps<{
  state: ParamState
  themes: ThemeInfo[]
  fthemes: string[]
}>()
const emit = defineEmits<{ update: [Partial<ParamState>] }>()

const update = (patch: Partial<ParamState>) => emit('update', patch)

// M9.5: themes are grouped by type. The user first picks a theme type
// (frame/character), then picks a theme within that type. The type is
// stored explicitly in state (not derived from the theme name) so the
// picker works even before the theme list has loaded.
const frameThemes = computed(() => props.themes.filter((t) => t.kind === 'frame'))
const characterThemes = computed(() => props.themes.filter((t) => t.kind === 'character'))
const availableThemes = computed(() => (props.state.kind === 'character' ? characterThemes.value : frameThemes.value))

const onSelectKind = (kind: 'frame' | 'character') => {
  const patch: Partial<ParamState> = { kind }
  // Auto-switch to the first theme of the new type when the current
  // theme does not belong to it. This keeps the dropdown in sync. If the
  // list is not loaded yet the theme is left as-is; once it loads the
  // dropdown will show the right set and the user can pick one.
  const list = kind === 'character' ? characterThemes.value : frameThemes.value
  const current = props.themes.find((t) => t.name === props.state.theme)
  if (list.length > 0 && (!current || current.kind !== kind)) {
    patch.theme = list[0]!.name
  }
  // Character themes are always random; coerce mode for consistency.
  if (kind === 'character') patch.mode = 'random'
  update(patch)
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <label class="block text-sm font-medium mb-1">计数器名称</label>
      <input :value="state.name" @input="update({ name: ($event.target as HTMLInputElement).value })" class="w-full border rounded px-2 py-1" />
    </div>
    <!-- M9.5: theme type selector comes first, then the theme dropdown
         is scoped to the chosen type. -->
    <div>
      <label class="block text-sm font-medium mb-1">主题类型</label>
      <div class="grid grid-cols-2 gap-2">
        <button
          type="button"
          :class="cn(
            'py-2 rounded border text-sm transition',
            state.kind === 'frame' ? 'border-loli-pink bg-loli-cream text-loli-pink' : 'border-gray-200 text-gray-600',
          )"
          @click="onSelectKind('frame')"
        >
          卡片主题
        </button>
        <button
          type="button"
          :class="cn(
            'py-2 rounded border text-sm transition',
            state.kind === 'character' ? 'border-loli-pink bg-loli-cream text-loli-pink' : 'border-gray-200 text-gray-600',
          )"
          @click="onSelectKind('character')"
        >
          立绘主题
        </button>
      </div>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="block text-sm font-medium mb-1">主题</label>
        <select :value="state.theme" @change="update({ theme: ($event.target as HTMLSelectElement).value })" class="w-full border rounded px-2 py-1">
          <option v-for="t in availableThemes" :key="t.name" :value="t.name">{{ t.name }}</option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">字体样式</label>
        <select :value="state.ftheme" @change="update({ ftheme: ($event.target as HTMLSelectElement).value })" class="w-full border rounded px-2 py-1">
          <option value="">默认</option>
          <option v-for="f in fthemes" :key="f" :value="f">{{ f }}</option>
        </select>
      </div>
    </div>
    <!-- M9: mode only applies to frame themes; character themes are
         always random, so the control is hidden for them. -->
    <div v-if="state.kind === 'frame'" class="grid grid-cols-2 gap-3">
      <div>
        <label class="block text-sm font-medium mb-1">帧模式 mode</label>
        <select :value="state.mode" @change="update({ mode: ($event.target as HTMLSelectElement).value as 'seq' | 'random' })" class="w-full border rounded px-2 py-1">
          <option value="seq">顺序 seq</option>
          <option value="random">随机 random</option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">说明</label>
        <p class="text-xs text-gray-500 mt-1">顺序模式随计数循环帧,随机模式每次请求随机抽帧。</p>
      </div>
    </div>
    <div v-else>
      <p class="text-xs text-gray-500">立绘主题固定随机模式,每次请求重新组合服装与表情。</p>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="block text-sm font-medium mb-1">字号 fsize</label>
        <input type="number" :value="state.fsize" @input="update({ fsize: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">图片缩放 scale</label>
        <input type="number" step="0.1" :value="state.scale" @input="update({ scale: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="block text-sm font-medium mb-1">像素 x</label>
        <input type="number" :value="state.x ?? 0" @input="update({ x: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">像素 y</label>
        <input type="number" :value="state.y ?? 0" @input="update({ y: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="block text-sm font-medium mb-1">比例 rx</label>
        <input type="number" step="0.1" :value="state.rx ?? 0" @input="update({ rx: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">比例 ry</label>
        <input type="number" step="0.1" :value="state.ry ?? 0" @input="update({ ry: Number(($event.target as HTMLInputElement).value) })" class="w-full border rounded px-2 py-1" />
      </div>
    </div>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" :checked="state.unshowf" @change="update({ unshowf: ($event.target as HTMLInputElement).checked })" />
      隐藏字体 (unshowf)
    </label>
  </div>
</template>
