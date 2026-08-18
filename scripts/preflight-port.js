// Pre-flight port check for dev:server. If the configured PORT is already
// in use, prints a warning and kills the occupying process(es) before the
// Go server starts, so `pnpm dev` does not fail with "bind: address
// already in use" when a stale server is left over from a previous run.
//
// Cross-platform: on Windows it uses netstat/tasklist/taskkill; on
// macOS/Linux it uses lsof/ps/kill. A listener is treated as "ours" if
// its working directory is this repo OR its own/parent command line
// matches the lolicount server. Unrelated listeners are reported with a
// hint to change PORT and left untouched.
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const PORT = Number.parseInt(process.env.PORT || '9721', 10);
const REPO = fileURLToPath(new URL('..', import.meta.url)).replace(/\/+$/, '');
const IS_WIN = process.platform === 'win32';

// A listener is "ours" if its cwd is this repo, or its own / parent
// command line matches the lolicount server entry points.
const ownByCmd = (cmd) =>
  cmd.includes('cmd/server') ||
  cmd.includes('lolicount') ||
  cmd.includes('.output/server') ||
  (cmd.includes('go run') && cmd.includes('cmd/server'));

const ownByCwd = (cwd) => cwd !== '' && cwd === REPO;

// --- Windows helpers -------------------------------------------------

function listenersOnPortWin(port) {
  // netstat -ano -p TCP, last column is PID. Filter LISTENING rows.
  let out;
  try {
    out = execFileSync('netstat', ['-ano', '-p', 'TCP'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
  } catch {
    return [];
  }
  const rows = [];
  for (const line of out.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed.includes('LISTENING')) continue;
    // "TCP    0.0.0.0:9721    0.0.0.0:0    LISTENING    1234"
    const cols = trimmed.split(/\s+/);
    const local = cols[1] ?? '';
    const pid = Number(cols[cols.length - 1]);
    if (local.endsWith(`:${port}`) && Number.isFinite(pid)) {
      rows.push({ command: '', pid });
    }
  }
  return rows;
}

function cmdlineOfWin(pid) {
  try {
    // wmic gives the full command line. Wrapped in quotes for safety.
    const out = execFileSync('wmic', ['process', 'where', `ProcessId=${pid}`, 'get', 'CommandLine', '/format:list'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const line = out.split('\n').find((l) => l.startsWith('CommandLine='));
    return line ? line.slice('CommandLine='.length).trim() : '';
  } catch {
    return '';
  }
}

function cwdOfWin(pid) {
  try {
    const out = execFileSync('wmic', ['process', 'where', `ProcessId=${pid}`, 'get', 'ExecutablePath', '/format:list'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const line = out.split('\n').find((l) => l.startsWith('ExecutablePath='));
    // ExecutablePath is not cwd, but on Windows we can't reliably get cwd
    // without extra tooling. Return empty so we fall back to cmd matching.
    return line ? '' : '';
  } catch {
    return '';
  }
}

function killPidWin(pid) {
  try {
    execFileSync('taskkill', ['/PID', String(pid), '/F', '/T'], { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

// --- Unix helpers ----------------------------------------------------

function listenersOnPortUnix(port) {
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

function cmdlineOfUnix(pid) {
  try {
    return execFileSync('ps', ['-o', 'command=', '-p', String(pid)], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
  } catch {
    return '';
  }
}

function cwdOfUnix(pid) {
  try {
    const out = execFileSync('lsof', ['-a', '-p', String(pid), '-d', 'cwd', '-Fn'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const line = out.split('\n').find((l) => l.startsWith('n'));
    return line ? line.slice(1).replace(/\/+$/, '') : '';
  } catch {
    return '';
  }
}

function ppidOfUnix(pid) {
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

function killPidUnix(pid) {
  try {
    process.kill(pid, 'SIGTERM');
    return true;
  } catch {
    try {
      execFileSync('kill', ['-9', String(pid)], { stdio: 'ignore' });
      return true;
    } catch {
      return false;
    }
  }
}

// --- Platform-dispatched wrappers -----------------------------------

const listenersOnPort = IS_WIN ? listenersOnPortWin : listenersOnPortUnix;
const cmdlineOf = IS_WIN ? cmdlineOfWin : cmdlineOfUnix;
const cwdOf = IS_WIN ? cwdOfWin : cwdOfUnix;
const killPid = IS_WIN ? killPidWin : killPidUnix;

function isOurs(row) {
  const cmd = cmdlineOf(row.pid);
  if (ownByCmd(cmd)) return { parent: false };
  const cwd = cwdOf(row.pid);
  if (ownByCwd(cwd)) return { parent: false };
  // Check the parent (go run spawns a temp binary; the parent matches).
  // Parent-walking is Unix-only; on Windows the cmd line already covers
  // "go run ./cmd/server" via wmic.
  if (!IS_WIN) {
    const parent = ppidOfUnix(row.pid);
    if (parent) {
      const parentCmd = cmdlineOfUnix(parent);
      if (ownByCmd(parentCmd)) return { parent: true, parentPid: parent };
      const parentCwd = cwdOfUnix(parent);
      if (ownByCwd(parentCwd)) return { parent: true, parentPid: parent };
    }
  }
  return null;
}

const rows = listenersOnPort(PORT);
if (rows.length === 0) {
  process.exit(0); // port free, nothing to do
}

console.warn(`\u26A0\uFE0F  Port ${PORT} is already in use by:`);
for (const row of rows) {
  console.warn(`   - ${row.command || 'process'} (PID ${row.pid})`);
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
      const ok = killPid(pid);
      console.warn(ok ? `   killed PID ${pid}` : `   failed to kill PID ${pid} — kill it manually`);
    }
  }
  await new Promise((r) => setTimeout(r, 800));
}

if (foreign.length > 0) {
  console.warn(
    `\u26A0\uFE0F  Port ${PORT} also has unrelated listener(s) that were NOT killed:\n` +
      foreign.map((r) => `   - ${r.command || 'process'} (PID ${r.pid})`).join('\n') +
      `\n   Set a different PORT, e.g. PORT=9722 pnpm dev:server\n`,
  );
  process.exit(1);
}
