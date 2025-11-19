import { svelteTesting } from '@testing-library/svelte/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import fs from 'fs';
import { fileURLToPath } from 'url';
import devtoolsJson from "vite-plugin-devtools-json";

const file = fileURLToPath(new URL('package.json', import.meta.url));
const pkg = JSON.parse(fs.readFileSync(file, 'utf8'));
// In Docker we can pass IMAGINE_CONFIG_PATH (e.g., /app/imagine.json)
const configPath = process.env.IMAGINE_CONFIG_PATH || '../imagine.json';
let config: any;
try {
	config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
} catch (err) {
	if (process.env.IMAGINE_CONFIG_PATH) {
		// Fail fast for unexpected Docker build context issues.
		throw new Error(`IMAGINE_CONFIG_PATH set to '${process.env.IMAGINE_CONFIG_PATH}' but file not found: ${err}`);
	}

	config = {
		servers: {
			'api': { host: 'localhost', port: Number(process.env.API_PORT ?? 7770) },
			viz: { port: Number(process.env.VITE_VIZ_PORT ?? 7777) }
		}
	};
}
const define = {
	'__APP_VERSION__': JSON.stringify(pkg.version)
};

// Expose runtime config (servers) as a global so the built frontend can read it without extra fetches.
(define as any).__RUNTIME_CONFIG__ = JSON.stringify({ servers: config.servers, version: pkg.version });

if (process.env.NODE_ENV !== 'production') {
	(define as any).__servers = config.servers;
}

export default defineConfig({
	plugins: [devtoolsJson(), sveltekit()],
	define: define,
	build: {
		reportCompressedSize: false
	},
	test: {
		workspace: [
			{
				extends: './vite.config.ts',
				plugins: [svelteTesting()],
				test: {
					name: 'client',
					environment: 'jsdom',
					clearMocks: true,
					include: ['src/**/*.svelte.{test,spec.{js,ts}'],
					exclude: ['src/lib/server/**'],
					setupFiles: ['./vitest-setup-client.ts']
				}
			},
			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec.{js,ts}']
				}
			}
		],
	},
	server: {
		port: config.servers.viz.port,
		cors: true
	},
	preview: {
		port: config.servers.viz.port,
		cors: true
	}
});
