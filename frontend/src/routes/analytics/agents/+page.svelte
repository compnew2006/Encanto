<script lang="ts">
	import { onMount } from 'svelte';
	import {
		apiFetch,
		apiFetchText,
		type AgentAnalyticsSummaryResponse,
		type AgentComparisonRow,
		type AnalyticsPoint,
		type CustomerRating,
		type WhatsAppInstance,
		type WorkspaceSnapshot,
		type WorkspaceUser
	} from '$lib/api';

	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let agentId = $state('');
	let instanceId = $state('');
	let users = $state<WorkspaceUser[]>([]);
	let instances = $state<WhatsAppInstance[]>([]);
	let summary = $state<AgentAnalyticsSummaryResponse | null>(null);
	let transferPoints = $state<AnalyticsPoint[]>([]);
	let sourcePoints = $state<AnalyticsPoint[]>([]);
	let comparisonRows = $state<AgentComparisonRow[]>([]);
	let ratings = $state<CustomerRating[]>([]);
	let exportCSV = $state('');

	function queryString() {
		const query = new URLSearchParams();
		if (agentId) query.set('agent_id', agentId);
		if (instanceId) query.set('instance_id', instanceId);
		return query.toString();
	}

	async function loadReferenceData() {
		const workspace = await apiFetch<WorkspaceSnapshot>('/api/chats?tab=assigned');
		users = workspace.users;
		instances = workspace.instances;
	}

	async function loadAnalytics() {
		loading = true;
		error = '';
		try {
			const query = queryString();
			const suffix = query ? `?${query}` : '';
			const [summaryResponse, transfersResponse, sourcesResponse, comparisonResponse, ratingsResponse] =
				await Promise.all([
					apiFetch<AgentAnalyticsSummaryResponse>(`/api/analytics/agents/summary${suffix}`),
					apiFetch<{ points: AnalyticsPoint[] }>(`/api/analytics/agents/transfers${suffix}`),
					apiFetch<{ points: AnalyticsPoint[] }>(`/api/analytics/agents/sources${suffix}`),
					apiFetch<{ rows: AgentComparisonRow[] }>(`/api/analytics/agents/comparison${suffix}`),
					apiFetch<{ rows: CustomerRating[] }>(`/api/analytics/agents/ratings${suffix}`)
				]);
			summary = summaryResponse;
			transferPoints = transfersResponse.points;
			sourcePoints = sourcesResponse.points;
			comparisonRows = comparisonResponse.rows;
			ratings = ratingsResponse.rows;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load analytics.';
		} finally {
			loading = false;
		}
	}

	async function exportAnalytics() {
		error = '';
		success = '';
		try {
			const query = queryString();
			exportCSV = await apiFetchText(`/api/analytics/agents/export${query ? `?${query}` : ''}`);
			success = 'Analytics exported.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to export analytics.';
		}
	}

	onMount(async () => {
		try {
			await loadReferenceData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load analytics filters.';
		}
		await loadAnalytics();
	});
</script>

