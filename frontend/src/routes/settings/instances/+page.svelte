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

	async function loadAll() {
		loading = true;
		try {
			const [instanceResponse, healthResponse, settingsResponse] = await Promise.all([
				apiFetch<{ instances: WhatsAppInstance[] }>('/api/instances'),
				apiFetch<{ health: InstanceHealthSummary[] }>('/api/instances/health'),
				apiFetch<SettingsSummary>('/api/settings/summary')
			]);
			instances = instanceResponse.instances;
			health = healthResponse.health;
			settings = settingsResponse;
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

	async function renameInstance(instance: WhatsAppInstance, name: string) {
		try {
			await apiFetch(`/api/instances/${instance.id}/name`, { method: 'PUT', body: { name } });
			await loadAll();
			success = 'Account name updated.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Rename failed.';
		}
	}

	async function runInstanceAction(path: string, message: string) {
		try {
			await apiFetch(path, { method: 'POST' });
			await loadAll();
			success = message;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Action failed.';
		}
	}

	async function deleteInstance(instance: WhatsAppInstance) {
		if (!window.confirm(`Delete ${instance.name}?`)) return;
		try {
			await apiFetch(`/api/instances/${instance.id}`, { method: 'DELETE' });
			await loadAll();
			success = 'Account deleted.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Delete failed.';
		}
	}

	async function saveInstanceSettings(instance: WhatsAppInstance) {
		try {
			await Promise.all([
				apiFetch(`/api/instances/${instance.id}/settings`, { method: 'PUT', body: instance.settings }),
				apiFetch(`/api/instances/${instance.id}/call-auto-reject`, { method: 'PUT', body: instance.call_policy }),
				apiFetch(`/api/instances/${instance.id}/auto-campaign`, { method: 'PUT', body: instance.auto_campaign })
			]);
			await loadAll();
			success = `${instance.name} settings saved.`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save account settings.';
		}
	}

	onMount(async () => {
		await loadAll();
		teardownRealtime = await connectRealtime(async (message) => {
			if (['instance_connected', 'instance_disconnected', 'instance_recovering', 'new_message'].includes(message.type)) {
				await loadAll();
			}
		});
	});

	onDestroy(() => {
		teardownRealtime?.();
	});
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Account Operations</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">WhatsApp Accounts & Health</h1>
			<p class="mt-2 text-sm text-gray-500">Operational catalog, lifecycle controls, health metrics, and per-account policies.</p>
		</div>
		<div class="flex gap-2">
			<a href="/settings/audit" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Audit</a>
			<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Back to Settings</a>
		</div>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading account operations...</div>
	{:else}
		<section class="mb-5 rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
			<div class="grid gap-4 md:grid-cols-[1fr_1fr_auto]">
				<input bind:value={newInstanceName} class="rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="New account name" />
				<input bind:value={newInstancePhone} class="rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="Phone number" />
				<button data-testid="create-instance" class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={createInstance}>Add Account</button>
			</div>
			{#if settings}
				<p class="mt-3 text-sm text-gray-500">Slots used: {settings.general.used_instances}/{settings.general.max_instances}</p>
			{/if}
		</section>

		<div class="grid gap-4 xl:grid-cols-[1.4fr_0.6fr]">
			<div class="space-y-4">
				{#each instances as instance}
					<div class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
						<div class="flex flex-wrap items-start justify-between gap-4">
							<div>
								<div class="flex items-center gap-3">
									<input bind:value={instance.name} class="rounded-[1.25rem] border border-gray-200 px-3 py-2 text-lg font-semibold text-gray-900 outline-none focus:border-blue-400" />
									<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${instance.status === 'connected' ? 'bg-emerald-50 text-emerald-700' : instance.status === 'recovering' ? 'bg-amber-50 text-amber-700' : 'bg-red-50 text-red-700'}`}>{instance.status}</span>
								</div>
								<p class="mt-2 text-sm text-gray-500">{instance.phone_number || 'No phone number yet'} · {instance.jid || 'JID pending'}</p>
							</div>
						<div class="flex flex-wrap gap-2">
							<button class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={() => renameInstance(instance, instance.name)}>Save Name</button>
							<button data-testid={`connect-${instance.id}`} class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={() => runInstanceAction(`/api/instances/${instance.id}/connect`, `${instance.name} connected.`)}>Connect / Scan QR</button>
							<button class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={() => runInstanceAction(`/api/instances/${instance.id}/disconnect`, `${instance.name} disconnected.`)}>Disconnect</button>
							<button class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={() => runInstanceAction(`/api/instances/${instance.id}/recover`, `${instance.name} recovery started.`)}>Recover</button>
							<button data-testid={`delete-instance-${instance.id}`} class="rounded-full border border-red-200 px-4 py-2 text-sm text-red-600" onclick={() => deleteInstance(instance)}>Delete</button>
						</div>
					</div>

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
								<h3 class="text-sm font-semibold text-gray-900">Auto Campaign & Reset</h3>
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

						<div class="mt-4 flex flex-wrap items-center justify-between gap-3">
							<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3 text-sm text-gray-600">
								Last health observation: {formatDateTime(instance.health.observed_at)}
							</div>
							<button data-testid={`save-instance-${instance.id}`} class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={() => saveInstanceSettings(instance)}>Save Policies</button>
						</div>

						{#if instance.qr_code}
							<div class="mt-4 rounded-[1.5rem] border border-dashed border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-700">
								QR pairing token: <span class="font-mono">{instance.qr_code}</span>
							</div>
						{/if}
					</div>
				{/each}
			</div>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<h2 class="text-lg font-semibold text-gray-900">Health Summary</h2>
				<div class="mt-4 space-y-3">
					{#each health as item}
						<div class="rounded-[1.5rem] border border-gray-200 px-4 py-3">
							<div class="flex items-center justify-between gap-2">
								<p class="font-medium text-gray-900">{item.name}</p>
								<span class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${item.status === 'connected' ? 'bg-emerald-50 text-emerald-700' : item.status === 'recovering' ? 'bg-amber-50 text-amber-700' : 'bg-red-50 text-red-700'}`}>{item.status}</span>
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
