<script lang="ts">
	import PermissionButton from '$lib/components/PermissionButton.svelte';

	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>{data.detail.chat.name} | Encanto</title>
</svelte:head>

<div class="grid gap-6 xl:grid-cols-[1.25fr_0.75fr]">
	<section class="space-y-6">
		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
				<div>
					<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">{data.detail.chat.status}</p>
					<h2 class="mt-2 text-3xl font-semibold text-[#14213d]">{data.detail.chat.name}</h2>
					<p class="mt-2 text-sm text-[#5a6777]">{data.detail.chat.visiblePhone} · {data.detail.chat.instanceName}</p>
				</div>

				<div class="rounded-[24px] border border-[#dfd6c6] bg-white/70 px-4 py-3 text-sm text-[#4d5e72]">
					<p class="font-semibold text-[#14213d]">Composer state</p>
					<p class="mt-1">{data.detail.composer.allowed ? 'Enabled' : 'Disabled by current permissions'}</p>
				</div>
			</div>
		</div>

		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-xl font-semibold text-[#14213d]">Message history</h3>
				<span class="rounded-full bg-[#14213d] px-3 py-1 text-xs font-semibold text-white">{data.detail.messages.length}</span>
			</div>

			<div class="space-y-3">
				{#each data.detail.messages as message (message.id)}
					<div class={`rounded-[24px] px-4 py-3 text-sm ${message.direction === 'outbound' ? 'ml-auto max-w-[75%] bg-[#14213d] text-white' : 'mr-auto max-w-[75%] bg-[#f7f3eb] text-[#14213d]'}`}>
						<p>{message.body}</p>
						<p class={`mt-2 text-[11px] uppercase tracking-[0.18em] ${message.direction === 'outbound' ? 'text-white/70' : 'text-[#7d8a99]'}`}>
							{message.direction} · {message.status}
						</p>
					</div>
				{/each}
			</div>
		</div>

		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<h3 class="text-xl font-semibold text-[#14213d]">Composer</h3>
			<p class="mt-2 text-sm text-[#5a6777]">
				This phase proves permission parity only. Sending is intentionally deferred, so the composer shows the correct enabled or disabled state without posting messages yet.
			</p>

			<textarea class="mt-4 min-h-36 w-full rounded-[24px] border border-[#dfd6c6] bg-white px-4 py-3 text-sm outline-none disabled:bg-[#f3eee4] disabled:text-[#9aa3ad]" disabled={data.detail.composer.disabled} placeholder={data.detail.composer.allowed ? 'Compose a message…' : data.detail.composer.denialReason}></textarea>

			<div class="mt-4 flex items-center justify-between gap-4">
				{#if data.detail.composer.denialReason}
					<p class="max-w-xl text-sm text-[#9b3d2a]">{data.detail.composer.denialReason}</p>
				{:else}
					<p class="text-sm text-[#5a6777]">Send actions arrive in the next implementation phase.</p>
				{/if}

				<PermissionButton allowed={data.detail.composer.allowed} label="Send" reason={data.detail.composer.denialReason} />
			</div>
		</div>
	</section>

	<aside class="space-y-6">
		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<h3 class="text-xl font-semibold text-[#14213d]">Notes</h3>
			<div class="mt-4 space-y-3">
				{#each data.detail.notes as note (note.id)}
					<div class="rounded-[24px] border border-[#e5ddcf] bg-white/80 px-4 py-3 text-sm">
						<p class="font-semibold text-[#14213d]">{note.authorName}</p>
						<p class="mt-2 text-[#4d5e72]">{note.body}</p>
					</div>
				{:else}
					<p class="rounded-[24px] border border-dashed border-[#d6cbb7] px-4 py-6 text-center text-sm text-[#5a6777]">No notes were recorded for this chat.</p>
				{/each}
			</div>
		</div>

		<div class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<h3 class="text-xl font-semibold text-[#14213d]">Contact info</h3>
			<dl class="mt-4 space-y-3 text-sm">
				<div>
					<dt class="text-[#6b7786]">Name</dt>
					<dd class="font-medium text-[#14213d]">{data.detail.chat.name}</dd>
				</div>
				<div>
					<dt class="text-[#6b7786]">Visible phone</dt>
					<dd class="font-medium text-[#14213d]">{data.detail.chat.visiblePhone}</dd>
				</div>
				<div>
					<dt class="text-[#6b7786]">Source</dt>
					<dd class="font-medium text-[#14213d]">{data.detail.chat.instanceName}</dd>
				</div>
			</dl>
		</div>
	</aside>
</div>
