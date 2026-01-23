import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PROJECT_ROOT = path.resolve(__dirname, '../../'); // scripts/js -> scripts -> root

const DOMAIN = 'imagine.local';
const TARGET_IP = '127.0.0.1';
const CADDYFILE_NAME = 'Caddyfile';

// ANSI colors for prettier output
const colors = {
    reset: "\x1b[0m",
    green: "\x1b[32m",
    yellow: "\x1b[33m",
    red: "\x1b[31m",
    cyan: "\x1b[36m",
    bold: "\x1b[1m"
};

const log = {
    info: (msg: string) => console.log(`${colors.cyan}ℹ ${msg}${colors.reset}`),
    success: (msg: string) => console.log(`${colors.green}✔ ${msg}${colors.reset}`),
    warn: (msg: string) => console.log(`${colors.yellow}⚠ ${msg}${colors.reset}`),
    error: (msg: string) => console.log(`${colors.red}✖ ${msg}${colors.reset}`),
    step: (msg: string) => console.log(`\n${colors.bold}${msg}${colors.reset}`),
};

function getHostsFilePath(): string {
    const platform = os.platform();
    if (platform === 'win32') {
        return 'C:\\Windows\\System32\\drivers\\etc\\hosts';
    }
    return '/etc/hosts';
}

function checkCaddyInstalled(): boolean {
    try {
        const cmd = os.platform() === 'win32' ? 'where caddy' : 'which caddy';
        execSync(cmd, { stdio: 'ignore' });
        return true;
    } catch {
        return false;
    }
}

function ensureCaddyfile() {
    const caddyfilePath = path.join(PROJECT_ROOT, CADDYFILE_NAME);
    const caddyContent = `${DOMAIN} {
	tls internal
	reverse_proxy localhost:7777
}
`;

    if (fs.existsSync(caddyfilePath)) {
        log.info(`Caddyfile already exists at ${caddyfilePath}`);
        // Optional: Check content? simpler to just leave it if it exists.
    } else {
        try {
            fs.writeFileSync(caddyfilePath, caddyContent);
            log.success(`Created Caddyfile at ${caddyfilePath}`);
        } catch (e) {
            log.error(`Failed to create Caddyfile: ${(e as Error).message}`);
            process.exit(1);
        }
    }
}

function updateHostsFile() {
    const hostsPath = getHostsFilePath();
    
    log.info(`Checking hosts file at: ${hostsPath}`);

    try {
        const content = fs.readFileSync(hostsPath, 'utf8');
        const lines = content.split(/\r?\n/);
        
        // Check if entry already exists (naively)
        const exists = lines.some(line => {
            const trimmed = line.trim();
            return !trimmed.startsWith('#') && trimmed.includes(DOMAIN) && trimmed.includes(TARGET_IP);
        });

        if (exists) {
            log.success(`Host entry '${DOMAIN}' already exists.`);
            return;
        }

        // Prepare new content
        const newEntry = `${os.EOL}${TARGET_IP} ${DOMAIN} # Added by Imagine setup script${os.EOL}`;
        
        try {
            fs.appendFileSync(hostsPath, newEntry);
            log.success(`Added '${DOMAIN}' to hosts file.`);
        } catch (err: any) {
            if (err.code === 'EACCES' || err.code === 'EPERM') {
                log.error("Permission denied writing to hosts file.");
                log.warn("Please run this script with Administrator privileges (sudo or Run as Administrator).");
                if (os.platform() !== 'win32') {
                    console.log(`\n  sudo npx tsx scripts/js/setup-https.ts\n`);
                }
                process.exit(1);
            } else {
                throw err;
            }
        }

    } catch (e) {
        log.error(`Failed to update hosts file: ${(e as Error).message}`);
        process.exit(1);
    }
}

function trustCertificate() {
    log.info("Attempting to trust Caddy root certificate...");
    try {
        // This command requires admin privileges to write to the trust store
        execSync('caddy trust', { stdio: 'inherit', cwd: PROJECT_ROOT });
        log.success("Certificate trust command executed.");
    } catch (e) {
        log.warn("Failed to automatically trust the certificate.");
        log.info("This is expected if you are not running as Administrator.");
        log.info("To fix 'Not Secure' errors, run this command manually as Administrator:");
        console.log(`\n  ${colors.green}caddy trust${colors.reset}\n`);
    }
}

async function main() {
    console.log(`${colors.bold}Imagine HTTPS Setup (Caddy)${colors.reset}`);
    console.log("==================================");

    // 1. Check Caddy
    log.step("1. Checking for Caddy...");
    if (checkCaddyInstalled()) {
        log.success("Caddy is installed.");
    } else {
        log.warn("Caddy is NOT found in your PATH.");
        console.log(`  Please install Caddy manually:`);
        if (os.platform() === 'win32') {
            console.log(`  > choco install caddy`);
        } else if (os.platform() === 'darwin') {
            console.log(`  > brew install caddy`);
        } else {
            console.log(`  > https://caddyserver.com/docs/install`);
        }
        // We don't exit here, we can still setup the config files.
    }

    // 2. Setup Caddyfile
    log.step("2. Configuring Caddyfile...");
    ensureCaddyfile();

    // 3. Update Hosts
    log.step("3. Updating hosts file...");
    updateHostsFile();

    // 4. Trust Certs
    log.step("4. Trusting Certificate...");
    trustCertificate();

    // 5. Summary
    log.step("Setup Complete!");
    console.log(`\nTo start the HTTPS proxy:`)
    console.log(`  ${colors.green}caddy run${colors.reset}`);
    console.log(`\nThen access the app at:`);
    console.log(`  ${colors.cyan}https://${DOMAIN}${colors.reset}`);
}

main().catch(console.error);
