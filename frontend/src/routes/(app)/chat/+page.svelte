<script lang="ts">
	import { resolve } from '$app/paths';

	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const assignedChats = $derived(data.chats.filter((chat) => chat.status !== 'pending'));
	const pendingChats = $derived(data.chats.filter((chat) => chat.status === 'pending'));
</script>

<svelte:head>
	<title>Chat | Encanto</title>
</svelte:head>

<div class="space-y-6">
	<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
		<div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
			<div>
				<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Scoped inbox</p>
				<h2 class="mt-2 text-3xl font-semibold text-[#14213d]">Chats you are actually allowed to see</h2>
				<p class="mt-2 max-w-2xl text-sm text-[#4d5e72]">
					List, search, and direct detail access all share the same backend visibility filter. Pending chats disappear entirely when the current role cannot view them.
				</p>
			</div>

			<form class="flex gap-3" method="GET">
				<input class="w-64 rounded-full border border-[#dfd6c6] bg-white px-4 py-3 text-sm outline-none focus:border-[#2a9d8f]" name="search" placeholder="Search name or phone" value={data.search} />
				<button class="rounded-full bg-[#f4a261] px-4 py-3 text-sm font-semibold text-[#14213d]" type="submit">Search</button>
			</form>
		</div>
	</section>

	<section class="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-xl font-semibold text-[#14213d]">Assigned chats</h3>
				<span class="rounded-full bg-[#14213d] px-3 py-1 text-xs font-semibold text-white">{assignedChats.length}</span>
			</div>

			<div class="space-y-3">
				{#each assignedChats as chat (chat.id)}
					<a class="block rounded-[24px] border border-[#e5ddcf] bg-white/80 p-4 transition hover:-translate-y-0.5 hover:shadow-md" href={resolve(`/chat/${chat.id}`)}>
						<div class="flex items-start justify-between gap-4">
							<div>
								<p class="font-semibold text-[#14213d]">{chat.name}</p>
								<p class="text-sm text-[#5a6777]">{chat.visiblePhone} · {chat.instanceName}</p>
							</div>
							<div class="flex gap-2">
								{#if chat.isPinned}
									<span class="rounded-full bg-[#14213d] px-2 py-1 text-[11px] font-semibold text-white">Pinned</span>
								{/if}
								<span class="rounded-full bg-[#2a9d8f]/15 px-2 py-1 text-[11px] font-semibold text-[#1f776c]">{chat.status}</span>
							</div>
						</div>
						<p class="mt-3 text-sm text-[#4d5e72]">{chat.lastMessagePreview}</p>
					</a>
				{:else}
					<p class="rounded-[24px] border border-dashed border-[#d6cbb7] px-4 py-10 text-center text-sm text-[#5a6777]">No assigned chats match the current scope.</p>
				{/each}
			</div>
		</div>

		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-xl font-semibold text-[#14213d]">Pending chats</h3>
				<span class="rounded-full bg-[#f4a261] px-3 py-1 text-xs font-semibold text-[#14213d]">{pendingChats.length}</span>
			</div>

			{#if pendingChats.length}
				<div class="space-y-3">
					{#each pendingChats as chat (chat.id)}
						<a class="block rounded-[24px] border border-[#e5ddcf] bg-white/80 p-4 transition hover:-translate-y-0.5 hover:shadow-md" href={resolve(`/chat/${chat.id}`)}>
							<p class="font-semibold text-[#14213d]">{chat.name}</p>
							<p class="mt-1 text-sm text-[#5a6777]">{chat.visiblePhone}</p>
							<p class="mt-3 text-sm text-[#4d5e72]">{chat.lastMessagePreview}</p>
						</a>
					{/each}
				</div>
			{:else}
				<p class="rounded-[24px] border border-dashed border-[#d6cbb7] px-4 py-10 text-center text-sm text-[#5a6777]">
					No pending chats are visible in the current permission scope.
				</p>
			{/if}
		</div>
	</section>
</div>
