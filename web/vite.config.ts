// Konfigurasi Vite untuk SPA SvelteKit 3: build statis dengan fallback index.html (disajikan
// oleh binary Go), proxy dev /api ke server Go, dan Vitest dengan jsdom.
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';
import { apiProxy } from './vite.proxy.ts';

export default defineConfig({
	plugins: [
		sveltekit({
			// Mode runes Svelte 5 dipaksa untuk kode project; pustaka di node_modules memakai
			// pengaturannya sendiri.
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// SPA murni: semua path yang tidak dikenal dijawab index.html (docs/03-architecture.md D4).
			adapter: adapter({ fallback: 'index.html' })
		}),
		// Membersihkan DOM otomatis dan menyiapkan Svelte untuk Testing Library di Vitest.
		svelteTesting()
	],
	server: {
		// `make dev` memuat .env, jadi HTTP_ADDR yang sama dipakai server Go dan proxy ini.
		proxy: apiProxy(process.env.HTTP_ADDR)
	},
	// Di Vitest komponen Svelte harus diambil versi browser-nya (bukan versi SSR) agar bisa
	// dirender di jsdom.
	resolve: process.env.VITEST ? { conditions: ['browser'] } : undefined,
	test: {
		environment: 'jsdom',
		include: ['src/**/*.test.ts', 'vite.proxy.test.ts'],
		// Test tanpa assertion dianggap gagal supaya test kosong tidak lolos diam-diam.
		expect: { requireAssertions: true }
	}
});
