#!/usr/bin/env node
const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');
const os = require('os');

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

function removeIfExists(file) {
  try {
    fs.unlinkSync(file);
  } catch (err) {
    if (err.code !== 'ENOENT') throw err;
  }
}

function download(url, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) return reject(new Error('too many redirects'));

    const file = fs.createWriteStream(dest);
    const request = https.get(url, { timeout: 30000 }, (res) => {
      const redirectCodes = new Set([301, 302, 303, 307, 308]);
      if (redirectCodes.has(res.statusCode)) {
        file.close();
        removeIfExists(dest);
        const location = res.headers.location;
        if (!location) return reject(new Error(`HTTP ${res.statusCode} without Location`));
        const nextUrl = new URL(location, url).toString();
        res.resume();
        return download(nextUrl, dest, redirects + 1).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        file.close();
        removeIfExists(dest);
        res.resume();
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      res.pipe(file);
      file.on('finish', () => { file.close(resolve); });
    });

    request.on('timeout', () => {
      request.destroy(new Error('download timed out'));
    });
    request.on('error', (err) => {
      file.close();
      removeIfExists(dest);
      reject(err);
    });
  });
}

function installedVersion(binPath) {
  if (!fs.existsSync(binPath)) return null;
  const result = spawnSync(binPath, ['version'], { encoding: 'utf8', stdio: 'pipe' });
  if (result.status !== 0) return null;
  try {
    return JSON.parse(result.stdout).version || null;
  } catch (_) {
    return null;
  }
}

function extractArchive(tarball, tmpDir, binName) {
  const result = spawnSync('tar', ['xzf', tarball, '-C', tmpDir], { encoding: 'utf8', stdio: 'pipe' });
  if (result.status !== 0) {
    const detail = result.stderr || result.stdout || `exit ${result.status}`;
    throw new Error(`tar extraction failed: ${detail.trim()}`);
  }

  const extracted = path.join(tmpDir, binName);
  if (!fs.existsSync(extracted)) {
    throw new Error(`archive did not contain expected binary ${binName}`);
  }
  return extracted;
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

  const currentVersion = installedVersion(dest);
  if (currentVersion === VERSION) {
    console.log(`[agent-asearch] binary already installed: ${dest}`);
    process.exit(0);
  }
  if (currentVersion) {
    console.log(`[agent-asearch] replacing ${dest} (${currentVersion} -> ${VERSION})`);
  }

  fs.mkdirSync(binDir, { recursive: true });

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'asearch-'));
  const tarball = path.join(tmpDir, archiveName);

  try {
    console.log(`[agent-asearch] downloading ${url}...`);
    await download(url, tarball);
    const extracted = extractArchive(tarball, tmpDir, binName);
    removeIfExists(dest);
    fs.renameSync(extracted, dest);
    fs.chmodSync(dest, 0o755);
    console.log(`[agent-asearch] installed to ${dest}`);
  } catch (err) {
    console.error(`[agent-asearch] install failed: ${err.message}`);
    console.error('[agent-asearch] install binary manually from https://github.com/izzzzzi/agent-asearch/releases');
    process.exitCode = 1;
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

install();
