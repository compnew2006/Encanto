<script lang="ts">
	import { onMount } from 'svelte';
	import {
		apiFetch,
		apiFetchText,
		type ChatContact,
		type ContactImportResult,
		type ContactsView
	} from '$lib/api';

	type ContactDraft = {
		name: string;
		phone_number: string;
		instance_id: string;
		tags_text: string;
	};

	const blankDraft = (): ContactDraft => ({
		name: '',
		phone_number: '',
		instance_id: '',
		tags_text: ''
	});

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let success = $state('');
	let contacts = $state<ChatContact[]>([]);
	let instances = $state<ContactsView['instances']>([]);
	let search = $state('');
	let instanceId = $state('');
	let editingId = $state<string | null>(null);
	let draft = $state<ContactDraft>(blankDraft());
	let exportCSV = $state('');
	let sampleCSV = $state('');
	let importCSV = $state('');
	let updateOnDuplicate = $state(true);
	let importResult = $state<ContactImportResult | null>(null);

	function draftPayload() {
		return {
			name: draft.name,
			phone_number: draft.phone_number,
			instance_id: draft.instance_id,
			tags: draft.tags_text
				.split(',')
				.map((value) => value.trim())
				.filter(Boolean)
		};
	}

	function resetDraft() {
		draft = {
			...blankDraft(),
			instance_id: instances[0]?.id ?? ''
		};
		editingId = null;
	}

	function beginEdit(contact: ChatContact) {
		editingId = contact.id;
		draft = {
			name: contact.name,
			phone_number: contact.phone_number,
			instance_id: contact.instance_id,
			tags_text: contact.tags.join(', ')
		};
		success = '';
		error = '';
	}

	async function loadContacts() {
		loading = true;
		error = '';
		try {
			const query = new URLSearchParams();
			if (search.trim()) query.set('search', search.trim());
			if (instanceId) query.set('instance_id', instanceId);

			const view = await apiFetch<ContactsView>(
				`/api/contacts${query.size ? `?${query.toString()}` : ''}`
			);
			contacts = view.contacts;
			instances = view.instances;
			if (!draft.instance_id) {
				draft.instance_id = view.instance_id || view.instances[0]?.id || '';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load contacts.';
		} finally {
			loading = false;
		}
	}

	async function saveContact() {
		saving = true;
		error = '';
		success = '';
		try {
			if (editingId) {
				await apiFetch(`/api/contacts/${editingId}`, {
					method: 'PUT',
					body: draftPayload()
				});
				success = 'Contact updated.';
			} else {
				await apiFetch('/api/contacts', {
					method: 'POST',
					body: draftPayload()
				});
				success = 'Contact created.';
			}
			await loadContacts();
			resetDraft();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save contact.';
		} finally {
			saving = false;
		}
	}

	async function deleteContact(contact: ChatContact) {
		if (!window.confirm(`Delete ${contact.name}?`)) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/contacts/${contact.id}`, { method: 'DELETE' });
			success = 'Contact deleted.';
			await loadContacts();
			if (editingId === contact.id) {
				resetDraft();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete contact.';
		}
	}

	async function exportContacts(sample = false) {
		error = '';
		success = '';
		try {
			const csv = await apiFetchText(
				`/api/contacts/export?columns=name,phone_number,instance_name,status,assigned_user_name,tags${sample ? '&sample=1' : ''}`
			);
			if (sample) {
				sampleCSV = csv;
				success = 'Import sample generated.';
			} else {
				exportCSV = csv;
				success = 'Contacts exported.';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to export contacts.';
		}
	}

	async function importContacts() {
		error = '';
		success = '';
		try {
			importResult = await apiFetch<ContactImportResult>('/api/contacts/import', {
				method: 'POST',
				body: {
					csv: importCSV,
					update_on_duplicate: updateOnDuplicate
				}
			});
			success = 'Contacts imported.';
			await loadContacts();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to import contacts.';
		}
	}

	function openChat(contact: ChatContact) {
		window.location.href = `/chat/${contact.id}`;
	}

	onMount(async () => {
		await loadContacts();
	});
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 11</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Contacts Directory</h1>
			<p class="mt-2 text-sm text-gray-500">Create, edit, export, import, and open conversations directly from the shared contact catalog.</p>
		</div>
		<div class="flex gap-2">
			<a href="/settings/closed-chats" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Closed Chats</a>
			<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Back to Settings</a>
		</div>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	<div class="grid gap-4 xl:grid-cols-[0.7fr_1.3fr]">
		<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
			<div class="flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Editor</p>
					<h2 class="mt-1 text-lg font-semibold text-gray-900">{editingId ? 'Edit Contact' : 'Add Contact'}</h2>
				</div>
				<button class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-600" onclick={resetDraft}>Clear</button>
			</div>

			<div class="mt-4 space-y-3">
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Name</span>
					<input data-testid="contact-name" bind:value={draft.name} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Contact name" />
				</label>
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Phone Number</span>
					<input data-testid="contact-phone" bind:value={draft.phone_number} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="+201..." />
				</label>
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Account</span>
					<select data-testid="contact-instance" bind:value={draft.instance_id} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
						{#each instances as instance}
							<option value={instance.id}>{instance.name}</option>
						{/each}
					</select>
				</label>
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Tags</span>
					<input data-testid="contact-tags" bind:value={draft.tags_text} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="vip, renewal, cairo" />
				</label>
			</div>

			<button data-testid="save-contact" class="mt-5 w-full rounded-full bg-gray-900 px-5 py-3 text-sm font-medium text-white disabled:opacity-60" onclick={saveContact} disabled={saving}>
				{editingId ? 'Save Contact' : 'Create Contact'}
			</button>

			<div class="mt-6 rounded-[1.5rem] border border-gray-200 bg-gray-50 p-4">
				<div class="flex flex-wrap gap-2">
					<button data-testid="contacts-export" class="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700" onclick={() => exportContacts(false)}>Export CSV</button>
					<button data-testid="contacts-sample" class="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700" onclick={() => exportContacts(true)}>Import Sample</button>
				</div>
				<textarea class="mt-3 min-h-[140px] w-full rounded-[1.25rem] border border-gray-200 bg-white px-4 py-3 text-xs text-gray-700 outline-none" readonly placeholder="CSV output appears here.">{exportCSV || sampleCSV}</textarea>
			</div>

			<div class="mt-6 rounded-[1.5rem] border border-gray-200 bg-gray-50 p-4">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-sm font-medium text-gray-900">CSV Import</p>
						<p class="mt-1 text-xs text-gray-500">Duplicates can be updated in-place when the same phone/account pair already exists.</p>
					</div>
					<label class="flex items-center gap-2 text-sm text-gray-700">
						<input data-testid="contacts-update-on-duplicate" bind:checked={updateOnDuplicate} type="checkbox" />
						Update duplicates
					</label>
				</div>
				<textarea data-testid="contacts-import-csv" bind:value={importCSV} class="mt-3 min-h-[160px] w-full rounded-[1.25rem] border border-gray-200 bg-white px-4 py-3 text-xs text-gray-700 outline-none focus:border-blue-400" placeholder="Paste CSV here..."></textarea>
				<button data-testid="contacts-import" class="mt-3 rounded-full bg-blue-600 px-4 py-2.5 text-sm font-medium text-white" onclick={importContacts}>Import Contacts</button>

				{#if importResult}
					<div class="mt-4 rounded-[1.25rem] bg-white p-4 text-sm text-gray-700">
						<p class="font-medium text-gray-900">Import Result</p>
						<p class="mt-2">Created: {importResult.created} · Updated: {importResult.updated} · Skipped: {importResult.skipped}</p>
						<p class="mt-1 text-xs text-gray-500">Job: {importResult.job.id} · {importResult.job.status}</p>
					</div>
				{/if}
			</div>
		</section>

		<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
			<div class="flex flex-wrap items-end justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Directory</p>
					<h2 class="mt-1 text-lg font-semibold text-gray-900">All Contacts</h2>
				</div>
				<div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_220px_auto]">
					<input data-testid="contacts-search" bind:value={search} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Search name or phone" />
					<select data-testid="contacts-instance-filter" bind:value={instanceId} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
						<option value="">All accounts</option>
						{#each instances as instance}
							<option value={instance.id}>{instance.name}</option>
						{/each}
					</select>
					<button class="rounded-full border border-gray-200 px-4 py-3 text-sm font-medium text-gray-700" onclick={loadContacts}>Apply</button>
				</div>
			</div>

			{#if loading}
				<div class="mt-4 rounded-[1.5rem] border border-gray-200 bg-gray-50 px-4 py-5 text-sm text-gray-600">Loading contacts...</div>
			{:else}
				<div class="mt-4 space-y-3">
					{#each contacts as contact}
						<div class="rounded-[1.5rem] border border-gray-200 px-4 py-4" data-contact-name={contact.name}>
							<div class="flex flex-wrap items-start justify-between gap-3">
								<div>
									<div class="flex items-center gap-3">
										<p class="text-base font-semibold text-gray-900">{contact.name}</p>
										<span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-gray-600">{contact.status}</span>
									</div>
									<p class="mt-1 text-sm text-gray-500">{contact.phone_display} · {contact.instance_name}</p>
									<p class="mt-1 text-xs text-gray-400">{contact.assigned_user_name || 'Unassigned'} · {contact.tags.join(', ') || 'No tags'}</p>
								</div>
								<div class="flex flex-wrap gap-2">
									<button data-testid={`open-chat-${contact.id}`} class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-700" onclick={() => openChat(contact)}>Open Chat</button>
									<button data-testid={`edit-contact-${contact.id}`} class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-700" onclick={() => beginEdit(contact)}>Edit</button>
									<button data-testid={`delete-contact-${contact.id}`} class="rounded-full border border-red-200 px-3 py-2 text-sm text-red-600" onclick={() => deleteContact(contact)}>Delete</button>
								</div>
							</div>
						</div>
					{/each}

					{#if contacts.length === 0}
						<div class="rounded-[1.5rem] border border-dashed border-gray-200 px-4 py-8 text-center text-sm text-gray-500">No contacts match the current filters.</div>
					{/if}
				</div>
			{/if}
		</section>
	</div>
</div>
