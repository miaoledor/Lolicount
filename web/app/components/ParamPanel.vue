<script setup lang="ts">
export type ParamState = {
  name: string
  theme: string
  ftheme: string
  fsize: number
  scale: number
  unshowf: boolean
  x: number | undefined
  y: number | undefined
  rx: number | undefined
  ry: number | undefined
  number: number
  text: string
}

defineProps<{
  state: ParamState
  themes: ThemeInfo[]
  fthemes: string[]
}>()
const emit = defineEmits<{ update: [Partial<ParamState>] }>()
const { t } = useI18n()

type PanelMode = 'quick' | 'fine' | 'expert'

const panelMode = ref<PanelMode>('quick')

const update = (patch: Partial<ParamState>) => emit('update', patch)

// Sanitize numeric inputs like Moe-Counter does (digits / sign / dot only).
const sanitizeInt = (v: string) => v.replace(/[^0-9-]/g, '')
const sanitizeFloat = (v: string) => v.replace(/[^0-9.]/g, '')
</script>

<template>
  <div class="loli-tool">
    <div class="field">
      <label class="field-label" for="loli-name">{{ t('param.nameLabel') }}</label>
      <div class="loli-name-bar">
        <span class="loli-name-prefix" aria-hidden="true">@</span>
        <input
          id="loli-name"
          :value="state.name"
          :placeholder="t('param.namePlaceholder')"
          class="loli-name-input"
          autocomplete="off"
          spellcheck="false"
          @input="update({ name: ($event.target as HTMLInputElement).value })"
        />
      </div>
    </div>

    <div class="loli-mode-bar">
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'quick' && 'is-active')"
        :aria-pressed="panelMode === 'quick'"
        @click="panelMode = 'quick'"
      >{{ t('playground.modeQuick') }}</button>
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'fine' && 'is-active')"
        :aria-pressed="panelMode === 'fine'"
        @click="panelMode = 'fine'"
      >{{ t('playground.modeFine') }}</button>
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'expert' && 'is-active')"
        :aria-pressed="panelMode === 'expert'"
        @click="panelMode = 'expert'"
      >{{ t('playground.modeExpert') }}</button>
    </div>

    <div class="param-list">
      <div class="param-row">
        <div class="param-meta">
          <span class="param-label">{{ t('param.theme') }}</span>
          <code class="param-code">theme</code>
        </div>
        <div class="param-control">
          <select
            :value="state.theme"
            class="loli-input"
            @change="update({ theme: ($event.target as HTMLSelectElement).value })"
          >
            <option v-for="tth in themes" :key="tth.name" :value="tth.name">
              {{ tth.animated ? `${tth.name} · ${t('param.animated')}` : tth.name }}
            </option>
          </select>
        </div>
      </div>

      <template v-if="panelMode !== 'quick'">
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.fontStyle') }}</span>
            <code class="param-code">ftheme</code>
          </div>
          <div class="param-control">
            <select
              :value="state.ftheme"
              class="loli-input"
              @change="update({ ftheme: ($event.target as HTMLSelectElement).value })"
            >
              <option value="">{{ t('param.fontDefault') }}</option>
              <option v-for="f in fthemes" :key="f" :value="f">{{ f }}</option>
            </select>
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.text') }}</span>
            <code class="param-code">text</code>
          </div>
          <div class="param-control">
            <input
              type="text"
              :value="state.text"
              :placeholder="t('param.textPlaceholder')"
              class="loli-input"
              @input="update({ text: ($event.target as HTMLInputElement).value })"
            >
          </div>
        </div>
      </template>

      <template v-if="panelMode === 'expert'">
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.fsize') }}</span>
            <code class="param-code">fsize</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.fsize"
              min="0"
              max="500"
              step="1"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ fsize: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.scale') }}</span>
            <code class="param-code">scale</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.scale"
              min="0.1"
              max="4"
              step="0.1"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ scale: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.px') }}</span>
            <code class="param-code">x</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.x ?? 0"
              min="-500"
              max="2000"
              step="1"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ x: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.py') }}</span>
            <code class="param-code">y</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.y ?? 0"
              min="-500"
              max="2000"
              step="1"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ y: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.rx') }}</span>
            <code class="param-code">rx</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.rx ?? 0"
              min="0"
              max="1"
              step="0.01"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ rx: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.ry') }}</span>
            <code class="param-code">ry</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.ry ?? 0"
              min="0"
              max="1"
              step="0.01"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ ry: Number((e.target as HTMLInputElement).value) }) }"
            />
          </div>
        </div>

        <p class="param-group-caption">{{ t('tool.unusual') }}</p>
        <!-- number lets the user preview a fixed value without +1 (the
             back-end early-returns when number > 0). -->
        <div class="param-row">
          <div class="param-meta">
            <span class="param-label">{{ t('param.number') }}</span>
            <code class="param-code">number</code>
          </div>
          <div class="param-control">
            <input
              type="number"
              :value="state.number ?? 0"
              min="0"
              max="999999"
              step="1"
              class="loli-input"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ number: Number((e.target as HTMLInputElement).value) }) }"
            />
            <p class="param-hint">{{ t('param.numberHint') }}</p>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.loli-tool {
  display: grid;
  gap: 1rem;
}

