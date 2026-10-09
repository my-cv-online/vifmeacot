// Semua teks antarmuka aplikasi dalam bahasa Inggris (CLAUDE.md, Language policy). Komponen
// tidak boleh menulis teks yang terlihat pengguna secara langsung; ambil dari objek en.
// Label mengikuti docs/08-screens.md §1 dan §14.

export const en = {
	// Identitas aplikasi di top bar dan judul tab browser.
	app: {
		name: 'PFD · PFMEA · CP System',
		// Huruf di kotak logo top bar (dekoratif, disembunyikan dari pembaca layar).
		logoMark: 'P'
	},
	// Top bar.
	topbar: {
		searchPlaceholder: 'Search packages, steps, characteristics',
		searchLabel: 'Search'
	},
	// Sidebar: judul grup ditulis kapital seperti di docs/08-screens.md §1.
	nav: {
		label: 'Main navigation',
		monitoring: 'MONITORING',
		dashboard: 'Dashboard',
		findings: 'Consistency check',
		packages: 'PACKAGES',
		allPackages: 'All packages',
		template: 'TEMPLATE',
		templateGeneral: 'Template General'
	},
	// Breadcrumb di atas setiap halaman.
	breadcrumb: {
		label: 'Breadcrumb'
	},
	// Judul halaman (dipakai di breadcrumb, heading dan judul tab).
	pages: {
		dashboard: 'Dashboard',
		findings: 'Consistency check',
		packages: 'Packages',
		templateGeneral: 'Template General'
	},
	// Isi halaman yang layarnya dibangun di milestone berikutnya.
	placeholder: {
		message: 'This screen is not available yet.'
	},
	// Halaman error SPA (route tidak dikenal atau error saat memuat halaman).
	error: {
		notFound: 'Page not found',
		generic: 'Something went wrong',
		backToDashboard: 'Back to Dashboard'
	}
} as const;

// pageTitle menyusun judul tab browser: "<halaman> · <nama aplikasi>".
export function pageTitle(page: string): string {
	return `${page} · ${en.app.name}`;
}

// errorTitle memilih judul halaman error dari status HTTP: 404 untuk route yang tidak dikenal,
// pesan umum untuk error lain.
export function errorTitle(status: number): string {
	return status === 404 ? en.error.notFound : en.error.generic;
}
