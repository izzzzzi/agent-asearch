#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

function getBinaryPath() {
  const platform = process.platform;
  const arch = process.arch === 'x64' ? 'amd64' : process.arch === 'arm64' ? 'arm64' : process.arch;
  const ext = platform === 'win32' ? '.exe' : '';

  const stateDir = process.env.ASEARCH_STATE_DIR
    || path.join(process.env.HOME || process.env.USERPROFILE || '', '.asearch');

  const binaryPath = path.join(stateDir, 'bin', `asearch-${platform}-${arch}${ext}`);

  if (fs.existsSync(binaryPath)) {
    return binaryPath;
  }

  const names = platform === 'win32'
    ? ['asearch.exe', 'asearch']
    : ['asearch'];

  for (const name of names) {
    const p = path.join(__dirname, name);
    if (fs.existsSync(p)) return p;
  }

  for (const name of names) {
    const p = path.join(stateDir, 'bin', name);
    if (fs.existsSync(p)) return p;
  }

  return 'asearch';
}

const bin = getBinaryPath();
const args = process.argv.slice(2);

const child = spawn(bin, args, { stdio: 'inherit' });
child.on('close', (code) => { process.exit(code); });
