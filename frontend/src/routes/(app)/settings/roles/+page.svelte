<script lang="ts">
	import { invalidateAll } from '$app/navigation';

	import PermissionButton from '$lib/components/PermissionButton.svelte';
	import { clientJSON } from '$lib/client/backend';
	import { groupPermissions } from '$lib/permissions';
	import type { RoleDefinition, VisibilityScopeMode } from '$lib/types';

	import type { PageData } from './$types';

	type RoleDraft = RoleDefinition & { saving: boolean; error: string };

	let { data }: { data: PageData } = $props();

	let drafts = $state<RoleDraft[]>([]);

	function toDraft(role: RoleDefinition): RoleDraft {
		return {
			...role,
			permissionKeys: [...role.permissionKeys],
			defaultVisibility: {
				...role.defaultVisibility,
				allowedPhoneNumbers: [...role.defaultVisibility.allowedPhoneNumbers],
				allowedInstanceIds: [...role.defaultVisibility.allowedInstanceIds]
			},
			saving: false,
			error: ''
		};
	}

	function createRoleDraft() {
		return {
			name: '',
			description: '',
			permissionKeys: ['chats.view', 'messages.view', 'contacts.view'],
			defaultVisibility: {
				mode: 'all_contacts' as VisibilityScopeMode,
				allowedPhoneNumbers: [] as string[],
				allowedInstanceIds: [] as string[],
				canViewUnmaskedPhone: true
			},
			error: ''
		};
	}

	let newRole = $state(createRoleDraft());

	$effect(() => {
		drafts = data.roles.map(toDraft);
	});

	const grouped = $derived(groupPermissions(data.permissions));

	function splitCsv(value: string) {
		return value.split(',').map((item) => item.trim()).filter(Boolean);
	}

	function updateAllowedPhoneNumbers(role: RoleDraft, value: string) {
		role.defaultVisibility.allowedPhoneNumbers = splitCsv(value);
	}

	function createPermissionToggle(role: RoleDraft, key: string) {
		return () => togglePermission(role, key);
	}

	function togglePermission(target: { permissionKeys: string[] }, key: string) {
		target.permissionKeys = target.permissionKeys.includes(key)
			? target.permissionKeys.filter((item) => item !== key)
			: [...target.permissionKeys, key];
	}

	async function saveRole(role: RoleDefinition & { saving: boolean; error: string }, index: number) {
		role.error = '';
		role.saving = true;
		try {
			const payload = {
				name: role.name,
				description: role.description,
				permissionKeys: role.permissionKeys,
				defaultVisibility: role.defaultVisibility
			};
			await clientJSON(`/api/roles/${role.id}`, {
				method: 'PUT',
				body: JSON.stringify(payload)
			});
			await invalidateAll();
		} catch (cause) {
			drafts[index].error = cause instanceof Error ? cause.message : 'Unable to save the role.';
		} finally {
			role.saving = false;
		}
	}

	async function deleteRole(roleId: string) {
		await clientJSON(`/api/roles/${roleId}`, { method: 'DELETE' });
		await invalidateAll();
	}

	async function createRole() {
		newRole.error = '';
		try {
			await clientJSON('/api/roles', {
				method: 'POST',
				body: JSON.stringify(newRole)
			});
			newRole = createRoleDraft();
			await invalidateAll();
		} catch (cause) {
			newRole.error = cause instanceof Error ? cause.message : 'Unable to create the role.';
		}
	}
</script>

<svelte:head>
	<title>Roles | Encanto</title>
</svelte:head>

