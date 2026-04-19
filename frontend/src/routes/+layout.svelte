<script lang="ts">
	import type { Snippet } from 'svelte';

	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';

	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	function applyTheme(theme: string | undefined) {
		const selected = theme ?? 'light';
		const nextTheme = selected === 'system'
			? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
			: selected;
		document.documentElement.dataset.theme = nextTheme;
	}

	$effect(() => {
		if (typeof window === 'undefined') {
			return;
		}
		const language = data.user?.settings.language ?? 'en';
		document.documentElement.lang = language;
		document.documentElement.dir = language === 'ar' ? 'rtl' : 'ltr';
		applyTheme(data.user?.settings.theme);
	});

</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
{@render children()}
