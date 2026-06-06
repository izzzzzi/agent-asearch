#!/usr/bin/env node
// Download native binary on install
const fs = require('fs');
const path = require('path');

const stateDir = process.env.ASEARCH_STATE_DIR
  || path.join(process.env.HOME || process.env.USERPROFILE || '', '.asearch');
const binDir = path.join(stateDir, 'bin');

// If SKIP_DOWNLOAD is set, create a placeholder
if (process.env.AGENT_ASEARCH_SKIP_DOWNLOAD === '1') {
  if (!fs.existsSync(binDir)) fs.mkdirSync(binDir, { recursive: true });
  const placeholder = path.join(binDir, 'asearch');
  if (!fs.existsSync(placeholder)) {
    fs.writeFileSync(placeholder, '#!/bin/sh\necho \'{"ok":false,"code":"not_downloaded","message":"binary not downloaded","hint":"run npm i -g agent-asearch or set ASEARCH_BINARY_PATH"}\'\nexit 1\n');
    try { fs.chmodSync(placeholder, 0o755); } catch (_) {}
  }
  console.log('[agent-asearch] skipped binary download (AGENT_ASEARCH_SKIP_DOWNLOAD=1)');
  process.exit(0);
}

// Create bin directory
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Build locally if go is available and source is present
const goModPath = path.join(__dirname, '..', 'go.mod');
if (fs.existsSync(goModPath)) {
  const { spawnSync } = require('child_process');
  const result = spawnSync('go', ['build', '-o', path.join(binDir, 'asearch'), './cmd/asearch'], {
    cwd: path.join(__dirname, '..'),
    stdio: 'pipe'
  });
  if (result.status === 0) {
    console.log('[agent-asearch] built from source');
    process.exit(0);
  }
}

// Not from npm - just exit quietly
console.log('[agent-asearch] install stub — binary will be built at first use');
process.exit(0);
