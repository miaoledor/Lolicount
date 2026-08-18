// Generates assets/themes.json: a manifest of built-in themes derived
// from the assets/theme directory tree. Each entry records the theme
// name, frame count, frame extension, and optional meta.json fields.
// The frontend and API consume this manifest to list themes.
import { readdir, readFile, writeFile, stat } from 'node:fs/promises';
import { join } from 'node:path';

const THEME_DIR = new URL('../assets/theme/', import.meta.url);
const OUT_FILE = new URL('../assets/themes.json', import.meta.url);

const SUPPORTED = new Set(['.gif', '.png', '.webp']);

function frameIndex(stem) {
  const n = Number.parseInt(stem, 10);
  return Number.isInteger(n) && n >= 0 ? n : -1;
}

async function loadMeta(dir) {
  const metaPath = join(THEME_DIR.pathname, dir, 'meta.json');
  try {
    await stat(metaPath);
  } catch {
    return {};
  }
  try {
    return JSON.parse(await readFile(metaPath, 'utf8'));
  } catch {
    return {};
  }
}

const themes = [];
const dirs = await readdir(THEME_DIR, { withFileTypes: true });
for (const d of dirs) {
  if (!d.isDirectory()) continue;
  const files = await readdir(join(THEME_DIR.pathname, d.name));
  let count = 0;
  let ext = '';
  for (const f of files) {
    const e = f.slice(f.lastIndexOf('.')).toLowerCase();
    if (!SUPPORTED.has(e)) continue;
    const stem = f.slice(0, f.lastIndexOf('.'));
    if (frameIndex(stem) < 0) continue;
    count++;
    if (!ext) ext = e;
  }
  if (count === 0) continue;
  const meta = await loadMeta(d.name);
  themes.push({
    name: d.name,
    frames: count,
    ext: ext.replace('.', ''),
    ...meta,
  });
}

themes.sort((a, b) => a.name.localeCompare(b.name));
const output = JSON.stringify({ themes }, null, 2) + '\n';
await writeFile(OUT_FILE, output, 'utf8');
console.log(`generated assets/themes.json with ${themes.length} theme(s)`);
