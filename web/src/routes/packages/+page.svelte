<!--
	Daftar paket (docs/08-screens.md §4). Dibangun di M4; di M0 hanya placeholder.
-->
<script lang="ts">
	import PagePlaceholder from '#lib/components/PagePlaceholder.svelte';
	import { en, pageTitle } from '#lib/i18n/en.ts';
</script>

<svelte:head>
	<title>{pageTitle(en.pages.packages)}</title>
</svelte:head>

<PagePlaceholder title={en.pages.packages} />
