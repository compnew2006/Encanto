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

<div class="px-5 py-5">
	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if infoMessage}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{infoMessage}</div>
	{/if}

	<div class="grid min-h-[78vh] grid-cols-12 gap-4">
		<aside class="col-span-12 lg:col-span-3 rounded-[2rem] border border-gray-200 bg-white p-4 shadow-sm">
			<div class="mb-4 flex items-center justify-between gap-3">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Conversation Workspace</p>
					<h2 class="text-2xl font-semibold text-gray-900">Inbox</h2>
				</div>
				<div class="flex items-center gap-2">
					<button data-testid="open-direct-chat" class="rounded-full bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700" onclick={openDirectChatDialog}>
						Start New Chat
					</button>
					<button class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => loadWorkspace()}>
						Refresh
					</button>
				</div>
			</div>

			<div class="mb-3 grid grid-cols-3 gap-2">
				<button data-testid="tab-assigned" class={`rounded-2xl px-3 py-2 text-sm font-medium ${currentTab === 'assigned' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`} onclick={() => openTab('assigned')}>
					Assigned
					<span class="block text-xs opacity-80">{workspace?.tab_counts.assigned ?? 0}</span>
				</button>
				<button data-testid="tab-pending" class={`rounded-2xl px-3 py-2 text-sm font-medium ${currentTab === 'pending' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`} onclick={() => openTab('pending')}>
					Pending
					<span class="block text-xs opacity-80">{workspace?.tab_counts.pending ?? 0}</span>
				</button>
				<button data-testid="tab-closed" class={`rounded-2xl px-3 py-2 text-sm font-medium ${currentTab === 'closed' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`} onclick={() => openTab('closed')}>
					Closed
					<span class="block text-xs opacity-80">{workspace?.tab_counts.closed ?? 0}</span>
				</button>
			</div>

			<div class="space-y-2 border-b border-gray-100 pb-4">
				<input bind:value={search} class="w-full rounded-2xl border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Search contacts..." />
				<div class="grid grid-cols-2 gap-2">
					<select bind:value={instanceFilter} class="rounded-2xl border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400">
						<option value="">All instances</option>
						{#each workspace?.instances ?? [] as instance}
							<option value={instance.id}>{instance.name}</option>
						{/each}
					</select>
					<input bind:value={tagFilter} class="rounded-2xl border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Tag" />
				</div>
				<button class="w-full rounded-2xl bg-gray-900 px-3 py-2 text-sm font-medium text-white" onclick={applyFilters}>Apply Filters</button>
			</div>

			<div class="mt-4 space-y-2 overflow-y-auto">
				{#each workspace?.conversations ?? [] as contact}
					<button
						data-testid={`conversation-${contact.id}`}
						class={`w-full rounded-[1.5rem] border px-3 py-3 text-left transition ${contact.id === currentContactId ? 'border-blue-400 bg-blue-50' : 'border-gray-200 bg-white hover:border-blue-200 hover:bg-gray-50'}`}
						onclick={() => selectContact(contact)}
					>
						<div class="mb-2 flex items-start justify-between gap-2">
							<div>
								<p class="font-medium text-gray-900">{contact.name}</p>
								<p class="text-xs text-gray-500">{contact.phone_display || contact.phone_number}</p>
							</div>
							<div class="flex items-center gap-1">
								{#if contact.is_pinned}
									<span class="rounded-full bg-amber-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-amber-700">Pinned</span>
								{/if}
								{#if contact.unread_count > 0}
									<span class="rounded-full bg-blue-600 px-2 py-1 text-[10px] font-semibold text-white">{contact.unread_count}</span>
								{/if}
							</div>
						</div>
						<p class="text-sm text-gray-600 line-clamp-2">{contact.last_message_preview}</p>
						<div class="mt-3 flex flex-wrap items-center gap-2 text-[11px] text-gray-500">
							<span class="rounded-full bg-gray-100 px-2 py-1 uppercase">{contact.status}</span>
							<span>{contact.instance_name}</span>
							{#each contact.tags as tag}
								<span class="rounded-full bg-emerald-50 px-2 py-1 text-emerald-700">{tag}</span>
							{/each}
						</div>
					</button>
				{/each}
				{#if (workspace?.conversations?.length ?? 0) === 0}
					<div class="rounded-[1.5rem] border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500">
						No conversations match the current tab and filters.
					</div>
				{/if}
			</div>
		</aside>

		<section class="col-span-12 lg:col-span-6 rounded-[2rem] border border-gray-200 bg-white shadow-sm">
			{#if loading}
				<div class="flex min-h-[70vh] items-center justify-center text-gray-500">Loading conversation...</div>
			{:else if workspace?.selected}
				<div class="border-b border-gray-100 px-5 py-4">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div>
							<div class="flex items-center gap-3">
								<img class="h-12 w-12 rounded-2xl object-cover" src={workspace.selected.contact.avatar} alt={workspace.selected.contact.name} />
								<div>
									<h3 class="text-xl font-semibold text-gray-900">{workspace.selected.contact.name}</h3>
									<p class="text-sm text-gray-500">{workspace.selected.contact.phone_display || workspace.selected.contact.phone_number}</p>
								</div>
							</div>
							<div class="mt-3 flex flex-wrap gap-2 text-xs text-gray-500">
								<span class="rounded-full bg-blue-50 px-3 py-1 text-blue-700">{workspace.selected.contact.instance_name}</span>
								<span class="rounded-full bg-gray-100 px-3 py-1">{workspace.selected.contact.status}</span>
								{#if workspace.selected.contact.assigned_user_name}
									<span class="rounded-full bg-emerald-50 px-3 py-1 text-emerald-700">Assigned to {workspace.selected.contact.assigned_user_name}</span>
								{/if}
							</div>
						</div>

						<div class="flex flex-wrap items-center gap-2">
							<button data-testid="notifications-toggle" class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => showNotifications = !showNotifications}>
								Notifications ({workspace.notifications.filter((item) => !item.is_read).length})
							</button>
							<button data-testid="statuses-toggle" class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => showStatuses = !showStatuses}>
								Statuses ({workspace.statuses.length})
							</button>
							<button data-testid="notes-toggle" class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => showNotes = !showNotes}>Notes</button>
							<button data-testid="info-toggle" class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => showInfo = !showInfo}>Info</button>
							<button data-testid="timeline-toggle" class="rounded-full border border-gray-200 px-3 py-2 text-sm text-gray-600 hover:border-blue-300 hover:text-blue-700" onclick={() => showTimeline = !showTimeline}>Timeline</button>
						</div>
					</div>

					<div class="mt-4 grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto_auto_auto]">
						<select bind:value={selectedAssignee} class="rounded-2xl border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400">
							<option value="">Select assignee</option>
							{#each workspace.users as user}
								<option value={user.id}>{user.name}</option>
							{/each}
						</select>
						<button data-testid="assign-action" class="rounded-2xl bg-blue-600 px-4 py-2 text-sm font-medium text-white" onclick={assignConversation}>Assign</button>
						<button data-testid="unassign-action" class="rounded-2xl border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={unassignConversation}>Unassign</button>
						<button data-testid="pin-action" class="rounded-2xl border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={togglePinnedConversation}>
							{selectedContact?.is_pinned ? 'Unpin' : 'Pin'}
						</button>
						<button data-testid="hide-action" class="rounded-2xl border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={toggleHiddenConversation}>
							{selectedContact?.is_hidden ? 'Unhide' : 'Hide'}
						</button>
						<button data-testid="close-action" class="rounded-2xl border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={toggleConversationState}>
							{selectedContact?.status === 'closed' ? 'Reopen' : 'Close'}
						</button>
					</div>
				</div>

				<div class="flex min-h-[56vh] flex-col justify-between">
					<div class="flex-1 overflow-y-auto px-5 py-5" style={chatBackgroundStyle()}>
						<div class="space-y-4">
							{#each workspace.selected.messages as message}
								<div class={`flex ${message.direction === 'outbound' ? 'justify-end' : 'justify-start'}`}>
									<div class={`max-w-[78%] rounded-[1.75rem] px-4 py-3 shadow-sm ${message.direction === 'outbound' ? 'bg-blue-600 text-white' : 'bg-white text-gray-800 border border-gray-200'}`}>
										<div class="mb-2 flex items-center justify-between gap-4 text-[11px] opacity-80">
											<span>{message.type === 'media' ? 'Attachment' : 'Message'}</span>
											<span>{formatDateTime(message.created_at)}</span>
										</div>

										{#if message.type === 'media'}
											<div class={`rounded-2xl border px-3 py-3 ${message.direction === 'outbound' ? 'border-blue-300/40 bg-white/10' : 'border-gray-200 bg-gray-50'}`}>
												{#if message.media_url && (message.media_url.startsWith('data:image/') || message.media_url.startsWith('http'))}
													<img class="mb-3 max-h-56 w-full rounded-xl object-cover" src={message.media_url} alt={message.file_name || 'Attachment preview'} />
												{/if}
												<p class="font-medium">{message.file_name}</p>
												<p class={`text-xs ${message.direction === 'outbound' ? 'text-blue-50' : 'text-gray-500'}`}>{message.file_size_label || 'Attachment preview'}</p>
											</div>
										{/if}

										{#if message.body}
											<p class="mt-2 whitespace-pre-wrap text-sm leading-6">{message.body}</p>
										{/if}

										<div class="mt-3 flex flex-wrap items-center gap-2 text-xs">
											<span class={`rounded-full px-2.5 py-1 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-gray-100 text-gray-600'}`}>{message.status}</span>
											{#if message.typed_for_ms}
												<span class={`rounded-full px-2.5 py-1 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-gray-100 text-gray-600'}`}>Typing {message.typed_for_ms}ms</span>
											{/if}
											{#if message.reaction}
												<span class={`rounded-full px-2.5 py-1 ${message.direction === 'outbound' ? 'bg-white/15' : 'bg-gray-100 text-gray-600'}`}>{message.reaction}</span>
											{/if}
										</div>

										{#if message.failure_reason}
											<p class="mt-2 text-xs font-medium text-amber-200">{message.failure_reason}</p>
										{/if}

										<div class="mt-3 flex flex-wrap gap-2">
											{#if workspace.settings.chat.show_print_buttons}
												<button class={`rounded-full border px-3 py-1.5 text-xs ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-gray-200 text-gray-600'}`} onclick={() => printMessage(message)}>Print</button>
											{/if}
											{#if workspace.settings.chat.show_download_buttons && message.type === 'media'}
												<button class={`rounded-full border px-3 py-1.5 text-xs ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-gray-200 text-gray-600'}`} onclick={() => downloadMessage(message)}>Download</button>
											{/if}
											{#if message.can_retry && message.status === 'failed'}
												<button data-testid={`retry-${message.id}`} class={`rounded-full border px-3 py-1.5 text-xs ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-gray-200 text-gray-600'}`} onclick={() => runAction(`/api/chats/${currentContactId}/messages/${message.id}/retry`, undefined, 'Retry requested.')}>Retry</button>
											{/if}
											{#if message.can_revoke && message.direction === 'outbound' && message.status !== 'revoked'}
												<button data-testid={`revoke-${message.id}`} class={`rounded-full border px-3 py-1.5 text-xs ${message.direction === 'outbound' ? 'border-white/30 text-white' : 'border-gray-200 text-gray-600'}`} onclick={() => runAction(`/api/chats/${currentContactId}/messages/${message.id}/revoke`, undefined, 'Message revoked.')}>Revoke</button>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<div class="border-t border-gray-100 px-5 py-4">
						<div class="mb-3 flex flex-wrap items-center gap-2">
							<button data-testid="composer-text" class={`rounded-full px-3 py-1.5 text-sm ${composerMode === 'text' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`} onclick={() => composerMode = 'text'}>Text</button>
							<button data-testid="composer-media" class={`rounded-full px-3 py-1.5 text-sm ${composerMode === 'media' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`} onclick={() => composerMode = 'media'}>Attachment</button>
							<button data-testid="quick-replies-toggle" class="rounded-full border border-gray-200 px-3 py-1.5 text-sm text-gray-600" onclick={() => showQuickReplies = !showQuickReplies}>Quick Replies</button>
						</div>

						{#if showQuickReplies}
							<div class="mb-3 grid gap-2 rounded-[1.5rem] border border-gray-200 bg-gray-50 p-3 md:grid-cols-3">
								{#each workspace.quick_replies as quickReply}
									<button class="rounded-2xl border border-gray-200 bg-white px-3 py-3 text-left" onclick={() => pickQuickReply(quickReply.body)}>
										<p class="text-xs font-semibold uppercase tracking-wide text-blue-600">{quickReply.shortcut}</p>
										<p class="mt-1 text-sm font-medium text-gray-900">{quickReply.title}</p>
										<p class="mt-1 text-sm text-gray-500 line-clamp-2">{quickReply.body}</p>
									</button>
								{/each}
							</div>
						{/if}

						{#if composerMode === 'text'}
							<textarea data-testid="composer-textarea" bind:value={composerText} class="min-h-[120px] w-full rounded-[1.75rem] border border-gray-200 px-4 py-4 text-sm outline-none focus:border-blue-400" placeholder="Write a reply. Text sends use typing simulation based on message length."></textarea>
						{:else}
							<div class="space-y-3">
								<input bind:this={attachmentInput} data-testid="attachment-file-input" class="hidden" type="file" onchange={handleAttachmentChange} />
								<div
									data-testid="attachment-dropzone"
									class={`rounded-[1.75rem] border-2 border-dashed px-4 py-5 text-sm transition ${isDragOver ? 'border-blue-400 bg-blue-50' : 'border-gray-200 bg-gray-50'}`}
									role="button"
									tabindex="0"
									onclick={() => attachmentInput?.click()}
									onkeydown={(event) => (event.key === 'Enter' || event.key === ' ') && attachmentInput?.click()}
									ondragenter={() => isDragOver = true}
									ondragleave={() => isDragOver = false}
									ondragover={(event) => event.preventDefault()}
									ondrop={handleAttachmentDrop}
								>
									<p class="font-medium text-gray-900">Drop a file here or click to choose one</p>
									<p class="mt-1 text-gray-500">Media sends immediately without typing simulation, matching the confirmed workflow.</p>
								</div>

								{#if attachmentName}
									<div class="rounded-[1.5rem] border border-gray-200 bg-white px-4 py-4">
										<div class="flex items-start justify-between gap-3">
											<div class="min-w-0">
												<p data-testid="attachment-preview-name" class="truncate text-sm font-medium text-gray-900">{attachmentName}</p>
												<p class="mt-1 text-xs text-gray-500">{attachmentFileSizeLabel}</p>
											</div>
											<button class="rounded-full border border-gray-200 px-3 py-1.5 text-xs text-gray-600" onclick={clearAttachment}>Remove</button>
										</div>
										{#if attachmentUrl && attachmentPreviewMime.startsWith('image/')}
											<img class="mt-3 max-h-56 w-full rounded-xl object-cover" src={attachmentUrl} alt={attachmentName} />
										{/if}
									</div>
								{/if}

								<textarea data-testid="composer-caption" bind:value={composerText} class="min-h-[96px] w-full rounded-[1.75rem] border border-gray-200 px-4 py-4 text-sm outline-none focus:border-blue-400" placeholder="Optional caption for the attachment..."></textarea>
							</div>
						{/if}

						<div class="mt-4 flex items-center justify-between gap-3">
							<p class="text-xs text-gray-500">
								{#if composerMode === 'text'}
									Text messages simulate provider typing.
								{:else}
									Media bypasses the typing path and can be retried independently.
								{/if}
							</p>
							<button data-testid="send-message" disabled={saving} class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white disabled:opacity-50" onclick={sendMessage}>Send</button>
						</div>
					</div>
				</div>
			{:else}
				<div class="flex min-h-[70vh] items-center justify-center text-gray-500">Select a conversation from the inbox.</div>
			{/if}
		</section>

		<aside class="col-span-12 lg:col-span-3 space-y-4">
			{#if workspace?.selected && showNotes}
				<section class="rounded-[2rem] border border-gray-200 bg-white p-4 shadow-sm">
					<div class="mb-3 flex items-center justify-between">
						<h3 class="text-lg font-semibold text-gray-900">Notes</h3>
						<span class="text-xs text-gray-400">{workspace.selected.notes.length}</span>
					</div>
					<div class="space-y-3">
						{#each workspace.selected.notes as note}
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-sm text-gray-800">{note.body}</p>
								<p class="mt-2 text-xs text-gray-500">{note.user_name} · {formatDateTime(note.created_at)}</p>
							</div>
						{/each}
					</div>
					<textarea bind:value={noteDraft} class="mt-4 min-h-[100px] w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Add an internal note..."></textarea>
					<button data-testid="add-note" class="mt-3 w-full rounded-full bg-blue-600 px-4 py-2.5 text-sm font-medium text-white" onclick={addNote}>Add Note</button>
				</section>
			{/if}

			{#if workspace?.selected && showInfo}
				<section class="rounded-[2rem] border border-gray-200 bg-white p-4 shadow-sm">
					<div class="mb-3 flex items-center justify-between">
						<h3 class="text-lg font-semibold text-gray-900">Contact Info</h3>
						<span class="rounded-full bg-gray-100 px-3 py-1 text-xs text-gray-500">{workspace.selected.contact.instance_source_label}</span>
					</div>
					<div class="space-y-3 text-sm text-gray-600">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Phone</p>
							<p class="mt-1">{workspace.selected.contact.phone_display || workspace.selected.contact.phone_number}</p>
						</div>
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Tags</p>
							<div class="mt-2 flex flex-wrap gap-2">
								{#each workspace.selected.contact.tags as tag}
									<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs text-emerald-700">{tag}</span>
								{/each}
							</div>
						</div>
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Collaborators</p>
							<div class="mt-2 space-y-2">
								{#each workspace.selected.collaborators as collaborator}
									<div class="rounded-[1.25rem] bg-gray-50 px-3 py-2">
										<p class="font-medium text-gray-800">{collaborator.user_name}</p>
										<p class="text-xs text-gray-500">{collaborator.status}</p>
									</div>
								{/each}
							</div>
						</div>
						<div class="rounded-[1.5rem] border border-dashed border-gray-200 px-3 py-3 text-xs text-gray-500">
							Configure panel display in the chatbot flow settings. The extra panel area is preserved here as a confirmed dependency from the audited design docs.
						</div>
						<select bind:value={collaboratorUserId} class="w-full rounded-[1.5rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400">
							<option value="">Invite collaborator</option>
							{#each workspace.users as user}
								<option value={user.id}>{user.name}</option>
							{/each}
						</select>
						<button data-testid="invite-collaborator" class="w-full rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700" onclick={inviteCollaborator}>Invite Collaborator</button>
					</div>
				</section>
			{/if}

			{#if workspace?.selected && showTimeline}
				<section class="rounded-[2rem] border border-gray-200 bg-white p-4 shadow-sm">
					<div class="mb-3 flex items-center justify-between">
						<h3 class="text-lg font-semibold text-gray-900">Timeline</h3>
						<span class="text-xs text-gray-400">{workspace.selected.events.length}</span>
					</div>
					<div class="space-y-3">
						{#each workspace.selected.events as event}
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="font-medium text-gray-800">{event.summary}</p>
								<p class="mt-1 text-xs text-gray-500">{event.actor_name} · {formatDateTime(event.occurred_at)}</p>
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
