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
  number: number
}

const props = defineProps<{
  state: ParamState
  themes: ThemeInfo[]
  fthemes: string[]
}>()
const emit = defineEmits<{ update: [Partial<ParamState>] }>()
const { t } = useI18n()

const update = (patch: Partial<ParamState>) => emit('update', patch)

// M9.5: themes are grouped by type. The user first picks a theme type
// (frame/character), then picks a theme within that type. The type is
// stored explicitly in state (not derived from the theme name) so the
// picker works even before the theme list has loaded.
const frameThemes = computed(() => props.themes.filter((tth) => tth.kind === 'frame'))
const characterThemes = computed(() => props.themes.filter((tth) => tth.kind === 'character'))
const availableThemes = computed(() => (props.state.kind === 'character' ? characterThemes.value : frameThemes.value))

const onSelectKind = (kind: 'frame' | 'character') => {
  const patch: Partial<ParamState> = { kind }
  // Auto-switch to the first theme of the new type when the current
  // theme does not belong to it. This keeps the dropdown in sync. If the
  // list is not loaded yet the theme is left as-is; once it loads the
  // dropdown will show the right set and the user can pick one.
  const list = kind === 'character' ? characterThemes.value : frameThemes.value
  const current = props.themes.find((tth) => tth.name === props.state.theme)
  if (list.length > 0 && (!current || current.kind !== kind)) {
    patch.theme = list[0]!.name
  }
  // Character themes are always random; coerce mode for consistency.
  if (kind === 'character') patch.mode = 'random'
  update(patch)
}

// Sanitize numeric inputs like Moe-Counter does (digits / sign / dot only).
const sanitizeInt = (v: string) => v.replace(/[^0-9-]/g, '')
const sanitizeFloat = (v: string) => v.replace(/[^0-9.]/g, '')
</script>

<template>
  <div class="loli-tool">
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
          <td><code>name</code></td>
          <td>{{ t('param.name') }}</td>
          <td>
            <input
              :value="state.name"
              :placeholder="t('param.namePlaceholder')"
              @input="update({ name: ($event.target as HTMLInputElement).value })"
              class="loli-input"
            />
          </td>
        </tr>
        <tr>
          <td><code>kind</code></td>
          <td>{{ t('param.kind') }}</td>
          <td>
            <div class="loli-kind-grid">
              <button
                type="button"
                :class="cn('loli-kind-btn', state.kind === 'frame' && 'is-active')"
                @click="onSelectKind('frame')"
              >{{ t('param.kindFrame') }}</button>
              <button
                type="button"
                :class="cn('loli-kind-btn', state.kind === 'character' && 'is-active')"
                @click="onSelectKind('character')"
              >{{ t('param.kindCharacter') }}</button>
            </div>
          </td>
        </tr>
        <tr>
          <td><code>theme</code></td>
          <td>{{ t('param.theme') }}</td>
          <td>
            <select :value="state.theme" @change="update({ theme: ($event.target as HTMLSelectElement).value })" class="loli-input">
              <option v-for="tth in availableThemes" :key="tth.name" :value="tth.name">{{ tth.name }}</option>
            </select>
          </td>
        </tr>
        <tr>
          <td><code>ftheme</code></td>
          <td>{{ t('param.fontStyle') }}</td>
          <td>
            <select :value="state.ftheme" @change="update({ ftheme: ($event.target as HTMLSelectElement).value })" class="loli-input">
              <option value="">{{ t('param.fontDefault') }}</option>
              <option v-for="f in fthemes" :key="f" :value="f">{{ f }}</option>
            </select>
          </td>
        </tr>
        <!-- M9: mode only applies to frame themes; character themes are
             always random, so the control is hidden for them. -->
        <tr v-if="state.kind === 'frame'">
          <td><code>mode</code></td>
          <td>{{ t('param.mode') }}</td>
          <td>
            <select :value="state.mode" @change="update({ mode: ($event.target as HTMLSelectElement).value as 'seq' | 'random' })" class="loli-input">
              <option value="seq">{{ t('param.modeSeq') }}</option>
              <option value="random">{{ t('param.modeRandom') }}</option>
            </select>
            <p class="loli-cell-hint">{{ t('param.modeHint') }}</p>
          </td>
        </tr>
        <tr v-else>
          <td colspan="3" class="loli-cell-hint-row">{{ t('param.characterHint') }}</td>
        </tr>
        <tr>
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
        <tr>
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
        <tr>
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
        <tr>
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
        <tr>
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
        <tr>
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
        <tr>
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

        <tr>
          <td colspan="3" class="loli-unusual-caption">{{ t('tool.unusual') }}</td>
        </tr>
        <!-- Unusual Options: number lets the user preview a fixed value
             without +1 (back-end early-returns when number > 0). -->
        <tr>
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
  overflow-x: auto;
}
.loli-tool-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}
.loli-tool-table th,
.loli-tool-table td {
  text-align: left;
  padding: 0.5rem 0.625rem;
  border-bottom: 1px solid #f0e3e8;
  vertical-align: middle;
}
.loli-tool-table th {
  font-weight: 600;
  color: #6b7280;
  border-bottom: 1px solid #e5d4dc;
  background: #fff5f7;
}
.loli-tool-table td code {
  font-weight: 600;
  color: #e91e63;
  background: transparent;
}
.loli-input {
  width: 100%;
  border: 1px solid #e5d4dc;
  border-radius: 0.375rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
  background: #fff;
}
.loli-input:focus {
  outline: none;
  border-color: #e91e63;
  box-shadow: 0 0 0 2px rgba(233, 30, 99, 0.15);
}
.loli-kind-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}
.loli-kind-btn {
  padding: 0.375rem 0;
  border-radius: 0.375rem;
  border: 1px solid #e5d4dc;
  font-size: 0.8rem;
  color: #6b7280;
  background: #fff;
  transition: all 0.2s;
  cursor: pointer;
}
.loli-kind-btn.is-active {
  border-color: #e91e63;
  background: #fff5f7;
  color: #e91e63;
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
  background: #e91e63;
}
.loli-switch-input:checked + .loli-switch-label::after {
  transform: translateX(1.8em);
}
</style>
