// Pre-flight port check for dev:server. If the configured PORT is already
// in use, prints a warning and kills the occupying process(es) before the
// Go server starts, so `pnpm dev` does not fail with "bind: address
// already in use" when a stale server is left over from a previous run.
//
// Only kills processes whose own command line or parent's command line
// matches this project's server (go run ./cmd/server, the compiled binary,
// or the Nuxt preview server). Unrelated listeners are reported with a
// hint to change PORT and are left untouched.
import { execFileSync } from 'node:child_process';

const PORT = Number.parseInt(process.env.PORT || '8721', 10);

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

// A listener is "ours" if its own command line or its parent's matches the
// lolicount server. go run ./cmd/server spawns a temp binary whose command
// line is an opaque path ending in /server, so matching the parent
// (which contains "go run ./cmd/server") is what catches the dev case.
const isOwnServer = (cmd) =>
  cmd.includes('cmd/server') ||
  cmd.includes('lolicount') ||
  cmd.includes('.output/server') ||
  (cmd.includes('go run') && cmd.includes('cmd/server'));

const rows = listenersOnPort(PORT);
if (rows.length === 0) {
  process.exit(0); // port free, nothing to do
}

console.warn(`\u26A0\uFE0F  Port ${PORT} is already in use by:`);
for (const row of rows) {
  console.warn(`   - ${row.command} (PID ${row.pid})`);
}

// Classify each listener: own (stale lolicount) vs foreign.
const ours = [];
const foreign = [];
for (const row of rows) {
  const cmd = cmdlineOf(row.pid);
  const parent = ppidOf(row.pid);
  const parentCmd = parent ? cmdlineOf(parent) : '';
  if (isOwnServer(cmd) || isOwnServer(parentCmd)) {
    // Also kill the parent go-run so it does not respawn the binary.
    ours.push({ ...row, parent: parent && isOwnServer(parentCmd) ? parent : 0 });
  } else {
    foreign.push(row);
  }
}

if (ours.length > 0) {
  console.warn(`\u26A0\uFE0F  Killing stale lolicount process(es) on port ${PORT}...`);
  for (const row of ours) {
    for (const pid of [row.parent, row.pid].filter((p) => p && p !== 0)) {
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
  // Give the OS a moment to release the socket.
  await new Promise((r) => setTimeout(r, 800));
}

if (foreign.length > 0) {
  console.warn(
    `\u26A0\uFE0F  Port ${PORT} also has unrelated listener(s) that were NOT killed:\n` +
      foreign.map((r) => `   - ${r.command} (PID ${r.pid})`).join('\n') +
      `\n   Set a different PORT, e.g. PORT=8730 pnpm dev:server\n`,
  );
  process.exit(1);
}