<div class="space-y-6">
	<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
		<p class="text-xs uppercase tracking-[0.28em] text-[#2a9d8f]">Settings · Roles</p>
		<h2 class="mt-2 text-3xl font-semibold text-[#14213d]">Action-based role matrix</h2>
		<p class="mt-2 max-w-3xl text-sm text-[#4d5e72]">
			These definitions drive both UI helpers and backend authorization. Scope permissions are normalized from the selected visibility mode so the matrix cannot drift.
		</p>
	</section>

	<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
		<h3 class="text-xl font-semibold text-[#14213d]">Create role</h3>
		<div class="mt-4 grid gap-4 lg:grid-cols-2">
			<label class="block">
				<span class="mb-1 block text-sm text-[#5a6777]">Name</span>
				<input bind:value={newRole.name} class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" />
			</label>
			<label class="block">
				<span class="mb-1 block text-sm text-[#5a6777]">Description</span>
				<input bind:value={newRole.description} class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2" />
			</label>
		</div>
		<div class="mt-4 flex justify-end">
			<PermissionButton allowed={Boolean(newRole.name)} label="Create role" onclick={createRole} />
		</div>
		{#if newRole.error}
			<p class="mt-3 text-sm text-[#9b3d2a]">{newRole.error}</p>
		{/if}
	</section>

	{#each drafts as role, index (role.id)}
		<section class="surface rounded-[30px] border border-white/60 p-6 shadow-xl">
			<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
				<div>
					<h3 class="text-2xl font-semibold text-[#14213d]">{role.name}</h3>
					<p class="mt-1 text-sm text-[#5a6777]">{role.description}</p>
				</div>
				<div class="flex gap-2">
					{#if role.isSystem}
						<span class="rounded-full bg-[#14213d] px-3 py-1 text-xs font-semibold text-white">System</span>
					{/if}
					{#if role.isDefault}
						<span class="rounded-full bg-[#2a9d8f]/20 px-3 py-1 text-xs font-semibold text-[#1f776c]">Default</span>
					{/if}
				</div>
			</div>

			<div class="mt-6 grid gap-6 xl:grid-cols-[0.65fr_1.35fr]">
				<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
					<h4 class="text-lg font-semibold text-[#14213d]">Visibility defaults</h4>
					<div class="mt-4 space-y-3 text-sm">
						<label class="block">
							<span class="mb-1 block text-[#5a6777]">Scope mode</span>
							<select bind:value={role.defaultVisibility.mode} class="w-full rounded-2xl border border-[#dfd6c6] bg-white px-3 py-2">
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
								value={role.defaultVisibility.allowedPhoneNumbers.join(', ')}
								onchange={(event) => updateAllowedPhoneNumbers(role, (event.currentTarget as HTMLTextAreaElement).value)}
							></textarea>
						</label>

						<label class="flex items-center gap-3">
							<input bind:checked={role.defaultVisibility.canViewUnmaskedPhone} class="h-4 w-4 rounded border-[#dfd6c6]" type="checkbox" />
							<span>Allow full phone visibility</span>
						</label>
					</div>
				</div>

				<div class="rounded-[24px] border border-[#e5ddcf] bg-white/75 p-4">
					<h4 class="text-lg font-semibold text-[#14213d]">Permission matrix</h4>
					<div class="mt-4 grid gap-4 xl:grid-cols-2">
						{#each Object.entries(grouped) as [resource, permissions] (resource)}
							<div class="rounded-[20px] border border-[#efe8da] bg-[#fbf8f1] p-4">
								<p class="text-sm font-semibold capitalize text-[#14213d]">{resource}</p>
								<div class="mt-3 space-y-2">
									{#each permissions as permission (permission.key)}
										<label class="flex gap-3 rounded-2xl px-2 py-2 hover:bg-white/80">
										<input checked={role.permissionKeys.includes(permission.key)} class="mt-1 h-4 w-4 rounded border-[#dfd6c6]" type="checkbox" onchange={createPermissionToggle(role, permission.key)} />
										<span class="text-sm text-[#4d5e72]">
											<span class="block font-medium text-[#14213d]">{permission.label}</span>
											<span class="block text-xs">{permission.description}</span>
											</span>
										</label>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			{#if role.error}
				<p class="mt-4 text-sm text-[#9b3d2a]">{role.error}</p>
			{/if}

			<div class="mt-5 flex flex-wrap justify-end gap-3">
				{#if !role.isSystem}
					<PermissionButton allowed={!role.saving} label="Delete" onclick={() => deleteRole(role.id)} variant="secondary" />
				{/if}
				<PermissionButton allowed={!role.saving} label={role.saving ? 'Saving…' : 'Save role'} onclick={() => saveRole(role, index)} />
			</div>
		</section>
	{/each}
</div>
