#!/usr/bin/env node
/**
 * generate-icons.js
 *
 * Scans the `viz/src` tree for `MaterialIcon` usages with literal `iconName`
 * values, attempts to fetch matching SVGs from the Google Material Icons
 * repository across common categories, runs SVGO optimization, and writes
 * Svelte components to `src/lib/components/icons/generated/`.
 */

import { readFileSync, existsSync, mkdirSync, writeFileSync } from 'fs';
import { resolve, join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { sync } from 'glob';
import { optimize } from 'svgo';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(__dirname, '../../');
const SRC = join(ROOT, 'src');
const OUT_DIR = join(SRC, 'lib/components/icons/generated');

const CATEGORIES = [
    'action', 'alert', 'av', 'communication', 'content', 'editor', 'file', 'hardware', 'image', 'maps', 'navigation', 'notification', 'places', 'social', 'toggle'
];

/**
 * Convert a string to PascalCase (e.g. "upload file" -> "UploadFile").
 * @param {string} s
 * @returns {string}
 */
function pascalCase (s) {
    return String(s)
        .replace(/[^a-zA-Z0-9]+/g, ' ')
        .trim()
        .split(/\s+/)
        .map(p => p[0].toUpperCase() + p.slice(1))
        .join('');
}

/**
 * Sanitize an icon name for use in file paths (lowercase, underscores).
 * @param {string} name
 * @returns {string}
 */
function sanitizeNameForFile (name) {
    return String(name).toLowerCase().replace(/\s+/g, '_').replace(/[^a-z0-9_]/g, '_');
}

/**
 * Attempt to obtain an SVG for the given icon name.
 * It will prefer a provided weight first, then fall back through common weights.
 * @param {string} name
 * @param {number|string} [preferredWeight]
 * @returns {Promise<string|null>} SVG text or null if not found
 */
async function fetchSvgForName (name, preferredWeight) {
    const s = sanitizeNameForFile(name);
    const variants = [s, `ic_${s}`, s.replace(/^ic_/, '')];
    const defaultWeights = [100, 200, 300, 400, 500, 600, 700];
    const styles = ['outlined', 'rounded', 'sharp'];

    // Build search order: preferredWeight first (if provided), then defaults
    /** @type {string[]} */
    const weights = [];
    if (preferredWeight) {
        const pw = String(preferredWeight).trim();
        if (pw && !weights.includes(pw)) {
            weights.push(pw);
        }
    }

    for (const w of defaultWeights) {
        if (!weights.includes(String(w))) {
            weights.push(String(w));
        }
    }

    // Prefer local @material-symbols/svg-{weight} packages when available.
    for (const weight of weights) {
        const pkgName = `@material-symbols/svg-${weight}`;
        const localPkg = resolve(ROOT, 'node_modules', pkgName);
        if (existsSync(localPkg)) {
            for (const style of styles) {
                for (const v of variants) {
                    const candidates = [
                        join(localPkg, style, `${v}.svg`),
                        join(localPkg, style, `${v}-fill.svg`),
                        join(localPkg, `${v}.svg`),
                        join(localPkg, `${v}-fill.svg`)
                    ];

                    for (const c of candidates) {
                        if (existsSync(c)) {
                            try {
                                const text = readFileSync(c, 'utf8');
                                console.log('Found local SVG in', pkgName, c);
                                return text;
                            } catch (e) {
                                // continue
                            }
                        }
                    }
                }
            }
        }
    }

    // Remote fallback: try marella/material-symbols svg layout first
    const materialSymbols = 'https://raw.githubusercontent.com/marella/material-symbols/main/svg';
    for (const weight of weights) {
        for (const style of styles) {
            for (const v of variants) {
                const nm = v.replace(/^ic_/, '');
                const patterns = [
                    `${materialSymbols}/${weight}/${style}/${nm}.svg`,
                    `${materialSymbols}/${weight}/${style}/${nm}-fill.svg`
                ];
                for (const url of patterns) {
                    try {
                        const res = await fetch(url);
                        if (res && res.ok) {
                            const text = await res.text();
                            console.log('Fetched', url);
                            return text;
                        }
                    } catch (e) {
                        // ignore
                    }
                }
            }
        }
    }

    // Final fallback: older google/material-design-icons repo
    const googleBase = 'https://raw.githubusercontent.com/google/material-design-icons/master';
    for (const cat of CATEGORIES) {
        for (const v of variants) {
            const filename = `ic_${v}_24px.svg`.replace(/ic_ic_/, 'ic_');
            const url = `${googleBase}/${cat}/svg/production/${filename}`;
            try {
                const res = await fetch(url);
                if (res && res.ok) {
                    const text = await res.text();
                    console.log('Fetched', url);
                    return text;
                }
            } catch (e) {
                // ignore
            }
        }
    }

    return null;
}

/**
 * @param {string} componentName
 * @param {string} svgContent
 */
function svelteTemplate (componentName, svgContent) {
    // remove width/height from root svg and ensure viewBox exists
    // strip xml prolog and size attrs, then extract inner svg content so we
    // can re-create the root <svg> with a single set of attributes (avoid
    // duplicate xmlns/viewBox attributes).
    const cleaned = svgContent
        .replace(/<\?xml[\s\S]*?\?>/g, '')
        .replace(/\swidth="[^"]+"/g, '')
        .replace(/\sheight="[^"]+"/g, '')
        .replace(/\sxmlns(:\w+)?="[^"]+"/g, '')
        .replace(/\sviewBox="[^"]+"/g, '')
        .trim();

    const innerMatch = cleaned.match(/<svg[^>]*>([\s\S]*?)<\/svg>/i);
    let inner = innerMatch ? innerMatch[1].trim() : cleaned;
    // ensure self-closing tags have a space before '/>' for consistent formatting
    inner = inner.replace(/\/\>/g, ' \/>');

    // Prepare the inner SVG content with consistent indentation
    /**
     * Indent each line of `str` by `n` spaces.
     * @param {string} str
     * @param {number} [n=2]
     * @returns {string}
     */
    const indent = (str, n = 2) => String(str).split('\n').map((l) => (l ? ''.repeat(n) + l : l)).join('\n');
    const innerIndented = indent(inner, 2);

    return `<script lang="ts">\n    let { size = "1.5em", className = "", title = "${componentName}", viewBox = "0 0 24 24" } = $props();\n</script>\n\n<svg\n    class={className}\n    width={size}\n    height={size}\n    {viewBox}\n    xmlns="http://www.w3.org/2000/svg"\n    aria-label={title}\n    focusable="false"\n>\n${innerIndented ? innerIndented + '\n' : ''}</svg>\n\n<style>\n    svg {\n        display: inline-block;\n        vertical-align: middle;\n    }\n</style>\n`;
}

/**
 * Scan source files for `iconName` usages, fetch matching SVGs and
 * generate Svelte components under `src/lib/components/icons/generated/`.
 * @returns {Promise<void>}
 */
async function main () {
    // gather candidate icon names from source - look for iconName="..." or iconName={'...'}
    // Use forward-slash patterns so glob works reliably on Windows.
    const pattern = `${SRC.replace(/\\/g, '/')}/**/*.{svelte,ts,js}`;
    const files = sync(pattern, { nodir: true });

    console.log('Scanning files with pattern:', pattern);
    console.log('Found files:', files.length);
    if (files.length > 0) console.log('Example file:', files[0]);
    // Map of iconName -> Set of weights (as strings). If no weight found, default to '400'.
    const names = new Map();

    const reNameDouble = /iconName\s*=\s*"([a-zA-Z0-9_\- ]+)"/g;
    const reNameSingle = /iconName\s*=\s*'([a-zA-Z0-9_\- ]+)'/g;
    const reNameExpr = /iconName\s*=\s*{\s*"([a-zA-Z0-9_\- ]+)"\s*}/g;
    const reNameObj = /iconName\s*:\s*"([a-zA-Z0-9_\- ]+)"/g; // object literal

    // Weight detection (common patterns). Capture 3-digit weights like 100..700
    const reWeightDouble = /(?:iconWeight|weight)\s*=\s*"([0-9]{3})"/g;
    const reWeightSingle = /(?:iconWeight|weight)\s*=\s*'([0-9]{3})'/g;
    const reWeightExpr = /(?:iconWeight|weight)\s*=\s*{\s*([0-9]{3})\s*}/g;
    const reWeightObj = /(?:iconWeight|weight)\s*:\s*"([0-9]{3})"/g;

    for (const f of files) {
        const text = readFileSync(f, 'utf8');
        let m;
        // collect names
        [reNameDouble, reNameSingle, reNameExpr, reNameObj].forEach((r) => {
            r.lastIndex = 0;
            while ((m = r.exec(text)) !== null) {
                const n = m[1];
                if (!names.has(n)) names.set(n, new Set());
            }
        });

        // collect weights nearby: a file may contain both name and weight; we conservatively
        // associate any found weight in the file with all discovered names in that file.
        const foundWeights = new Set();
        [reWeightDouble, reWeightSingle, reWeightExpr, reWeightObj].forEach((r) => {
            r.lastIndex = 0;
            while ((m = r.exec(text)) !== null) {
                foundWeights.add(m[1]);
            }
        });

        if (foundWeights.size > 0) {
            for (const n of names.keys()) {
                const set = names.get(n);
                foundWeights.forEach(w => set.add(w));
            }
        }
    }

    // Ensure at least a default weight for icons without detected weights
    for (const [n, set] of names.entries()) {
        if (set.size === 0) {
            set.add('400');
        }
    }

    if (!existsSync(OUT_DIR)) {
        mkdirSync(OUT_DIR, { recursive: true });
    }

    const generated = [];

    // debug: list discovered literal icon names
    console.log('Discovered icon names:', Array.from(names).sort());
    console.log('Total distinct icon names:', names.size);

    // Generate a single multi-weight component per icon name. For each name we try to
    // fetch SVGs for the detected weights and embed them in a `variants` map used
    // at runtime via a `weight` prop.
    for (const [name, weightSet] of names.entries()) {
        const compName = 'Icon' + pascalCase(name);
        const outFile = join(OUT_DIR, `${compName}.svelte`);

        let skip = false;
        if (existsSync(outFile)) {
            try {
                const content = readFileSync(outFile, 'utf8');
                // Regex to find keys in the variants object, e.g. "400":
                // We look for quoted digits followed by a colon.
                const existingWeights = new Set();
                const reKeys = /"([0-9]{3})"\s*:/g;
                let m;
                while ((m = reKeys.exec(content)) !== null) {
                    existingWeights.add(m[1]);
                }

                // Check if we have all required weights
                const required = Array.from(weightSet);
                const missing = required.filter(w => !existingWeights.has(w));

                if (missing.length === 0) {
                    // We still need to add to 'generated' so index.ts includes it.
                    console.log(`Skipping ${name} (up to date)`);
                    generated.push({ name, compName });
                    skip = true;
                }
            } catch (e) {
                // If read fails maybe just regenerate???
            }
        }

        if (skip) {
            continue;
        }

        console.log('Processing', name, 'weights', Array.from(weightSet).join(','));
        /** @type {Record<string,string>} */
        const variants = {};
        let finalViewBox = '0 0 24 24';

        for (const weight of Array.from(weightSet)) {
            const rawSvg = await fetchSvgForName(name, weight);
            if (!rawSvg) {
                console.warn('No SVG found for', name, 'weight', weight);
                continue;
            }

            const optimized = optimize(rawSvg, { multipass: true }).data;

            // extract viewBox if present (prefer the first one found)
            const vbMatch = optimized.match(/viewBox="([^\"]+)"/i);
            if (vbMatch) {
                finalViewBox = vbMatch[1];
            }

            // remove outer svg wrapper and any width/height/xmlns/viewBox attrs
            const cleaned = optimized
                .replace(/<\?xml[\s\S]*?\?>/g, '')
                .replace(/\swidth="[^"]+"/g, '')
                .replace(/\sheight="[^"]+"/g, '')
                .replace(/\sxmlns(:\w+)?="[^"]+"/g, '')
                .replace(/\sviewBox="[^"]+"/g, '')
                .trim();
            const innerMatch = cleaned.match(/<svg[^>]*>([\s\S]*?)<\/svg>/i);
            const inner = innerMatch ? innerMatch[1].trim() : cleaned;

            variants[String(weight)] = inner.replace(/\/>/g, ' \/>');
        }

        if (Object.keys(variants).length === 0) {
            console.warn('No variants available for', name, '; skipping component generation.');
            continue;
        }

        // Prepare the variants object as JSON so it becomes a JS object literal in the output file.
        const variantsJson = JSON.stringify(variants, null, 4);

        const svelte = `
<script lang="ts">
    const variants: Record<string, string> = ${variantsJson};
    let { size = "1.5em", className = "", title = "${compName}", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = variants[String(weight)] || variants["400"] || Object.values(variants)[0];
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox="${finalViewBox}"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
`;

        writeFileSync(outFile, svelte.trimStart(), 'utf8');
        console.log('Wrote', outFile);
        generated.push({ name, compName });
    }

    // write index file
    const indexPath = join(OUT_DIR, 'index.ts');
    const lines = generated.map(g => `export { default as ${g.compName} } from './${g.compName}.svelte';`);
    writeFileSync(indexPath, lines.join('\n') + '\n', 'utf8');
    console.log('Wrote index.ts with', generated.length, 'icons');
}

main().catch(e => { console.error(e); process.exit(1); });
