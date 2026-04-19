<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { apiFetch, type WorkspaceSnapshot } from '$lib/api';

	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const workspace = await apiFetch<WorkspaceSnapshot>('/api/chats?tab=assigned');
			const fallback =
				workspace.conversations[0] ??
				(workspace.current_tab !== 'pending'
					? (await apiFetch<WorkspaceSnapshot>('/api/chats?tab=pending')).conversations[0]
					: undefined);

			if (fallback) {
				await goto(`/chat/${fallback.id}?tab=${workspace.current_tab || 'assigned'}`, { replaceState: true });
				return;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load inbox.';
		} finally {
			loading = false;
		}
	});
</script>

<div class="min-h-[70vh] flex items-center justify-center px-6">
	{#if loading}
		<div class="rounded-3xl border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Opening the conversation workspace...</div>
	{:else if error}
		<div class="rounded-3xl border border-red-200 bg-red-50 px-6 py-5 shadow-sm text-red-700">{error}</div>
	{:else}
		<div class="rounded-3xl border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">No conversations are available yet.</div>
	{/if}
</div>
