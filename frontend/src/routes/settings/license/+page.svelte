<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { apiFetch, formatDateTime, type LicenseBootstrapView } from '$lib/api';

	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');
	let success = $state('');
	let securityKey = $state('');
	let bootstrap = $state<LicenseBootstrapView | null>(null);

	async function loadBootstrap() {
		loading = true;
		error = '';
		try {
			bootstrap = await apiFetch<LicenseBootstrapView>('/api/license/bootstrap');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load license bootstrap.';
		} finally {
			loading = false;
		}
	}

	async function activateLicense() {
		submitting = true;
		error = '';
		success = '';
		try {
			bootstrap = await apiFetch<LicenseBootstrapView>('/api/license/activate', {
				method: 'POST',
				body: {
					security_key: securityKey
				}
			});
			success = 'License activated.';
			if (bootstrap.restricted_cleanup) {
				await goto('/license-cleanup');
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to activate license.';
		} finally {
			submitting = false;
		}
	}

	onMount(loadBootstrap);
</script>

<div class="mx-auto max-w-6xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Milestone 12</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">License & Limits</h1>
			<p class="mt-2 text-sm text-gray-500">Offline activation, device bootstrap, and live entitlement tracking for contacts, campaigns, and WhatsApp accounts.</p>
		</div>
		<div class="flex gap-2">
			<a href="/license-cleanup" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Cleanup Mode</a>
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
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading license bootstrap...</div>
	{:else if bootstrap}
		<div class="grid gap-4 xl:grid-cols-[0.8fr_1.2fr]">
			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Bootstrap</p>
				<div class="mt-3 flex items-center gap-3">
					<span data-testid="license-status" class={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${bootstrap.restricted_cleanup ? 'bg-amber-50 text-amber-700' : bootstrap.status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>{bootstrap.status}</span>
					<span class="text-sm text-gray-500">{bootstrap.tier} · {bootstrap.kind}</span>
				</div>
				<p class="mt-4 text-sm text-gray-600">{bootstrap.message}</p>

				<div class="mt-5 grid gap-3 md:grid-cols-2">
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">HWID</p>
						<p class="mt-2 break-all text-sm font-medium text-gray-900">{bootstrap.hwid}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Short ID</p>
						<p class="mt-2 text-sm font-medium text-gray-900">{bootstrap.short_id}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Activated</p>
						<p class="mt-2 text-sm font-medium text-gray-900">{formatDateTime(bootstrap.activated_at)}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Expires</p>
						<p class="mt-2 text-sm font-medium text-gray-900">{formatDateTime(bootstrap.expires_at)}</p>
					</div>
				</div>

				{#if bootstrap.restricted_cleanup}
					<div class="mt-5 rounded-[1.5rem] border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
						Normal operations are restricted until the over-limit resources are cleaned up.
					</div>
				{/if}
			</section>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<div class="grid gap-4 md:grid-cols-3">
					{#each bootstrap.quotas as quota}
						<div class={`rounded-[1.5rem] border px-4 py-4 ${quota.over_quota ? 'border-amber-200 bg-amber-50' : 'border-gray-200 bg-gray-50'}`}>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">{quota.label}</p>
							<p class="mt-2 text-2xl font-semibold text-gray-900">{quota.current}/{quota.limit}</p>
							<p class="mt-2 text-xs {quota.over_quota ? 'text-amber-700' : 'text-gray-500'}">{quota.over_quota ? 'Cleanup required' : 'Within entitlement'}</p>
						</div>
					{/each}
				</div>

				<div class="mt-6 rounded-[1.75rem] border border-gray-200 bg-gray-50 p-5">
					<p class="text-sm font-semibold text-gray-900">Activate Offline Security Key</p>
					<p class="mt-1 text-sm text-gray-500">Use a normal key to expand quotas. Use a cleanup/restrict key to simulate reduced entitlements and restricted cleanup mode.</p>
					<textarea data-testid="license-key" bind:value={securityKey} class="mt-3 min-h-[150px] w-full rounded-[1.25rem] border border-gray-200 bg-white px-4 py-3 text-sm text-gray-700 outline-none focus:border-blue-400" placeholder="Paste offline security key here..."></textarea>
					<div class="mt-3 flex flex-wrap gap-2">
						<button data-testid="activate-license" class="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white disabled:opacity-60" onclick={activateLicense} disabled={submitting}>Activate License</button>
						<button class="rounded-full border border-gray-200 px-5 py-2.5 text-sm font-medium text-gray-700" onclick={loadBootstrap}>Refresh</button>
					</div>
				</div>
			</section>
		</div>
	{/if}
</div>
