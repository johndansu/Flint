#!/usr/bin/env node
// Postinstall script — downloads the correct Go binary for the current platform.
// Pattern mirrors esbuild, Turbo, and other Go-in-npm tools.

'use strict';

const https   = require('https');
const fs      = require('fs');
const path    = require('path');
const os      = require('os');
const zlib    = require('zlib');

const pkg     = require('./package.json');
const VERSION = pkg.version;
const REPO    = 'flintlang/flint';
const BIN_DIR = path.join(__dirname, 'bin');

const PLATFORM_MAP = {
  darwin:  'darwin',
  linux:   'linux',
  win32:   'windows',
};

const ARCH_MAP = {
  x64:   'amd64',
  arm64: 'arm64',
};

function main() {
  const platform = PLATFORM_MAP[os.platform()];
  const arch     = ARCH_MAP[os.arch()];

  if (!platform) throw new Error(`Unsupported platform: ${os.platform()}`);
  if (!arch)     throw new Error(`Unsupported arch: ${os.arch()}`);

  const ext = platform === 'windows' ? '.exe' : '';

  const binaries = [
    { name: 'flint',  outFile: `flint${ext}`  },
    { name: 'flintd', outFile: `flintd${ext}` },
  ];

  fs.mkdirSync(BIN_DIR, { recursive: true });

  Promise.all(binaries.map(({ name, outFile }) => {
    const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${name}-${platform}-${arch}${ext}`;
    const dest = path.join(BIN_DIR, outFile);
    return download(url, dest);
  })).catch(err => {
    // Graceful degradation — if download fails, warn but don't break npm install.
    // The shim will report the missing binary when invoked.
    process.stderr.write(`\nFlint: binary download failed — ${err.message}\n`);
    process.stderr.write(`Run 'flint' after ensuring internet access, or build from source.\n\n`);
  });
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest, { mode: 0o755 });

    function get(url) {
      https.get(url, res => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return get(res.headers.location); // follow redirects
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        res.pipe(file);
        file.on('finish', () => { file.close(); resolve(); });
        file.on('error', reject);
      }).on('error', reject);
    }

    get(url);
  });
}

main();
