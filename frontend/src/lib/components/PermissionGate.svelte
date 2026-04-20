<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		allowed,
		fallback,
		children
	}: {
		allowed: boolean;
		fallback?: Snippet;
		children: Snippet;
	} = $props();
</script>

{#if allowed}
	{@render children()}
{:else if fallback}
	{@render fallback()}
{/if}
