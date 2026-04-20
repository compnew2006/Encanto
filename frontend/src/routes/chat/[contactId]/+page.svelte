<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		apiFetch,
		formatDateTime,
		type ChatContact,
		type ChatMessage,
		type WorkspaceSnapshot
	} from '$lib/api';
	import { connectRealtime } from '$lib/realtime/ws';
	import { onDestroy, onMount } from 'svelte';

	let workspace = $state<WorkspaceSnapshot | null>(null);
	let loading = $state(true);
	let error = $state('');
	let search = $state('');
	let currentTab = $state('assigned');
	let instanceFilter = $state('');
	let tagFilter = $state('');
	let currentContactId = $state('');
	let noteDraft = $state('');
	let statusDraft = $state('');
	let selectedAssignee = $state('');
	let collaboratorUserId = $state('');
	let composerMode = $state<'text' | 'media'>('text');
	let composerText = $state('');
	let attachmentName = $state('');
	let attachmentFileSizeLabel = $state('');
	let attachmentUrl = $state('');
	let attachmentPreviewMime = $state('');
	let showNotes = $state(true);
	let showInfo = $state(true);
	let showTimeline = $state(false);
	let showNotifications = $state(false);
	let showStatuses = $state(false);
	let showQuickReplies = $state(false);
	let showDirectChatDialog = $state(false);
	let creatingDirectChat = $state(false);
	let newContactPhone = $state('');
	let newContactName = $state('');
	let newContactInstanceId = $state('');
	let isDragOver = $state(false);
	let saving = $state(false);
	let infoMessage = $state('');
	let attachmentInput = $state<HTMLInputElement | null>(null);

	let unsubscribePage: (() => void) | undefined;
	let teardownRealtime: (() => void) | undefined;
	let selectedConversation = $derived(workspace?.selected ?? null);
	let selectedContact = $derived(selectedConversation?.contact ?? null);

	function workspacePath(contactId: string) {
		const params = new URLSearchParams();
		params.set('tab', currentTab);
		if (search.trim()) params.set('search', search.trim());
		if (instanceFilter) params.set('instance_id', instanceFilter);
		if (tagFilter) params.set('tag', tagFilter);
		const query = params.toString();
		return `/chat/${contactId}${query ? `?${query}` : ''}`;
	}

	function chatBackgroundStyle() {
		const background = workspace?.settings.chat.chat_background ?? 'paper-grid';
		switch (background) {
			case 'linen':
				return 'background-image: linear-gradient(90deg, rgba(148,163,184,0.09) 1px, transparent 1px), linear-gradient(rgba(148,163,184,0.09) 1px, transparent 1px); background-size: 26px 26px;';
			case 'plain':
				return 'background-image: none;';
			default:
				return 'background-image: linear-gradient(rgba(59,130,246,0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(59,130,246,0.06) 1px, transparent 1px); background-size: 24px 24px;';
		}
	}

	function formatFileSize(size: number) {
		if (size < 1024) return `${size} B`;
		if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
		return `${(size / (1024 * 1024)).toFixed(1)} MB`;
	}

	function clearAttachment() {
		attachmentName = '';
		attachmentFileSizeLabel = '';
		attachmentUrl = '';
		attachmentPreviewMime = '';
		if (attachmentInput) {
			attachmentInput.value = '';
		}
	}

	function openDirectChatDialog() {
		if (!newContactInstanceId) {
			newContactInstanceId = workspace?.instances[0]?.id ?? '';
		}
		showDirectChatDialog = true;
	}

	function closeDirectChatDialog() {
		showDirectChatDialog = false;
		newContactPhone = '';
		newContactName = '';
	}

	function readFileAsDataURL(file: File) {
		return new Promise<string>((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = () => resolve(typeof reader.result === 'string' ? reader.result : '');
			reader.onerror = () => reject(new Error('Failed to read the selected file.'));
			reader.readAsDataURL(file);
		});
	}

	async function applyAttachmentFile(file: File) {
		attachmentName = file.name;
		attachmentFileSizeLabel = formatFileSize(file.size);
		attachmentPreviewMime = file.type;
		attachmentUrl = await readFileAsDataURL(file);
	}

	async function handleAttachmentChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		try {
			await applyAttachmentFile(file);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to prepare the attachment.';
		}
	}

	async function handleAttachmentDrop(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
		const file = event.dataTransfer?.files?.[0];
		if (!file) return;
		try {
			await applyAttachmentFile(file);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to prepare the attachment.';
		}
	}

	async function loadWorkspace() {
		if (!currentContactId) return;
		loading = true;
		error = '';

		try {
			const params = new URLSearchParams();
			params.set('tab', currentTab);
			if (search.trim()) params.set('search', search.trim());
			if (instanceFilter) params.set('instance_id', instanceFilter);
			if (tagFilter) params.set('tag', tagFilter);

			workspace = await apiFetch<WorkspaceSnapshot>(`/api/chats/${currentContactId}?${params.toString()}`);
			selectedAssignee = workspace.selected?.contact.assigned_user_id ?? '';
			if (
				workspace.instances &&
				workspace.instances.length > 0 &&
				(!newContactInstanceId || !workspace.instances.some((instance) => instance.id === newContactInstanceId))
			) {
				newContactInstanceId = workspace.instances[0].id;
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load the workspace.';
		} finally {
			loading = false;
		}
	}

	async function refreshAndKeepMessage(message: string) {
		infoMessage = message;
		await loadWorkspace();
		window.setTimeout(() => {
			if (infoMessage === message) infoMessage = '';
		}, 2200);
	}

	async function runAction(path: string, body?: object, successMessage = 'Updated.') {
		saving = true;
		try {
			await apiFetch(path, { method: 'POST', body });
			await refreshAndKeepMessage(successMessage);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Action failed.';
		} finally {
			saving = false;
		}
	}

	async function assignConversation() {
		if (!selectedAssignee) return;
		await runAction(`/api/chats/${currentContactId}/assign`, { assignee_id: selectedAssignee }, 'Conversation assigned.');
	}

	async function unassignConversation() {
		await runAction(`/api/chats/${currentContactId}/unassign`, undefined, 'Conversation returned to pending.');
	}

	async function togglePinnedConversation() {
		if (!selectedContact) return;
		await runAction(`/api/chats/${currentContactId}/pin`, undefined, 'Pin state updated.');
	}

	async function toggleHiddenConversation() {
		if (!selectedContact) return;
		await runAction(`/api/chats/${currentContactId}/hide`, undefined, 'Conversation visibility updated.');
	}

	async function toggleConversationState() {
		if (!selectedContact) return;
		const action = selectedContact.status === 'closed' ? 'reopen' : 'close';
		const message = selectedContact.status === 'closed' ? 'Conversation reopened.' : 'Conversation closed.';
		await runAction(`/api/chats/${currentContactId}/${action}`, undefined, message);
	}

	async function sendMessage() {
		if (composerMode === 'text' && !composerText.trim()) return;
		if (composerMode === 'media' && !attachmentName.trim()) return;

		saving = true;
		try {
			await apiFetch(`/api/chats/${currentContactId}/messages`, {
				method: 'POST',
				body: {
					type: composerMode,
					body: composerText.trim(),
					file_name: attachmentName.trim(),
					file_size_label: attachmentFileSizeLabel,
					media_url: attachmentUrl.trim()
				}
			});
			composerText = '';
			clearAttachment();
			await refreshAndKeepMessage('Message queued.');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to send.';
		} finally {
			saving = false;
		}
	}

	async function addNote() {
		if (!noteDraft.trim()) return;
		saving = true;
		try {
			await apiFetch(`/api/chats/${currentContactId}/notes`, {
				method: 'POST',
				body: { body: noteDraft }
			});
			noteDraft = '';
			await refreshAndKeepMessage('Note added.');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add note.';
		} finally {
			saving = false;
		}
	}

	async function inviteCollaborator() {
		if (!collaboratorUserId) return;
		saving = true;
		try {
			await apiFetch(`/api/chats/${currentContactId}/collaborators`, {
				method: 'POST',
				body: { user_id: collaboratorUserId }
			});
			collaboratorUserId = '';
			await refreshAndKeepMessage('Collaborator invited.');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to invite collaborator.';
		} finally {
			saving = false;
		}
	}

	async function addStatus() {
		const activeInstanceId = workspace?.selected?.contact.instance_id;
		if (!statusDraft.trim() || !activeInstanceId) return;
		saving = true;
		try {
			await apiFetch('/api/statuses', {
				method: 'POST',
				body: {
					contact_id: currentContactId,
					instance_id: activeInstanceId,
					body: statusDraft
				}
			});
			statusDraft = '';
			await refreshAndKeepMessage('Status posted.');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to post status.';
		} finally {
			saving = false;
		}
	}

	async function markNotificationsRead() {
		await runAction('/api/notifications/read-all', undefined, 'Notifications cleared.');
	}

	async function createDirectChat() {
		if (!newContactPhone.trim() || !newContactInstanceId) return;
		creatingDirectChat = true;
		error = '';
		try {
			const response = await apiFetch<{ contact: ChatContact }>('/api/chats/direct', {
				method: 'POST',
				body: {
					phone_number: newContactPhone.trim(),
					profile_name: newContactName.trim(),
					instance_id: newContactInstanceId
				}
			});
			const nextTab =
				response.contact.status === 'closed'
					? 'closed'
					: response.contact.status === 'pending'
						? 'pending'
						: 'assigned';
			currentTab = nextTab;
			closeDirectChatDialog();
			await goto(workspacePath(response.contact.id));
			infoMessage = 'Direct chat created.';
			window.setTimeout(() => {
				if (infoMessage === 'Direct chat created.') infoMessage = '';
			}, 2200);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create the chat.';
		} finally {
			creatingDirectChat = false;
		}
	}

	async function selectContact(contact: ChatContact) {
		await goto(workspacePath(contact.id));
	}

	async function applyFilters() {
		await goto(workspacePath(currentContactId), { replaceState: true });
	}

	async function openTab(tab: string) {
		currentTab = tab;
		await goto(workspacePath(currentContactId), { replaceState: true });
	}

	async function pickQuickReply(body: string) {
		composerMode = 'text';
		composerText = body;
		showQuickReplies = false;
	}

	function printMessage(message: ChatMessage) {
		const content = message.body || message.file_name || 'Message';
		const popup = window.open('', '_blank', 'width=640,height=480');
		if (!popup) return;
		popup.document.write(`<pre style="font-family: monospace; padding: 24px;">${content}</pre>`);
		popup.document.close();
		popup.print();
	}

	function downloadMessage(message: ChatMessage) {
		const href = message.media_url || '#';
		window.open(href, '_blank');
	}

	function contactInitials(name: string) {
		return name
			.split(' ')
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() ?? '')
			.join('');
	}

	function statusBadgeClass(status: string) {
		switch (status) {
			case 'assigned':
				return 'bg-sky-100 text-sky-700';
			case 'pending':
				return 'bg-amber-100 text-amber-700';
			case 'closed':
				return 'bg-slate-200 text-slate-600';
			default:
				return 'bg-slate-100 text-slate-600';
		}
	}

	function cleanIdentifier(value: string) {
		let cleaned = value.trim();
		const atIndex = cleaned.indexOf('@');
		if (atIndex !== -1) cleaned = cleaned.slice(0, atIndex);
		const colonIndex = cleaned.indexOf(':');
		if (colonIndex !== -1) cleaned = cleaned.slice(0, colonIndex);
		return cleaned;
	}

	function formatPhoneDisplay(value: string) {
		const cleaned = cleanIdentifier(value);
		if (!cleaned) return '';
		if (cleaned.startsWith('+')) return cleaned;
		if (/^\d{10,15}$/.test(cleaned)) return `+${cleaned}`;
		return cleaned;
	}

	function contactTitle(contact: ChatContact) {
		const label = contact.name?.trim();
		const displayPhone = formatPhoneDisplay(contact.phone_display || contact.phone_number);
		if (!label || label === contact.phone_number || label === contact.phone_display) {
			return displayPhone || label || 'Conversation';
		}
		return label;
	}

	onMount(async () => {
		unsubscribePage = page.subscribe(($page) => {
			const nextContact = $page.params.contactId ?? '';
			currentContactId = nextContact;
			currentTab = $page.url.searchParams.get('tab') ?? 'assigned';
			search = $page.url.searchParams.get('search') ?? '';
			instanceFilter = $page.url.searchParams.get('instance_id') ?? '';
			tagFilter = $page.url.searchParams.get('tag') ?? '';
			void loadWorkspace();
		});

		teardownRealtime = await connectRealtime(async (message) => {
			if (
				[
					'new_message',
					'status_update',
					'conversation_event',
					'notification',
					'notification_read',
					'status_feed_update',
					'instance_connected',
					'instance_disconnected',
					'instance_recovering'
				].includes(message.type)
			) {
				await loadWorkspace();
			}
		});
	});

	onDestroy(() => {
		unsubscribePage?.();
		teardownRealtime?.();
	});
