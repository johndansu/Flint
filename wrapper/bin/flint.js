#!/usr/bin/env node
// Shim — locates the downloaded Go binary and execs it with the user's arguments.
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs   = require('fs');
const os   = require('os');

const ext    = os.platform() === 'win32' ? '.exe' : '';
const binary = path.join(__dirname, `flint${ext}`);

if (!fs.existsSync(binary)) {
  process.stderr.write(
    `Flint binary not found at ${binary}.\n` +
    `Try reinstalling: npm install -g @flintlang/cli\n` +
    `Or build from source: https://github.com/flintlang/flint\n`
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status ?? 1);
