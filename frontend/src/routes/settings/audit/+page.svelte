<script lang="ts">
	import { onMount } from 'svelte';
	import {
		apiFetch,
		formatDateTime,
		type AuditLogEntry,
		type BackgroundJob,
		type WebhookDelivery,
		type WebhookEndpoint
	} from '$lib/api';

	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let jobs = $state<BackgroundJob[]>([]);
	let auditEntries = $state<AuditLogEntry[]>([]);
	let webhooks = $state<WebhookEndpoint[]>([]);
	let deliveries = $state<WebhookDelivery[]>([]);
	let selectedWebhookId = $state('');

	async function loadDeliveries(webhookId: string) {
		if (!webhookId) {
			deliveries = [];
			return;
		}
		const response = await apiFetch<{ deliveries: WebhookDelivery[] }>(
			`/api/webhooks/${webhookId}/deliveries`
		);
		deliveries = response.deliveries;
	}

	async function loadAll() {
		loading = true;
		error = '';
		try {
			const [jobsResponse, webhooksResponse, auditResponse] = await Promise.all([
				apiFetch<{ jobs: BackgroundJob[] }>('/api/jobs'),
				apiFetch<{ webhooks: WebhookEndpoint[] }>('/api/webhooks'),
				apiFetch<{ entries: AuditLogEntry[] }>('/api/audit-logs')
			]);
			jobs = jobsResponse.jobs;
			webhooks = webhooksResponse.webhooks;
			auditEntries = auditResponse.entries;
			selectedWebhookId = selectedWebhookId || webhooksResponse.webhooks[0]?.id || '';
			await loadDeliveries(selectedWebhookId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load reliability surfaces.';
		} finally {
			loading = false;
		}
	}

	async function retryDelivery(delivery: WebhookDelivery) {
		error = '';
		success = '';
		try {
			await apiFetch(`/api/webhooks/${delivery.webhook_id}/deliveries/${delivery.id}/retry`, {
				method: 'POST'
			});
			success = 'Webhook delivery retried.';
			await loadDeliveries(delivery.webhook_id);
			await loadAll();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to retry webhook delivery.';
		}
	}

	onMount(loadAll);
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 15</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Audit & Reliability</h1>
			<p class="mt-2 text-sm text-gray-500">Review recent background jobs, webhook delivery retries, and administrative audit entries.</p>
		</div>
		<a href="/settings" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Back to Settings</a>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading reliability data...</div>
	{:else}
		<div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
			<div class="space-y-4">
				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Job Runs</p>
					<div class="mt-4 space-y-3">
						{#each jobs as job}
							<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3">
								<div class="flex items-center justify-between gap-3">
									<p class="font-medium text-gray-900">{job.kind}</p>
									<span class="text-xs text-gray-500">{job.status}</span>
								</div>
								<p class="mt-1 text-sm text-gray-600">{job.summary}</p>
								<p class="mt-1 text-xs text-gray-400">{formatDateTime(job.started_at)} · {job.entity_type}:{job.entity_id}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<div class="flex items-center justify-between gap-3">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Webhook Deliveries</p>
							<h2 class="mt-1 text-lg font-semibold text-gray-900">Retry Failed Deliveries</h2>
						</div>
						<select data-testid="audit-webhook-select" bind:value={selectedWebhookId} class="rounded-[1.25rem] border border-gray-200 px-4 py-2 text-sm outline-none focus:border-blue-400" onchange={() => loadDeliveries(selectedWebhookId)}>
							{#each webhooks as webhook}
								<option value={webhook.id}>{webhook.name}</option>
							{/each}
						</select>
					</div>

					<div class="mt-4 space-y-3">
						{#each deliveries as delivery}
							<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3">
								<div class="flex items-center justify-between gap-3">
									<p class="font-medium text-gray-900">{delivery.event_id}</p>
									<span class="text-xs text-gray-500">{delivery.status}</span>
								</div>
								<p class="mt-1 text-sm text-gray-600">HTTP {delivery.response_code} · attempt {delivery.attempt}</p>
								<p class="mt-1 text-xs text-gray-400">{delivery.response_body} · {formatDateTime(delivery.last_attempt_at)}</p>
								<button data-testid={`retry-delivery-${delivery.id}`} class="mt-3 rounded-full border border-gray-200 px-3 py-2 text-xs font-medium text-gray-700" onclick={() => retryDelivery(delivery)}>Retry Delivery</button>
							</div>
						{/each}
					</div>
				</section>
			</div>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Audit Log</p>
				<div class="mt-4 space-y-3">
					{#each auditEntries as entry}
						<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3">
							<div class="flex items-center justify-between gap-3">
								<p class="font-medium text-gray-900">{entry.action}</p>
								<span class="text-xs text-gray-500">{formatDateTime(entry.occurred_at)}</span>
							</div>
							<p class="mt-1 text-sm text-gray-600">{entry.summary}</p>
							<p class="mt-1 text-xs text-gray-400">{entry.actor_name} · {entry.entity_type}:{entry.entity_id}</p>
						</div>
					{/each}
				</div>
			</section>
		</div>
	{/if}
</div>
