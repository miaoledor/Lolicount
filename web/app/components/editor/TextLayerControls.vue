<script setup lang="ts">
defineProps<{
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
    <div class="control-row">
      <label class="control-label">{{ t('editor.counterText') }}</label>
      <input
        :value="text"
        type="text"
        class="control-input"
        @input="emit('update:text', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="control-row-pair">
      <div class="control-half">
        <label class="control-label">{{ t('editor.fontSize') }}</label>
        <input
          :value="fontSize"
          type="number"
          class="control-input"
          @input="emit('update:fontSize', Number(($event.target as HTMLInputElement).value))"
        >
      </div>
      <div class="control-half">
        <label class="control-label">{{ t('editor.scale') }}</label>
        <input
          :value="scale"
          type="number"
          step="0.1"
          class="control-input"
          @input="emit('update:scale', Number(($event.target as HTMLInputElement).value))"
        >
      </div>
    </div>
    <label class="control-checkbox">
      <input
        :checked="unshowFont"
        type="checkbox"
        @change="emit('update:unshowFont', ($event.target as HTMLInputElement).checked)"
      >
      <span>{{ t('editor.hideText') }}</span>
    </label>
  </div>
</template>

<style scoped>
.text-controls {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.control-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.control-row-pair {
  display: flex;
  gap: 0.5rem;
}

.control-half {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.control-label {
  font-size: 0.6875rem;
  color: var(--text-muted, #999);
}

.control-input {
  width: 100%;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color, #444);
  border-radius: 4px;
  background: #fff;
  color: #111;
  font-size: 0.8125rem;
  box-sizing: border-box;
}

.control-checkbox {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--text-color, #eee);
  cursor: pointer;
  user-select: none;
}

.control-checkbox input {
  cursor: pointer;
}
</style>
