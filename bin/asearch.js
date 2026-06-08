#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

function executableExists(file) {
  try {
    fs.accessSync(file, fs.constants.X_OK);
    return true;
  } catch (_) {
    return false;
  }
}

function getBinaryPath() {
  const ext = process.platform === 'win32' ? '.exe' : '';

  if (process.env.ASEARCH_BIN) {
    if (executableExists(process.env.ASEARCH_BIN)) return process.env.ASEARCH_BIN;
    console.error(`[asearch] ASEARCH_BIN does not point to an executable: ${process.env.ASEARCH_BIN}`);
    process.exit(1);
  }

  // 1. Same directory as wrapper. This keeps package-local/dev installs deterministic.
  const localPath = path.join(__dirname, 'asearch' + ext);
  if (executableExists(localPath)) return localPath;

  const stateDir = process.env.ASEARCH_STATE_DIR
    || path.join(process.env.HOME || process.env.USERPROFILE || '', '.asearch');

  // 2. ~/.asearch/bin/asearch (installed by postinstall)
  const installedPath = path.join(stateDir, 'bin', 'asearch' + ext);
  if (executableExists(installedPath)) return installedPath;

  console.error('[asearch] binary not found — run: npm explore agent-asearch -g -- npm run postinstall');
  process.exit(1);
}

const bin = getBinaryPath();
const args = process.argv.slice(2);

const child = spawn(bin, args, { stdio: 'inherit' });
child.on('error', (err) => {
  console.error(`[asearch] failed to start binary: ${err.message}`);
  process.exit(1);
});
child.on('close', (code, signal) => {
  if (typeof code === 'number') {
    process.exit(code);
  }
  if (signal) {
    const signalNumber = Number(String(signal).replace(/^SIG/, ''));
    process.exit(Number.isInteger(signalNumber) ? 128 + signalNumber : 1);
  }
  process.exit(1);
});
