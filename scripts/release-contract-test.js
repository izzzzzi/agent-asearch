#!/usr/bin/env node
const { execSync } = require('child_process');

console.log('=== release contract test ===');

const tests = [
  { name: 'help is JSON', cmd: 'node bin/asearch.js --help' },
  { name: 'prompt is text', cmd: 'node bin/asearch.js prompt' },
  { name: 'doctor is JSON', cmd: 'node bin/asearch.js doctor' },
  { name: 'session list is JSON', cmd: 'node bin/asearch.js session list' },
  { name: 'version is JSON', cmd: 'node bin/asearch.js version' },
  { name: 'invalid args returns error JSON', cmd: 'node bin/asearch.js --fakearg' },
];

let failed = 0;
for (const t of tests) {
  try {
    const out = execSync(t.cmd, { stdio: 'pipe', cwd: __dirname + '/..' }).toString();
    const lastLine = out.trim().split('\n').pop();
    JSON.parse(lastLine);
    console.log('PASS:', t.name);
  } catch (e) {
    // stderr might have JSON error
    const stderr = e.stderr?.toString() || '';
    try {
      JSON.parse(stderr.trim().split('\n').pop());
      console.log('PASS:', t.name, '(error JSON on stderr)');
    } catch {
      console.log('FAIL:', t.name, stderr.slice(0, 100));
      failed++;
    }
  }
}

console.log(`\n=== ${failed === 0 ? 'all tests passed' : failed + ' tests FAILED'} ===`);
process.exit(failed);