</script>

<div class="h-[calc(100vh-4rem)] overflow-hidden bg-[#eef3fb] p-3 md:p-4">
	{#if error}
		<div class="mb-3 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if infoMessage}
		<div class="mb-3 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{infoMessage}</div>
	{/if}

	<div class="grid h-full min-h-0 grid-cols-1 gap-3 xl:grid-cols-[auto_minmax(20rem,24rem)_minmax(0,1fr)_22rem]">
		<aside class="group hidden h-full min-h-0 overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)] transition-[width] duration-300 ease-out xl:block xl:w-20 hover:xl:w-72">
			<div class="flex h-full flex-col px-3 py-4">
				<div class="mb-4 flex items-center gap-3 overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3">
					<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-sky-600 text-sm font-semibold text-white">E1</div>
					<div class="min-w-0 opacity-0 transition duration-200 group-hover:opacity-100">
						<p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400">Current view</p>
						<p class="truncate text-sm font-semibold text-slate-900">Conversations</p>
					</div>
				</div>

				<nav class="space-y-1.5">
					<button class="flex w-full items-center gap-3 overflow-hidden rounded-2xl px-3 py-3 text-left text-slate-600 transition hover:bg-slate-50 hover:text-sky-700">
						<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-sky-100 text-lg text-sky-700">💬</span>
						<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">All Conversations</span>
					</button>
					<button class="flex w-full items-center gap-3 overflow-hidden rounded-2xl px-3 py-3 text-left text-slate-600 transition hover:bg-slate-50 hover:text-sky-700">
						<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-lg">＠</span>
						<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">Mentions</span>
					</button>
					<button class="flex w-full items-center gap-3 overflow-hidden rounded-2xl px-3 py-3 text-left text-slate-600 transition hover:bg-slate-50 hover:text-sky-700">
						<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-lg">⏳</span>
						<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">Waiting</span>
					</button>
				</nav>

				<div class="mt-6 overflow-hidden">
					<p class="px-3 text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400 opacity-0 transition duration-200 group-hover:opacity-100">Inboxes</p>
					<div class="mt-2 space-y-1.5">
						{#each workspace?.instances ?? [] as instance}
							<button class="flex w-full items-center gap-3 overflow-hidden rounded-2xl px-3 py-3 text-left text-slate-600 transition hover:bg-slate-50 hover:text-slate-900">
								<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-xs font-semibold text-slate-600">
									{contactInitials(instance.name)}
								</span>
								<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">{instance.name}</span>
							</button>
						{/each}
					</div>
				</div>

				<div class="mt-auto space-y-1.5 overflow-hidden">
					<button data-testid="open-direct-chat" class="flex w-full items-center gap-3 rounded-2xl bg-sky-600 px-3 py-3 text-left text-white transition hover:bg-sky-700" onclick={openDirectChatDialog}>
						<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-white/15 text-lg">＋</span>
						<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">New conversation</span>
					</button>
					<button class="flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left text-slate-600 transition hover:bg-slate-50 hover:text-slate-900" onclick={() => loadWorkspace()}>
						<span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-lg">↻</span>
						<span class="truncate text-sm font-medium opacity-0 transition duration-200 group-hover:opacity-100">Refresh workspace</span>
					</button>
				</div>
			</div>
		</aside>

		<aside class="flex h-full min-h-0 flex-col overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
			<div class="border-b border-slate-200 px-4 py-4">
				<div class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5">
					<span class="text-slate-400">⌕</span>
					<input bind:value={search} class="w-full bg-transparent text-sm text-slate-700 outline-none placeholder:text-slate-400" placeholder="Search for messages in conversations" />
				</div>
				<div class="mt-4 flex items-center justify-between gap-3">
					<div>
						<h2 class="text-[2rem] font-semibold leading-none text-slate-900">Conversations</h2>
						<p class="mt-1 text-sm text-slate-400">Inbox workspace</p>
					</div>
					<div class="flex items-center gap-2">
						<select bind:value={instanceFilter} class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600 outline-none focus:border-sky-400">
							<option value="">All inboxes</option>
							{#each workspace?.instances ?? [] as instance}
								<option value={instance.id}>{instance.name}</option>
							{/each}
						</select>
						<button class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:border-sky-300 hover:text-sky-700" onclick={applyFilters}>Apply</button>
					</div>
				</div>

				<div class="mt-4 flex items-center gap-2 text-sm">
					<button data-testid="tab-assigned" class={`rounded-xl px-3 py-2 font-medium transition ${currentTab === 'assigned' ? 'bg-sky-50 text-sky-700' : 'text-slate-500 hover:bg-slate-100'}`} onclick={() => openTab('assigned')}>Mine {workspace?.tab_counts.assigned ?? 0}</button>
					<button data-testid="tab-pending" class={`rounded-xl px-3 py-2 font-medium transition ${currentTab === 'pending' ? 'bg-sky-50 text-sky-700' : 'text-slate-500 hover:bg-slate-100'}`} onclick={() => openTab('pending')}>Unassigned {workspace?.tab_counts.pending ?? 0}</button>
					<button data-testid="tab-closed" class={`rounded-xl px-3 py-2 font-medium transition ${currentTab === 'closed' ? 'bg-sky-50 text-sky-700' : 'text-slate-500 hover:bg-slate-100'}`} onclick={() => openTab('closed')}>All {workspace?.tab_counts.closed ?? 0}</button>
				</div>

				<div class="mt-3 flex items-center gap-2">
					<input bind:value={tagFilter} class="min-w-0 flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-sky-400" placeholder="Filter by label" />
					<button class="rounded-xl bg-slate-900 px-3 py-2 text-sm font-medium text-white transition hover:bg-slate-800" onclick={applyFilters}>Filter</button>
				</div>
			</div>

			<div class="min-h-0 flex-1 overflow-y-auto">
				{#each workspace?.conversations ?? [] as contact}
					<button
						data-testid={`conversation-${contact.id}`}
						class={`flex w-full items-start gap-3 border-l-2 px-4 py-4 text-left transition ${contact.id === currentContactId ? 'border-sky-500 bg-sky-50/70' : 'border-transparent hover:bg-slate-50'}`}
						onclick={() => selectContact(contact)}
					>
						<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-sky-100 text-sm font-semibold text-sky-700">
							{contactInitials(contact.name)}
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="truncate text-sm font-semibold text-slate-900">{contactTitle(contact)}</p>
									<p class="mt-0.5 truncate text-xs text-slate-400">{contact.instance_name}</p>
								</div>
								<div class="shrink-0 text-right">
									<p class="text-xs text-slate-400">{contact.last_message_at ? formatDateTime(contact.last_message_at) : ''}</p>
									{#if contact.unread_count > 0}
										<span class="mt-1 inline-flex min-w-6 items-center justify-center rounded-full bg-emerald-500 px-2 py-0.5 text-[11px] font-semibold text-white">{contact.unread_count}</span>
									{/if}
								</div>
							</div>
							<p class="mt-2 truncate text-sm text-slate-600">{contact.last_message_preview || 'No recent message yet.'}</p>
							<div class="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
								<span class={`rounded-full px-2.5 py-1 font-medium uppercase ${statusBadgeClass(contact.status)}`}>{contact.status}</span>
								{#if contact.is_pinned}
									<span class="rounded-full bg-amber-100 px-2.5 py-1 font-medium text-amber-700">Pinned</span>
								{/if}
								{#each contact.tags.slice(0, 2) as tag}
									<span class="rounded-full bg-slate-100 px-2.5 py-1 text-slate-500">{tag}</span>
								{/each}
							</div>
						</div>
					</button>
				{/each}
				{#if (workspace?.conversations?.length ?? 0) === 0}
					<div class="px-4 py-8 text-center text-sm text-slate-500">No conversations match the current filters.</div>
				{/if}
			</div>
		</aside>

		<section class="flex h-full min-h-0 flex-col overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
			{#if loading}
				<div class="flex min-h-[70vh] items-center justify-center text-slate-500">Loading conversation...</div>
			{:else if workspace?.selected}
				<div class="border-b border-slate-200 px-5 py-4">
					<div class="flex flex-wrap items-center justify-between gap-4">
						<div class="flex items-center gap-3">
							{#if workspace.selected.contact.avatar}
								<img class="h-11 w-11 rounded-full object-cover" src={workspace.selected.contact.avatar} alt={workspace.selected.contact.name} />
							{:else}
								<div class="flex h-11 w-11 items-center justify-center rounded-full bg-sky-100 text-sm font-semibold text-sky-700">
									{contactInitials(workspace.selected.contact.name)}
								</div>
							{/if}
							<div>
								<div class="flex items-center gap-2">
									<h3 class="text-lg font-semibold text-slate-900">{contactTitle(workspace.selected.contact)}</h3>
									<span class={`rounded-full px-2.5 py-1 text-[11px] font-medium uppercase ${statusBadgeClass(workspace.selected.contact.status)}`}>{workspace.selected.contact.status}</span>
								</div>
								<p class="mt-1 text-sm text-slate-500">{workspace.selected.contact.phone_display || workspace.selected.contact.phone_number}</p>
							</div>
						</div>

						<div class="flex flex-wrap items-center gap-2">
							<button data-testid="notifications-toggle" class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:border-sky-300 hover:text-sky-700" onclick={() => showNotifications = !showNotifications}>
								Alerts {workspace.notifications.filter((item) => !item.is_read).length}
							</button>
							<button data-testid="statuses-toggle" class="rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:border-sky-300 hover:text-sky-700" onclick={() => showStatuses = !showStatuses}>
								Statuses {workspace.statuses.length}
							</button>
							<button class="rounded-xl bg-emerald-500 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-600" onclick={toggleConversationState}>
								{selectedContact?.status === 'closed' ? 'Reopen' : 'Resolve'}
							</button>
						</div>
					</div>

					<div class="mt-4 flex flex-wrap items-center gap-2">
						<span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-500">{workspace.selected.contact.instance_name}</span>
						{#if workspace.selected.contact.assigned_user_name}
							<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700">Assigned to {workspace.selected.contact.assigned_user_name}</span>
						{/if}
						{#each workspace.selected.contact.tags as tag}
							<span class="rounded-full bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700">{tag}</span>
						{/each}
					</div>

					<div class="mt-4 grid gap-2 md:grid-cols-[minmax(0,1fr)_repeat(4,auto)]">
						<select bind:value={selectedAssignee} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-sky-400">
							<option value="">Select assignee</option>
							{#each workspace.users as user}
								<option value={user.id}>{user.name}</option>
							{/each}
						</select>
						<button data-testid="assign-action" class="rounded-xl bg-sky-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-sky-700" onclick={assignConversation}>Assign</button>
						<button data-testid="unassign-action" class="rounded-xl border border-slate-200 px-4 py-2 text-sm text-slate-700 transition hover:bg-slate-50" onclick={unassignConversation}>Unassign</button>
						<button data-testid="pin-action" class="rounded-xl border border-slate-200 px-4 py-2 text-sm text-slate-700 transition hover:bg-slate-50" onclick={togglePinnedConversation}>
							{selectedContact?.is_pinned ? 'Unpin' : 'Pin'}
						</button>
						<button data-testid="hide-action" class="rounded-xl border border-slate-200 px-4 py-2 text-sm text-slate-700 transition hover:bg-slate-50" onclick={toggleHiddenConversation}>
							{selectedContact?.is_hidden ? 'Unhide' : 'Hide'}
						</button>
					</div>
				</div>

				<div class="flex min-h-0 flex-1 flex-col justify-between">
					<div class="min-h-0 flex-1 overflow-y-auto px-5 py-6" style={`${chatBackgroundStyle()} background-color: #f8fafc;`}>
						<div class="space-y-4">
							{#each workspace.selected.messages as message}
								<div class={`flex ${message.direction === 'outbound' ? 'justify-end' : 'justify-start'}`}>
									<div class={`max-w-[78%] rounded-[22px] px-4 py-3 shadow-sm ${message.direction === 'outbound' ? 'bg-[#5b3fd3] text-white' : 'border border-slate-200 bg-white text-slate-800'}`}>
										<div class="mb-2 flex items-center justify-between gap-4 text-[11px] opacity-75">
											<span>{message.type === 'media' ? 'Attachment' : 'Message'}</span>
											<span>{formatDateTime(message.created_at)}</span>
										</div>

										{#if message.type === 'media'}
											<div class={`rounded-2xl border px-3 py-3 ${message.direction === 'outbound' ? 'border-white/15 bg-white/10' : 'border-slate-200 bg-slate-50'}`}>
												{#if message.media_url && (message.media_url.startsWith('data:image/') || message.media_url.startsWith('http'))}
													<img class="mb-3 max-h-56 w-full rounded-xl object-cover" src={message.media_url} alt={message.file_name || 'Attachment preview'} />
												{/if}
												<p class="font-medium">{message.file_name}</p>
												<p class={`text-xs ${message.direction === 'outbound' ? 'text-white/80' : 'text-slate-500'}`}>{message.file_size_label || 'Attachment preview'}</p>
											</div>
										{/if}

										{#if message.body}
											<p class="mt-2 whitespace-pre-wrap text-sm leading-6">{message.body}</p>
										{/if}

										<div class="mt-3 flex flex-wrap items-center gap-1.5 text-[11px]">
											<span class={`rounded-full px-2 py-0.5 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-slate-100 text-slate-600'}`}>{message.status}</span>
											{#if message.typed_for_ms}
												<span class={`rounded-full px-2 py-0.5 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-slate-100 text-slate-600'}`}>Typing {message.typed_for_ms}ms</span>
											{/if}
											{#if message.reaction}
												<span class={`rounded-full px-2 py-0.5 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-slate-100 text-slate-600'}`}>{message.reaction}</span>
											{/if}
										</div>

										{#if message.failure_reason}
											<p class="mt-2 text-xs font-medium text-amber-200">{message.failure_reason}</p>
										{/if}

										<div class="mt-3 flex flex-nowrap items-center gap-1.5 overflow-x-auto pb-0.5">
											{#if workspace.settings.chat.show_print_buttons}
												<button class={`shrink-0 rounded-full border px-2.5 py-1 text-[11px] ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-slate-200 text-slate-600'}`} onclick={() => printMessage(message)}>Print</button>
											{/if}
											{#if workspace.settings.chat.show_download_buttons && message.type === 'media'}
												<button class={`shrink-0 rounded-full border px-2.5 py-1 text-[11px] ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-slate-200 text-slate-600'}`} onclick={() => downloadMessage(message)}>Download</button>
											{/if}
											{#if message.can_retry && message.status === 'failed'}
												<button data-testid={`retry-${message.id}`} class={`shrink-0 rounded-full border px-2.5 py-1 text-[11px] ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-slate-200 text-slate-600'}`} onclick={() => runAction(`/api/chats/${currentContactId}/messages/${message.id}/retry`, undefined, 'Retry requested.')}>Retry</button>
											{/if}
											{#if message.can_revoke && message.direction === 'outbound' && message.status !== 'revoked'}
												<button data-testid={`revoke-${message.id}`} class={`shrink-0 rounded-full border px-2.5 py-1 text-[11px] ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-slate-200 text-slate-600'}`} onclick={() => runAction(`/api/chats/${currentContactId}/messages/${message.id}/revoke`, undefined, 'Message revoked.')}>Revoke</button>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<div class="border-t border-slate-200 bg-white px-5 py-4">
						<div class="mb-3 flex flex-wrap items-center gap-2">
							<button data-testid="composer-text" class={`rounded-xl px-3 py-1.5 text-sm font-medium ${composerMode === 'text' ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-600'}`} onclick={() => composerMode = 'text'}>Reply</button>
							<button data-testid="composer-media" class={`rounded-xl px-3 py-1.5 text-sm font-medium ${composerMode === 'media' ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-600'}`} onclick={() => composerMode = 'media'}>Attachment</button>
							<button data-testid="quick-replies-toggle" class="rounded-xl border border-slate-200 px-3 py-1.5 text-sm text-slate-600" onclick={() => showQuickReplies = !showQuickReplies}>Quick Replies</button>
						</div>

						{#if showQuickReplies}
							<div class="mb-3 grid gap-2 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-3 md:grid-cols-3">
								{#each workspace.quick_replies as quickReply}
									<button class="rounded-2xl border border-slate-200 bg-white px-3 py-3 text-left" onclick={() => pickQuickReply(quickReply.body)}>
										<p class="text-xs font-semibold uppercase tracking-wide text-sky-600">{quickReply.shortcut}</p>
										<p class="mt-1 text-sm font-medium text-slate-900">{quickReply.title}</p>
										<p class="mt-1 text-sm text-slate-500 line-clamp-2">{quickReply.body}</p>
									</button>
								{/each}
							</div>
						{/if}

						{#if composerMode === 'text'}
							<textarea data-testid="composer-textarea" bind:value={composerText} class="min-h-[120px] w-full rounded-[22px] border border-slate-200 px-4 py-4 text-sm outline-none focus:border-sky-400" placeholder="Type a reply..."></textarea>
						{:else}
							<div class="space-y-3">
								<input bind:this={attachmentInput} data-testid="attachment-file-input" class="hidden" type="file" onchange={handleAttachmentChange} />
								<div
									data-testid="attachment-dropzone"
									class={`rounded-[22px] border-2 border-dashed px-4 py-5 text-sm transition ${isDragOver ? 'border-sky-400 bg-sky-50' : 'border-slate-200 bg-slate-50'}`}
									role="button"
									tabindex="0"
									onclick={() => attachmentInput?.click()}
									onkeydown={(event) => (event.key === 'Enter' || event.key === ' ') && attachmentInput?.click()}
									ondragenter={() => isDragOver = true}
									ondragleave={() => isDragOver = false}
									ondragover={(event) => event.preventDefault()}
									ondrop={handleAttachmentDrop}
								>
									<p class="font-medium text-slate-900">Drop a file here or click to choose one</p>
									<p class="mt-1 text-slate-500">Media messages are sent without typing simulation.</p>
								</div>

								{#if attachmentName}
									<div class="rounded-[1.5rem] border border-slate-200 bg-white px-4 py-4">
										<div class="flex items-start justify-between gap-3">
											<div class="min-w-0">
												<p data-testid="attachment-preview-name" class="truncate text-sm font-medium text-slate-900">{attachmentName}</p>
												<p class="mt-1 text-xs text-slate-500">{attachmentFileSizeLabel}</p>
											</div>
											<button class="rounded-full border border-slate-200 px-3 py-1.5 text-xs text-slate-600" onclick={clearAttachment}>Remove</button>
										</div>
										{#if attachmentUrl && attachmentPreviewMime.startsWith('image/')}
											<img class="mt-3 max-h-56 w-full rounded-xl object-cover" src={attachmentUrl} alt={attachmentName} />
										{/if}
									</div>
								{/if}

								<textarea data-testid="composer-caption" bind:value={composerText} class="min-h-[96px] w-full rounded-[22px] border border-slate-200 px-4 py-4 text-sm outline-none focus:border-sky-400" placeholder="Optional caption for the attachment..."></textarea>
							</div>
						{/if}

						<div class="mt-4 flex items-center justify-between gap-3">
							<p class="text-xs text-slate-500">
								{#if composerMode === 'text'}
									Text replies follow the current typing simulation flow.
								{:else}
									Media can still be retried independently after send.
								{/if}
							</p>
							<button data-testid="send-message" disabled={saving} class="rounded-xl bg-[#f8b833] px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-[#e6a61f] disabled:opacity-50" onclick={sendMessage}>Send</button>
						</div>
					</div>
				</div>
			{:else}
				<div class="flex min-h-[70vh] items-center justify-center text-slate-500">Select a conversation from the inbox.</div>
			{/if}
		</section>

		<aside class="hidden h-full min-h-0 flex-col gap-3 xl:flex">
			{#if workspace?.selected}
				<section class="overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
					<div class="border-b border-slate-200 px-5 py-5">
						<div class="flex items-start justify-between gap-3">
							<div class="flex items-center gap-3">
								{#if workspace.selected.contact.avatar}
									<img class="h-14 w-14 rounded-full object-cover" src={workspace.selected.contact.avatar} alt={workspace.selected.contact.name} />
								{:else}
									<div class="flex h-14 w-14 items-center justify-center rounded-full bg-sky-100 text-lg font-semibold text-sky-700">
										{contactInitials(workspace.selected.contact.name)}
									</div>
								{/if}
								<div>
									<h3 class="text-lg font-semibold text-slate-900">{contactTitle(workspace.selected.contact)}</h3>
									<p class="mt-1 text-sm text-slate-500">{workspace.selected.contact.phone_display || workspace.selected.contact.phone_number}</p>
									<p class="mt-1 text-sm text-slate-400">{workspace.selected.contact.instance_source_label}</p>
								</div>
							</div>
							<button class="rounded-full border border-slate-200 px-2.5 py-1.5 text-slate-400 transition hover:text-slate-700" onclick={() => showInfo = !showInfo}>›</button>
						</div>
						<div class="mt-4 flex items-center gap-2">
							<button data-testid="notes-toggle" class={`rounded-xl px-3 py-2 text-sm font-medium ${showNotes ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-600'}`} onclick={() => showNotes = !showNotes}>Notes</button>
							<button data-testid="info-toggle" class={`rounded-xl px-3 py-2 text-sm font-medium ${showInfo ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-600'}`} onclick={() => showInfo = !showInfo}>Info</button>
							<button data-testid="timeline-toggle" class={`rounded-xl px-3 py-2 text-sm font-medium ${showTimeline ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-600'}`} onclick={() => showTimeline = !showTimeline}>Timeline</button>
						</div>
					</div>
				</section>
			{/if}

			{#if workspace?.selected && showNotes}
				<section class="overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
					<button class="flex w-full items-center justify-between border-b border-slate-200 px-5 py-4 text-left" onclick={() => showNotes = !showNotes}>
						<h3 class="text-base font-semibold text-slate-900">Conversation Notes</h3>
						<span class="text-xs text-slate-400">{workspace.selected.notes.length}</span>
					</button>
					<div class="space-y-3 p-4">
						{#each workspace.selected.notes as note}
							<div class="rounded-[20px] bg-slate-50 px-4 py-3">
								<p class="text-sm text-slate-800">{note.body}</p>
								<p class="mt-2 text-xs text-slate-500">{note.user_name} · {formatDateTime(note.created_at)}</p>
							</div>
						{/each}
						<textarea bind:value={noteDraft} class="min-h-[100px] w-full rounded-[20px] border border-slate-200 px-4 py-3 text-sm outline-none focus:border-sky-400" placeholder="Add an internal note..."></textarea>
						<button data-testid="add-note" class="w-full rounded-xl bg-sky-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-sky-700" onclick={addNote}>Add Note</button>
					</div>
				</section>
			{/if}

			{#if workspace?.selected && showInfo}
				<section class="overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
					<button class="flex w-full items-center justify-between border-b border-slate-200 px-5 py-4 text-left" onclick={() => showInfo = !showInfo}>
						<h3 class="text-base font-semibold text-slate-900">Conversation Information</h3>
						<span class="text-slate-400">+</span>
					</button>
					<div class="space-y-3 p-4 text-sm text-slate-600">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Phone</p>
							<p class="mt-1">{workspace.selected.contact.phone_display || workspace.selected.contact.phone_number}</p>
						</div>
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Tags</p>
							<div class="mt-2 flex flex-wrap gap-2">
								{#each workspace.selected.contact.tags as tag}
									<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs text-emerald-700">{tag}</span>
								{/each}
							</div>
						</div>
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Collaborators</p>
							<div class="mt-2 space-y-2">
								{#each workspace.selected.collaborators as collaborator}
									<div class="rounded-[1.25rem] bg-slate-50 px-3 py-2">
										<p class="font-medium text-slate-800">{collaborator.user_name}</p>
										<p class="text-xs text-slate-500">{collaborator.status}</p>
									</div>
								{/each}
							</div>
						</div>
						<select bind:value={collaboratorUserId} class="w-full rounded-[1.5rem] border border-slate-200 px-3 py-2 text-sm outline-none focus:border-sky-400">
							<option value="">Invite collaborator</option>
							{#each workspace.users as user}
								<option value={user.id}>{user.name}</option>
							{/each}
						</select>
						<button data-testid="invite-collaborator" class="w-full rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:bg-slate-50" onclick={inviteCollaborator}>Invite Collaborator</button>
					</div>
				</section>
			{/if}

			{#if workspace?.selected && showTimeline}
				<section class="overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-[0_18px_60px_rgba(15,23,42,0.08)]">
					<button class="flex w-full items-center justify-between border-b border-slate-200 px-5 py-4 text-left" onclick={() => showTimeline = !showTimeline}>
						<h3 class="text-base font-semibold text-slate-900">Previous Conversations</h3>
						<span class="text-xs text-slate-400">{workspace.selected.events.length}</span>
					</button>
					<div class="space-y-3 p-4">
						{#each workspace.selected.events as event}
							<div class="rounded-[1.5rem] bg-slate-50 px-4 py-3">
								<p class="font-medium text-slate-800">{event.summary}</p>
								<p class="mt-1 text-xs text-slate-500">{event.actor_name} · {formatDateTime(event.occurred_at)}</p>
							</div>
						{/each}
					</div>
				</section>
			{/if}
		</aside>
	</div>

	{#if showDirectChatDialog}
		<div
			class="fixed inset-0 z-40 bg-gray-900/20 backdrop-blur-sm"
			role="button"
			tabindex="0"
			aria-label="Close direct chat dialog"
			onclick={closeDirectChatDialog}
			onkeydown={(event) => event.key === 'Escape' && closeDirectChatDialog()}
		></div>
		<div class="fixed inset-x-4 top-[10vh] z-50 mx-auto max-w-lg rounded-[2rem] border border-gray-200 bg-white p-6 shadow-2xl">
			<div class="flex items-start justify-between gap-4">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Start New Chat</p>
					<h3 class="mt-2 text-2xl font-semibold text-gray-900">Create a direct conversation</h3>
					<p class="mt-2 text-sm text-gray-500">Match the reference inbox flow by creating a phone-based chat directly from the workspace.</p>
				</div>
				<button class="rounded-full border border-gray-200 px-3 py-1.5 text-sm text-gray-600" onclick={closeDirectChatDialog}>Close</button>
			</div>

			<div class="mt-5 grid gap-4">
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Phone Number</span>
					<input data-testid="direct-chat-phone" bind:value={newContactPhone} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="+201234567890" />
				</label>
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Profile Name</span>
					<input data-testid="direct-chat-name" bind:value={newContactName} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Optional display name" />
				</label>
				<label class="block">
					<span class="mb-2 block text-sm font-medium text-gray-700">Sending Account</span>
					<select data-testid="direct-chat-instance" bind:value={newContactInstanceId} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
						<option value="">Select an account</option>
						{#each workspace?.instances ?? [] as instance}
							<option value={instance.id}>{instance.name}</option>
						{/each}
					</select>
				</label>
			</div>

			<div class="mt-6 flex items-center justify-end gap-3">
				<button class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700" onclick={closeDirectChatDialog}>Cancel</button>
				<button data-testid="create-direct-chat" disabled={creatingDirectChat} class="rounded-full bg-blue-600 px-5 py-2.5 text-sm font-medium text-white disabled:opacity-50" onclick={createDirectChat}>Create Chat</button>
			</div>
		</div>
	{/if}

	{#if showNotifications}
		<div
			class="fixed inset-0 z-40 bg-gray-900/20 backdrop-blur-sm"
			role="button"
			tabindex="0"
			aria-label="Close notifications panel"
			onclick={() => showNotifications = false}
			onkeydown={(event) => event.key === 'Escape' && (showNotifications = false)}
		>
			<div class="absolute right-6 top-24 w-full max-w-md rounded-[2rem] border border-gray-200 bg-white p-5 shadow-2xl" role="dialog" tabindex="-1" aria-modal="true" onclick={(event) => event.stopPropagation()} onkeydown={(event) => event.stopPropagation()}>
				<div class="mb-4 flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Notifications</h3>
					<button class="rounded-full border border-gray-200 px-3 py-1.5 text-sm text-gray-600" onclick={markNotificationsRead}>Mark all as read</button>
				</div>
				<div class="space-y-3">
					{#each workspace?.notifications ?? [] as item}
						<button class={`w-full rounded-[1.5rem] border px-4 py-3 text-left ${item.is_read ? 'border-gray-200 bg-white' : 'border-blue-200 bg-blue-50'}`} onclick={() => item.related_contact_id ? goto(workspacePath(item.related_contact_id)) : undefined}>
							<p class="font-medium text-gray-900">{item.title}</p>
							<p class="mt-1 text-sm text-gray-600">{item.body}</p>
							<p class="mt-2 text-xs text-gray-400">{formatDateTime(item.created_at)}</p>
						</button>
					{/each}
				</div>
			</div>
		</div>
	{/if}

	{#if showStatuses}
		<div
			class="fixed inset-0 z-40 bg-gray-900/20 backdrop-blur-sm"
			role="button"
			tabindex="0"
			aria-label="Close statuses panel"
			onclick={() => showStatuses = false}
			onkeydown={(event) => event.key === 'Escape' && (showStatuses = false)}
		>
			<div class="absolute right-6 top-24 w-full max-w-md rounded-[2rem] border border-gray-200 bg-white p-5 shadow-2xl" role="dialog" tabindex="-1" aria-modal="true" onclick={(event) => event.stopPropagation()} onkeydown={(event) => event.stopPropagation()}>
				<div class="mb-4 flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Statuses</h3>
					<span class="text-xs text-gray-400">{workspace?.statuses.length ?? 0} items</span>
				</div>
				<textarea bind:value={statusDraft} class="mb-3 min-h-[100px] w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Add a status update for the current conversation or instance..."></textarea>
				<button data-testid="add-status" class="mb-4 w-full rounded-full bg-blue-600 px-4 py-2.5 text-sm font-medium text-white" onclick={addStatus}>Post Status</button>
				<div class="space-y-3">
					{#each workspace?.statuses ?? [] as item}
						<div class="rounded-[1.5rem] border border-gray-200 px-4 py-3">
							<div class="flex items-center justify-between gap-2">
								<p class="font-medium text-gray-900">{item.contact_name}</p>
								<span class="text-xs text-gray-400">{item.instance_name}</span>
							</div>
							<p class="mt-1 text-sm text-gray-600">{item.body}</p>
							<p class="mt-2 text-xs text-gray-400">{formatDateTime(item.created_at)}</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}
</div>
