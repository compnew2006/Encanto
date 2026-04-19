<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch, formatDateTime, type SettingsSummary } from '$lib/api';

	type SettingsTab = 'general' | 'appearance' | 'chat' | 'notifications';

	let settings = $state<SettingsSummary | null>(null);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let activeTab = $state<SettingsTab>('general');

	onMount(async () => {
		try {
			settings = await apiFetch<SettingsSummary>('/api/settings/summary');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load settings.';
		} finally {
			loading = false;
		}
	});

	async function save(path: string, body: object, message: string) {
		error = '';
		success = '';
		try {
			await apiFetch(path, { method: 'PUT', body });
			settings = await apiFetch<SettingsSummary>('/api/settings/summary');
			success = message;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Save failed.';
		}
	}

	async function runCleanup() {
		error = '';
		success = '';
		try {
			await apiFetch('/api/settings/uploads-cleanup/run', { method: 'POST' });
			settings = await apiFetch<SettingsSummary>('/api/settings/summary');
			success = 'Cleanup job completed.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Cleanup failed.';
		}
	}

	async function saveGeneralSettings() {
		if (!settings) return;
		await save('/api/settings/general', settings.general, 'General settings saved.');
	}

	async function saveAppearanceSettings() {
		if (!settings) return;
		await save('/api/settings/appearance', settings.appearance, 'Appearance settings saved.');
	}

	async function saveChatSettings() {
		if (!settings) return;
		await save('/api/settings/chat', settings.chat, 'Chat settings saved.');
	}

	async function saveNotificationSettings() {
		if (!settings) return;
		await save('/api/settings/notifications', settings.notifications, 'Notification settings saved.');
	}
</script>

