// Test teks halaman error SPA (test case TC-M00-027, docs/test-cases/M00-scaffold.md).
import { describe, expect, test } from 'vitest';
import { en, errorTitle } from './en.ts';

describe('errorTitle', () => {
	// TC-M00-027: route yang tidak dikenal menampilkan "Page not found"; error lain menampilkan
	// pesan umum. Keduanya bahasa Inggris dari en.ts, bukan teks bawaan SvelteKit.
	test('TC-M00-027 unknown SPA routes show an English error page', () => {
		expect(errorTitle(404)).toBe('Page not found');
		expect(errorTitle(500)).toBe('Something went wrong');
		expect(en.error.backToDashboard).toBe('Back to Dashboard');
	});
});
