// Proxy server dev Vite: /api (termasuk WebSocket /api/v1/ws) diteruskan ke server Go supaya
// browser di :5173 bisa memanggil API tanpa CORS (docs/10-milestones.md M0).
import type { ProxyOptions } from 'vite';

// defaultPort adalah port server Go bila HTTP_ADDR tidak di-set (docs/03-architecture.md §8).
const defaultPort = 8080;

// Host yang berarti "semua antarmuka"; proxy harus menuju alamat loopback yang nyata.
const anyHost = new Set(['', '0.0.0.0', '::', '[::]']);

// apiProxy membuat konfigurasi proxy dari HTTP_ADDR server Go (misalnya ":8080" atau
// "0.0.0.0:9090"). changeOrigin dibiarkan false supaya header Origin asli browser sampai ke
// server, karena mulai M2 server membandingkannya dengan APP_BASE_URL.
export function apiProxy(httpAddr: string | undefined): Record<string, ProxyOptions> {
	let host = '127.0.0.1';
	let port = defaultPort;
	const addr = (httpAddr ?? '').trim();
	const sep = addr.lastIndexOf(':');
	if (sep >= 0) {
		const parsed = Number(addr.slice(sep + 1));
		if (Number.isInteger(parsed) && parsed > 0) {
			port = parsed;
		}
		const h = addr.slice(0, sep);
		if (!anyHost.has(h)) {
			host = h;
		}
	}
	return { '/api': { target: `http://${host}:${port}`, ws: true, changeOrigin: false } };
}
