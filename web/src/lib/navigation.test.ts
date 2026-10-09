// Test navigasi app shell: menu aktif dan breadcrumb mengikuti path halaman
// (test case TC-M00-019, docs/test-cases/M00-scaffold.md).
import { cleanup, render, screen, within } from '@testing-library/svelte';
import { afterEach, describe, expect, test } from 'vitest';
import AppShell from '#lib/components/AppShell.svelte';
import { en } from '#lib/i18n/en.ts';
import { breadcrumbFor, isActive } from './navigation.ts';

// DOM dibersihkan setelah setiap test yang merender komponen.
afterEach(() => cleanup());

describe('navigation', () => {
	// TC-M00-019: breadcrumb setiap halaman placeholder M0 diawali Dashboard; menu aktif ditandai
	// aria-current="page" hanya pada halaman yang sedang dibuka.
	test('TC-M00-019 active navigation item and breadcrumb follow the path', () => {
		// labels mengubah breadcrumb menjadi teks "A / B" supaya mudah dibandingkan.
		const labels = (path: string) =>
			breadcrumbFor(path)
				.map((c) => c.label)
				.join(' / ');
		expect(labels('/')).toBe('Dashboard');
		expect(labels('/findings')).toBe('Dashboard / Consistency check');
		expect(labels('/packages')).toBe('Dashboard / Packages');
		expect(labels('/template-general')).toBe('Dashboard / Template General');
		// Halaman detail paket baru ada di M4; sampai saat itu breadcrumb berhenti di Packages.
		expect(labels('/packages/PS-07')).toBe('Dashboard / Packages');

		// Dashboard hanya aktif di "/" persis; menu lain juga aktif untuk sub-path-nya.
		expect(isActive('/', '/')).toBe(true);
		expect(isActive('/findings', '/')).toBe(false);
		expect(isActive('/packages/PS-07', '/packages')).toBe(true);
		expect(isActive('/packages-old', '/packages')).toBe(false);

		render(AppShell, { props: { pathname: '/findings' } });
		const sidebar = within(screen.getByRole('navigation', { name: en.nav.label }));
		const current = sidebar
			.getAllByRole('link')
			.filter((a) => a.getAttribute('aria-current') === 'page')
			.map((a) => a.textContent?.trim());
		expect(current).toEqual([en.nav.findings]);

		// Breadcrumb tampil di halaman; item terakhir adalah halaman saat ini (bukan tautan).
		const crumbs = within(screen.getByRole('navigation', { name: en.breadcrumb.label }));
		expect(crumbs.getByRole('link', { name: en.pages.dashboard }).getAttribute('href')).toBe('/');
		const last = crumbs.getByText(en.pages.findings);
		expect(last.getAttribute('aria-current')).toBe('page');
	});
});
