// Test komponen app shell: teks yang tampil harus bahasa Inggris dan berasal dari en.ts
// (test case TC-M00-018, docs/test-cases/M00-scaffold.md).
import { cleanup, render, screen, within } from '@testing-library/svelte';
import { afterEach, describe, expect, test } from 'vitest';
import { en } from '#lib/i18n/en.ts';
import AppShell from './AppShell.svelte';

// Setiap test merender shell baru; DOM dibersihkan supaya query tidak menemukan sisa test lain.
afterEach(() => cleanup());

describe('AppShell', () => {
	// TC-M00-018: top bar, kotak pencarian dan semua grup sidebar memakai teks dari en.ts; tidak ada
	// lonceng notifikasi (Tahap 3) dan tidak ada label bahasa Indonesia dari mockup.
	test('TC-M00-018 app shell shows the English texts from en.ts', () => {
		render(AppShell, { props: { pathname: '/' } });

		// Logo dan nama aplikasi menautkan ke dashboard.
		const brand = screen.getByRole('link', { name: en.app.name });
		expect(brand.getAttribute('href')).toBe('/');
		expect(en.app.name).toBe('PFD · PFMEA · CP System');

		// Pencarian global tampil tetapi nonaktif sampai M11.
		const search = screen.getByPlaceholderText(en.topbar.searchPlaceholder);
		expect(en.topbar.searchPlaceholder).toBe('Search packages, steps, characteristics');
		expect((search as HTMLInputElement).disabled).toBe(true);

		// Grup dan menu sidebar sesuai docs/08-screens.md §1.
		const sidebar = within(screen.getByRole('navigation', { name: en.nav.label }));
		for (const title of [en.nav.monitoring, en.nav.packages, en.nav.template]) {
			expect(sidebar.getByText(title)).toBeTruthy();
		}
		expect([en.nav.monitoring, en.nav.packages, en.nav.template]).toEqual([
			'MONITORING',
			'PACKAGES',
			'TEMPLATE'
		]);
		const links: [string, string, string][] = [
			[en.nav.dashboard, '/', 'Dashboard'],
			[en.nav.findings, '/findings', 'Consistency check'],
			[en.nav.allPackages, '/packages', 'All packages'],
			[en.nav.templateGeneral, '/template-general', 'Template General']
		];
		for (const [label, href, english] of links) {
			expect(label).toBe(english);
			expect(sidebar.getByRole('link', { name: label }).getAttribute('href')).toBe(href);
		}

		// Lonceng notifikasi adalah fitur Tahap 3 dan tidak boleh dirender.
		expect(screen.queryByRole('button', { name: /notification/i })).toBeNull();
		// Label mockup berbahasa Indonesia tidak boleh masuk ke aplikasi.
		const text = document.body.textContent ?? '';
		for (const indonesian of ['Cek konsistensi', 'Paket', 'Cari paket', 'Sistem PFD']) {
			expect(text).not.toContain(indonesian);
		}
	});
});
