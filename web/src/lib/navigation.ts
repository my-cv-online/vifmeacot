// Struktur navigasi app shell (sidebar dan breadcrumb) sesuai docs/08-screens.md §1. Grup paket
// yang terakhir dibuka dan MASTER DATA ditambahkan di milestone berikutnya.
import { en } from '#lib/i18n/en.ts';

// NavItem adalah satu tautan di sidebar.
export interface NavItem {
	href: string;
	label: string;
}

// NavSection adalah satu grup sidebar dengan judulnya.
export interface NavSection {
	title: string;
	items: NavItem[];
}

// Crumb adalah satu bagian breadcrumb.
export interface Crumb {
	href: string;
	label: string;
}

// navSections adalah isi sidebar Tahap 1 yang sudah ada di M0.
// TODO(M4): grup "PACKAGE <code> · AIAG 4TH" untuk paket yang terakhir dibuka (local storage).
// TODO(M3): grup MASTER DATA, hanya untuk admin (perlu peran dari getMe, M2).
export const navSections: readonly NavSection[] = [
	{
		title: en.nav.monitoring,
		items: [
			{ href: '/', label: en.nav.dashboard },
			{ href: '/findings', label: en.nav.findings }
		]
	},
	{
		title: en.nav.packages,
		items: [{ href: '/packages', label: en.nav.allPackages }]
	},
	{
		title: en.nav.template,
		items: [{ href: '/template-general', label: en.nav.templateGeneral }]
	}
];

// sectionPages memetakan segmen pertama path ke judul halaman untuk breadcrumb.
const sectionPages: Record<string, Crumb> = {
	findings: { href: '/findings', label: en.pages.findings },
	packages: { href: '/packages', label: en.pages.packages },
	'template-general': { href: '/template-general', label: en.pages.templateGeneral }
};

// isActive menentukan apakah tautan sidebar mewakili halaman saat ini. Dashboard hanya aktif
// di "/" persis; tautan lain juga aktif untuk sub-path-nya (misalnya /packages/PS-07).
export function isActive(pathname: string, href: string): boolean {
	if (href === '/') {
		return pathname === '/';
	}
	return pathname === href || pathname.startsWith(`${href}/`);
}

// breadcrumbFor menyusun breadcrumb untuk path: selalu diawali Dashboard, lalu halaman bagian.
// TODO(M4): tambahkan "Package <code> <name>" dan dokumen (PFD, PFMEA, Control Plan).
export function breadcrumbFor(pathname: string): Crumb[] {
	const crumbs: Crumb[] = [{ href: '/', label: en.pages.dashboard }];
	const first = pathname.split('/').filter(Boolean)[0];
	const section = first ? sectionPages[first] : undefined;
	if (section) {
		crumbs.push(section);
	}
	return crumbs;
}
