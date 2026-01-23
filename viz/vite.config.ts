import { svelteTesting } from '@testing-library/svelte/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, type ProxyOptions } from 'vite';
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

// ideally a user/developer NEVER gets to the hardcoded defaults
const host = process.env.IMAGINE_API_SERVER_HOST || config.servers.api.host || "localhost";
const port = process.env.IMAGINE_API_SERVER_PORT || config.servers.api.port || 7770;

const apiServer: ProxyOptions = {
	target: `http://${host}:${port}`,
	secure: true,
	changeOrigin: true,
	ws: true
};

const viteProxy: Record<string, string | ProxyOptions> = {
	'/api': apiServer
};

// Expose runtime config (servers) as a global so the built frontend can read it without extra fetches.
(define as any).__RUNTIME_CONFIG__ = JSON.stringify({ version: pkg.version });

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
		globals: true, // This enables global APIs like 'expect', 'vi'
		environment: 'jsdom', // Default to jsdom for client-side tests
		setupFiles: ['./vitest-setup-client.ts'], // Apply this to all client tests
		// Common include/exclude patterns for all test projects
		include: [
			'src/tests/**/*.spec.ts',
			'src/**/*.spec.{js,ts}',
			'src/**/*.test.{js,ts}',
			'src/**/*.svelte.spec.{js,ts}',
			'src/**/*.svelte.test.{js,ts}'
		],
		exclude: [
			'src/lib/server/**',
			'src/lib/third-party/**',
			'src/lib/third-party/**/tests/**',
			'e2e/**', // Exclude Playwright E2E tests
		],
		projects: [
			{
				extends: './vite.config.ts', // Extend the base config
				plugins: [svelteTesting()],
				test: {
					name: 'client',
					clearMocks: true,
					// No need to redefine include/exclude or setupFiles here if they are in the top-level
				}
			},
			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					// keep server tests narrow to server-specific locations to avoid loading UI/component tests
					include: [
						'src/lib/server/**',
						'src/**/*.server.spec.{js,ts}',
						'src/**/*.server.test.{js,ts}'
					],
					exclude: [
						'src/tests/**',
						'src/**/*.svelte.spec.{js,ts}',
						'src/lib/third-party/**/tests/**',
						'e2e/**',
					]
				}
			}
		],
	},
	server: {
		port: config.servers.viz.port,
		cors: true,
		proxy: viteProxy,
		allowedHosts: ['imagine.local']
	},
	preview: {
		port: config.servers.viz.port,
		cors: true,
		proxy: viteProxy
	}
});
