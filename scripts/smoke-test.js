#!/usr/bin/env node
const { spawnSync } = require('child_process');
const fs = require('fs');
const os = require('os');
const path = require('path');

const root = path.join(__dirname, '..');
const pkg = require('../package.json');

console.log('=== asearch smoke test ===');

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'asearch-smoke-'));
const stateDir = path.join(tmpDir, 'state');
const binPath = path.join(tmpDir, process.platform === 'win32' ? 'asearch.exe' : 'asearch');
let failed = 0;

function run(command, args, options = {}) {
  return spawnSync(command, args, {
    cwd: root,
    encoding: 'utf8',
    env: {
      ...process.env,
      ASEARCH_BIN: options.useWrapper === false ? process.env.ASEARCH_BIN : binPath,
      ASEARCH_STATE_DIR: stateDir,
      ...options.env,
    },
  });
}

function record(name, ok, detail = '') {
  console.log(`${name}: ${ok ? 'PASS' : 'FAIL'}${detail ? ` ${detail}` : ''}`);
  if (!ok) failed++;
}

function parseJson(text) {
  return JSON.parse(String(text || '').trim());
}

function buildTestBinary() {
  const result = run('go', [
    'build',
    '-ldflags', `-X github.com/izzzzzi/agent-asearch/internal/cli.Version=${pkg.version}`,
    '-o', binPath,
    './cmd/asearch',
  ], { useWrapper: false });
  if (result.status !== 0) {
    process.stderr.write(result.stdout || '');
    process.stderr.write(result.stderr || '');
    throw new Error(`failed to build test binary (exit ${result.status})`);
  }
}

try {
  buildTestBinary();

  const ver = run('node', ['bin/asearch.js', 'version']);
  const v = parseJson(ver.stdout);
  record('version', ver.status === 0 && v.ok === true && v.version === pkg.version, v.version);

  const help = run('node', ['bin/asearch.js', '--help']);
  const h = parseJson(help.stdout);
  record('help', help.status === 0 && h.ok === true && h.tool === 'asearch', h.tool);

  const prompt = run('node', ['bin/asearch.js', 'prompt']);
  record('prompt', prompt.status === 0 && prompt.stdout.length > 100, `${prompt.stdout.length} chars`);

  const doc = run('node', ['bin/asearch.js', 'doctor']);
  const d = parseJson(doc.stdout);
  record('doctor', doc.status === 0 && Array.isArray(d.checks), `(${Array.isArray(d.checks) ? d.checks.length : 0} backends)`);

  const sl = run('node', ['bin/asearch.js', 'session', 'list']);
  const s = parseJson(sl.stdout);
  record('session list', sl.status === 0 && s.ok === true && s.count === 0, `count: ${s.count}`);
} catch (e) {
  console.error('FAIL:', e.message);
  failed++;
} finally {
  fs.rmSync(tmpDir, { recursive: true, force: true });
}

console.log(`\n=== ${failed === 0 ? 'all smoke tests passed' : failed + ' smoke tests FAILED'} ===`);
process.exit(failed === 0 ? 0 : 1);
