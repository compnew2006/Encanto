<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
	let email = $state('');

	$effect(() => {
		if (typeof form?.email === 'string') {
			email = form.email;
		}
	});
</script>

<svelte:head>
	<title>Login - Encanto</title>
</svelte:head>

<div class="flex items-center justify-center min-h-[80vh]">
	<div class="bg-white p-8 shadow rounded-lg max-w-sm w-full">
		<h2 class="text-2xl font-semibold mb-6 text-center text-gray-800">Sign in</h2>
		
		<form method="POST" use:enhance class="space-y-4">
			<div>
				<label for="email" class="block text-sm text-gray-600 mb-1">Email</label>
				<input id="email" name="email" type="email" bind:value={email} class="w-full border-gray-300 rounded px-3 py-2 border outline-none focus:ring-2 focus:ring-blue-500" required />
			</div>
			<div>
				<label for="password" class="block text-sm text-gray-600 mb-1">Password</label>
				<input id="password" name="password" type="password" class="w-full border-gray-300 rounded px-3 py-2 border outline-none focus:ring-2 focus:ring-blue-500" required />
			</div>
			
			{#if form?.missing}
				<p class="text-red-500 text-sm">Please fill out all fields.</p>
			{/if}
			{#if form?.incorrect}
				<p class="text-red-500 text-sm">{form?.message || 'Invalid credentials!'}</p>
			{/if}
			{#if form?.error}
				<p class="text-red-500 text-sm">{form?.error}</p>
			{/if}

			<!-- Hint for UI testing -->
			<p class="text-xs text-gray-400 mb-4">Hint: admin@example.com / password123</p>

			<button type="submit" class="w-full bg-blue-600 text-white font-medium py-2 rounded hover:bg-blue-700 transition">
				Sign In
			</button>
		</form>
	</div>
</div>
