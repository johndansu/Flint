#!/usr/bin/env node
// Postinstall script — downloads the correct Go binaries and verifies their SHA256.
// Pattern mirrors esbuild, Turbo, and other Go-in-npm tools.

'use strict';

const https  = require('https');
const fs     = require('fs');
const path   = require('path');
const os     = require('os');
const crypto = require('crypto');

const pkg     = require('./package.json');
const VERSION = pkg.version;
const REPO    = 'johndansu/Flint';
const BIN_DIR = path.join(__dirname, 'bin');

const PLATFORM_MAP = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
const ARCH_MAP     = { x64: 'amd64', arm64: 'arm64' };

async function main() {
  const platform = PLATFORM_MAP[os.platform()];
  const arch     = ARCH_MAP[os.arch()];

  if (!platform) throw new Error(`Unsupported platform: ${os.platform()}`);
  if (!arch)     throw new Error(`Unsupported arch: ${os.arch()}`);

  const ext  = platform === 'windows' ? '.exe' : '';
  const base = `https://github.com/${REPO}/releases/download/v${VERSION}`;

  // Fetch checksums.txt before downloading any binary
  const checksums = await fetchChecksums(`${base}/checksums.txt`);

  fs.mkdirSync(BIN_DIR, { recursive: true });

  const binaries = [
    { name: 'flint',  outFile: `flint${ext}`  },
    { name: 'flintd', outFile: `flintd${ext}` },
  ];

  await Promise.all(binaries.map(async ({ name, outFile }) => {
    const fileName     = `${name}-${platform}-${arch}${ext}`;
    const url          = `${base}/${fileName}`;
    const dest         = path.join(BIN_DIR, outFile);
    const expectedHash = checksums.get(fileName);

    if (!expectedHash) {
      throw new Error(`No checksum entry found for ${fileName} — checksums.txt may be stale.`);
    }

    const data       = await fetchBuffer(url);
    const actualHash = crypto.createHash('sha256').update(data).digest('hex');

    if (actualHash !== expectedHash) {
      throw new Error(
        `Checksum mismatch for ${fileName}.\n` +
        `  Expected: ${expectedHash}\n` +
        `  Got:      ${actualHash}\n` +
        `Refusing to write binary. Download may be corrupted or tampered with.`,
      );
    }

    fs.writeFileSync(dest, data, { mode: 0o755 });
  }));
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// Download checksums.txt and parse into Map<filename, sha256hex>
function fetchChecksums(url) {
  return fetchBuffer(url).then(buf => {
    const map = new Map();
    for (const line of buf.toString('utf8').split('\n')) {
      const parts = line.trim().split(/\s+/);
      if (parts.length === 2) {
        map.set(parts[1], parts[0]); // filename → hash
      }
    }
    return map;
  });
}

// Download url as a Buffer, following redirects (GitHub releases redirect to S3)
function fetchBuffer(url) {
  return new Promise((resolve, reject) => {
    const chunks = [];

    function get(u) {
      https.get(u, res => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
        }
        res.on('data', chunk => chunks.push(chunk));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      }).on('error', reject);
    }

    get(url);
  });
}

// ---------------------------------------------------------------------------
// Entry point — degrade gracefully so npm install never fails hard
// ---------------------------------------------------------------------------

main().catch(err => {
  process.stderr.write(`\nFlint: binary installation failed — ${err.message}\n`);
  process.stderr.write(`You can build from source or download manually:\n`);
  process.stderr.write(`  https://github.com/${REPO}/releases/tag/v${VERSION}\n\n`);
  // Exit 0: a failed download shouldn't break the developer's npm install.
  // The CLI shim will report the missing binary with a clear message when invoked.
  process.exit(0);
});
