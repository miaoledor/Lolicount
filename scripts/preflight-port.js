// Pre-flight port check for dev:server. If the configured PORT is already
// in use, prints a warning and kills the occupying process(es) before the
// Go server starts, so `pnpm dev` does not fail with "bind: address
// already in use" when a stale server is left over from a previous run.
//
// A listener is treated as "ours" if its working directory is this repo
// OR its own/parent command line matches the lolicount server. This
// catches: go run ./cmd/server (parent matches), its compiled temp binary
// (cwd is the repo), the compiled release binary (cmdline has lolicount),
// and the Nuxt preview server (cmdline has .output/server). Unrelated
// listeners are reported with a hint to change PORT and left untouched.
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const PORT = Number.parseInt(process.env.PORT || '9721', 10);
const REPO = fileURLToPath(new URL('..', import.meta.url)).replace(/\/+$/, '');

// lsof returns one line per listener. Column layout (with -P -n):
// COMMAND   PID  USER  FD  TYPE  DEVICE  SIZE/OFF  NODE  NAME
function listenersOnPort(port) {
  try {
    const out = execFileSync('lsof', ['-nP', `-iTCP:${port}`, '-sTCP:LISTEN'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    return out
      .split('\n')
      .slice(1) // skip header
      .filter(Boolean)
      .map((line) => {
        const cols = line.trim().split(/\s+/);
        return { command: cols[0] ?? '', pid: Number(cols[1]) };
      })
      .filter((row) => Number.isFinite(row.pid));
  } catch {
    return []; // lsof found nothing (port free) or not installed
  }
}

// Full command line of a PID (best-effort; empty if the process is gone).
function cmdlineOf(pid) {
  try {
    return execFileSync('ps', ['-o', 'command=', '-p', String(pid)], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
  } catch {
    return '';
  }
}

// Working directory of a PID via lsof (best-effort; empty if unavailable).
function cwdOf(pid) {
  try {
    const out = execFileSync('lsof', ['-a', '-p', String(pid), '-d', 'cwd', '-Fn'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    // lsof -Fn prints "n<path>"; take the first n line.
    const line = out.split('\n').find((l) => l.startsWith('n'));
    return line ? line.slice(1).replace(/\/+$/, '') : '';
  } catch {
    return '';
  }
}

// Parent PID of a process (0 if unknown / gone).
function ppidOf(pid) {
  try {
    const out = execFileSync('ps', ['-o', 'ppid=', '-p', String(pid)], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
    const n = Number(out);
    return Number.isFinite(n) ? n : 0;
  } catch {
    return 0;
  }
}

// A listener is "ours" if its cwd is this repo, or its own / parent
// command line matches the lolicount server entry points.
const ownByCmd = (cmd) =>
  cmd.includes('cmd/server') ||
  cmd.includes('lolicount') ||
  cmd.includes('.output/server') ||
  (cmd.includes('go run') && cmd.includes('cmd/server'));

const ownByCwd = (cwd) => cwd !== '' && cwd === REPO;

function isOurs(row) {
  const cmd = cmdlineOf(row.pid);
  if (ownByCmd(cmd)) return { parent: false };
  const cwd = cwdOf(row.pid);
  if (ownByCwd(cwd)) return { parent: false };
  // Check the parent (go run spawns a temp binary; the parent matches).
  const parent = ppidOf(row.pid);
  if (parent) {
    const parentCmd = cmdlineOf(parent);
    if (ownByCmd(parentCmd)) return { parent: true, parentPid: parent };
    const parentCwd = cwdOf(parent);
    if (ownByCwd(parentCwd)) return { parent: true, parentPid: parent };
  }
  return null;
}

const rows = listenersOnPort(PORT);
if (rows.length === 0) {
  process.exit(0); // port free, nothing to do
}

console.warn(`\u26A0\uFE0F  Port ${PORT} is already in use by:`);
for (const row of rows) {
  console.warn(`   - ${row.command} (PID ${row.pid})`);
}

const ours = [];
const foreign = [];
for (const row of rows) {
  const res = isOurs(row);
  if (res) ours.push({ ...row, ...res });
  else foreign.push(row);
}

if (ours.length > 0) {
  console.warn(`\u26A0\uFE0F  Killing stale lolicount process(es) on port ${PORT}...`);
  for (const row of ours) {
    const pids = [row.parentPid, row.pid].filter((p) => p && p !== 0);
    for (const pid of pids) {
      try {
        process.kill(pid, 'SIGTERM');
        console.warn(`   killed PID ${pid}`);
      } catch {
        try {
          execFileSync('kill', ['-9', String(pid)], { stdio: 'ignore' });
          console.warn(`   force-killed PID ${pid}`);
        } catch {
          console.warn(`   failed to kill PID ${pid} \u2014 kill it manually`);
        }
      }
    }
  }
  await new Promise((r) => setTimeout(r, 800));
}

if (foreign.length > 0) {
  console.warn(
    `\u26A0\uFE0F  Port ${PORT} also has unrelated listener(s) that were NOT killed:\n` +
      foreign.map((r) => `   - ${r.command} (PID ${r.pid})`).join('\n') +
      `\n   Set a different PORT, e.g. PORT=9722 pnpm dev:server\n`,
  );
  process.exit(1);
}
