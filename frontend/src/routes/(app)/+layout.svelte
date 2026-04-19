<script lang="ts">
	import type { Snippet } from 'svelte';

	import { invalidateAll, goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import PermissionGate from '$lib/components/PermissionGate.svelte';
	import { clientJSON } from '$lib/client/backend';
	import { copy } from '$lib/i18n';
	import { canViewSettings } from '$lib/permissions';
	import { setUserContext } from '$lib/user.svelte';
	import type { CurrentUserContext } from '$lib/types';

	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
	let user = $state<CurrentUserContext | null>(null);
	let currentPath = $state('');
	const userState = setUserContext(null);

	$effect(() => {
		const nextUser = data.user;
		user = nextUser;
		currentPath = data.pathname;
		userState.update(nextUser);
	});

	const language = $derived(user?.settings.language ?? 'en');
	const labels = $derived(copy[language]);
	const showSettings = $derived(canViewSettings(user));

	async function switchOrganization(organizationId: string) {
		if (!organizationId) return;
		await clientJSON<{ user: CurrentUserContext }>('/api/auth/switch-org', {
			method: 'POST',
			body: JSON.stringify({ organizationId })
		});
		await invalidateAll();
		await goto(resolve('/chat'));
	}

	async function updateAvailability(status: string) {
		await clientJSON<{ user: CurrentUserContext }>('/api/me/availability', {
			method: 'PUT',
			body: JSON.stringify({ availabilityStatus: status })
		});
		await invalidateAll();
	}

	async function updatePreferences(next: Partial<CurrentUserContext['settings']>) {
		if (!user) return;
		await clientJSON<{ user: CurrentUserContext }>('/api/me/settings', {
			method: 'PUT',
			body: JSON.stringify({
				theme: next.theme ?? user.settings.theme,
				language: next.language ?? user.settings.language,
				sidebarPinned: next.sidebarPinned ?? user.settings.sidebarPinned
			})
		});
		await invalidateAll();
	}

	function linkClass(path: string) {
		return currentPath.startsWith(path)
			? 'bg-[#14213d] text-white'
			: 'text-[#3b4d63] hover:bg-white/70';
	}

	function currentValue(event: Event) {
		return (event.currentTarget as HTMLSelectElement).value;
	}
</script>

<div class="app-shell-grid">
	<aside class="surface border-r border-white/60 px-5 py-6">
		<div class="mb-8">
			<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Encanto</p>
			<h1 class="mt-2 text-2xl font-semibold text-[#14213d]">Phases 1-4</h1>
			<p class="mt-2 text-sm text-[#4d5e72]">Only implemented surfaces stay visible.</p>
		</div>

		<nav class="space-y-2">
			<a class={`block rounded-2xl px-4 py-3 text-sm font-medium transition ${linkClass('/chat')}`} href={resolve('/chat')}>{labels.chat}</a>
			<PermissionGate allowed={showSettings}>
				<a class={`block rounded-2xl px-4 py-3 text-sm font-medium transition ${linkClass('/settings')}`} href={resolve('/settings')}>{labels.settings}</a>
			</PermissionGate>
			<a class={`block rounded-2xl px-4 py-3 text-sm font-medium transition ${linkClass('/profile')}`} href={resolve('/profile')}>{labels.profile}</a>
		</nav>

		{#if user}
			<div class="mt-8 rounded-[28px] border border-[#dfd6c6] bg-white/70 p-4">
				<div class="flex items-center gap-3">
					<img alt={user.fullName} class="h-12 w-12 rounded-2xl border border-[#e5ddcf]" src={user.avatarUrl} />
					<div>
						<p class="font-semibold text-[#14213d]">{user.fullName}</p>
						<p class="text-xs text-[#5e6c7d]">{user.currentOrganization.name}</p>
					</div>
				</div>

				<div class="mt-4 space-y-3 text-sm">
					<label class="block">
						<span class="mb-1 block text-xs uppercase tracking-[0.16em] text-[#6b7786]">Organization</span>
						<select class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" onchange={(event) => switchOrganization(currentValue(event))}>
							{#each user.organizations as organization (organization.id)}
								<option selected={organization.id === user.currentOrganization.id} value={organization.id}>
									{organization.name} · {organization.roleName}
								</option>
							{/each}
						</select>
					</label>

					<label class="block">
						<span class="mb-1 block text-xs uppercase tracking-[0.16em] text-[#6b7786]">Availability</span>
						<select class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" onchange={(event) => updateAvailability(currentValue(event))}>
							<option selected={user.availabilityStatus === 'available'} value="available">Available</option>
							<option selected={user.availabilityStatus === 'busy'} value="busy">Busy</option>
							<option selected={user.availabilityStatus === 'unavailable'} value="unavailable">Unavailable</option>
						</select>
					</label>

						<div class="grid grid-cols-2 gap-3">
							<label class="block">
								<span class="mb-1 block text-xs uppercase tracking-[0.16em] text-[#6b7786]">Theme</span>
								<select class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" onchange={(event) => updatePreferences({ theme: currentValue(event) as CurrentUserContext['settings']['theme'] })}>
									<option selected={user.settings.theme === 'light'} value="light">Light</option>
									<option selected={user.settings.theme === 'dark'} value="dark">Dark</option>
									<option selected={user.settings.theme === 'system'} value="system">System</option>
							</select>
						</label>

							<label class="block">
								<span class="mb-1 block text-xs uppercase tracking-[0.16em] text-[#6b7786]">Language</span>
								<select class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" onchange={(event) => updatePreferences({ language: currentValue(event) as CurrentUserContext['settings']['language'] })}>
									<option selected={user.settings.language === 'en'} value="en">English</option>
									<option selected={user.settings.language === 'ar'} value="ar">العربية</option>
								</select>
						</label>
					</div>
				</div>

				<form action="/logout" class="mt-4" method="POST">
					<button class="w-full rounded-full border border-[#d6cbb7] px-4 py-2 text-sm font-semibold text-[#14213d] transition hover:bg-[#14213d] hover:text-white" type="submit">
						{labels.logout}
					</button>
				</form>
			</div>
		{/if}
	</aside>

	<main class="px-4 py-4 md:px-6 md:py-6">
		{@render children()}
	</main>
</div>
