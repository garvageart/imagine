import fs from 'fs';

export function readConfig(): any {
    const fileData = fs.readFileSync('../../../config/imagine.json');
    return JSON.parse(fileData.toString());
}

export * from './utils'; 