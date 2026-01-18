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
 * @param {string} [preferredStyle]
 * @returns {Promise<string|null>} SVG text or null if not found
 */
async function fetchSvgForName (name, preferredWeight, preferredStyle) {
    const s = sanitizeNameForFile(name);
    const variants = [s, `ic_${s}`, s.replace(/^ic_/, '')];
    // Prioritize 400 (standard) first, then go to nearby weights.
    const defaultWeights = [400, 300, 200, 100, 500, 600, 700];
    const defaultStyles = ['sharp', 'rounded', 'outlined'];

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

    /** @type {string[]} */
    const styles = [];
    if (preferredStyle) {
        styles.push(preferredStyle);
    }
    for (const st of defaultStyles) {
        if (!styles.includes(st)) {
            styles.push(st);
        }
    }

    // 1. Try official Google Symbols repo first (direct access, high quality)
    // The main branch has symbols organized by name and style.
    // Note: The official repo often only has 400 weight as standard SVG export in these folders.
    for (const style of styles) {
        for (const v of variants) {
            const nm = v.replace(/^ic_/, '');
            const url = `https://raw.githubusercontent.com/google/material-design-icons/master/symbols/web/${nm}/materialsymbols${style}/${nm}_24px.svg`;

            // Only use this generic URL if we are looking for 400 weight (or it's our preferredWeight)
            // to avoid getting 400 when we specifically wanted 300/600 etc.
            if (!preferredWeight || preferredWeight === '400') {
                try {
                    const res = await fetch(url);
                    if (res && res.ok) {
                        const text = await res.text();
                        console.log(`Fetched from Google Symbols (400, ${style}):`, url);
                        return text;
                    }
                } catch (e) {
                    // continue
                }
            }
        }
    }

    // 2. Remote fallback: marella/material-symbols (EXCELLENT for multiple weights)
    // This repo specifically exports all weight variants as SVGs.
    for (const weight of weights) {
        for (const style of styles) {
            for (const v of variants) {
                const nm = v.replace(/^ic_/, '');
                const url = `https://raw.githubusercontent.com/marella/material-symbols/main/svg/${weight}/${style}/${nm}.svg`;
                try {
                    const res = await fetch(url);
                    if (res && res.ok) {
                        const text = await res.text();
                        console.log(`Fetched from Marella repo (${weight}, ${style}):`, url);
                        return text;
                    }
                } catch (e) {
                    // ignore
                }
            }
        }
    }

    // 3. Local fallback: @material-symbols/svg-{weight} packages when available.
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

    // 4. Final fallback: older google/material-design-icons repo
    const googleBase = 'https://raw.githubusercontent.com/google/material-design-icons/master';
    for (const cat of CATEGORIES) {
        for (const v of variants) {
            const filename = `ic_${v}_24px.svg`.replace(/ic_ic_/, 'ic_');
            const url = `${googleBase}/${cat}/svg/production/${filename}`;
            try {
                const res = await fetch(url);
                if (res && res.ok) {
                    const text = await res.text();
                    console.log('Fetched from old Google repo:', url);
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
 * Fetches the official list of Material Symbols codepoints.
 * @returns {Promise<Set<string>>}
 */
async function fetchCodepoints () {
    const url = 'https://raw.githubusercontent.com/google/material-design-icons/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.codepoints';
    console.log('Fetching codepoints from:', url);
    try {
        const res = await fetch(url);
        if (!res.ok) {
            throw new Error(`Failed to fetch codepoints: ${res.statusText}`);
        }

        const text = await res.text();
        const iconNames = new Set();
        text.split('\n').forEach(line => {
            const trimmed = line.trim();
            if (trimmed) {
                const [name] = trimmed.split(' ');
                if (name) iconNames.add(name);
            }
        });

        console.log(`Fetched ${iconNames.size} valid Material Symbols.`);
        return iconNames;
    } catch (e) {
        console.warn('Failed to fetch codepoints, skipping type generation update.', e);
        return new Set();
    }
}

/**
 * Generates the TypeScript definition for Material Symbols.
 * @param {Set<string>} iconNames 
 */
function generateTypeDefinition (iconNames) {
    if (iconNames.size === 0) return;

    const typeDefPath = join(SRC, 'lib/types/MaterialSymbol.ts');
    const sortedNames = Array.from(iconNames).sort();

    // Create a union type of all string literals
    const typeContent = `// Auto-generated by tools/icon-gen/generate-icons.js
// Do not edit manually.

export type MaterialSymbol = 
${sortedNames.map(n => `    | "${n}"`).join('\n')};
`;

    if (!existsSync(dirname(typeDefPath))) {
        mkdirSync(dirname(typeDefPath), { recursive: true });
    }

    writeFileSync(typeDefPath, typeContent, 'utf8');
    console.log('Wrote type definitions to', typeDefPath);
}

/**
 * Scan source files for `iconName` usages, fetch matching SVGs and
 * generate Svelte components under `src/lib/components/icons/generated/`.
 * @returns {Promise<void>}
 */
async function main () {
    // 1. Fetch codepoints and update types
    const validIcons = await fetchCodepoints();
    generateTypeDefinition(validIcons);

    // gather candidate icon names from source - look for iconName="..." or iconName={'...'}
    // Use forward-slash patterns so glob works reliably on Windows.
    const pattern = `${SRC.replace(/\\/g, '/')}/**/*.{svelte,ts,js}`;
    const files = sync(pattern, { nodir: true });

    console.log('Scanning files with pattern:', pattern);
    console.log('Found files:', files.length);
    if (files.length > 0) console.log('Example file:', files[0]);
    
    // Map of iconName -> { weights: Set, styles: Set }
    const names = new Map();

    const reNameDouble = /iconName\s*=\s*"([a-zA-Z0-9_\- ]+)"/g;
    const reNameSingle = /iconName\s*=\s*'([a-zA-Z0-9_\- ]+)'/g;
    const reNameExpr = /iconName\s*=\s*{\s*"([a-zA-Z0-9_\- ]+)"\s*}/g;
    const reNameObj = /iconName\s*:\s*"([a-zA-Z0-9_\- ]+)"/g; // object literal

    // Weight detection
    const reWeightDouble = /(?:iconWeight|weight)\s*=\s*"([0-9]{3})"/g;
    const reWeightSingle = /(?:iconWeight|weight)\s*=\s*'([0-9]{3})'/g;
    const reWeightExpr = /(?:iconWeight|weight)\s*=\s*{\s*([0-9]{3})\s*}/g;
    const reWeightObj = /(?:iconWeight|weight)\s*:\s*"([0-9]{3})"/g;

    // Style detection
    const reStyleDouble = /iconStyle\s*=\s*"([a-z]+)"/g;
    const reStyleSingle = /iconStyle\s*=\s*'([a-z]+)'/g;
    const reStyleExpr = /iconStyle\s*=\s*{\s*"([a-z]+)"\s*}/g;
    const reStyleObj = /iconStyle\s*:\s*"([a-z]+)"/g;

    for (const f of files) {
        if (f.includes('/generated/')) continue;
        const text = readFileSync(f, 'utf8');
        let m;
        
        const fileNames = new Set();
        [reNameDouble, reNameSingle, reNameExpr, reNameObj].forEach((r) => {
            r.lastIndex = 0;
            while ((m = r.exec(text)) !== null) {
                fileNames.add(m[1]);
            }
        });

        if (fileNames.size === 0) continue;

        const foundWeights = new Set();
        [reWeightDouble, reWeightSingle, reWeightExpr, reWeightObj].forEach((r) => {
            r.lastIndex = 0;
            while ((m = r.exec(text)) !== null) {
                foundWeights.add(m[1]);
            }
        });

        const foundStyles = new Set();
        [reStyleDouble, reStyleSingle, reStyleExpr, reStyleObj].forEach((r) => {
            r.lastIndex = 0;
            while ((m = r.exec(text)) !== null) {
                foundStyles.add(m[1]);
            }
        });

        for (const n of fileNames) {
            if (!names.has(n)) names.set(n, { weights: new Set(), styles: new Set() });
            const entry = names.get(n);
            foundWeights.forEach(w => entry.weights.add(w));
            foundStyles.forEach(s => entry.styles.add(s));
        }
    }

    if (!existsSync(OUT_DIR)) {
        mkdirSync(OUT_DIR, { recursive: true });
    }

    const generated = [];

    // debug: list discovered literal icon names
    console.log('Discovered icon names:', Array.from(names.keys()).sort());
    console.log('Total distinct icon names:', names.size);

    for (const [name, entry] of names.entries()) {
        const requiredWeights = new Set(['400', ...Array.from(entry.weights)]);
        const requiredStyles = new Set(['sharp', ...Array.from(entry.styles)]);
        
        // Filter out "filled" from styles as it's handled via the "fill" prop in our components
        requiredStyles.delete('filled');

        for (const style of requiredStyles) {
            const isDefaultStyle = style === 'sharp';
            const styleSuffix = isDefaultStyle ? '' : pascalCase(style);
            const compName = 'Icon' + pascalCase(name) + styleSuffix;
            const outFile = join(OUT_DIR, `${compName}.svelte`);

            console.log(`Processing ${name} (${style}) weights ${Array.from(requiredWeights).join(',')}`);
            
            /** @type {Record<string,string>} */
            const variants = {};
            /** @type {Record<string,string>} */
            const filledVariants = {};
            let finalViewBox = '0 0 24 24';

            for (const weight of Array.from(requiredWeights)) {
                // 1. Try normal version
                const rawSvg = await fetchSvgForName(name, weight, style);
                if (rawSvg) {
                    const optimized = optimize(rawSvg, { multipass: true }).data;
                    const vbMatch = optimized.match(/viewBox="([^\"]+)"/i);
                    if (vbMatch) finalViewBox = vbMatch[1];

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

                // 2. Try filled version
                const s = sanitizeNameForFile(name);
                const nm = s.replace(/^ic_/, '');
                
                // Try Marella repo for filled variant
                const filledUrl = `https://raw.githubusercontent.com/marella/material-symbols/main/svg/${weight}/${style}/${nm}-fill.svg`;
                try {
                    const res = await fetch(filledUrl);
                    if (res && res.ok) {
                        const filledSvg = await res.text();
                        console.log('Fetched filled variant:', filledUrl);
                        
                        const optimized = optimize(filledSvg, { multipass: true }).data;
                        const cleaned = optimized
                            .replace(/<\?xml[\s\S]*?\?>/g, '')
                            .replace(/\swidth="[^"]+"/g, '')
                            .replace(/\sheight="[^"]+"/g, '')
                            .replace(/\sxmlns(:\w+)?="[^"]+"/g, '')
                            .replace(/\sviewBox="[^"]+"/g, '')
                            .trim();
                        const innerMatch = cleaned.match(/<svg[^>]*>([\s\S]*?)<\/svg>/i);
                        const inner = innerMatch ? innerMatch[1].trim() : cleaned;
                        filledVariants[String(weight)] = inner.replace(/\/>/g, ' \/>');
                    }
                } catch (e) { }
            }

            if (Object.keys(variants).length === 0) {
                console.warn(`No variants available for ${name} (${style}); skipping component generation.`);
                continue;
            }

            const variantsJson = JSON.stringify(variants, null, 4);
            const filledJson = JSON.stringify(filledVariants, null, 4);

            const svelte = `
<script lang="ts">
    const variants: Record<string, string> = ${variantsJson};
    const filledVariants: Record<string, string> = ${filledJson};
    let { size = "1.5em", className = "", title = "${compName}", viewBox = "${finalViewBox}", weight = "400", fill = false, ...rest } = $props();
    
    const activeMap = $derived(fill ? filledVariants : variants);
    const inner = $derived(activeMap[String(weight)] || activeMap["400"] || Object.values(activeMap)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox={viewBox}
    fill="currentColor"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
    {...rest}
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
            generated.push({ name, compName, style, isDefault: isDefaultStyle });
        }
    }

    // write index file
    const indexPath = join(OUT_DIR, 'index.ts');
    const lines = generated.map(g => `export { default as ${g.compName} } from './${g.compName}.svelte';`);
    writeFileSync(indexPath, lines.join('\n') + '\n', 'utf8');
    console.log('Wrote index.ts with', generated.length, 'icons');
}

main().catch(e => { console.error(e); process.exit(1); });
