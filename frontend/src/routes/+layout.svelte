<script lang="ts">
	import '../app.css';
	import type { PageData } from './$types';
	import { API_BASE } from '$lib/api';
	import { setUserContext } from '$lib/user.svelte';

	let { data, children }: { data: PageData, children: any } = $props();

	// Initialize Svelte 5 User Context
	// We check if data.user exists (since layout wraps public pages too)
	let userState = $state<ReturnType<typeof setUserContext> | null>(null);

	// Update the existing context if page data user changes (e.g. after login)
	$effect(() => {
		if (data.user) {
			if (!userState) {
				userState = setUserContext(data.user);
			} else {
				userState.update(data.user);
			}
		} else {
			userState = null;
		}
	});

	// Dropdown menu state
	let showMenu = $state(false);
	let showOrgMenu = $state(false);

	// Map status to color
	function getStatusColor(status: string) {
		switch (status) {
			case 'online': return 'bg-green-500';
			case 'busy': return 'bg-red-500';
			case 'offline': return 'bg-gray-400';
			default: return 'bg-gray-400';
		}
	}

	async function switchOrganization(orgId: string) {
		const res = await fetch(`${API_BASE}/api/auth/switch-org`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ org_id: orgId })
		});
		if (res.ok) {
			window.location.reload();
		} else {
			alert("Failed to switch context.");
		}
	}
</script>

<div class="min-h-screen bg-gray-50 flex flex-col">
	<header class="bg-white shadow relative z-40">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex justify-between items-center">

			<div class="flex items-center gap-2">
				{#if userState && userState.current_organization}
					{#if userState.organizations && userState.organizations.length > 1}
						<div class="relative">
							<button 
								onclick={() => showOrgMenu = !showOrgMenu} 
								class="text-xl font-bold text-gray-900 flex items-center gap-1 hover:bg-gray-50 p-1 rounded transition-colors focus:outline-none"
							>
								{userState.current_organization.name}
								<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
								</svg>
							</button>
							
							{#if showOrgMenu}
								<div 
									class="absolute left-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 overflow-hidden text-sm z-50"
									role="menu"
									tabindex="-1"
									onmouseleave={() => showOrgMenu = false}
								>
									<div class="px-4 py-2 border-b border-gray-100 bg-gray-50 text-xs font-semibold text-gray-500 uppercase tracking-wider">
										Switch Workspace
									</div>
									{#each userState.organizations as org}
										<button 
											onclick={() => switchOrganization(org.id)}
											class="w-full text-left px-4 py-3 hover:bg-blue-50 flex items-center justify-between"
										>
											<span class={org.id === userState.current_organization.id ? 'font-bold text-blue-600' : 'text-gray-700'}>
												{org.name}
											</span>
											{#if org.id === userState.current_organization.id}
												<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
											{/if}
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{:else}
						<h1 class="text-xl font-bold text-gray-900 px-1">{userState.current_organization.name}</h1>
					{/if}
				{:else}
					<h1 class="text-xl font-bold text-gray-900">Encanto Workspace</h1>
				{/if}
			</div>

			<nav class="hidden md:flex items-center gap-2 text-sm">
				<a href="/chat" class="rounded-full border border-gray-200 px-3 py-1.5 text-gray-600 hover:border-blue-300 hover:text-blue-700 transition-colors">Inbox</a>
				<a href="/profile" class="rounded-full border border-gray-200 px-3 py-1.5 text-gray-600 hover:border-blue-300 hover:text-blue-700 transition-colors">Profile</a>
				<a href="/settings" class="rounded-full border border-gray-200 px-3 py-1.5 text-gray-600 hover:border-blue-300 hover:text-blue-700 transition-colors">Settings</a>
				<a href="/settings/instances" class="rounded-full border border-gray-200 px-3 py-1.5 text-gray-600 hover:border-blue-300 hover:text-blue-700 transition-colors">Accounts</a>
			</nav>

			{#if userState}
				<div class="relative">
					<button 
						onclick={() => showMenu = !showMenu}
						class="flex items-center gap-3 focus:outline-none p-1 rounded-lg hover:bg-gray-50 transition-colors"
					>
						<div class="text-right hidden sm:block">
							<p class="text-sm font-medium text-gray-900">{userState.name}</p>
							<p class="text-xs text-gray-500 capitalize">{userState.role}</p>
						</div>
						
						<div class="relative">
							<img 
								src={userState.avatar} 
								alt={userState.name} 
								class="h-10 w-10 rounded-full object-cover border border-gray-200"
							/>
							<span 
								class={`absolute bottom-0 right-0 block h-3 w-3 rounded-full ring-2 ring-white ${getStatusColor(userState.status)}`}
							></span>
						</div>
					</button>

					<!-- Dropdown Menu -->
					{#if showMenu}
						<div 
							class="absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-50 overflow-hidden text-sm"
							role="menu"
							tabindex="-1"
							onmouseleave={() => showMenu = false}
						>
							<div class="px-4 py-3 border-b border-gray-100 flex flex-col">
								<span class="text-gray-900 font-medium truncate">{userState.name}</span>
								<span class="text-gray-500 truncate">{userState.email}</span>
							</div>

							<div class="py-1">
								<div class="px-4 py-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Status</div>
								<button class="w-full text-left px-4 py-2 hover:bg-gray-50 flex items-center gap-2">
									<span class="h-2 w-2 rounded-full bg-green-500"></span> Online
								</button>
								<button class="w-full text-left px-4 py-2 hover:bg-gray-50 flex items-center gap-2">
									<span class="h-2 w-2 rounded-full bg-red-500"></span> Busy
								</button>
								<button class="w-full text-left px-4 py-2 hover:bg-gray-50 flex items-center gap-2">
									<span class="h-2 w-2 rounded-full bg-gray-400"></span> Offline
								</button>
							</div>

							<div class="border-t border-gray-100 py-1">
								<a href="/profile" class="block px-4 py-2 text-gray-700 hover:bg-gray-50">Profile Settings</a>
								{#if userState.role === 'admin'}
									<a href="/settings" class="block px-4 py-2 text-gray-700 hover:bg-gray-50">App Default Settings</a>
								{/if}
							</div>
							
							<div class="border-t border-gray-100 py-1">
								<form action="/logout" method="POST">
									<button type="submit" class="w-full text-left px-4 py-2 text-red-600 hover:bg-red-50 font-medium">
										Sign out
									</button>
								</form>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</header>

	<main class="flex-1 w-full relative">
		{@render children()}
	</main>
</div>
