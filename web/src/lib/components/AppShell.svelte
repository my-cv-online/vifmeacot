<!--
	Kerangka aplikasi (docs/08-screens.md §1): top bar, sidebar navigasi, breadcrumb dan area
	halaman. Path dikirim lewat prop supaya komponen tidak bergantung pada $app/state dan bisa
	diuji di Vitest.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { en } from '#lib/i18n/en.ts';
	import { breadcrumbFor, isActive, navSections } from '#lib/navigation.ts';

	// pathname adalah path halaman saat ini; children adalah isi halaman.
	let { pathname, children }: { pathname: string; children?: Snippet } = $props();

	// Breadcrumb dihitung ulang setiap kali path berubah.
	const crumbs = $derived(breadcrumbFor(pathname));
</script>

<div class="shell">
	<header class="topbar">
		<a class="brand" href="/">
			<span class="logo" aria-hidden="true">{en.app.logoMark}</span>
			<span>{en.app.name}</span>
		</a>
		<!-- Pencarian global dihubungkan ke GET /search di M11; sampai saat itu nonaktif. -->
		<input
			class="search"
			type="search"
			placeholder={en.topbar.searchPlaceholder}
			aria-label={en.topbar.searchLabel}
			disabled
		/>
		<!-- TODO(M2): badge "DEMO DATA" bila DEV_MODE (kontrak API belum punya penandanya) dan
		     menu pengguna (nama, peran, "Change password", "Log out"). Lonceng notifikasi adalah
		     Tahap 3 dan tidak dirender. -->
	</header>

	<nav class="sidebar" aria-label={en.nav.label}>
		{#each navSections as section (section.title)}
			<div class="section">
				<div class="section-title">{section.title}</div>
				<ul>
					{#each section.items as item (item.href)}
						<li>
							<a
								href={item.href}
								aria-current={isActive(pathname, item.href) ? 'page' : undefined}
							>
								{item.label}
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>

	<main class="content">
		<!-- TODO(M5): banner "Live updates disconnected. Reconnecting…" saat WebSocket putus. -->
		<nav class="breadcrumb" aria-label={en.breadcrumb.label}>
			<ol>
				{#each crumbs as crumb, i (crumb.href)}
					<li>
						{#if i === crumbs.length - 1}
							<span aria-current="page">{crumb.label}</span>
						{:else}
							<a href={crumb.href}>{crumb.label}</a>
						{/if}
					</li>
				{/each}
			</ol>
		</nav>
		{@render children?.()}
	</main>
</div>

<style>
	/* Tata letak: top bar selebar layar, sidebar di kiri, isi halaman di kanan. */
	.shell {
		display: grid;
		grid-template-columns: 240px 1fr;
		grid-template-rows: 56px 1fr;
		min-height: 100vh;
	}

	.topbar {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 0 16px;
		background: var(--color-topbar);
		color: #ffffff;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: var(--touch-target);
		color: inherit;
		font-weight: 600;
		text-decoration: none;
	}

	.logo {
		display: grid;
		place-items: center;
		width: 28px;
		height: 28px;
		border-radius: 6px;
		background: var(--color-accent);
		font-family: var(--font-mono);
	}

	.search {
		width: min(420px, 40vw);
		min-height: var(--touch-target);
		padding: 0 12px;
		border: 1px solid #2c333b;
		border-radius: 6px;
		background: #1f252c;
		color: #ffffff;
		font: inherit;
	}

	.search:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.sidebar {
		padding: 16px 8px;
		border-right: 1px solid var(--color-border-row);
		background: var(--color-surface);
	}

	.section + .section {
		margin-top: 16px;
	}

	/* Judul grup kecil dan redup supaya tautan tetap menonjol. */
	.section-title {
		padding: 0 12px 4px;
		color: var(--color-text-muted);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.06em;
	}

	.sidebar ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.sidebar a {
		display: flex;
		align-items: center;
		min-height: var(--touch-target);
		padding: 0 12px;
		border-radius: 6px;
		color: var(--color-text);
		text-decoration: none;
	}

	.sidebar a:hover {
		background: var(--color-surface-subtle);
	}

	/* Menu halaman aktif memakai warna navigasi terpilih dari token desain. */
	.sidebar a[aria-current='page'] {
		background: var(--color-nav-selected-bg);
		color: var(--color-nav-selected-text);
		font-weight: 600;
	}

	.content {
		padding: 16px 24px;
		min-width: 0;
	}

	.breadcrumb ol {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin: 0 0 16px;
		padding: 0;
		list-style: none;
		color: var(--color-text-secondary);
	}

	/* Pemisah "/" antar item breadcrumb. */
	.breadcrumb li + li::before {
		content: '/';
		margin-right: 6px;
		color: var(--color-text-muted);
	}
</style>
