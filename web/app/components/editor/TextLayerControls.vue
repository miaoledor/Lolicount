<script setup lang="ts">
const props = defineProps<{
  text: string
  fontSize: number
  scale: number
  unshowFont: boolean
}>()

const emit = defineEmits<{
  'update:text': [v: string]
  'update:fontSize': [v: number]
  'update:scale': [v: number]
  'update:unshowFont': [v: boolean]
}>()

const { t } = useI18n()
</script>

<template>
  <div class="text-controls">
    <h3 class="text-controls-title">{{ t('editor.textLayer') }}</h3>
    <div class="text-field">
      <label>{{ t('editor.counterText') }}</label>
      <input
        :value="text"
        type="text"
        class="text-input"
        @input="emit('update:text', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="text-field-row">
      <div class="text-field">
        <label>{{ t('editor.fontSize') }}</label>
        <input
          :value="fontSize"
          type="number"
          class="text-input"
          @input="emit('update:fontSize', Number(($event.target as HTMLInputElement).value))"
        >
      </div>
      <div class="text-field">
        <label>{{ t('editor.scale') }}</label>
        <input
          :value="scale"
          type="number"
          step="0.1"
          class="text-input"
          @input="emit('update:scale', Number(($event.target as HTMLInputElement).value))"
        >
      </div>
    </div>
    <label class="text-checkbox">
      <input
        :checked="unshowFont"
        type="checkbox"
        @change="emit('update:unshowFont', ($event.target as HTMLInputElement).checked)"
      >
      {{ t('editor.hideText') }}
    </label>
  </div>
</template>

<style scoped>
.text-controls {
  border: 1px solid var(--border-color, #333);
  border-radius: 6px;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.text-controls-title {
  font-size: 0.875rem;
  margin: 0;
}

.text-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.text-field-row {
  display: flex;
  gap: 0.5rem;
}

.text-field-row .text-field {
  flex: 1;
}

.text-input {
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.875rem;
}

.text-checkbox {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  cursor: pointer;
}
</style>
