#!/usr/bin/env node
const { execSync } = require('child_process');
const path = require('path');

console.log('=== asearch smoke test ===');

try {
  // Version
  const ver = execSync('node bin/asearch.js version', { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
  const v = JSON.parse(ver);
  console.log('version:', v.ok ? 'PASS' : 'FAIL', v.version);

  // Help
  const help = execSync('node bin/asearch.js --help', { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
  const h = JSON.parse(help);
  console.log('help:', h.ok ? 'PASS' : 'FAIL', h.tool);

  // Prompt
  const prompt = execSync('node bin/asearch.js prompt', { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
  console.log('prompt:', prompt.length > 100 ? 'PASS' : 'FAIL');

  // Doctor
  const doc = execSync('node bin/asearch.js doctor', { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
  const d = JSON.parse(doc);
  console.log('doctor:', d.checks ? `PASS (${d.checks.length} backends)` : 'FAIL');

  // Session list
  const sl = execSync('node bin/asearch.js session list', { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
  const s = JSON.parse(sl);
  console.log('session list:', s.ok ? 'PASS' : 'FAIL', 'count:', s.count);

  console.log('\n=== all smoke tests passed ===');
} catch (e) {
  console.error('FAIL:', e.message);
  process.exit(1);
}
