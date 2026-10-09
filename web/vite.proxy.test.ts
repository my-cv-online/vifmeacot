// Test konfigurasi proxy dev Vite (test case TC-M00-020, docs/test-cases/M00-scaffold.md).
import { describe, expect, test } from 'vitest';
import { apiProxy } from './vite.proxy.ts';

describe('apiProxy', () => {
	// TC-M00-020: /api (termasuk WebSocket /api/v1/ws) diteruskan ke server Go di port HTTP_ADDR,
	// dan header Origin asli dipertahankan untuk cek Origin mulai M2.
	test('TC-M00-020 dev proxy forwards /api and WebSocket to the Go server', () => {
		// expected membuat konfigurasi proxy yang diharapkan untuk port server Go tertentu.
		const expected = (port: number) => ({
			'/api': { target: `http://127.0.0.1:${port}`, ws: true, changeOrigin: false }
		});

		// Tanpa HTTP_ADDR dan dengan nilai bawaan: port 8080.
		expect(apiProxy(undefined)).toEqual(expected(8080));
		expect(apiProxy('')).toEqual(expected(8080));
		expect(apiProxy(':8080')).toEqual(expected(8080));

		// Port mengikuti HTTP_ADDR; host kosong atau "semua antarmuka" dituju lewat 127.0.0.1
		// supaya Node tidak mencoba ::1 sementara server hanya mendengar di IPv4.
		expect(apiProxy(':8081')).toEqual(expected(8081));
		expect(apiProxy('0.0.0.0:9090')).toEqual(expected(9090));
		expect(apiProxy('127.0.0.1:7000')).toEqual(expected(7000));
	});
});
