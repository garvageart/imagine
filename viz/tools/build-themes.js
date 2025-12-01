import fs from 'fs';
import path from 'path';
import * as sass from 'sass';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const SCSS_DIR = path.resolve(__dirname, '../../src/lib/styles/scss');
const OUTPUT_DIR = path.resolve(__dirname, '../../static/themes');

if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

// Find all viz-*.scss files
const files = fs.readdirSync(SCSS_DIR).filter(file => file.startsWith('viz-') && file.endsWith('.scss') && !file.includes('mixins'));

console.log(`Found ${files.length} theme files.`);

files.forEach(file => {
    const inputPath = path.join(SCSS_DIR, file);
    const outputFilename = file.replace('.scss', '.css');
    const outputPath = path.join(OUTPUT_DIR, outputFilename);

    try {
        const result = sass.compile(inputPath, {
            style: 'compressed',
            loadPaths: [SCSS_DIR]
        });

        fs.writeFileSync(outputPath, result.css);
        console.log(`Compiled ${file} -> ${outputFilename}`);
    } catch (error) {
        console.error(`Error compiling ${file}:`, error);
        process.exit(1);
    }
});
