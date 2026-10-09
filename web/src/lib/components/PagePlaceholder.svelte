<!--
	Isi sementara untuk halaman yang layarnya dibangun di milestone berikutnya: judul halaman dan
	pesan bahwa layar belum tersedia (tanpa data palsu, docs/02-phase1-scope.md §2).
-->
<script lang="ts">
	import { en } from '#lib/i18n/en.ts';

	// title adalah judul halaman dari en.pages.
	let { title }: { title: string } = $props();
</script>

<section>
	<h1>{title}</h1>
	<p>{en.placeholder.message}</p>
</section>

<style>
	h1 {
		margin: 0 0 8px;
		font-size: 20px;
		font-weight: 600;
	}

	p {
		margin: 0;
		color: var(--color-text-secondary);
	}
</style>
