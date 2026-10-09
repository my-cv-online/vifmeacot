// Konfigurasi Playwright untuk test E2E user story (docs/11-testing.md §4). Spec E2E baru
// ditulis mulai M2 di tests/e2e; di M0 hanya konfigurasi dasarnya.
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './tests/e2e',
	// Satu worker: semua spec memakai satu server dan satu database yang di-reset per file.
	workers: 1,
	fullyParallel: false,
	reporter: [['list'], ['html', { open: 'never' }]],
	use: {
		// Server hasil `make build` dengan DEV_MODE=true dan DEV_FAKE_TODAY=2026-10-08.
		baseURL: 'http://localhost:8080',
		trace: 'retain-on-failure'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
	// TODO(M2): helper resetDb() (pfmea seed-demo --reset) dan start/stop server per file spec.
});