.field {
  display: grid;
  gap: 0.4rem;
}

.field-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
}

.loli-name-bar {
  display: flex;
  align-items: center;
  gap: 0.15rem;
  padding: 0 0.75rem;
  border: 1px solid var(--loli-border);
  border-radius: 0.75rem;
  background: var(--bg-card);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.loli-name-bar:focus-within {
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.loli-name-prefix {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--loli-pink);
  user-select: none;
}

.loli-name-input {
  flex: 1;
  min-width: 0;
  padding: 0.7rem 0.25rem;
  border: 0;
  outline: none;
  font-size: 1rem;
  color: var(--text-color);
  background: transparent;
}

.loli-name-input::placeholder {
  color: var(--loli-hint);
}

.loli-mode-bar {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.25rem;
  padding: 0.25rem;
  border: 1px solid color-mix(in srgb, var(--loli-border) 70%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--loli-cream) 45%, var(--bg-card));
}

.loli-mode-btn {
  padding: 0.45rem 0.5rem;
  border: 0;
  border-radius: 0.55rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
  transition: color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.loli-mode-btn:hover:not(.is-active) {
  color: var(--text-color);
}

.loli-mode-btn.is-active {
  color: var(--loli-pink);
  background: var(--bg-card);
  font-weight: 650;
  box-shadow: 0 1px 5px rgba(76, 79, 105, 0.1);
}

.param-list {
  display: grid;
}

.param-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
  gap: 0.85rem;
  align-items: center;
  padding: 0.65rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--loli-border) 50%, transparent);
}

.param-row:last-child {
  border-bottom: 0;
}

.param-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.4rem;
  min-width: 0;
}

.param-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-color);
  line-height: 1.4;
}

.param-code {
  padding: 0.05rem 0.35rem;
  border-radius: 0.35rem;
  font-size: 0.7rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: var(--text-muted);
  background: color-mix(in srgb, var(--loli-cream) 60%, transparent);
}

.param-control {
  min-width: 0;
}

.loli-input {
  width: 100%;
  box-sizing: border-box;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--loli-border);
  border-radius: 0.55rem;
  font-size: 0.85rem;
  color: var(--text-color);
  background: var(--bg-card);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.loli-input:hover {
  border-color: color-mix(in srgb, var(--loli-pink) 45%, var(--loli-border));
}

.loli-input:focus {
  outline: none;
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.param-hint {
  margin: 0.35rem 0 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--text-muted);
}

.param-group-caption {
  margin: 0.85rem 0 0.1rem;
  font-size: 0.72rem;
  font-weight: 650;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

/* Mobile: stack labels above controls, and keep 16px control text so iOS
   Safari does not auto-zoom the viewport on focus. */
@media (max-width: 767px) {
  .param-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.4rem;
  }

  .loli-mode-btn {
    min-height: 36px;
  }

  .loli-input {
    min-height: 38px;
    padding: 0.4rem 0.6rem;
    font-size: 16px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .loli-mode-btn,
  .loli-input,
  .loli-name-bar {
    transition: none;
  }
}
</style>
