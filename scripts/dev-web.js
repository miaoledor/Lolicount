// Cross-platform launcher for the Nuxt dev server. Sets the API base env
// var before spawning `nuxt dev` so the same `pnpm dev:web` command works
// on Windows (cmd/PowerShell do not support inline `VAR=val cmd` syntax),
// macOS and Linux.
import { spawn } from 'node:child_process';

process.env.NUXT_PUBLIC_API_BASE = process.env.NUXT_PUBLIC_API_BASE || 'http://127.0.0.1:9721';

const child = spawn('pnpm', ['--dir', 'web', 'dev'], {
  stdio: 'inherit',
  shell: process.platform === 'win32',
});

child.on('exit', (code) => process.exit(code ?? 0));
