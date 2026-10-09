<!--
	Halaman error SPA, dirender di dalam app shell untuk route yang tidak dikenal (misalnya layar
	milestone berikutnya) atau error saat memuat halaman. Menggantikan halaman error bawaan
	SvelteKit supaya semua teks berasal dari en.ts.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { en, errorTitle, pageTitle } from '#lib/i18n/en.ts';

	// Judul dihitung dari status error halaman saat ini.
	const title = $derived(errorTitle(page.status));
</script>

<svelte:head>
	<title>{pageTitle(title)}</title>
</svelte:head>

<section>
	<h1>{title}</h1>
	<p><a href="/">{en.error.backToDashboard}</a></p>
</section>

<style>
	h1 {
		margin: 0 0 8px;
		font-size: 20px;
		font-weight: 600;
	}

	p {
		margin: 0;
	}
</style>
