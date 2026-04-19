<script lang="ts">
	import { invalidateAll } from '$app/navigation';

	import PermissionButton from '$lib/components/PermissionButton.svelte';
	import { clientJSON } from '$lib/client/backend';
	import type { UserSummary } from '$lib/types';

	import type { PageData } from './$types';

	type UserDraft = PageData['users'][number] & { saving: boolean; error: string };

	let { data }: { data: PageData } = $props();

	let drafts = $state<UserDraft[]>([]);

	function toDraft(user: PageData['users'][number]): UserDraft {
		return {
			...user,
			sendRestrictions: {
				...user.sendRestrictions
			},
			visibilityRule: {
				...user.visibilityRule,
				allowedPhoneNumbers: [...user.visibilityRule.allowedPhoneNumbers],
				allowedInstanceIds: [...user.visibilityRule.allowedInstanceIds]
			},
			saving: false,
			error: ''
		};
	}

	$effect(() => {
		drafts = data.users.map(toDraft);
	});

	function updateRestriction(index: number, key: string, mode: 'inherit' | 'allow' | 'deny') {
		const draft = drafts[index];
		draft.sendRestrictions.allowPermissionKeys = draft.sendRestrictions.allowPermissionKeys.filter((item) => item !== key);
		draft.sendRestrictions.denyPermissionKeys = draft.sendRestrictions.denyPermissionKeys.filter((item) => item !== key);
		if (mode === 'allow') draft.sendRestrictions.allowPermissionKeys = [...draft.sendRestrictions.allowPermissionKeys, key];
		if (mode === 'deny') draft.sendRestrictions.denyPermissionKeys = [...draft.sendRestrictions.denyPermissionKeys, key];
	}

	async function saveUser(index: number) {
		const draft = drafts[index];
		draft.error = '';
		draft.saving = true;

		try {
			await clientJSON(`/api/users/${draft.id}`, {
				method: 'PUT',
				body: JSON.stringify({ availabilityStatus: draft.availabilityStatus } satisfies Partial<UserSummary>)
			});
			await clientJSON(`/api/users/${draft.id}/send-restrictions`, {
				method: 'PUT',
				body: JSON.stringify(draft.sendRestrictions)
			});
			await clientJSON(`/api/users/${draft.id}/contact-visibility`, {
				method: 'PUT',
				body: JSON.stringify(draft.visibilityRule)
			});
			await invalidateAll();
		} catch (cause) {
			draft.error = cause instanceof Error ? cause.message : 'Unable to save the user overrides.';
		} finally {
			draft.saving = false;
		}
	}

	function joined(values: string[]) {
		return values.join(', ');
	}

	function splitCsv(value: string) {
		return value.split(',').map((item) => item.trim()).filter(Boolean);
	}
</script>

<svelte:head>
	<title>Users | Encanto</title>
</svelte:head>

<div class="space-y-6">
	<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
		<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Settings · Users</p>
		<h2 class="mt-2 text-3xl font-semibold text-[#14213d]">User-level overrides</h2>
		<p class="mt-2 max-w-3xl text-sm text-[#4d5e72]">
			This surface exercises Phase 4 directly: operator-level send restrictions and per-user visibility rules can change the effective UI state without bypassing backend enforcement.
		</p>
	</section>

	<div class="space-y-5">
		{#each drafts as draft, index (draft.id)}
			<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
				<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
					<div>
						<h3 class="text-2xl font-semibold text-[#14213d]">{draft.fullName}</h3>
						<p class="mt-1 text-sm text-[#5a6777]">{draft.email} · {draft.roleName}</p>
					</div>
					<label class="block min-w-44">
						<span class="mb-1 block text-xs uppercase tracking-[0.16em] text-[#6b7786]">Availability</span>
						<select bind:value={draft.availabilityStatus} class="w-full rounded-full border border-[#dfd6c6] bg-white px-3 py-2 text-sm font-semibold text-[#14213d]">
							<option value="available">Available</option>
							<option value="busy">Busy</option>
							<option value="unavailable">Unavailable</option>
						</select>
					</label>
				</div>

				<div class="mt-6 grid gap-6 xl:grid-cols-2">
					<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
						<h4 class="text-lg font-semibold text-[#14213d]">Send restrictions</h4>
						<div class="mt-4 space-y-3 text-sm">
							{#each ['messages.send', 'chats.unclaimed.send'] as key (key)}
								<label class="block">
									<span class="mb-1 block text-[#5a6777]">{key}</span>
									<select
										class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2"
										onchange={(event) => updateRestriction(index, key, (event.currentTarget as HTMLSelectElement).value as 'inherit' | 'allow' | 'deny')}
									>
										<option selected={!draft.sendRestrictions.allowPermissionKeys.includes(key) && !draft.sendRestrictions.denyPermissionKeys.includes(key)} value="inherit">Inherit role</option>
										<option selected={draft.sendRestrictions.allowPermissionKeys.includes(key)} value="allow">Force allow</option>
										<option selected={draft.sendRestrictions.denyPermissionKeys.includes(key)} value="deny">Force deny</option>
									</select>
								</label>
							{/each}
						</div>
					</div>

					<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
						<h4 class="text-lg font-semibold text-[#14213d]">Contact visibility</h4>
						<div class="mt-4 space-y-3 text-sm">
							<label class="block">
								<span class="mb-1 block text-[#5a6777]">Scope mode</span>
								<select bind:value={draft.visibilityRule.scopeMode} class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2">
									<option value="all_contacts">All contacts</option>
									<option value="instances_only">Instance-only</option>
									<option value="allowed_numbers_only">Allowed numbers only</option>
									<option value="instances_plus_allowed_numbers">Instances + allowed numbers</option>
								</select>
							</label>

						<label class="block">
							<span class="mb-1 block text-[#5a6777]">Allowed phone numbers</span>
							<textarea
								class="min-h-24 w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2"
								value={joined(draft.visibilityRule.allowedPhoneNumbers)}
								onchange={(event) => draft.visibilityRule.allowedPhoneNumbers = splitCsv((event.currentTarget as HTMLTextAreaElement).value)}
							></textarea>
						</label>

							<label class="flex items-center gap-3">
								<input bind:checked={draft.visibilityRule.inheritRoleScope} class="h-4 w-4 rounded border-[#dfd6c6]" type="checkbox" />
								<span>Inherit role scope before applying overrides</span>
							</label>

							<label class="flex items-center gap-3">
								<input bind:checked={draft.visibilityRule.canViewUnmaskedPhone} class="h-4 w-4 rounded border-[#dfd6c6]" type="checkbox" />
								<span>Allow full phone number visibility</span>
							</label>
						</div>
					</div>
				</div>

				{#if draft.error}
					<p class="mt-4 text-sm text-[#9b3d2a]">{draft.error}</p>
				{/if}

				<div class="mt-5 flex justify-end">
					<PermissionButton allowed={!draft.saving} label={draft.saving ? 'Saving…' : 'Save user overrides'} onclick={() => saveUser(index)} />
				</div>
			</section>
		{/each}
	</div>
</div>
