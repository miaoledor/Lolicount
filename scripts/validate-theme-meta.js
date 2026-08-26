// Validates every meta.json under assets/theme/<name>/ against the theme
// metadata schema. Run locally and in CI (theme-check workflow). meta.json
// is optional; directories without one are skipped.
//
// Schema (all fields optional except `name`):
//   name        string  must equal the parent directory name
//   author      string  non-empty
//   description string  non-empty
//   tags        array   of non-empty strings
//   version     string  semver-ish, non-empty
import { readdir, readFile, stat } from 'node:fs/promises';
import { join } from 'node:path';

const THEME_DIR = new URL('../assets/theme/', import.meta.url);

const errors = [];
let checked = 0;

function isNonEmptyString(v) {
  return typeof v === 'string' && v.trim().length > 0;
}

function validate(meta, dir) {
  if (!isNonEmptyString(meta.name)) {
    errors.push(`${dir}/meta.json: missing or empty "name"`);
  } else if (meta.name !== dir) {
    errors.push(`${dir}/meta.json: name "${meta.name}" does not match directory "${dir}"`);
  }
  if ('author' in meta && !isNonEmptyString(meta.author)) {
    errors.push(`${dir}/meta.json: "author" must be a non-empty string`);
  }
  if ('description' in meta && !isNonEmptyString(meta.description)) {
    errors.push(`${dir}/meta.json: "description" must be a non-empty string`);
  }
  if ('version' in meta && !isNonEmptyString(meta.version)) {
    errors.push(`${dir}/meta.json: "version" must be a non-empty string`);
  }
  if ('tags' in meta) {
    if (!Array.isArray(meta.tags)) {
      errors.push(`${dir}/meta.json: "tags" must be an array`);
    } else {
      meta.tags.forEach((t, i) => {
        if (!isNonEmptyString(t)) {
          errors.push(`${dir}/meta.json: tags[${i}] must be a non-empty string`);
        }
      });
    }
  }
  // Reject unknown top-level keys to keep the schema tight.
  const allowed = new Set(['name', 'author', 'description', 'tags', 'version']);
  for (const key of Object.keys(meta)) {
    if (!allowed.has(key)) {
      errors.push(`${dir}/meta.json: unknown field "${key}"`);
    }
  }
}

try {
  let dirs;
  try {
    dirs = await readdir(THEME_DIR, { withFileTypes: true });
  } catch {
    console.error('validate-theme-meta: assets/theme/ not found');
    process.exit(2);
  }
  for (const d of dirs) {
    if (!d.isDirectory()) continue;
    const metaPath = join(THEME_DIR.pathname, d.name, 'meta.json');
    try {
      await stat(metaPath);
    } catch {
      continue; // meta.json is optional
    }
    checked++;
    let raw;
    try {
      raw = await readFile(metaPath, 'utf8');
    } catch (e) {
      errors.push(`${d.name}/meta.json: cannot read: ${e.message}`);
      continue;
    }
    let meta;
    try {
      meta = JSON.parse(raw);
    } catch (e) {
      errors.push(`${d.name}/meta.json: invalid JSON: ${e.message}`);
      continue;
    }
    if (typeof meta !== 'object' || meta === null || Array.isArray(meta)) {
      errors.push(`${d.name}/meta.json: root must be an object`);
      continue;
    }
    validate(meta, d.name);
  }
} catch (e) {
  console.error(`validate-theme-meta: ${e.message}`);
  process.exit(2);
}

console.log(`checked ${checked} meta.json file(s)`);
if (errors.length > 0) {
  for (const msg of errors) console.error(`  - ${msg}`);
  process.exit(1);
}
console.log('all meta.json valid');
