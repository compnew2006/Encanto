<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch, type ProfileView } from '$lib/api';

	let profile = $state<ProfileView | null>(null);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');

	onMount(async () => {
		try {
			profile = await apiFetch<ProfileView>('/api/profile');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load profile.';
		} finally {
			loading = false;
		}
	});

	async function saveProfile() {
		if (!profile) return;
		error = '';
		success = '';
		try {
			profile = await apiFetch<ProfileView>('/api/profile', {
				method: 'PUT',
				body: {
					name: profile.user.name,
					status: profile.user.status,
					language: profile.user.settings.language,
					theme_preset: profile.settings.appearance.theme_preset
				}
			});
			success = 'Profile saved.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save profile.';
		}
	}
</script>

<div class="mx-auto max-w-5xl px-5 py-6">
	<div class="mb-6">
		<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Personal Settings</p>
		<h1 class="mt-2 text-3xl font-semibold text-gray-900">Profile</h1>
		<p class="mt-2 text-sm text-gray-500">Edit the current-user context that feeds the protected layout, status badge, and personal preferences.</p>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading profile...</div>
	{:else if profile}
		<div class="grid gap-4 lg:grid-cols-[1.3fr_0.7fr]">
			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<div class="mb-5 flex items-center gap-4">
					<img class="h-20 w-20 rounded-[2rem] object-cover" src={profile.user.avatar} alt={profile.user.name} />
					<div>
						<p class="text-sm text-gray-500">{profile.user.email}</p>
						<h2 class="text-2xl font-semibold text-gray-900">{profile.user.name}</h2>
						<p class="text-sm text-gray-500">Active role: {profile.user.role} · Org: {profile.user.current_organization.name}</p>
					</div>
				</div>

				<div class="grid gap-4 md:grid-cols-2">
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Display Name</span>
						<input bind:value={profile.user.name} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Availability Status</span>
						<select bind:value={profile.user.status} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="online">Online</option>
							<option value="busy">Busy</option>
							<option value="offline">Offline</option>
						</select>
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Language</span>
						<select bind:value={profile.user.settings.language} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="en">English</option>
							<option value="ar">Arabic</option>
						</select>
					</label>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-gray-700">Theme Preset</span>
						<select bind:value={profile.settings.appearance.theme_preset} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
							<option value="ocean-breeze">Ocean Breeze</option>
							<option value="soft-pop">Soft Pop</option>
							<option value="amber-minimal">Amber Minimal</option>
						</select>
					</label>
				</div>

				<button data-testid="save-profile" class="mt-6 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveProfile}>Save Profile</button>
			</section>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				<h3 class="text-lg font-semibold text-gray-900">Live Context</h3>
				<div class="mt-4 space-y-3 text-sm text-gray-600">
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Theme</p>
						<p class="mt-1">{profile.user.settings.theme}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Sidebar</p>
						<p class="mt-1">{profile.user.settings.sidebar_pinned ? 'Pinned by default' : 'Auto-collapsed'}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Organizations</p>
						<p class="mt-1">{profile.user.organizations.map((org) => org.name).join(', ')}</p>
					</div>
				</div>
			</section>
		</div>
	{/if}
</div>
