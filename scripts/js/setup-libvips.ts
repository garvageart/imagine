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
    throw new Error(`Command '${command} ${args.join(' ')}' failed with code ${result.status}`);
  }
}

async function installWindows() {
  log('Detecting Windows environment for MSYS2/pacman installation...');

  // Check if pacman is available (implies MSYS2 environment)
  try {
    execSync('which pacman', { stdio: 'ignore' });
    log('MSYS2/pacman detected. Proceeding with installation.');
  } catch (e) {
    error('pacman not found in PATH.');
    error('Please ensure MSYS2 is installed and this script is run from an MSYS2 terminal (MinGW 64-bit recommended).');
    process.exit(1);
  }

  // Cleanup old dormant installation directory if it exists
  const localAppData = process.env.LOCALAPPDATA;
  if (localAppData) {
    const oldVipsDir = path.join(localAppData, 'Programs', 'vips');
    try {
      await fs.access(oldVipsDir);
      log(`Cleaning up old dormant installation at ${oldVipsDir}...`);
      await fs.rm(oldVipsDir, { recursive: true, force: true });
    } catch { } // Ignore errors during cleanup
  }

  log('Updating MSYS2 packages...');
  try {
    runCommand('pacman', ['-Syu', '--noconfirm']);
  } catch (e: any) {
    log('pacman -Syu finished (may require terminal restart if core packages were updated).');
  }

  log(`Installing mingw-w64-x86_64-vips via pacman...`);
  try {
    runCommand('pacman', ['-S', '--noconfirm', 'mingw-w64-x86_64-vips']);
    success('libvips installed via pacman.');
  } catch (e) {
    error(`Failed to install libvips via pacman. Ensure you've run 'pacman -Syu' recently.`);
    throw e;
  }

  // Configure Environment Variables (Crucial for Go build)
  log('Configuring persistent environment variables for Go build...');

  let msys2Root;
  try {
    msys2Root = execSync('cygpath -m /', { encoding: 'utf-8' }).trim();
  } catch (e) {
    error("Could not determine MSYS2 root path using 'cygpath -m /'.");
    throw e;
  }

  const mingwBinPath = `${msys2Root}/mingw64/bin`;
  const mingwPkgConfigPath = `${msys2Root}/mingw64/lib/pkgconfig`;

  // Escape single quotes for PowerShell strings
  const esc = (s: string) => s.replace(/'/g, "''");

  const psScript = `
    $MingwBin = '${esc(mingwBinPath)}'
    $MingwPkgConfig = '${esc(mingwPkgConfigPath)}'
    
    # Update PATH (User level)
    $CurrentPath = [System.Environment]::GetEnvironmentVariable('Path', 'User')
    $pathEntries = $CurrentPath -split ';' | Where-Object { $_ -and ($_ -notlike "*vips\bin*") }
    if ($pathEntries -notcontains $MingwBin) { $pathEntries += $MingwBin }
    [System.Environment]::SetEnvironmentVariable('Path', ($pathEntries -join ';'), 'User')

    # Update PKG_CONFIG_PATH (User level)
    $CurrentPkgConfigPath = [System.Environment]::GetEnvironmentVariable('PKG_CONFIG_PATH', 'User')
    $pkgConfigEntries = if ($CurrentPkgConfigPath) { $CurrentPkgConfigPath -split ';' | Where-Object { $_ -and ($_ -notlike "*vips\lib\pkgconfig*") } } else { @() }
    if ($pkgConfigEntries -notcontains $MingwPkgConfig) { $pkgConfigEntries += $MingwPkgConfig }
    [System.Environment]::SetEnvironmentVariable('PKG_CONFIG_PATH', ($pkgConfigEntries -join ';'), 'User')
  `;

  // Execute PowerShell text directly via stdin
  const psResult = spawnSync('powershell', ['-NoProfile', '-NonInteractive', '-Command', '-'], {
    input: psScript,
    encoding: 'utf-8',
    stdio: ['pipe', 'inherit', 'inherit']
  });

  if (psResult.status === 0) {
    log('Environment variables updated successfully.');
  } else {
    error('Failed to update environment variables via PowerShell.');
  }

  // Update current process environment so verification works immediately
  process.env.PATH = `${mingwBinPath}${path.delimiter}${process.env.PATH}`;
  process.env.PKG_CONFIG_PATH = mingwPkgConfigPath;

  success('Environment configured. Restart your terminal/IDE for changes to fully apply.');
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

  try {
    execSync('which apt-get', { stdio: 'ignore' });
    log('Detected apt-based system. Installing libvips-dev...');
    runCommand('sudo', ['apt-get', 'update']);
    runCommand('sudo', ['apt-get', 'install', '-y', 'libvips-dev', 'pkg-config']);
    success('Installation complete via apt-get.');
    return;
  } catch (e) { } // Ignore errors and try next package manager

  try {
    execSync('which dnf', { stdio: 'ignore' });
    log('Detected dnf-based system. Installing vips-devel...');
    runCommand('sudo', ['dnf', 'install', '-y', 'vips-devel', 'pkg-config']);
    success('Installation complete via dnf.');
    return;
  } catch (e) { } // Ignore errors and try next package manager

  try {
    execSync('which pacman', { stdio: 'ignore' });
    log('Detected pacman-based system. Installing libvips...');
    runCommand('sudo', ['pacman', '-S', '--noconfirm', 'libvips', 'pkgconf']);
    success('Installation complete via pacman.');
    return;
  } catch (e) { } // Ignore errors and throw if no package manager found

  throw new Error('Unsupported Linux distribution. Please install libvips-dev and pkg-config manually.');
}

async function verify() {
  log('Verifying installation...');
  try {
    const vipsVersion = execSync('vips --version', { encoding: 'utf-8' }).trim();
    success(`Found binary: ${vipsVersion}`);
  } catch (e) {
    error('vips binary not found in PATH.');
  }

  try {
    const libs = execSync('pkg-config --cflags --libs vips', { encoding: 'utf-8' }).trim();
    if (libs) success('pkg-config verified.');
    else error('pkg-config returned empty output for vips.');
  } catch (e) {
    error('pkg-config check failed. Is PKG_CONFIG_PATH set?');
  }
}

async function main() {
  const version = await getRequiredVersion();
  log(`Target libvips version: ${version}`);

  const platform = os.platform();

  try {
    if (platform === 'win32') {
      await installWindows();
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