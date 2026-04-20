<script lang="ts">
	import { onMount } from 'svelte';
	import {
		apiFetch,
		type Campaign,
		type ContactsView,
		type LicenseBootstrapView
	} from '$lib/api';

	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let bootstrap = $state<LicenseBootstrapView | null>(null);
	let contacts = $state<ContactsView['contacts']>([]);
	let instances = $state<ContactsView['instances']>([]);
	let campaigns = $state<Campaign[]>([]);

	async function loadAll() {
		loading = true;
		error = '';
		try {
			const [bootstrapResponse, contactsResponse, campaignsResponse] = await Promise.all([
				apiFetch<LicenseBootstrapView>('/api/license/bootstrap'),
				apiFetch<ContactsView>('/api/contacts'),
				apiFetch<{ campaigns: Campaign[] }>('/api/campaigns')
			]);
			bootstrap = bootstrapResponse;
			contacts = contactsResponse.contacts;
			instances = contactsResponse.instances;
			campaigns = campaignsResponse.campaigns;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load cleanup workspace.';
		} finally {
			loading = false;
		}
	}

	function isOver(resource: string) {
		return bootstrap?.quotas.find((quota) => quota.resource === resource)?.over_quota ?? false;
	}

	async function deleteContact(contactId: string, name: string) {
		if (!window.confirm(`Delete ${name}?`)) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/contacts/${contactId}`, { method: 'DELETE' });
			success = `${name} deleted from contacts.`;
			await loadAll();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete contact.';
		}
	}

	async function deleteInstance(instanceId: string, name: string) {
		if (!window.confirm(`Delete ${name}?`)) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/instances/${instanceId}`, { method: 'DELETE' });
			success = `${name} deleted from accounts.`;
			await loadAll();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete account.';
		}
	}

	async function deleteCampaign(campaignId: string, name: string) {
		if (!window.confirm(`Delete ${name}?`)) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/campaigns/${campaignId}`, { method: 'DELETE' });
			success = `${name} deleted from campaigns.`;
			await loadAll();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete campaign.';
		}
	}

	onMount(loadAll);
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-amber-600">Restricted Cleanup</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">License Cleanup Workspace</h1>
			<p class="mt-2 text-sm text-gray-500">Only cleanup actions should be used here until all over-limit resources drop back within the active license entitlements.</p>
		</div>
		<div class="flex gap-2">
			<a href="/settings/license" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">License</a>
			<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Settings</a>
		</div>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading cleanup actions...</div>
	{:else if bootstrap}
		<div class="grid gap-4 lg:grid-cols-3">
			{#each bootstrap.quotas as quota}
				<div class={`rounded-[1.75rem] border px-5 py-5 shadow-sm ${quota.over_quota ? 'border-amber-200 bg-amber-50' : 'border-gray-200 bg-white'}`}>
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">{quota.label}</p>
					<p class="mt-2 text-3xl font-semibold text-gray-900">{quota.current}/{quota.limit}</p>
					<p class="mt-2 text-sm {quota.over_quota ? 'text-amber-800' : 'text-gray-500'}">{quota.over_quota ? 'Over limit: remove items below.' : 'Within limit.'}</p>
				</div>
			{/each}
		</div>

		{#if !bootstrap.restricted_cleanup}
			<div class="mt-5 rounded-[1.75rem] border border-emerald-200 bg-emerald-50 px-5 py-4 text-sm text-emerald-800">
				Cleanup is complete. Normal operations are available again.
			</div>
		{/if}

		<div class="mt-6 grid gap-4 xl:grid-cols-3">
			<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Contacts</p>
						<h2 class="mt-1 text-lg font-semibold text-gray-900">Delete Contacts</h2>
					</div>
					<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${isOver('contacts') ? 'bg-amber-50 text-amber-700' : 'bg-gray-100 text-gray-500'}`}>{isOver('contacts') ? 'Required' : 'Optional'}</span>
				</div>
				<div class="mt-4 space-y-3">
					{#each contacts as contact}
						<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3" data-contact-name={contact.name}>
							<p class="font-medium text-gray-900">{contact.name}</p>
							<p class="mt-1 text-xs text-gray-500">{contact.phone_display} · {contact.instance_name}</p>
							<button data-testid={`cleanup-delete-contact-${contact.id}`} class="mt-3 rounded-full border border-red-200 px-3 py-2 text-xs font-medium text-red-600" onclick={() => deleteContact(contact.id, contact.name)}>Delete Contact</button>
						</div>
					{/each}
				</div>
			</section>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Accounts</p>
						<h2 class="mt-1 text-lg font-semibold text-gray-900">Delete Disconnected Accounts</h2>
					</div>
					<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${isOver('instances') ? 'bg-amber-50 text-amber-700' : 'bg-gray-100 text-gray-500'}`}>{isOver('instances') ? 'Required' : 'Optional'}</span>
				</div>
				<div class="mt-4 space-y-3">
					{#each instances as instance}
						<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3" data-instance-name={instance.name}>
							<p class="font-medium text-gray-900">{instance.name}</p>
							<p class="mt-1 text-xs text-gray-500">{instance.status} · {instance.phone_number || 'No phone number'}</p>
							<button data-testid={`cleanup-delete-instance-${instance.id}`} class="mt-3 rounded-full border border-red-200 px-3 py-2 text-xs font-medium text-red-600" onclick={() => deleteInstance(instance.id, instance.name)}>Delete Account</button>
						</div>
					{/each}
				</div>
			</section>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Campaigns</p>
						<h2 class="mt-1 text-lg font-semibold text-gray-900">Delete Campaigns</h2>
					</div>
					<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${isOver('campaigns') ? 'bg-amber-50 text-amber-700' : 'bg-gray-100 text-gray-500'}`}>{isOver('campaigns') ? 'Required' : 'Optional'}</span>
				</div>
				<div class="mt-4 space-y-3">
					{#each campaigns as campaign}
						<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3" data-campaign-name={campaign.name}>
							<p class="font-medium text-gray-900">{campaign.name}</p>
							<p class="mt-1 text-xs text-gray-500">{campaign.status} · {campaign.source}</p>
							<button data-testid={`cleanup-delete-campaign-${campaign.id}`} class="mt-3 rounded-full border border-red-200 px-3 py-2 text-xs font-medium text-red-600" onclick={() => deleteCampaign(campaign.id, campaign.name)}>Delete Campaign</button>
						</div>
					{/each}
				</div>
			</section>
		</div>
	{/if}
</div>
