<script setup lang="ts">
// LoliCharacter renders a randomly assembled character portrait by
// overlaying five transparent layer images (body/eye/brow/mouth/face) at
// absolute coordinates, then scaling + translating the canvas so the face
// is centered in the stage. Clicking the portrait re-rolls a new random
// combination. The theme prop selects which character theme to load.
import { getLoli, type LoliParts } from '~/composables/useLoli'

const props = defineProps<{
  theme: string
}>()

const { t } = useI18n()

// Default canvas/range for lian (504x925); overridden per-theme from
// ren.config.json once the manifest loads. The frame box is the crop
// window centered on the assembled portrait's bounding box.
const FRAME_W = 367
const FRAME_H = 602
const CANVAS_W_DEFAULT = 504
const CANVAS_H_DEFAULT = 925
const SCALE = 0.7
const stageW = Math.round(FRAME_W * SCALE)
const stageH = Math.round(FRAME_H * SCALE)

const canvasW = ref(CANVAS_W_DEFAULT)
const canvasH = ref(CANVAS_H_DEFAULT)

const empty: LoliParts = {
  loliBodyLeft: '',
  loliBodyTop: '',
  loliEyeLeft: '',
  loliEyeTop: '',
  loliBrowLeft: '',
  loliBrowTop: '',
  loliMouthLeft: '',
  loliMouthTop: '',
  loliFaceLeft: '',
  loliFaceTop: '',
  body: '',
  eye: '',
  brow: '',
  mouth: '',
  face: '',
  bbox: { left: 137, top: 323, width: 367, height: 602 },
}

const loliData = ref<LoliParts>({ ...empty })
const ready = ref(false)
const rolling = ref(false)

const canvasStyle = computed(() => {
  const b = loliData.value.bbox
  const x0 = b.left + b.width / 2 - FRAME_W / 2
  const y0 = b.top + b.height / 2 - FRAME_H / 2
  return {
    width: `${canvasW.value}px`,
    height: `${canvasH.value}px`,
    transform: `scale(${SCALE}) translate(${-x0}px, ${-y0}px)`,
    transformOrigin: 'top left',
  }
})

const decode = (src: string) => {
  if (!src) return Promise.resolve()
  const img = new Image()
  img.src = src
  return img.decode().catch(() => {})
}

const revokeOld = () => {
  const old = loliData.value
  ;[old.body, old.eye, old.brow, old.mouth, old.face].forEach((u) => {
    if (u && u.startsWith('blob:')) URL.revokeObjectURL(u)
  })
}

const reroll = async () => {
  if (rolling.value) return
  rolling.value = true
  try {
    const data = await getLoli(props.theme)
    await Promise.all([data.body, data.eye, data.brow, data.mouth, data.face].map(decode))
    revokeOld()
    loliData.value = data
    ready.value = true
  } finally {
    rolling.value = false
  }
}

// Re-roll when the theme changes.
watch(() => props.theme, () => {
  ready.value = false
  revokeOld()
  loliData.value = { ...empty }
  reroll()
})

onMounted(reroll)
onUnmounted(revokeOld)
</script>

<template>
  <div
    class="relative cursor-pointer overflow-hidden rounded-lg"
    :style="{ width: `${stageW}px`, height: `${stageH}px` }"
    :title="rolling ? t('loli.rolling') : t('loli.reroll')"
    @click="reroll"
  >
    <div
      v-if="ready && loliData.body"
      class="absolute top-0 left-0 transition-opacity duration-300"
      :class="cn(rolling && 'opacity-40')"
      :style="canvasStyle"
    >
      <img
        class="absolute max-w-none"
        :src="loliData.body"
        alt="ren"
        :style="{ left: loliData.loliBodyLeft, top: loliData.loliBodyTop }"
      />
      <img
        class="absolute max-w-none"
        :src="loliData.eye"
        alt="ren"
        :style="{ left: loliData.loliEyeLeft, top: loliData.loliEyeTop }"
      />
      <img
        class="absolute max-w-none"
        :src="loliData.brow"
        alt="ren"
        :style="{ left: loliData.loliBrowLeft, top: loliData.loliBrowTop }"
      />
      <img
        class="absolute max-w-none"
        :src="loliData.mouth"
        alt="ren"
        :style="{ left: loliData.loliMouthLeft, top: loliData.loliMouthTop }"
      />
      <img
        class="absolute max-w-none"
        :src="loliData.face"
        alt="ren"
        :style="{ left: loliData.loliFaceLeft, top: loliData.loliFaceTop }"
      />
    </div>
    <div
      v-else
      class="absolute inset-0 flex items-center justify-center text-sm text-gray-400"
    >
      {{ t('loli.loading') }}
    </div>
  </div>
</template>