<div class="mx-auto max-w-7xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 13</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Agent Analytics</h1>
			<p class="mt-2 text-sm text-gray-500">Operational KPIs, transfer trends, source breakdown, comparison tables, and customer ratings built from conversation facts.</p>
		</div>
		<a href="/campaigns" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Open Campaigns</a>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
		<div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto_auto]">
			<select data-testid="analytics-agent-filter" bind:value={agentId} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
				<option value="">All agents</option>
				{#each users as user}
					<option value={user.id}>{user.name}</option>
				{/each}
			</select>
			<select data-testid="analytics-instance-filter" bind:value={instanceId} class="rounded-[1.25rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
				<option value="">All accounts</option>
				{#each instances as instance}
					<option value={instance.id}>{instance.name}</option>
				{/each}
			</select>
			<button class="rounded-full border border-gray-200 px-4 py-3 text-sm font-medium text-gray-700" onclick={loadAnalytics}>Apply</button>
			<button data-testid="analytics-export" class="rounded-full bg-gray-900 px-4 py-3 text-sm font-medium text-white" onclick={exportAnalytics}>Export CSV</button>
		</div>
	</section>

	{#if loading}
		<div class="mt-4 rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading analytics...</div>
	{:else if summary}
		<div class="mt-4 space-y-4">
			<div class="grid gap-4 md:grid-cols-4">
				{#each summary.cards as card}
					<div class="rounded-[1.75rem] border border-gray-200 bg-white px-5 py-5 shadow-sm">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">{card.label}</p>
						<p class="mt-3 text-3xl font-semibold text-gray-900">{card.value}</p>
						<p class="mt-2 text-sm text-blue-700">{card.trend}</p>
						<p class="mt-2 text-xs text-gray-500">{card.description}</p>
					</div>
				{/each}
			</div>

			<div class="grid gap-4 xl:grid-cols-[0.85fr_1.15fr]">
				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<div class="flex items-center justify-between gap-3">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Validation</p>
							<h2 class="mt-1 text-lg font-semibold text-gray-900">Evidence Trail</h2>
						</div>
						<p class="text-xs text-gray-400">{new Date(summary.generated).toLocaleString()}</p>
					</div>
					<div class="mt-4 space-y-3">
						{#each summary.validation as item}
							<div class="rounded-[1.25rem] border border-gray-200 px-4 py-3">
								<p class="text-sm font-medium text-gray-900">{item.label}</p>
								<p class="mt-1 text-sm text-gray-600">{item.value}</p>
								<p class="mt-1 text-xs text-gray-400">{item.source}</p>
							</div>
						{/each}
					</div>
				</section>

				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Breakdowns</p>
					<div class="mt-4 grid gap-4 md:grid-cols-2">
						<div class="rounded-[1.5rem] border border-gray-200 p-4">
							<h2 class="text-base font-semibold text-gray-900">Transfer Trend</h2>
							<div class="mt-3 space-y-2 text-sm text-gray-600">
								{#each transferPoints as point}
									<div class="flex items-center justify-between gap-3">
										<span>{point.label}</span>
										<span class="font-semibold text-gray-900">{point.value}</span>
									</div>
								{/each}
							</div>
						</div>
						<div class="rounded-[1.5rem] border border-gray-200 p-4">
							<h2 class="text-base font-semibold text-gray-900">Source Breakdown</h2>
							<div class="mt-3 space-y-2 text-sm text-gray-600">
								{#each sourcePoints as point}
									<div class="flex items-center justify-between gap-3">
										<span>{point.label}</span>
										<span class="font-semibold text-gray-900">{point.value}</span>
									</div>
								{/each}
							</div>
						</div>
					</div>
				</section>
			</div>

			<div class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Comparison</p>
					<h2 class="mt-1 text-lg font-semibold text-gray-900">Agent Comparison Table</h2>
					<div class="mt-4 overflow-x-auto">
						<table class="min-w-full text-left text-sm">
							<thead class="text-gray-400">
								<tr>
									<th class="pb-3 pr-4 font-medium">Agent</th>
									<th class="pb-3 pr-4 font-medium">Active</th>
									<th class="pb-3 pr-4 font-medium">Closed</th>
									<th class="pb-3 pr-4 font-medium">Transfers</th>
									<th class="pb-3 pr-4 font-medium">Queue</th>
									<th class="pb-3 font-medium">Rating</th>
								</tr>
							</thead>
							<tbody class="text-gray-700">
								{#each comparisonRows as row}
									<tr class="border-t border-gray-100">
										<td class="py-3 pr-4 font-medium text-gray-900">{row.agent_name}</td>
										<td class="py-3 pr-4">{row.active_conversations}</td>
										<td class="py-3 pr-4">{row.closed_conversations}</td>
										<td class="py-3 pr-4">{row.transfers}</td>
										<td class="py-3 pr-4">{row.average_queue_minutes.toFixed(1)}m</td>
										<td class="py-3">{row.average_rating.toFixed(1)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</section>

				<section class="rounded-[2rem] border border-gray-200 bg-white p-5 shadow-sm">
					<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Customer Ratings</p>
					<h2 class="mt-1 text-lg font-semibold text-gray-900">Recent Feedback</h2>
					<div class="mt-4 space-y-3">
						{#each ratings as rating}
							<a href={`/chat/${rating.contact_id}`} class="block rounded-[1.25rem] border border-gray-200 px-4 py-3 transition hover:border-blue-300">
								<div class="flex items-center justify-between gap-3">
									<p class="font-medium text-gray-900">{rating.contact_name}</p>
									<span class="rounded-full bg-blue-50 px-3 py-1 text-xs font-semibold text-blue-700">{rating.rating}/5</span>
								</div>
								<p class="mt-1 text-sm text-gray-600">{rating.comment}</p>
								<p class="mt-1 text-xs text-gray-400">{rating.agent_name} · {new Date(rating.submitted_at).toLocaleString()}</p>
							</a>
						{/each}
					</div>
				</section>
			</div>

			<textarea bind:value={exportCSV} class="min-h-[140px] w-full rounded-[1.75rem] border border-gray-200 bg-white px-4 py-3 text-xs text-gray-700 outline-none" readonly placeholder="Analytics CSV export appears here."></textarea>
		</div>
	{/if}
</div>
