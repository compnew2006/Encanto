<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<svelte:head>
	<title>Settings | Encanto</title>
</svelte:head>

<div class="grid gap-6 md:grid-cols-2">
	<a class="surface rounded-[30px] border border-white/60 p-6 shadow-xl transition hover:-translate-y-0.5 hover:shadow-2xl" href={resolve('/settings/users')}>
		<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Phase 4</p>
		<h2 class="mt-3 text-2xl font-semibold text-[#14213d]">Users</h2>
		<p class="mt-2 text-sm text-[#4d5e72]">Availability, send restrictions, and per-user visibility overrides.</p>
	</a>

	<a class="surface rounded-[30px] border border-white/60 p-6 shadow-xl transition hover:-translate-y-0.5 hover:shadow-2xl" href={resolve('/settings/roles')}>
		<p class="text-xs uppercase tracking-[0.28em] text-[#f4a261]">Phase 4</p>
		<h2 class="mt-3 text-2xl font-semibold text-[#14213d]">Roles</h2>
		<p class="mt-2 text-sm text-[#4d5e72]">Action-based permissions, default scopes, and UI/backend parity.</p>
	</a>
</div>
