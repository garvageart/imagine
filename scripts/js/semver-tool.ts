import fs from 'fs/promises';
import semver, { ReleaseType } from 'semver';

const versionFile = '../../version.txt';
const version = await fs.readFile(versionFile, 'utf8');
const releases = ["major", "premajor", "minor", "preminor", "patch", "prepatch", "prerelease", "release"] as const;
interface Args {
    command: string;
    args: string[];
}

const usage = `
Usage:
  node semver-tool.js <command> [options]

Commands:
  bump      Bump the version (patch, minor, major)
  get       Get the current version
  validate  Validate the version format
  compare   Compare two versions
  increment Increment the version by a specific amount
`;

async function bumpVersion(args: string[]) {
    const versionType = args[0] as ReleaseType; // e.g. "patch", "minor", "major"
    if (!versionType) {
        console.error('Please provide a version type');
        process.exit(1);
    }

    if (!releases.includes(versionType)) {
        console.error('Invalid version type');
        console.log("Available version types:", releases.join(", "));
        process.exit(1);
    }

    const newVersion = semver.inc(version.trim(), versionType);

    if (!newVersion) {
        console.error('Invalid version type');
        process.exit(1);
    }

    await fs.writeFile(versionFile, newVersion);
    console.log(`Version bumped to ${newVersion}`);
}

async function getVersion() {
    console.log(version.trim());
}

async function validateVersion() {
    if (semver.valid(version.trim())) {
        console.log('Version is valid');
    } else {
        console.error('Invalid version format');
    }
}

async function compareVersions(args: string[]) {
    const version1 = args[0];
    const version2 = args[1];

    if (!version1 || !version2) {
        console.error('Please provide two versions to compare');
        process.exit(1);
    }

    const result = semver.compare(version1, version2);
    console.log(result);
}

async function incrementVersion(args: string[]) {
    const increment = args[0] as ReleaseType; // e.g. "1.2.3"

    if (!increment) {
        console.error('Please provide an increment (e.g. 1.2.3)');
        process.exit(1);
    }

    const newVersion = semver.inc(version.trim(), increment);

    if (!newVersion) {
        console.error('Invalid increment');
        process.exit(1);
    }

    await fs.writeFile(versionFile, newVersion);
    console.log(`Version incremented to ${newVersion}`);
}

(async function main() {
    const args: Args = {
        command: process.argv[2],
        args: process.argv.slice(3),
    };

    if (!args.command) {
        console.log(usage);
        process.exit(0);
    }

    switch (args.command) {
        case 'bump':
            await bumpVersion(args.args);
            break;
        case 'get':
            await getVersion();
            break;
        case 'validate':
            await validateVersion();
            break;
        case 'compare':
            await compareVersions(args.args);
            break;
        case 'increment':
            await incrementVersion(args.args);
            break;
        default:
            console.error(`Unknown command: ${args.command}`);
            console.log(usage);
            process.exit(1);
    }
})();
