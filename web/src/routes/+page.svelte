<!--
	Dashboard (docs/08-screens.md §3). Isinya dibangun di M11; di M0 hanya placeholder.
-->
<script lang="ts">
	import PagePlaceholder from '#lib/components/PagePlaceholder.svelte';
	import { en, pageTitle } from '#lib/i18n/en.ts';
</script>

<svelte:head>
	<title>{pageTitle(en.pages.dashboard)}</title>
</svelte:head>

<PagePlaceholder title={en.pages.dashboard} />
