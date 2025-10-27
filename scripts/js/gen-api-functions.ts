#!/usr/bin/env node
/**
 * Auto-generates TypeScript API wrapper functions from OpenAPI spec
 * Run: bun scripts/js/gen-api-functions.ts
 */

import { readFileSync, writeFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { load as loadYaml } from 'js-yaml';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const root = join(__dirname, '..', '..');
const specPath = join(root, 'api', 'openapi', 'openapi.yaml');
const outputPath = join(root, 'viz', 'src', 'lib', 'api', 'functions.gen.ts');

interface OpenAPISpec {
    paths: Record<string, Record<string, {
        operationId?: string;
        summary?: string;
        parameters?: Array<{
            in: string;
            name: string;
            required?: boolean;
            schema: any;
        }>;
        requestBody?: any;
    }>>;
}

function pascalCase(str: string): string {
    return str.replace(/(^|[_-])([a-z])/g, (_, __, c) => c.toUpperCase());
}

function camelCase(str: string): string {
    const pascal = pascalCase(str);
    return pascal.charAt(0).toLowerCase() + pascal.slice(1);
}

function generateFunction(
    path: string,
    method: string,
    operation: any
): string | null {
    const operationId = operation.operationId;
    if (!operationId) return null;

    const functionName = camelCase(operationId);
    const summary = operation.summary || '';

    // Extract path parameters
    const pathParams = path.match(/\{([^}]+)\}/g)?.map(p => p.slice(1, -1)) || [];

    // Extract query parameters
    const queryParams = operation.parameters
        ?.filter((p: any) => p.in === 'query')
        .map((p: any) => ({ name: p.name, required: p.required })) || [];

    const hasRequestBody = !!operation.requestBody;
    const httpMethod = method.toUpperCase();

    // Build parameter type
    let paramType = '';
    let hasParams = pathParams.length > 0 || queryParams.length > 0 || hasRequestBody;

    if (hasParams) {
        const paramFields: string[] = [];

        // Path parameters
        pathParams.forEach(param => {
            paramFields.push(`    ${param}: string;`);
        });

        // Query parameters
        queryParams.forEach((param: any) => {
            const optional = param.required ? '' : '?';
            paramFields.push(`    ${param.name}${optional}: any;`);
        });

        // Request body
        if (hasRequestBody) {
            paramFields.push(`    body: any;`);
        }

        paramType = `{\n${paramFields.join('\n')}\n}`;
    }

    // Build function signature
    const paramString = hasParams ? `params: ${paramType}` : '';

    // Build function body
    let functionBody = '';

    if (pathParams.length > 0 || queryParams.length > 0) {
        const pathObj = pathParams.length > 0
            ? `path: { ${pathParams.map(p => `${p}: params.${p}`).join(', ')} }`
            : '';
        const queryObj = queryParams.length > 0
            ? `query: { ${queryParams.map((p: any) => `${p.name}: params.${p.name}`).join(', ')} }`
            : '';

        const paramsObj = [pathObj, queryObj].filter(Boolean).join(',\n            ');

        if (hasRequestBody) {
            functionBody = `    return apiClient.${httpMethod}("${path}", {
        params: {
            ${paramsObj}
        },
        body: params.body,
    });`;
        } else {
            functionBody = `    return apiClient.${httpMethod}("${path}", {
        params: {
            ${paramsObj}
        },
    });`;
        }
    } else if (hasRequestBody) {
        functionBody = `    return apiClient.${httpMethod}("${path}", {
        body: params.body,
    });`;
    } else {
        functionBody = `    return apiClient.${httpMethod}("${path}");`;
    }

    return `/**
 * ${summary}
 */
export async function ${functionName}(${paramString}) {
${functionBody}
}`;
}


console.log('Reading OpenAPI spec...');
const specContent = readFileSync(specPath, 'utf-8');
const spec = loadYaml(specContent) as OpenAPISpec;

console.log('Generating API functions...');
const functions: string[] = [];

for (const [path, methods] of Object.entries(spec.paths)) {
    for (const [method, operation] of Object.entries(methods)) {
        if (['get', 'post', 'put', 'delete', 'patch'].includes(method)) {
            const func = generateFunction(path, method, operation);
            if (func) {
                functions.push(func);
            }
        }
    }
}

const output = `/**
 * Auto-generated API functions from OpenAPI spec.
 * DO NOT EDIT MANUALLY - run 'bun scripts/js/gen-api-functions.ts' to regenerate.
 */
import { apiClient } from "./client";

${functions.join('\n\n')}
`;

writeFileSync(outputPath, output, 'utf-8');
console.log(`✅ Generated ${functions.length} functions -> ${outputPath}`);
