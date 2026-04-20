<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import {
		apiFetch,
		type ClosedChatPage,
		type WhatsAppInstance,
		type WorkspaceSnapshot,
		type WorkspaceUser
	} from '$lib/api';

	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let users = $state<WorkspaceUser[]>([]);
	let instances = $state<WhatsAppInstance[]>([]);
	let pageNumber = $state(1);
	let selectedAgent = $state('');
	let selectedInstance = $state('');
	let closed = $state<ClosedChatPage | null>(null);

	async function loadReferenceData() {
		const workspace = await apiFetch<WorkspaceSnapshot>('/api/chats?tab=assigned');
		users = workspace.users;
		instances = workspace.instances;
	}

	async function loadClosedChats(targetPage = pageNumber) {
		loading = true;
		error = '';
		try {
			const query = new URLSearchParams({
				page: String(targetPage),
				page_size: '10'
			});
			if (selectedAgent) query.set('agent_id', selectedAgent);
			if (selectedInstance) query.set('instance_id', selectedInstance);
			closed = await apiFetch<ClosedChatPage>(`/api/chats/closed?${query.toString()}`);
			pageNumber = closed.page;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load closed chats.';
		} finally {
			loading = false;
		}
	}

	async function reopenChat(contactId: string) {
		error = '';
		success = '';
		try {
			await apiFetch(`/api/chats/${contactId}/reopen`, { method: 'PUT' });
			success = 'Conversation reopened.';
			await loadClosedChats(pageNumber);
			await goto(`/chat/${contactId}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to reopen conversation.';
		}
	}

	onMount(async () => {
		try {
			await loadReferenceData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load reference data.';
		}
		await loadClosedChats(1);
	});
</script>

<div class="mx-auto max-w-6xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 11</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Closed Conversations</h1>
			<p class="mt-2 text-sm text-gray-500">Filter closed chats by agent or account, refresh the queue, and reopen a conversation straight into the inbox.</p>
		</div>
		<div class="flex gap-2">
			<a href="/settings/contacts" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Contacts</a>
			<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Back to Settings</a>
		</div>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
		<div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
			<select data-testid="closed-agent-filter" bind:value={selectedAgent} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
				<option value="">All agents</option>
				{#each users as user}
					<option value={user.id}>{user.name}</option>
				{/each}
			</select>
			<select data-testid="closed-instance-filter" bind:value={selectedInstance} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
				<option value="">All accounts</option>
				{#each instances as instance}
					<option value={instance.id}>{instance.name}</option>
				{/each}
			</select>
			<button data-testid="closed-refresh" class="rounded-full border border-gray-200 px-4 py-3 text-sm font-medium text-gray-700" onclick={() => loadClosedChats(1)}>Refresh</button>
		</div>

		{#if loading}
			<div class="mt-4 rounded-[1.5rem] border border-gray-200 bg-gray-50 px-4 py-5 text-sm text-gray-600">Loading closed conversations...</div>
		{:else if closed}
			<div class="mt-4 space-y-3">
				{#each closed.items as row}
					<div class="rounded-[1.5rem] border border-gray-200 px-4 py-4" data-contact-name={row.contact_name}>
						<div class="flex flex-wrap items-start justify-between gap-3">
							<div>
								<p class="text-base font-semibold text-gray-900">{row.contact_name}</p>
								<p class="mt-1 text-sm text-gray-500">{row.phone_display} · {row.instance_name}</p>
								<p class="mt-1 text-xs text-gray-400">Assigned to {row.assigned_user_name || 'n/a'} · Closed by {row.closed_by || 'n/a'} · {new Date(row.closed_at).toLocaleString()}</p>
							</div>
							<button data-testid={`closed-reopen-${row.id}`} class="rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white" onclick={() => reopenChat(row.id)}>Reopen</button>
						</div>
					</div>
				{/each}

				{#if closed.items.length === 0}
					<div class="rounded-[1.5rem] border border-dashed border-gray-200 px-4 py-8 text-center text-sm text-gray-500">No closed conversations match the current filters.</div>
				{/if}
			</div>

			<div class="mt-5 flex items-center justify-between gap-3 border-t border-gray-100 pt-4 text-sm text-gray-600">
				<p>Showing page {closed.page} · {closed.total} total rows</p>
				<div class="flex gap-2">
					<button data-testid="closed-page-prev" class="rounded-full border border-gray-200 px-4 py-2 disabled:opacity-50" onclick={() => loadClosedChats(pageNumber - 1)} disabled={!closed.has_previous}>Previous</button>
					<button data-testid="closed-page-next" class="rounded-full border border-gray-200 px-4 py-2 disabled:opacity-50" onclick={() => loadClosedChats(pageNumber + 1)} disabled={!closed.has_next}>Next</button>
				</div>
			</div>
		{/if}
	</section>
</div>
