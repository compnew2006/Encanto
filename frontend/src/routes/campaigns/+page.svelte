<script lang="ts">
	import { onMount } from 'svelte';
	import {
		apiFetch,
		type Campaign,
		type CampaignRecord,
		type CampaignRecipient,
		type CampaignRun,
		type WhatsAppInstance
	} from '$lib/api';

	type CampaignDraft = {
		name: string;
		content: string;
		status: string;
		source: string;
		instance_id: string;
		status_filter: string;
		tag: string;
		include_closed: boolean;
		every_days: number;
		time_of_day: string;
	};

	const blankDraft = (): CampaignDraft => ({
		name: '',
		content: '',
		status: 'draft',
		source: 'manual',
		instance_id: '',
		status_filter: '',
		tag: '',
		include_closed: false,
		every_days: 7,
		time_of_day: '10:00'
	});

	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let campaigns = $state<Campaign[]>([]);
	let instances = $state<WhatsAppInstance[]>([]);
	let selectedCampaignId = $state('');
	let selectedCampaign = $state<CampaignRecord | null>(null);
	let recipients = $state<CampaignRecipient[]>([]);
	let activeRunId = $state('');
	let draft = $state<CampaignDraft>(blankDraft());

	function payloadFromDraft() {
		return {
			name: draft.name,
			content: draft.content,
			status: draft.status,
			source: draft.source,
			filters: {
				instance_id: draft.instance_id,
				status: draft.status_filter,
				tag: draft.tag,
				include_closed: draft.include_closed
			},
			schedule: {
				mode: 'every_n_days',
				every_days: Number(draft.every_days) || 1,
				time_of_day: draft.time_of_day
			}
		};
	}

	function resetDraft() {
		draft = {
			...blankDraft(),
			instance_id: instances[0]?.id ?? ''
		};
		selectedCampaignId = '';
		selectedCampaign = null;
		recipients = [];
		activeRunId = '';
	}

	function hydrateDraft(record: CampaignRecord) {
		draft = {
			name: record.campaign.name,
			content: record.campaign.content,
			status: record.campaign.status,
			source: record.campaign.source,
			instance_id: record.campaign.filters.instance_id || '',
			status_filter: record.campaign.filters.status || '',
			tag: record.campaign.filters.tag || '',
			include_closed: record.campaign.filters.include_closed || false,
			every_days: record.campaign.schedule.every_days || 7,
			time_of_day: record.campaign.schedule.time_of_day || '10:00'
		};
	}

	async function loadCampaigns() {
		const response = await apiFetch<{ campaigns: Campaign[] }>('/api/campaigns');
		campaigns = response.campaigns;
	}

	async function loadInstances() {
		const response = await apiFetch<{ instances: WhatsAppInstance[] }>('/api/instances');
		instances = response.instances;
		if (!draft.instance_id) {
			draft.instance_id = response.instances[0]?.id ?? '';
		}
	}

	async function loadRecipients(campaignId: string, runId: string) {
		if (!runId) {
			recipients = [];
			return;
		}
		const response = await apiFetch<{ recipients: CampaignRecipient[] }>(
			`/api/campaigns/${campaignId}/recipients?run_id=${runId}`
		);
		recipients = response.recipients;
	}

	async function selectCampaign(campaignId: string) {
		error = '';
		try {
			selectedCampaignId = campaignId;
			selectedCampaign = await apiFetch<CampaignRecord>(`/api/campaigns/${campaignId}`);
			hydrateDraft(selectedCampaign);
			activeRunId = selectedCampaign.runs[0]?.id ?? '';
			await loadRecipients(campaignId, activeRunId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load campaign detail.';
		}
	}

	async function saveCampaign() {
		error = '';
		success = '';
		try {
			if (selectedCampaignId) {
				await apiFetch(`/api/campaigns/${selectedCampaignId}`, {
					method: 'PUT',
					body: payloadFromDraft()
				});
				success = 'Campaign updated.';
				await selectCampaign(selectedCampaignId);
			} else {
				const response = await apiFetch<{ campaign: Campaign }>('/api/campaigns', {
					method: 'POST',
					body: payloadFromDraft()
				});
				success = 'Campaign created.';
				await loadCampaigns();
				await selectCampaign(response.campaign.id);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save campaign.';
		}
	}

	async function deleteCampaign() {
		if (!selectedCampaign || !window.confirm(`Delete ${selectedCampaign.campaign.name}?`)) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/campaigns/${selectedCampaign.campaign.id}`, { method: 'DELETE' });
			success = 'Campaign deleted.';
			await loadCampaigns();
			resetDraft();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete campaign.';
		}
	}

	async function launchCampaign() {
		if (!selectedCampaignId) return;
		error = '';
		success = '';
		try {
			const response = await apiFetch<{ run: CampaignRun; recipients: CampaignRecipient[] }>(
				`/api/campaigns/${selectedCampaignId}/launch`,
				{ method: 'POST' }
			);
			success = 'Campaign launched.';
			activeRunId = response.run.id;
			recipients = response.recipients;
			await loadCampaigns();
			await selectCampaign(selectedCampaignId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to launch campaign.';
		}
	}

	async function pauseCampaign() {
		if (!selectedCampaignId) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/campaigns/${selectedCampaignId}/pause`, { method: 'POST' });
			success = 'Campaign paused.';
			await loadCampaigns();
			await selectCampaign(selectedCampaignId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to pause campaign.';
		}
	}

	async function resumeCampaign() {
		if (!selectedCampaignId) return;
		error = '';
		success = '';
		try {
			await apiFetch(`/api/campaigns/${selectedCampaignId}/resume`, { method: 'POST' });
			success = 'Campaign resumed.';
			await loadCampaigns();
			await selectCampaign(selectedCampaignId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to resume campaign.';
		}
	}

	onMount(async () => {
		loading = true;
		try {
			await Promise.all([loadCampaigns(), loadInstances()]);
			if (campaigns[0]) {
				await selectCampaign(campaigns[0].id);
			} else {
				resetDraft();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load campaigns.';
		} finally {
			loading = false;
		}
	});
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 14</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Campaigns</h1>
			<p class="mt-2 text-sm text-gray-500">Reusable campaign definitions, launch history, recipients, and linked account automations in one operational surface.</p>
		</div>
		<a href="/analytics/agents" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Analytics</a>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading campaigns...</div>
	{:else}
		<div class="grid gap-4 xl:grid-cols-[0.75fr_1.25fr]">
			<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
				<div class="flex items-center justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Campaign Library</p>
						<h2 class="mt-1 text-lg font-semibold text-gray-900">Definitions</h2>
					</div>
					<button data-testid="campaign-create-new" class="rounded-full border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700" onclick={resetDraft}>New Campaign</button>
				</div>

				<div class="mt-4 space-y-3">
					{#each campaigns as campaign}
						<button data-testid={`campaign-item-${campaign.id}`} class={`w-full rounded-[1.5rem] border px-4 py-4 text-left transition ${selectedCampaignId === campaign.id ? 'border-blue-300 bg-blue-50' : 'border-gray-200 bg-white hover:border-blue-200'}`} onclick={() => selectCampaign(campaign.id)}>
							<div class="flex items-center justify-between gap-3">
								<p class="font-medium text-gray-900">{campaign.name}</p>
								<span class="rounded-full bg-white px-3 py-1 text-xs font-semibold uppercase tracking-wide text-gray-600">{campaign.status}</span>
							</div>
							<p class="mt-2 text-sm text-gray-500">{campaign.last_run_summary}</p>
							<p class="mt-1 text-xs text-gray-400">{campaign.source} · {new Date(campaign.updated_at).toLocaleString()}</p>
						</button>
					{/each}
				</div>
			</section>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Editor</p>
						<h2 class="mt-1 text-lg font-semibold text-gray-900">{selectedCampaignId ? 'Update Campaign' : 'Create Campaign'}</h2>
					</div>
					<div class="flex flex-wrap gap-2">
						{#if selectedCampaignId}
							<button data-testid="campaign-launch" class="rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white" onclick={launchCampaign}>Launch</button>
							<button data-testid="campaign-pause" class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={pauseCampaign}>Pause</button>
							<button data-testid="campaign-resume" class="rounded-full border border-gray-200 px-4 py-2 text-sm text-gray-700" onclick={resumeCampaign}>Resume</button>
							<button data-testid="campaign-delete" class="rounded-full border border-red-200 px-4 py-2 text-sm text-red-600" onclick={deleteCampaign}>Delete</button>
						{/if}
					</div>
				</div>

				<div class="mt-5 grid gap-4 md:grid-cols-2">
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Campaign Name</span>
						<input data-testid="campaign-name" bind:value={draft.name} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Source</span>
						<select bind:value={draft.source} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="manual">manual</option>
							<option value="instance_auto_campaign">instance_auto_campaign</option>
						</select>
					</label>
					<label class="block md:col-span-2">
						<span class="mb-2 block text-sm font-medium text-gray-700">Content</span>
						<textarea data-testid="campaign-content" bind:value={draft.content} class="min-h-[120px] w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400"></textarea>
					</label>
				</div>

				<div class="mt-4 grid gap-4 md:grid-cols-3">
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Definition Status</span>
						<select bind:value={draft.status} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="draft">draft</option>
							<option value="scheduled">scheduled</option>
							<option value="active">active</option>
							<option value="paused">paused</option>
							<option value="disabled">disabled</option>
						</select>
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Account Filter</span>
						<select data-testid="campaign-instance-filter" bind:value={draft.instance_id} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="">All accounts</option>
							{#each instances as instance}
								<option value={instance.id}>{instance.name}</option>
							{/each}
						</select>
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Conversation Status Filter</span>
						<select data-testid="campaign-status-filter" bind:value={draft.status_filter} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="">Any status</option>
							<option value="assigned">assigned</option>
							<option value="pending">pending</option>
							<option value="closed">closed</option>
						</select>
					</label>
				</div>

				<div class="mt-4 grid gap-4 md:grid-cols-3">
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Tag Filter</span>
						<input data-testid="campaign-tag-filter" bind:value={draft.tag} class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" placeholder="renewal" />
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Every N Days</span>
						<input data-testid="campaign-every-days" bind:value={draft.every_days} type="number" min="1" class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Time of Day</span>
						<input bind:value={draft.time_of_day} type="time" class="w-full rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
					</label>
				</div>

				<label class="mt-4 flex items-center justify-between rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
					<span>Include closed conversations in target audience</span>
					<input bind:checked={draft.include_closed} type="checkbox" />
				</label>

				<button data-testid="campaign-save" class="mt-5 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveCampaign}>{selectedCampaignId ? 'Save Campaign' : 'Create Campaign'}</button>

				{#if selectedCampaign}
					<div class="mt-6 grid gap-4 xl:grid-cols-[0.8fr_1.2fr]">
						<div class="rounded-[1.5rem] border border-gray-200 p-4">
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Runs</p>
							<div class="mt-3 space-y-2">
								{#each selectedCampaign.runs as run}
									<button class={`w-full rounded-[1.25rem] border px-4 py-3 text-left ${activeRunId === run.id ? 'border-blue-300 bg-blue-50' : 'border-gray-200 bg-white'}`} onclick={() => { activeRunId = run.id; loadRecipients(selectedCampaignId, run.id); }}>
										<div class="flex items-center justify-between gap-3">
											<p class="font-medium text-gray-900">{run.trigger}</p>
											<span class="text-xs text-gray-500">{run.status}</span>
										</div>
										<p class="mt-1 text-xs text-gray-400">{run.delivered}/{run.recipient_total} delivered · job {run.job_id}</p>
									</button>
								{/each}
								{#if selectedCampaign.runs.length === 0}
									<div class="rounded-[1.25rem] border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500">No runs yet.</div>
								{/if}
							</div>
						</div>

						<div class="rounded-[1.5rem] border border-gray-200 p-4">
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Recipients</p>
							<div class="mt-3 space-y-2">
								{#each recipients as recipient}
									<a href={`/chat/${recipient.contact_id}`} class="block rounded-[1.25rem] border border-gray-200 px-4 py-3 transition hover:border-blue-300">
										<div class="flex items-center justify-between gap-3">
											<p class="font-medium text-gray-900">{recipient.contact_name}</p>
											<span class="text-xs text-gray-500">{recipient.status}</span>
										</div>
										<p class="mt-1 text-sm text-gray-600">{recipient.message_preview}</p>
										<p class="mt-1 text-xs text-gray-400">{recipient.phone_number}{recipient.failure_reason ? ` · ${recipient.failure_reason}` : ''}</p>
									</a>
								{/each}
								{#if recipients.length === 0}
									<div class="rounded-[1.25rem] border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500">Select a run to inspect recipients.</div>
								{/if}
							</div>
						</div>
					</div>
				{/if}
			</section>
		</div>
	{/if}
</div>
