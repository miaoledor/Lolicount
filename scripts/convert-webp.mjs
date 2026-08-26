// Convert PNG/JPG/JPEG images in theme asset directories to WebP.
// Uses sharp (libvips) for encoding. WebP offers significant size
// savings over PNG for the same visual quality, and is already the
// preferred format per AGENTS.md theme conventions.
//
// This script converts in place: the original file is replaced by a
// .webp file (same directory, same basename). Files that are already
// WebP are skipped. GIFs are skipped by default because sharp's WebP
// encoder does not support animation — use --force-gif to convert the
// first frame only (loses animation).
//
// Scope: assets/theme/** and assets/character/**
// assets/dist/ (SSG output) and assets/f-theme/ (JSON) are skipped.
//
// Usage:
//   pnpm convert:webp              convert all PNG/JPG to WebP in place
//   pnpm convert:webp --check      dry-run, report what would change (exit 1 if convertible)
//   pnpm convert:webp --quality 90 set WebP quality (1-100, default 90)
//   pnpm convert:webp --force-gif  also convert GIFs (first frame only, loses animation)
//   pnpm convert:webp --verbose    show per-file output
//
// Iron Rule 4 (server-side re-encode on upload) is unaffected: this
// script only touches pre-built-in assets, not the upload channel.
import { readdir, stat, unlink } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join, extname, basename } from 'node:path';
import sharp from 'sharp';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..');

const SCAN_DIRS = ['assets/theme', 'assets/character'];

// Extensions eligible for conversion.
const CONVERT_EXTS = ['.png', '.jpg', '.jpeg'];
const GIF_EXT = '.gif';

const args = process.argv.slice(2);
const checkOnly = args.includes('--check');
const verbose = args.includes('--verbose');
const forceGif = args.includes('--force-gif');
const qualityIdx = args.indexOf('--quality');
const quality = qualityIdx >= 0 && args[qualityIdx + 1] ? Number(args[qualityIdx + 1]) : 90;

if (!Number.isInteger(quality) || quality < 1 || quality > 100) {
  console.error(`invalid --quality: ${quality} (expected 1-100)`);
  process.exit(2);
}

const log = (...a) => { if (verbose) console.log(...a); };

// Recursively collect convertible images under a directory.
const collectImages = async (dir) => {
  const out = [];
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return out;
  }
  for (const e of entries) {
    const full = join(dir, e.name);
    if (e.isDirectory()) {
      out.push(...(await collectImages(full)));
    } else if (e.isFile()) {
      const ext = extname(e.name).toLowerCase();
      if (CONVERT_EXTS.includes(ext)) {
        out.push({ path: full, ext, type: 'raster' });
      } else if (ext === GIF_EXT && forceGif) {
        out.push({ path: full, ext, type: 'gif' });
      }
    }
  }
  return out;
};

const main = async () => {
  const files = [];
  for (const d of SCAN_DIRS) {
    files.push(...(await collectImages(join(repoRoot, d))));
  }

  if (files.length === 0) {
    console.log('no convertible images found (PNG/JPG/GIF)');
    return;
  }

  let beforeTotal = 0;
  for (const f of files) {
    beforeTotal += (await stat(f.path)).size;
  }

  const mode = checkOnly ? 'check (dry-run)' : 'convert';
  console.log(`${mode}: ${files.length} images, ${(beforeTotal / 1024 / 1024).toFixed(1)} MiB total, WebP quality ${quality}${forceGif ? ' (GIF: first frame)' : ''}`);

  let converted = 0;
  let skipped = 0;
  let savedBytes = 0;

  for (const f of files) {
    const outPath = join(dirname(f.path), basename(f.path, f.ext) + '.webp');

    // Skip if a .webp with the same basename already exists.
    try {
      await stat(outPath);
      log(`  skip (webp exists): ${f.path}`);
      skipped++;
      continue;
    } catch { /* no existing webp, proceed */ }

    const beforeSize = (await stat(f.path)).size;

    if (checkOnly) {
      log(`  would convert: ${f.path} -> ${outPath}`);
      converted++;
      continue;
    }

    try {
      await sharp(f.path)
        .webp({ quality, lossless: false, alphaQuality: 90 })
        .toFile(outPath);

      const afterSize = (await stat(outPath)).size;
      const delta = beforeSize - afterSize;
      savedBytes += delta;

      // Remove the original file after successful conversion.
      await unlink(f.path);

      log(`  converted: ${f.path} -> ${outPath} (${(delta / 1024).toFixed(1)} KiB saved)`);
      converted++;
    } catch (err) {
      console.error(`  failed: ${f.path}: ${err.message}`);
      // Clean up partial output if conversion failed.
      try { await unlink(outPath); } catch { /* ignore */ }
    }
  }

  if (checkOnly) {
    if (converted > 0) {
      console.log(`convertible: ${converted} files would be converted to WebP`);
      console.log('run `pnpm convert:webp` to apply');
      process.exit(1);
    }
    console.log('all images already WebP; nothing to do');
    return;
  }

  const pct = beforeTotal > 0 ? (savedBytes * 100 / beforeTotal).toFixed(1) : '0.0';
  console.log(`done: ${converted} converted, ${skipped} skipped, ~${(savedBytes / 1024 / 1024).toFixed(1)} MiB saved (${pct}% smaller)`);
};

main().catch((err) => {
  console.error('convert-webp failed:', err.message);
  process.exit(1);
});
