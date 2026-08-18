// Cross-platform launcher for the Nuxt dev server. Sets the API base env
// var before spawning `nuxt dev` so the same `pnpm dev:web` command works
// on Windows (cmd/PowerShell do not support inline `VAR=val cmd` syntax),
// macOS and Linux.
//
// It also loads the repo-root .env (the same file the Go backend reads
// via godotenv) and forwards BASE_URL as NUXT_PUBLIC_BASE_URL so the SSG
// runtime config picks up the public embed-link domain in dev too.
import { spawn } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..');

// Minimal .env loader (KEY=VALUE lines, skips comments/blank). We do not
// pull in dotenv to avoid an extra dependency; the format is simple and
// this only needs to read the repo's own .env.example-shaped file.
const loadDotEnv = (path) => {
  let text;
  try {
    text = readFileSync(path, 'utf8');
  } catch {
    return; // missing .env is fine
  }
  for (const line of text.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const eq = trimmed.indexOf('=');
    if (eq < 0) continue;
    const key = trimmed.slice(0, eq).trim();
    const val = trimmed.slice(eq + 1).trim().replace(/^["']|["']$/g, '');
    // Never override vars already set in the real environment.
    if (process.env[key] === undefined) process.env[key] = val;
  }
};

loadDotEnv(join(repoRoot, '.env'));

process.env.NUXT_PUBLIC_API_BASE = process.env.NUXT_PUBLIC_API_BASE || 'http://127.0.0.1:9721';
// Forward BASE_URL to Nuxt's public runtime config (see nuxt.config.ts).
process.env.NUXT_PUBLIC_BASE_URL = process.env.NUXT_PUBLIC_BASE_URL || process.env.BASE_URL || '';

const child = spawn('pnpm', ['--dir', 'web', 'dev'], {
  stdio: 'inherit',
  shell: process.platform === 'win32',
});

child.on('exit', (code) => process.exit(code ?? 0));
