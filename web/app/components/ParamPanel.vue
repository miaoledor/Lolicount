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

const props = defineProps<{
  state: ParamState
  themes: ThemeInfo[]
  fthemes: string[]
}>()
const emit = defineEmits<{ update: [Partial<ParamState>] }>()
const { t } = useI18n()

type PanelMode = 'quick' | 'fine' | 'expert'

const panelMode = ref<PanelMode>('quick')

const nameEmpty = computed(() => !props.state.name.trim())

const update = (patch: Partial<ParamState>) => emit('update', patch)

// Sanitize numeric inputs like Moe-Counter does (digits / sign / dot only).
const sanitizeInt = (v: string) => v.replace(/[^0-9-]/g, '')
const sanitizeFloat = (v: string) => v.replace(/[^0-9.]/g, '')
</script>

<template>
  <div class="loli-tool">
    <div class="loli-name-bar" :class="{ 'is-empty': nameEmpty }">
      <span class="loli-name-prefix">@</span>
      <input
        :value="state.name"
        :placeholder="t('param.namePlaceholder')"
        @input="update({ name: ($event.target as HTMLInputElement).value })"
        class="loli-name-input"
      />
    </div>
    <div class="loli-mode-bar">
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'quick' && 'is-active')"
        @click="panelMode = 'quick'"
      >{{ t('playground.modeQuick') }}</button>
      <span class="loli-mode-sep">&gt;&gt;</span>
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'fine' && 'is-active')"
        @click="panelMode = 'fine'"
      >{{ t('playground.modeFine') }}</button>
      <span class="loli-mode-sep">&gt;&gt;</span>
      <button
        type="button"
        :class="cn('loli-mode-btn', panelMode === 'expert' && 'is-active')"
        @click="panelMode = 'expert'"
      >{{ t('playground.modeExpert') }}</button>
    </div>
    <table class="loli-tool-table">
      <thead>
        <tr>
          <th>{{ t('tool.param') }}</th>
          <th>{{ t('tool.description') }}</th>
          <th>{{ t('tool.value') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>theme</code></td>
          <td>{{ t('param.theme') }}</td>
          <td>
            <select :value="state.theme" @change="update({ theme: ($event.target as HTMLSelectElement).value })" class="loli-input">
              <option v-for="tth in themes" :key="tth.name" :value="tth.name">
                {{ tth.animated ? `${tth.name} · ${t('param.animated')}` : tth.name }}
              </option>
            </select>
          </td>
        </tr>
        <tr v-if="panelMode !== 'quick'">
          <td><code>ftheme</code></td>
          <td>{{ t('param.fontStyle') }}</td>
          <td>
            <select :value="state.ftheme" @change="update({ ftheme: ($event.target as HTMLSelectElement).value })" class="loli-input">
              <option value="">{{ t('param.fontDefault') }}</option>
              <option v-for="f in fthemes" :key="f" :value="f">{{ f }}</option>
            </select>
          </td>
        </tr>
        <tr v-if="panelMode !== 'quick'">
          <td><code>text</code></td>
          <td>{{ t('param.text') }}</td>
          <td>
            <input
              type="text"
              :value="state.text"
              :placeholder="t('param.textPlaceholder')"
              class="loli-input"
              @input="update({ text: ($event.target as HTMLInputElement).value })"
            >
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>fsize</code></td>
          <td>{{ t('param.fsize') }}</td>
          <td>
            <input
              type="number"
              :value="state.fsize"
              min="0"
              max="500"
              step="1"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ fsize: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>scale</code></td>
          <td>{{ t('param.scale') }}</td>
          <td>
            <input
              type="number"
              :value="state.scale"
              min="0.1"
              max="4"
              step="0.1"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ scale: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>x</code></td>
          <td>{{ t('param.px') }}</td>
          <td>
            <input
              type="number"
              :value="state.x ?? 0"
              min="-500"
              max="2000"
              step="1"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ x: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>y</code></td>
          <td>{{ t('param.py') }}</td>
          <td>
            <input
              type="number"
              :value="state.y ?? 0"
              min="-500"
              max="2000"
              step="1"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ y: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>rx</code></td>
          <td>{{ t('param.rx') }}</td>
          <td>
            <input
              type="number"
              :value="state.rx ?? 0"
              min="0"
              max="1"
              step="0.01"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ rx: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode === 'expert'">
          <td><code>ry</code></td>
          <td>{{ t('param.ry') }}</td>
          <td>
            <input
              type="number"
              :value="state.ry ?? 0"
              min="0"
              max="1"
              step="0.01"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeFloat((e.target as HTMLInputElement).value); update({ ry: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
          </td>
        </tr>
        <tr v-if="panelMode !== 'quick'">
          <td><code>unshowf</code></td>
          <td>{{ t('param.unshowf') }}</td>
          <td>
            <input
              :id="'loli-unshowf'"
              type="checkbox"
              role="switch"
              :checked="state.unshowf"
              @change="update({ unshowf: ($event.target as HTMLInputElement).checked })"
              class="loli-switch-input"
            />
            <label :for="'loli-unshowf'" class="loli-switch-label"><span>ON</span><span>OFF</span></label>
          </td>
        </tr>

        <tr v-if="panelMode === 'expert'">
          <td colspan="3" class="loli-unusual-caption">{{ t('tool.unusual') }}</td>
        </tr>
        <!-- Unusual Options: number lets the user preview a fixed value
             without +1 (back-end early-returns when number > 0). -->
        <tr v-if="panelMode === 'expert'">
          <td><code>number</code></td>
          <td>{{ t('param.number') }}</td>
          <td>
            <input
              type="number"
              :value="state.number ?? 0"
              min="0"
              max="999999"
              step="1"
              @input="(e) => { (e.target as HTMLInputElement).value = sanitizeInt((e.target as HTMLInputElement).value); update({ number: Number((e.target as HTMLInputElement).value) }) }"
              class="loli-input"
            />
            <p class="loli-cell-hint">{{ t('param.numberHint') }}</p>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.loli-tool {
}
.loli-name-bar {
  display: flex;
  align-items: center;
  border: 2px solid var(--loli-pink);
  border-radius: 0.5rem;
  background: #fff;
  overflow: hidden;
  transition: border-color 0.2s;
}
.loli-name-bar.is-empty {
  border-color: #f0a0b0;
  animation: loli-shake 0.4s ease;
}
.loli-name-prefix {
  display: flex;
  align-items: center;
  padding: 0 0.625rem;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--loli-pink);
  background: var(--loli-cream);
  height: 100%;
  user-select: none;
}
.loli-name-input {
  flex: 1;
  border: none;
  outline: none;
  padding: 0.625rem 0.5rem;
  font-size: 1rem;
  background: transparent;
  box-sizing: border-box;
}
.loli-name-input::placeholder {
  color: #c4b0bb;
}
@keyframes loli-shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}
.loli-mode-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.625rem;
  border-bottom: 1px solid #e5d4dc;
  background: var(--loli-cream);
}
.loli-mode-btn {
  padding: 0.25rem 0.75rem;
  border-radius: 0.375rem;
  border: 1px solid transparent;
  font-size: 0.8rem;
  color: #6b7280;
  background: transparent;
  transition: all 0.2s;
  cursor: pointer;
}
.loli-mode-btn.is-active {
  border-color: var(--loli-pink);
  background: #fff;
  color: var(--loli-pink);
  font-weight: 600;
}
.loli-mode-sep {
  font-size: 0.7rem;
  color: #c4b0bb;
  user-select: none;
}
.loli-tool-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-size: 0.875rem;
}
.loli-tool-table th,
.loli-tool-table td {
  text-align: left;
  padding: 0.5rem 0.625rem;
  border-bottom: 1px solid #f0e3e8;
  vertical-align: middle;
  word-break: break-word;
}
.loli-tool-table th {
  font-weight: 600;
  color: #6b7280;
  border-bottom: 1px solid #e5d4dc;
  background: var(--loli-cream);
}
.loli-tool-table td code {
  font-weight: 600;
  color: var(--loli-pink);
  background: transparent;
}
.loli-input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #e5d4dc;
  border-radius: 0.375rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
  background: #fff;
}
.loli-input:focus {
  outline: none;
  border-color: var(--loli-pink);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--loli-pink) 25%, transparent);
}
.loli-cell-hint {
  margin: 0.25rem 0 0;
  font-size: 0.7rem;
  color: #9ca3af;
}
.loli-cell-hint-row {
  font-size: 0.75rem;
  color: #9ca3af;
}
.loli-unusual-caption {
  font-weight: 600;
  font-size: 0.85rem;
  color: #6b7280;
  padding-top: 0.75rem;
  border-bottom: 2px solid #e5d4dc;
}
/* Switch toggle, mirroring Moe-Counter's role=switch style. */
.loli-switch-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.loli-switch-label {
  position: relative;
  display: inline-block;
  width: 3.6em;
  height: 1.8em;
  border-radius: 1.8em;
  background: #9ca3af;
  cursor: pointer;
  transition: background 0.3s;
  user-select: none;
}
.loli-switch-label::after {
  content: "";
  position: absolute;
  top: 0.1em;
  left: 0.1em;
  width: 1.6em;
  height: 1.6em;
  background: #fff;
  border-radius: 50%;
  transition: transform 0.3s;
}
.loli-switch-label > span {
  position: absolute;
  inset: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 12.5%;
  font-size: 10px;
  font-weight: bold;
  color: #fff;
}
.loli-switch-input:checked + .loli-switch-label {
  background: var(--loli-pink);
}
.loli-switch-input:checked + .loli-switch-label::after {
  transform: translateX(1.8em);
}
</style>
