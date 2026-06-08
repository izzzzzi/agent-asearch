#!/usr/bin/env node
const { spawnSync } = require('child_process');
const fs = require('fs');
const os = require('os');
const path = require('path');

const root = path.join(__dirname, '..');
const pkg = require('../package.json');

console.log('=== release contract test ===');

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'asearch-contract-'));
const stateDir = path.join(tmpDir, 'state');
const binPath = path.join(tmpDir, process.platform === 'win32' ? 'asearch.exe' : 'asearch');

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

function parseLastJson(text) {
  const lastLine = String(text || '').trim().split('\n').filter(Boolean).pop();
  if (!lastLine) throw new Error('no output to parse');
  return JSON.parse(lastLine);
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

const tests = [
  { name: 'help is JSON', args: ['--help'], expect: 'json-stdout', status: 0 },
  { name: 'prompt is text', args: ['prompt'], expect: 'text-stdout', status: 0, minLength: 100 },
  { name: 'doctor is JSON', args: ['doctor'], expect: 'json-stdout', status: 0, allowOkFalse: true },
  { name: 'session list is JSON', args: ['session', 'list'], expect: 'json-stdout', status: 0 },
  { name: 'version is JSON', args: ['version'], expect: 'json-stdout', status: 0, version: pkg.version },
  { name: 'invalid args returns error JSON', args: ['--fakearg'], expect: 'json-stderr', status: 1, code: 'invalid_args' },
];

let failed = 0;

function fail(test, message, result) {
  console.log('FAIL:', test.name, message);
  if (result) {
    if (result.stdout) console.log('  stdout:', result.stdout.trim().slice(0, 200));
    if (result.stderr) console.log('  stderr:', result.stderr.trim().slice(0, 200));
  }
  failed++;
}

try {
  buildTestBinary();

  for (const test of tests) {
    const result = run('node', ['bin/asearch.js', ...test.args]);
    const actualStatus = result.status === null ? 1 : result.status;
    if (actualStatus !== test.status) {
      fail(test, `expected exit ${test.status}, got ${actualStatus}`, result);
      continue;
    }

    try {
      if (test.expect === 'json-stdout') {
        const payload = parseLastJson(result.stdout);
        if (payload.ok === false && !test.allowOkFalse) throw new Error('expected non-error JSON response');
        if (test.version && payload.version !== test.version) {
          throw new Error(`expected version ${test.version}, got ${payload.version}`);
        }
      } else if (test.expect === 'json-stderr') {
        const payload = parseLastJson(result.stderr);
        if (payload.ok !== false) throw new Error('expected ok=false');
        if (test.code && payload.code !== test.code) {
          throw new Error(`expected code ${test.code}, got ${payload.code}`);
        }
      } else if (test.expect === 'text-stdout') {
        if (result.stdout.trim().startsWith('{')) throw new Error('expected text, got JSON-looking output');
        if (result.stdout.length < test.minLength) {
          throw new Error(`expected at least ${test.minLength} chars, got ${result.stdout.length}`);
        }
      }
      console.log('PASS:', test.name);
    } catch (err) {
      fail(test, err.message, result);
    }
  }
} catch (err) {
  console.log('FAIL: setup', err.message);
  failed++;
} finally {
  fs.rmSync(tmpDir, { recursive: true, force: true });
}

console.log(`\n=== ${failed === 0 ? 'all tests passed' : failed + ' tests FAILED'} ===`);
process.exit(failed === 0 ? 0 : 1);
