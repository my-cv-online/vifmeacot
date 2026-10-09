<!--
	Layout akar: memuat font IBM Plex (self-hosted, jaringan pabrik bisa offline), gaya dasar dan
	app shell untuk semua halaman.
-->
<script lang="ts">
	import '@fontsource/ibm-plex-sans/400.css';
	import '@fontsource/ibm-plex-sans/600.css';
	import '@fontsource/ibm-plex-mono/400.css';
	import '../app.css';
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import AppShell from '#lib/components/AppShell.svelte';

	// children adalah halaman aktif yang dirender di dalam shell.
	let { children }: { children: Snippet } = $props();
</script>

<AppShell pathname={page.url.pathname}>
	{@render children()}
</AppShell>
