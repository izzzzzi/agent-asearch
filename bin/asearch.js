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

  // 1. ~/.asearch/bin/asearch (installed by postinstall)
  const installedPath = path.join(stateDir, 'bin', 'asearch' + ext);
  if (fs.existsSync(installedPath)) return installedPath;

  // 2. Same directory as wrapper
  const localPath = path.join(__dirname, 'asearch' + ext);
  if (fs.existsSync(localPath)) return localPath;

  // 3. PATH lookup (skip if it's this same script)
  try {
    const which = require('child_process').execSync('which asearch 2>/dev/null || command -v asearch', { encoding: 'utf8' }).trim();
    if (which && !which.includes('bin/asearch.js')) {
      return which;
    }
  } catch (_) {}

  console.error('[asearch] binary not found — run: npm explore agent-asearch -g -- npm run postinstall');
  process.exit(1);
}

const bin = getBinaryPath();
const args = process.argv.slice(2);

const child = spawn(bin, args, { stdio: 'inherit' });
child.on('close', (code) => { process.exit(code); });
