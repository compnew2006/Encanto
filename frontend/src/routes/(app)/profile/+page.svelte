<script lang="ts">
	import { getUserContext } from '$lib/user.svelte';

	const userState = getUserContext();
	const user = $derived(userState.value);
</script>

<svelte:head>
	<title>Profile | Encanto</title>
</svelte:head>

{#if user}
	<div class="surface max-w-3xl rounded-[30px] border border-white/60 p-8 shadow-xl">
		<div class="flex flex-col gap-5 md:flex-row md:items-center">
			<img alt={user.fullName} class="h-24 w-24 rounded-[28px] border border-[#e5ddcf]" src={user.avatarUrl} />
			<div>
				<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Profile</p>
				<h2 class="mt-2 text-3xl font-semibold text-[#14213d]">{user.fullName}</h2>
				<p class="mt-2 text-sm text-[#5a6777]">{user.email}</p>
				<p class="mt-1 text-sm text-[#5a6777]">{user.currentOrganization.name} · {user.currentOrganization.roleName}</p>
			</div>
		</div>

		<div class="mt-8 grid gap-4 md:grid-cols-2">
			<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
				<p class="text-sm text-[#6b7786]">Availability</p>
				<p class="mt-1 text-lg font-semibold text-[#14213d]">{user.availabilityStatus}</p>
			</div>
			<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
				<p class="text-sm text-[#6b7786]">Theme / language</p>
				<p class="mt-1 text-lg font-semibold text-[#14213d]">{user.settings.theme} · {user.settings.language}</p>
			</div>
		</div>
	</div>
{/if}

