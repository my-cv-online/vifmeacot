// Konfigurasi ESLint (flat config) untuk TypeScript dan Svelte di web/; dijalankan oleh
// `make lint` lewat `npm run lint`.
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import { defineConfig } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';

export default defineConfig(
	// Hasil build, cache SvelteKit dan laporan test bukan kode sumber.
	{ ignores: ['build/', '.svelte-kit/', 'node_modules/', 'test-results/', 'playwright-report/'] },
	js.configs.recommended,
	ts.configs.recommended,
	svelte.configs.recommended,
	{
		languageOptions: { globals: { ...globals.browser, ...globals.node } },
		rules: {
			// TypeScript sudah memeriksa variabel yang tidak dikenal; no-undef memberi alarm palsu
			// pada tipe global (saran typescript-eslint).
			'no-undef': 'off'
		}
	},
	{
		// File Svelte di-parse dengan parser TypeScript supaya blok <script lang="ts"> dipahami.
		files: ['**/*.svelte', '**/*.svelte.ts'],
		languageOptions: {
			parserOptions: {
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser
			}
		}
	}
);
