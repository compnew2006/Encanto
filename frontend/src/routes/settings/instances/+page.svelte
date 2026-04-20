<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import {
		apiFetch,
		formatDateTime,
		type InstanceHealthSummary,
		type SettingsSummary,
		type WhatsAppInstance
	} from '$lib/api';
	import { connectRealtime } from '$lib/realtime/ws';

	let instances = $state<WhatsAppInstance[]>([]);
	let health = $state<InstanceHealthSummary[]>([]);
	let settings = $state<SettingsSummary | null>(null);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let newInstanceName = $state('');
	let newInstancePhone = $state('');
	let teardownRealtime: (() => void) | undefined;

	// QR modal state
	let qrModal = $state<{ open: boolean; instanceName: string; qrCode: string; instanceId: string }>({
		open: false,
		instanceName: '',
		qrCode: '',
		instanceId: ''
	});

	// Per-instance action loading state
	let actionLoading = $state<Record<string, boolean>>({});

	async function loadAll() {
		loading = true;
		try {
			const [instanceResponse, healthResponse, settingsResponse] = await Promise.all([
				apiFetch<{ instances: WhatsAppInstance[] }>('/api/instances'),
				apiFetch<{ health: InstanceHealthSummary[] }>('/api/instances/health'),
				apiFetch<SettingsSummary>('/api/settings/summary')
			]);
			instances = instanceResponse?.instances || [];
			health = healthResponse?.health || [];
			settings = settingsResponse;

			// If the QR modal is open for an instance, refresh the QR code in it
			if (qrModal.open) {
				const refreshed = instances.find((i) => i.id === qrModal.instanceId);
				if (refreshed) {
					if (refreshed.status === 'connected') {
						// Connected — close the modal automatically
						closeQrModal();
						success = `${refreshed.name} connected successfully!`;
					} else if (refreshed.qr_code) {
						qrModal.qrCode = refreshed.qr_code;
					}
				}
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load account operations.';
		} finally {
			loading = false;
		}
	}

	async function createInstance() {
		error = '';
		success = '';
		try {
			await apiFetch('/api/instances', {
				method: 'POST',
				body: { name: newInstanceName, phone_number: newInstancePhone }
			});
			newInstanceName = '';
			newInstancePhone = '';
			await loadAll();
			success = 'Account created.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Create failed.';
		}
	}

	async function renameInstance(instance: WhatsAppInstance) {
		error = '';
		success = '';
		actionLoading[`rename-${instance.id}`] = true;
		try {
			await apiFetch(`/api/instances/${instance.id}/name`, {
				method: 'PUT',
				body: { name: instance.name }
			});
			await loadAll();
			success = 'Account name updated.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Rename failed.';
		} finally {
			actionLoading[`rename-${instance.id}`] = false;
		}
	}

	/**
	 * Connect / Scan QR:
	 * - If the instance already has a QR code (pairing_state === 'needs_qr' / 'disconnected'),
	 *   show it in the modal so the user can scan it.
	 * - After showing the QR, also call the backend to start pairing. In a real whameow
	 *   implementation the backend would push a new QR via WebSocket; here we simply
	 *   call connect, which the mock backend resolves immediately.
	 */
	async function connectInstance(instance: WhatsAppInstance) {
		error = '';
		success = '';

		// If already has QR code, show it while also starting connect
		if (instance.qr_code || instance.status !== 'connected') {
			// Open QR modal immediately with the current QR (may be stale)
			qrModal = {
				open: true,
				instanceName: instance.name,
				qrCode: instance.qr_code || 'QR-CODE-LOADING',
				instanceId: instance.id
			};
		}

		actionLoading[`connect-${instance.id}`] = true;
		try {
			const res = await apiFetch<{ instance: WhatsAppInstance }>(
				`/api/instances/${instance.id}/connect`,
				{ method: 'POST' }
			);
			// Update local instance data immediately from the response
			const idx = instances.findIndex((i) => i.id === instance.id);
			if (idx !== -1) {
				instances[idx] = res.instance;
			}

			if (res.instance.status === 'connected') {
				// Pairing succeeded (mock backend connects immediately)
				closeQrModal();
				success = `${res.instance.name} connected successfully!`;
			} else if (res.instance.qr_code) {
				// Backend returned a fresh QR — update the modal
				qrModal.qrCode = res.instance.qr_code;
			}

			await loadAll();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Connect failed.';
			closeQrModal();
		} finally {
			actionLoading[`connect-${instance.id}`] = false;
		}
	}

	async function disconnectInstance(instance: WhatsAppInstance) {
		error = '';
		success = '';
		actionLoading[`disconnect-${instance.id}`] = true;
		try {
			await apiFetch(`/api/instances/${instance.id}/disconnect`, { method: 'POST' });
			await loadAll();
			success = `${instance.name} disconnected.`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Disconnect failed.';
		} finally {
			actionLoading[`disconnect-${instance.id}`] = false;
		}
	}

	async function recoverInstance(instance: WhatsAppInstance) {
		error = '';
		success = '';
		actionLoading[`recover-${instance.id}`] = true;
		try {
			await apiFetch(`/api/instances/${instance.id}/recover`, { method: 'POST' });
			await loadAll();
			success = `${instance.name} recovery started.`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Recovery failed.';
		} finally {
			actionLoading[`recover-${instance.id}`] = false;
		}
	}

	async function deleteInstance(instance: WhatsAppInstance) {
		if (!window.confirm(`Delete "${instance.name}"? This action cannot be undone.`)) return;
		error = '';
		success = '';
		actionLoading[`delete-${instance.id}`] = true;
		try {
			await apiFetch(`/api/instances/${instance.id}`, { method: 'DELETE' });
			await loadAll();
			success = 'Account deleted.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Delete failed.';
		} finally {
			actionLoading[`delete-${instance.id}`] = false;
		}
	}

	async function saveInstanceSettings(instance: WhatsAppInstance) {
		error = '';
		success = '';
		actionLoading[`save-${instance.id}`] = true;
		try {
			await Promise.all([
				apiFetch(`/api/instances/${instance.id}/settings`, {
					method: 'PUT',
					body: instance.settings
				}),
				apiFetch(`/api/instances/${instance.id}/call-auto-reject`, {
					method: 'PUT',
					body: instance.call_policy
				}),
				apiFetch(`/api/instances/${instance.id}/auto-campaign`, {
					method: 'PUT',
					body: instance.auto_campaign
				})
			]);
			await loadAll();
			success = `${instance.name} settings saved.`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save account settings.';
		} finally {
			actionLoading[`save-${instance.id}`] = false;
		}
	}

	function closeQrModal() {
		qrModal = { open: false, instanceName: '', qrCode: '', instanceId: '' };
	}

	/**
	 * Generate a QR code SVG using the qrcode-svg approach via a 3rd-party API.
	 * We use api.qrserver.com which returns a PNG given any data string.
	 */
	function qrImageUrl(data: string): string {
		return `https://api.qrserver.com/v1/create-qr-code/?size=256x256&ecc=M&data=${encodeURIComponent(data)}`;
	}

	onMount(async () => {
		await loadAll();
		teardownRealtime = await connectRealtime(async (message) => {
			if (message.type === 'qr_updated') {
				if (qrModal.open && qrModal.instanceId === message.payload.instance_id) {
					qrModal.qrCode = message.payload.qr_code;
				}
				return;
			}

			if (
				[
					'instance_connected',
					'instance_disconnected',
					'instance_recovering',
					'new_message'
				].includes(message.type)
			) {
				await loadAll();
			}
		});
	});

	onDestroy(() => {
		teardownRealtime?.();
	});
</script>

<!-- ─── QR Code Modal ─────────────────────────────────────────────────────── -->
{#if qrModal.open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
		role="dialog"
		aria-modal="true"
		aria-label="Scan QR to connect WhatsApp"
	>
		<div class="w-full max-w-sm mx-4 rounded-3xl bg-white shadow-2xl overflow-hidden">
			<!-- Header -->
			<div class="bg-gradient-to-r from-emerald-500 to-teal-600 px-6 py-5 text-white">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-xs font-semibold uppercase tracking-widest opacity-80">WhatsApp Pairing</p>
						<h2 class="mt-1 text-lg font-bold">{qrModal.instanceName}</h2>
					</div>
					<button
						class="rounded-full bg-white/20 p-2 hover:bg-white/30 transition-colors"
						onclick={closeQrModal}
						aria-label="Close QR modal"
					>
						<svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="px-6 py-6 text-center">
				<p class="mb-4 text-sm text-gray-500">
					Open WhatsApp on your phone → <strong>Linked Devices</strong> → <strong>Link a Device</strong> → scan this code.
				</p>

				<!-- QR Image -->
				<div class="mx-auto flex h-64 w-64 items-center justify-center rounded-2xl border-2 border-dashed border-emerald-200 bg-emerald-50 overflow-hidden">
					{#if qrModal.qrCode && qrModal.qrCode !== 'QR-CODE-LOADING'}
						<img
							src={qrImageUrl(qrModal.qrCode)}
							alt="WhatsApp QR Code"
							class="h-56 w-56 rounded-xl object-contain"
						/>
					{:else}
						<div class="flex flex-col items-center gap-3 text-emerald-500">
							<svg class="h-10 w-10 animate-spin" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							<span class="text-sm font-medium">Generating QR code…</span>
						</div>
					{/if}
				</div>

				<!-- Raw token (small, for debugging) -->
				{#if qrModal.qrCode && qrModal.qrCode !== 'QR-CODE-LOADING'}
					<p class="mt-3 text-xs text-gray-400 font-mono break-all leading-relaxed">
						Token: {qrModal.qrCode}
					</p>
				{/if}

				<p class="mt-4 text-xs text-gray-400">
					This window will close automatically once the device is linked.
				</p>

				<button
					class="mt-5 w-full rounded-full border border-gray-200 py-2.5 text-sm font-medium text-gray-600 hover:border-red-300 hover:text-red-600 transition-colors"
					onclick={closeQrModal}
				>
					Cancel Pairing
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ─── Main Content ──────────────────────────────────────────────────────── -->
<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Account Operations</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">WhatsApp Accounts &amp; Health</h1>
			<p class="mt-2 text-sm text-gray-500">Operational catalog, lifecycle controls, health metrics, and per-account policies.</p>
		</div>
		<div class="flex gap-2">
			<a href="/settings/audit" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Audit</a>
			<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Back to Settings</a>
		</div>
	</div>

	{#if error}
		<div class="mb-4 flex items-start gap-3 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
			<svg class="mt-0.5 h-4 w-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
				<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
			</svg>
			<span>{error}</span>
		</div>
	{/if}
	{#if success}
		<div class="mb-4 flex items-start gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
			<svg class="mt-0.5 h-4 w-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
				<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
			</svg>
			<span>{success}</span>
		</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading account operations…</div>
	{:else}
		<!-- Add Account -->
		<section class="mb-5 rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
			<div class="grid gap-4 md:grid-cols-[1fr_1fr_auto]">
				<input
					bind:value={newInstanceName}
					class="rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400"
					placeholder="New account name"
				/>
				<input
					bind:value={newInstancePhone}
					class="rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400"
					placeholder="Phone number"
				/>
				<button
					data-testid="create-instance"
					class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white disabled:opacity-60"
					onclick={createInstance}
					disabled={!newInstanceName.trim()}
				>Add Account</button>
			</div>
			{#if settings}
				<p class="mt-3 text-sm text-gray-500">Slots used: {settings.general.used_instances}/{settings.general.max_instances}</p>
			{/if}
		</section>

		<div class="grid gap-4 xl:grid-cols-[1.4fr_0.6fr]">
			<div class="space-y-4">
				{#each instances as instance (instance.id)}
					<div class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
						<!-- Header row -->
						<div class="flex flex-wrap items-start justify-between gap-4">
							<div>
								<div class="flex items-center gap-3">
									<input
										bind:value={instance.name}
										class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-lg font-semibold text-gray-900 outline-none focus:border-blue-400"
									/>
									<span
										class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
											instance.status === 'connected'
												? 'bg-emerald-50 text-emerald-700'
												: instance.status === 'recovering'
												? 'bg-amber-50 text-amber-700'
												: 'bg-red-50 text-red-700'
										}`}
									>{instance.status}</span>
								</div>
								<p class="mt-2 text-sm text-gray-500">
									{instance.phone_number || 'No phone number yet'} · {instance.jid || 'JID pending'}
								</p>
							</div>

							<!-- Action buttons -->
							<div class="flex flex-wrap gap-2">
								<!-- Save Name -->
								<button
									class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700 hover:border-blue-300 hover:text-blue-700 disabled:opacity-50 transition-colors"
									onclick={() => renameInstance(instance)}
									disabled={actionLoading[`rename-${instance.id}`]}
								>
									{actionLoading[`rename-${instance.id}`] ? 'Saving…' : 'Save Name'}
								</button>

								<!-- Connect / Scan QR -->
								<button
									data-testid={`connect-${instance.id}`}
									class={`rounded-full border px-4 py-2 text-sm font-medium transition-colors disabled:opacity-50 ${
										instance.status === 'connected'
											? 'border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
											: 'border-blue-200 bg-blue-50 text-blue-700 hover:bg-blue-100'
									}`}
									onclick={() => connectInstance(instance)}
									disabled={actionLoading[`connect-${instance.id}`]}
								>
									{#if actionLoading[`connect-${instance.id}`]}
										<span class="flex items-center gap-1.5">
											<svg class="h-3 w-3 animate-spin" fill="none" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
											</svg>
											Connecting…
										</span>
									{:else if instance.status === 'connected'}
										✓ Connected
									{:else}
										📱 Connect / Scan QR
									{/if}
								</button>

								<!-- Disconnect -->
								<button
									class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700 hover:border-amber-300 hover:text-amber-700 disabled:opacity-50 transition-colors"
									onclick={() => disconnectInstance(instance)}
									disabled={actionLoading[`disconnect-${instance.id}`] || instance.status === 'disconnected'}
								>
									{actionLoading[`disconnect-${instance.id}`] ? 'Disconnecting…' : 'Disconnect'}
								</button>

								<!-- Recover -->
								<button
									class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700 hover:border-violet-300 hover:text-violet-700 disabled:opacity-50 transition-colors"
									onclick={() => recoverInstance(instance)}
									disabled={actionLoading[`recover-${instance.id}`]}
								>
									{actionLoading[`recover-${instance.id}`] ? 'Recovering…' : 'Recover'}
								</button>

								<!-- Delete -->
								<button
									data-testid={`delete-instance-${instance.id}`}
									class="rounded-full border border-red-200 px-4 py-2 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50 transition-colors"
									onclick={() => deleteInstance(instance)}
									disabled={actionLoading[`delete-${instance.id}`]}
								>
									{actionLoading[`delete-${instance.id}`] ? 'Deleting…' : 'Delete'}
								</button>
							</div>
						</div>

						<!-- QR code inline banner (when not in modal) -->
						{#if instance.qr_code && instance.status !== 'connected' && !qrModal.open}
							<div class="mt-4 flex items-center gap-4 rounded-[1.5rem] border border-dashed border-blue-200 bg-blue-50 px-4 py-3">
								<div class="shrink-0">
									<img
										src={qrImageUrl(instance.qr_code)}
										alt="WhatsApp QR Code"
										class="h-20 w-20 rounded-xl"
									/>
								</div>
								<div>
									<p class="text-sm font-semibold text-blue-700">QR Code Ready</p>
									<p class="mt-0.5 text-xs text-blue-600">Open WhatsApp → Linked Devices → Link a Device → scan this code.</p>
									<button
										class="mt-2 rounded-full bg-blue-600 px-4 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 transition-colors"
										onclick={() => connectInstance(instance)}
									>Open QR in Full Screen</button>
								</div>
							</div>
						{/if}

						<!-- Health metrics -->
						<div class="mt-4 grid gap-3 md:grid-cols-5">
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Uptime</p>
								<p class="mt-1 text-sm font-medium text-gray-900">{instance.health.uptime_label}</p>
							</div>
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Queue</p>
								<p class="mt-1 text-sm font-medium text-gray-900">{instance.health.queue_depth}</p>
							</div>
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Sent</p>
								<p class="mt-1 text-sm font-medium text-gray-900">{instance.health.sent_today}</p>
							</div>
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Received</p>
								<p class="mt-1 text-sm font-medium text-gray-900">{instance.health.received_today}</p>
							</div>
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Error Rate</p>
								<p class="mt-1 text-sm font-medium text-gray-900">{instance.health.error_rate}</p>
							</div>
						</div>

						<!-- Settings panels -->
						<div class="mt-4 grid gap-4 lg:grid-cols-3">
							<div class="space-y-3 rounded-[1.5rem] border border-gray-200 p-4">
								<h3 class="text-sm font-semibold text-gray-900">Quick Toggles</h3>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Auto-sync history</span>
									<input bind:checked={instance.settings.auto_sync_history} type="checkbox" />
								</label>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Auto-download incoming media</span>
									<input bind:checked={instance.settings.auto_download_incoming_media} type="checkbox" />
								</label>
								<input bind:value={instance.settings.source_tag_label} class="w-full rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Source tag label" />
								<div class="grid grid-cols-2 gap-2">
									<input bind:value={instance.settings.source_tag_display_mode} class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Display mode" />
									<input bind:value={instance.settings.source_tag_color} class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Color" />
								</div>
							</div>

							<div class="space-y-3 rounded-[1.5rem] border border-gray-200 p-4">
								<h3 class="text-sm font-semibold text-gray-900">Call Auto-Reject</h3>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Enabled</span>
									<input bind:checked={instance.call_policy.enabled} type="checkbox" />
								</label>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Reject individual calls</span>
									<input bind:checked={instance.call_policy.reject_individual_calls} type="checkbox" />
								</label>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Reject group calls</span>
									<input bind:checked={instance.call_policy.reject_group_calls} type="checkbox" />
								</label>
								<input bind:value={instance.call_policy.reply_mode} class="w-full rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Reply mode" />
								<textarea bind:value={instance.call_policy.reply_message} class="min-h-[88px] w-full rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Reply message"></textarea>
							</div>

							<div class="space-y-3 rounded-[1.5rem] border border-gray-200 p-4">
								<h3 class="text-sm font-semibold text-gray-900">Auto Campaign &amp; Reset</h3>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Auto campaign enabled</span>
									<input bind:checked={instance.auto_campaign.enabled} type="checkbox" />
								</label>
								<input bind:value={instance.auto_campaign.campaign_name_prefix} class="w-full rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Campaign prefix" />
								<div class="grid grid-cols-3 gap-2">
									<input bind:value={instance.auto_campaign.schedule_every_days} type="number" min="1" class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Days" />
									<input bind:value={instance.auto_campaign.delay_from_minutes} type="number" min="0" class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="From" />
									<input bind:value={instance.auto_campaign.delay_to_minutes} type="number" min="0" class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="To" />
								</div>
								<textarea bind:value={instance.auto_campaign.message_body} class="min-h-[88px] w-full rounded-[1.25rem] border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400" placeholder="Campaign body"></textarea>
								<label class="flex items-center justify-between text-sm text-gray-700">
									<span>Assigned chat reset enabled</span>
									<input bind:checked={instance.assignment_reset.enabled} type="checkbox" />
								</label>
							</div>
						</div>

						<!-- Footer: observed at + save -->
						<div class="mt-4 flex flex-wrap items-center justify-between gap-3">
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3 text-sm text-gray-600">
								Last health observation: {formatDateTime(instance.health.observed_at)}
							</div>
							<button
								data-testid={`save-instance-${instance.id}`}
								class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white disabled:opacity-60 hover:bg-gray-700 transition-colors"
								onclick={() => saveInstanceSettings(instance)}
								disabled={actionLoading[`save-${instance.id}`]}
							>
								{actionLoading[`save-${instance.id}`] ? 'Saving…' : 'Save Policies'}
							</button>
						</div>
					</div>
				{/each}

				{#if instances.length === 0}
					<div class="rounded-[2rem] border border-dashed border-gray-200 bg-gray-50 px-6 py-12 text-center text-gray-400">
						<p class="text-base font-medium">No WhatsApp accounts yet.</p>
						<p class="mt-1 text-sm">Add your first account using the form above.</p>
					</div>
				{/if}
			</div>

			<!-- Health summary sidebar -->
			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<h2 class="text-lg font-semibold text-gray-900">Health Summary</h2>
				<div class="mt-4 space-y-3">
					{#each health as item (item.id)}
						<div class="rounded-[1.5rem] border border-gray-200 px-4 py-3">
							<div class="flex items-center justify-between gap-2">
								<p class="font-medium text-gray-900">{item.name}</p>
								<span
									class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
										item.status === 'connected'
											? 'bg-emerald-50 text-emerald-700'
											: item.status === 'recovering'
											? 'bg-amber-50 text-amber-700'
											: 'bg-red-50 text-red-700'
									}`}
								>{item.status}</span>
							</div>
							<div class="mt-3 grid grid-cols-2 gap-2 text-sm text-gray-600">
								<div>Uptime: {item.uptime_label}</div>
								<div>Queue: {item.queue_depth}</div>
								<div>Sent: {item.sent_today}</div>
								<div>Received: {item.received_today}</div>
								<div>Failed: {item.failed_today}</div>
								<div>Error rate: {item.error_rate}</div>
							</div>
							<p class="mt-2 text-xs text-gray-400">{formatDateTime(item.observed_at)}</p>
						</div>
					{/each}
				</div>
			</section>
		</div>
	{/if}
</div>
