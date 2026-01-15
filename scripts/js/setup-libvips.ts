import fs from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import { execSync, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const PROJECT_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const VERSION_FILE = path.join(PROJECT_ROOT, '.libvips-version');

const log = (msg: string) => console.log(`\x1b[36m[setup-libvips]\x1b[0m ${msg}`);
const error = (msg: string) => console.error(`\x1b[31m[Error]\x1b[0m ${msg}`);
const success = (msg: string) => console.log(`\x1b[32m[Success]\x1b[0m ${msg}`);

async function getRequiredVersion(): Promise<string> {
  try {
    const content = await fs.readFile(VERSION_FILE, 'utf-8');
    return content.trim();
  } catch (e) {
    error(`Could not read .libvips-version at ${VERSION_FILE}`);
    process.exit(1);
  }
}

function runCommand(command: string, args: string[], opts: any = {}) {
  const result = spawnSync(command, args, { stdio: 'inherit', shell: true, ...opts });
  if (result.error) {
    throw result.error;
  }

  if (result.status !== 0) {
    throw new Error(`Command failed with code ${result.status}`);
  }
}

function getEnvVar(key: string): string | undefined {
  return process.env[key];
}


async function installWindows(version: string) {
  log('Detecting Windows environment...');

  const localAppData = getEnvVar('LOCALAPPDATA');
  if (!localAppData) {
    throw new Error('LOCALAPPDATA environment variable is missing.');
  }

  const installDir = path.join(localAppData, 'Programs');
  const vipsHome = path.join(installDir, 'vips');
  const vipsBin = path.join(vipsHome, 'bin');
  const pkgConfigPath = path.join(vipsHome, 'lib', 'pkgconfig');

  // Check existing installation
  let currentVersion = '';
  try {
    const output = execSync(`"${path.join(vipsBin, 'vips.exe')}" --version`, { encoding: 'utf-8' }).trim();
    const match = output.match(/vips-(\d+\.\d+\.\d+)/);
    if (match) currentVersion = match[1];
  } catch (e) {
    // Not installed or error
  }

  if (currentVersion === version) {
    log(`libvips ${version} is already installed at ${vipsHome}`);
  } else {
    if (currentVersion) {
      log(`Found libvips ${currentVersion}, upgrading to ${version}...`);
    } else {
      log(`libvips not found. Installing ${version}...`);
    }

    // Prepare URL and paths
    const url = `https://github.com/libvips/build-win64-mxe/releases/download/v${version}/vips-dev-w64-all-${version}.zip`;
    const zipPath = path.join(installDir, 'vips.zip');

    // Create Programs dir if needed
    await fs.mkdir(installDir, { recursive: true });

    // Download
    log(`Downloading ${url}...`);
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to download libvips: ${response.statusText}`);
    }

    const arrayBuffer = await response.arrayBuffer();
    await fs.writeFile(zipPath, Buffer.from(arrayBuffer));

    // Remove old install
    log('Removing old installation...');
    try {
      await fs.rm(vipsHome, { recursive: true, force: true });
    } catch (e) { } // ignore if doesn't exist

    // Extract using PowerShell (built-in, reliable on Windows)
    log('Extracting...');
    runCommand('powershell', [
      '-Command',
      `Expand-Archive -Path "${zipPath}" -DestinationPath "${installDir}" -Force`
    ]);

    // Cleanup Zip
    await fs.rm(zipPath);

    // Rename extracted folder (it usually comes as vips-dev-8.18.0)
    // We need to find what it was extracted as.
    const files = await fs.readdir(installDir);
    const extractedFolder = files.find(f => f.startsWith('vips-dev-'));
    if (extractedFolder) {
      await fs.rename(path.join(installDir, extractedFolder), vipsHome);
    } else {
      // It might have extracted directly if the zip structure changed, but usually it has a root folder.
      // If 'vips' folder exists (from a direct extract), we are good, otherwise error.
      if (!(await fs.stat(vipsHome).catch(() => false))) {
        throw new Error('Could not identify extracted folder.');
      }
    }

    success('Installation files placed.');
  }

  // Set Persistent Environment Variables (User Level)
  log('Configuring persistent environment variables...');

  // Create a temporary PowerShell script to update environment variables cleanly
  const psScriptPath = path.join(os.tmpdir(), 'update_env.ps1');
  const psScriptContent = `
    $VipsBin = "${vipsBin}"
    $PkgConfigPath = "${pkgConfigPath}"
    
    # Update PATH
    $CurrentPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
    if ($CurrentPath -notlike "*$VipsBin*") {
        [System.Environment]::SetEnvironmentVariable('Path', "$CurrentPath;$VipsBin", 'User')
        Write-Host "Added to PATH."
    } else {
        Write-Host "Already in PATH."
    }

    # Update PKG_CONFIG_PATH
    [System.Environment]::SetEnvironmentVariable('PKG_CONFIG_PATH', $PkgConfigPath, 'User')
    Write-Host "Set PKG_CONFIG_PATH."
  `;

  await fs.writeFile(psScriptPath, psScriptContent);

  try {
    runCommand('powershell', ['-ExecutionPolicy', 'Bypass', '-File', psScriptPath]);
  } finally {
    await fs.unlink(psScriptPath);
  }

  // Update current process environment so verification works immediately
  process.env.PATH = `${vipsBin}${path.delimiter}${process.env.PATH}`;
  process.env.PKG_CONFIG_PATH = pkgConfigPath;

  log('Environment configured.');
}

async function installMacOS() {
  log('Detecting macOS environment...');
  try {
    execSync('brew --version', { stdio: 'ignore' });
  } catch (e) {
    throw new Error('Homebrew is not installed. Please install Homebrew first: https://brew.sh/');
  }

  log('Installing vips and pkg-config via Homebrew...');
  runCommand('brew', ['install', 'vips', 'pkg-config']);
  success('Installation complete via Homebrew.');
}

async function installLinux() {
  log('Detecting Linux environment...');

  // Naive check for apt-get (Debian/Ubuntu)
  try {
    execSync('which apt-get', { stdio: 'ignore' });
    log('Detected apt-based system. Installing libvips-dev...');
    runCommand('sudo', ['apt-get', 'update']);
    runCommand('sudo', ['apt-get', 'install', '-y', 'libvips-dev', 'pkg-config']);
    success('Installation complete via apt-get.');
    return;
  } catch (e) { }

  // Naive check for dnf (Fedora)
  try {
    execSync('which dnf', { stdio: 'ignore' });
    log('Detected dnf-based system. Installing vips-devel...');
    runCommand('sudo', ['dnf', 'install', '-y', 'vips-devel', 'pkg-config']);
    success('Installation complete via dnf.');
    return;
  } catch (e) { }

  // Naive check for pacman (Arch)
  try {
    execSync('which pacman', { stdio: 'ignore' });
    log('Detected pacman-based system. Installing libvips...');
    runCommand('sudo', ['pacman', '-S', '--noconfirm', 'libvips', 'pkgconf']);
    success('Installation complete via pacman.');
    return;
  } catch (e) { }

  throw new Error('Unsupported Linux distribution. Please install libvips-dev and pkg-config manually.');
}

async function verify() {
  log('Verifying installation...');

  // Verify vips binary
  try {
    const vipsVersion = execSync('vips --version', { encoding: 'utf-8' }).trim();
    success(`Found binary: ${vipsVersion}`);
  } catch (e) {
    error('vips binary not found in PATH (you might need to restart terminal on Windows).');
  }

  // Verify pkg-config
  try {
    const libs = execSync('pkg-config --cflags --libs vips', { encoding: 'utf-8' }).trim();
    if (libs) {
      success('pkg-config verified.');
    } else {
      error('pkg-config returned empty output for vips.');
    }
  } catch (e) {
    error('pkg-config check failed. Is PKG_CONFIG_PATH set?');
  }

  // Verify RAW support
  try {
    const modules = execSync('vips -l foreign', { encoding: 'utf-8' });
    if (modules.includes('dcrawload') || modules.includes('VipsForeignLoadDcRaw')) {
      success('Native RAW support detected (dcrawload).');
    } else if (modules.includes('VipsForeignLoadRaw')) {
      log('Generic Raw support detected (might be raw data, not camera RAW).');
    } else {
      error('RAW support NOT detected.');
    }
  } catch (e) {
    error('Failed to check modules.');
  }
}


async function main() {
  const version = await getRequiredVersion();
  log(`Target libvips version: ${version}`);

  const platform = os.platform();

  try {
    if (platform === 'win32') {
      await installWindows(version);
    } else if (platform === 'darwin') {
      await installMacOS();
    } else if (platform === 'linux') {
      await installLinux();
    } else {
      throw new Error(`Unsupported platform: ${platform}`);
    }

    await verify();

  } catch (e: any) {
    error(e.message || e);
    process.exit(1);
  }
}

main();
