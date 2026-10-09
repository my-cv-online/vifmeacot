<!--
	Hasil cek konsistensi semua paket (docs/08-screens.md §9). Dibangun di M8; di M0 hanya placeholder.
-->
<script lang="ts">
	import PagePlaceholder from '#lib/components/PagePlaceholder.svelte';
	import { en, pageTitle } from '#lib/i18n/en.ts';
</script>

<svelte:head>
	<title>{pageTitle(en.pages.findings)}</title>
</svelte:head>

<PagePlaceholder title={en.pages.findings} />