<div class="mx-auto max-w-6xl px-5 py-6">
	<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.25em] text-blue-600">Settings Center</p>
			<h1 class="mt-2 text-3xl font-semibold text-gray-900">Settings</h1>
			<p class="mt-2 text-sm text-gray-500">General, appearance, chat, and notification settings that persist and alter the product surface.</p>
		</div>
		<a href="/settings/instances" class="rounded-full border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-700 hover:border-blue-300 hover:text-blue-700">Open Account Operations</a>
	</div>

	{#if error}
		<div class="mb-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
	{/if}
	{#if success}
		<div class="mb-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{success}</div>
	{/if}

	{#if loading}
		<div class="rounded-[2rem] border border-gray-200 bg-white px-6 py-5 shadow-sm text-gray-600">Loading settings...</div>
	{:else if settings}
		<div class="grid gap-4 lg:grid-cols-[0.75fr_1.25fr]">
			<aside class="rounded-[2rem] border border-gray-200 bg-white p-4 shadow-sm">
				<nav class="space-y-2">
					<button class={`w-full rounded-[1.5rem] px-4 py-3 text-left text-sm font-medium ${activeTab === 'general' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`} onclick={() => activeTab = 'general'}>General</button>
					<button class={`w-full rounded-[1.5rem] px-4 py-3 text-left text-sm font-medium ${activeTab === 'appearance' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`} onclick={() => activeTab = 'appearance'}>Appearance</button>
					<button class={`w-full rounded-[1.5rem] px-4 py-3 text-left text-sm font-medium ${activeTab === 'chat' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`} onclick={() => activeTab = 'chat'}>Chat</button>
					<button class={`w-full rounded-[1.5rem] px-4 py-3 text-left text-sm font-medium ${activeTab === 'notifications' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700'}`} onclick={() => activeTab = 'notifications'}>Notifications</button>
				</nav>

				<div class="mt-5 grid grid-cols-2 gap-3">
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Members</p>
						<p class="mt-1 text-lg font-semibold text-gray-900">{settings.general.active_members}/{settings.general.max_members}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Instances</p>
						<p class="mt-1 text-lg font-semibold text-gray-900">{settings.general.used_instances}/{settings.general.max_instances}</p>
					</div>
					<div class="rounded-[1.5rem] bg-gray-50 px-4 py-3 col-span-2">
						<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Storage</p>
						<p class="mt-1 text-lg font-semibold text-gray-900">{settings.general.storage_used_label} / {settings.general.storage_limit_label}</p>
					</div>
				</div>
			</aside>

			<section class="rounded-[2rem] border border-gray-200 bg-white p-6 shadow-sm">
				{#if activeTab === 'general'}
					<div class="grid gap-4 md:grid-cols-2">
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Organization Name</span>
							<input bind:value={settings.general.organization_name} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Slug</span>
							<input bind:value={settings.general.slug} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Timezone</span>
							<input bind:value={settings.general.timezone} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Date Format</span>
							<input bind:value={settings.general.date_format} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Locale</span>
							<select bind:value={settings.general.locale} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="en">English</option>
								<option value="ar">Arabic</option>
							</select>
						</label>
						<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
							<span>Mask phone numbers</span>
							<input bind:checked={settings.general.mask_phone_numbers} type="checkbox" />
						</label>
					</div>
					<button data-testid="save-general-settings" class="mt-6 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveGeneralSettings}>Save General</button>
				{:else if activeTab === 'appearance'}
					<div class="grid gap-4 md:grid-cols-2">
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Color Mode</span>
							<select bind:value={settings.appearance.color_mode} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="light">Light</option>
								<option value="dark">Dark</option>
								<option value="system">System</option>
							</select>
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Theme Preset</span>
							<select bind:value={settings.appearance.theme_preset} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="ocean-breeze">Ocean Breeze</option>
								<option value="soft-pop">Soft Pop</option>
								<option value="amber-minimal">Amber Minimal</option>
							</select>
						</label>
					</div>
					<button data-testid="save-appearance-settings" class="mt-6 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveAppearanceSettings}>Save Appearance</button>
				{:else if activeTab === 'chat'}
					<div class="grid gap-4 md:grid-cols-2">
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Media Grouping Window (minutes)</span>
							<input bind:value={settings.chat.media_grouping_window_minutes} type="number" min="1" class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Sidebar Contact View</span>
							<select bind:value={settings.chat.sidebar_contact_view} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="compact">Compact</option>
								<option value="comfortable">Comfortable</option>
							</select>
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Chat Background</span>
							<select bind:value={settings.chat.chat_background} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="paper-grid">Paper Grid</option>
								<option value="linen">Linen</option>
								<option value="plain">Plain</option>
							</select>
						</label>
						<div class="grid gap-3">
							<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
								<span>Hover expand sidebar</span>
								<input bind:checked={settings.chat.sidebar_hover_expand} type="checkbox" />
							</label>
							<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
								<span>Pin sidebar</span>
								<input bind:checked={settings.chat.pin_sidebar} type="checkbox" />
							</label>
							<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
								<span>Show print buttons</span>
								<input bind:checked={settings.chat.show_print_buttons} type="checkbox" />
							</label>
							<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
								<span>Show download buttons</span>
								<input bind:checked={settings.chat.show_download_buttons} type="checkbox" />
							</label>
						</div>
					</div>
					<button data-testid="save-chat-settings" class="mt-6 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveChatSettings}>Save Chat Settings</button>
				{:else}
					<div class="grid gap-4 md:grid-cols-2">
						<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
							<span>Email Notifications</span>
							<input bind:checked={settings.notifications.email_notifications} type="checkbox" />
						</label>
						<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
							<span>New Message Alerts</span>
							<input bind:checked={settings.notifications.new_message_alerts} type="checkbox" />
						</label>
						<label class="block">
							<span class="mb-2 block text-sm font-medium text-gray-700">Notification Sound</span>
							<select bind:value={settings.notifications.notification_sound} class="w-full rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm outline-none focus:border-blue-400">
								<option value="soft-bell">Soft Bell</option>
								<option value="digital-ping">Digital Ping</option>
								<option value="none">None</option>
							</select>
						</label>
						<label class="flex items-center justify-between rounded-[1.5rem] border border-gray-200 px-4 py-3 text-sm text-gray-700">
							<span>Campaign Updates</span>
							<input bind:checked={settings.notifications.campaign_updates} type="checkbox" />
						</label>
					</div>
					<button data-testid="save-notification-settings" class="mt-6 rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white" onclick={saveNotificationSettings}>Save Notifications</button>
				{/if}

				<div class="mt-8 rounded-[1.75rem] border border-gray-200 bg-gray-50 px-5 py-4">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Uploads Cleanup</p>
							<p class="mt-1 text-sm text-gray-600">Run cleanup manually or review the scheduled configuration.</p>
						</div>
						<button data-testid="run-cleanup" class="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700" onclick={runCleanup}>Run Cleanup Now</button>
					</div>
					<div class="mt-4 grid gap-3 md:grid-cols-3 text-sm text-gray-600">
						<div class="rounded-[1.5rem] bg-white px-4 py-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Retention</p>
							<p class="mt-1">{settings.cleanup.retention_days} days</p>
						</div>
						<div class="rounded-[1.5rem] bg-white px-4 py-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Schedule</p>
							<p class="mt-1">{settings.cleanup.run_hour}:00 {settings.cleanup.timezone}</p>
						</div>
						<div class="rounded-[1.5rem] bg-white px-4 py-3">
							<p class="text-xs font-semibold uppercase tracking-wide text-gray-400">Last Run</p>
							<p class="mt-1">{formatDateTime(settings.cleanup.last_run_at)}</p>
							<p class="text-xs text-gray-400">{settings.cleanup.last_job_status}</p>
						</div>
					</div>
				</div>
			</section>
		</div>
	{/if}
</div>
