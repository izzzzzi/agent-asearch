#!/usr/bin/env node
const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');

const PKG = 'agent-asearch';
const VERSION = require('../package.json').version;

const stateDir = process.env.ASEARCH_STATE_DIR
  || path.join(process.env.HOME || process.env.USERPROFILE || '', '.asearch');
const binDir = path.join(stateDir, 'bin');

const platformMap = {
  darwin: 'darwin', win32: 'windows', linux: 'linux',
};
const archMap = {
  x64: 'amd64', arm64: 'arm64',
};

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, { timeout: 30000 }, (res) => {
      if (res.statusCode === 302 || res.statusCode === 301) {
        file.close();
        fs.unlinkSync(dest);
        return download(res.headers.location, dest).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        file.close();
        fs.unlinkSync(dest);
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      res.pipe(file);
      file.on('finish', () => { file.close(); resolve(); });
    }).on('error', (err) => {
      file.close();
      fs.unlinkSync(dest, () => {});
      reject(err);
    });
  });
}

async function install() {
  const plat = platformMap[process.platform];
  const arch = archMap[process.arch];
  if (!plat || !arch) {
    console.log(`[agent-asearch] unsupported platform: ${process.platform}/${process.arch}`);
    process.exit(0);
  }

  const binName = process.platform === 'win32' ? 'asearch.exe' : 'asearch';
  const archiveName = `asearch_${VERSION}_${plat}_${arch}.tar.gz`;
  const url = `https://github.com/izzzzzi/agent-asearch/releases/download/v${VERSION}/${archiveName}`;
  const dest = path.join(binDir, binName);

  if (fs.existsSync(dest)) {
    console.log(`[agent-asearch] binary already exists: ${dest}`);
    process.exit(0);
  }

  if (!fs.existsSync(binDir)) fs.mkdirSync(binDir, { recursive: true });

  const tmpDir = fs.mkdtempSync('asearch-');
  const tarball = path.join(tmpDir, archiveName);

  try {
    console.log(`[agent-asearch] downloading ${url}...`);
    await download(url, tarball);
    const { spawnSync } = require('child_process');
    spawnSync('tar', ['xzf', tarball, '-C', tmpDir], { stdio: 'pipe' });
    const extracted = path.join(tmpDir, binName);
    if (fs.existsSync(extracted)) {
      fs.renameSync(extracted, dest);
      fs.chmodSync(dest, 0o755);
      console.log(`[agent-asearch] installed to ${dest}`);
    }
  } catch (err) {
    console.log(`[agent-asearch] download failed: ${err.message}`);
    console.log('[agent-asearch] install binary manually from https://github.com/izzzzzi/agent-asearch/releases');
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

install();
