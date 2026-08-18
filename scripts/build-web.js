// Cross-platform launcher for `nuxt generate`. Loads the repo-root .env
// (same file the Go backend reads) and forwards BASE_URL as
// NUXT_PUBLIC_BASE_URL so the SSG output bakes the public embed-link
// domain into the bundle at build time.
import { spawn } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..');

const loadDotEnv = (path) => {
  let text;
  try {
    text = readFileSync(path, 'utf8');
  } catch {
    return;
  }
  for (const line of text.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const eq = trimmed.indexOf('=');
    if (eq < 0) continue;
    const key = trimmed.slice(0, eq).trim();
    const val = trimmed.slice(eq + 1).trim().replace(/^["']|["']$/g, '');
    if (process.env[key] === undefined) process.env[key] = val;
  }
};

loadDotEnv(join(repoRoot, '.env'));
process.env.NUXT_PUBLIC_BASE_URL = process.env.NUXT_PUBLIC_BASE_URL || process.env.BASE_URL || '';

const child = spawn('pnpm', ['--dir', 'web', 'generate'], {
  stdio: 'inherit',
  shell: process.platform === 'win32',
});

child.on('exit', (code) => process.exit(code ?? 0));
