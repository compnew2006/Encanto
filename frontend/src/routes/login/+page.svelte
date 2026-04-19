<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import { clientJSON } from '$lib/client/backend';

	let email = $state('admin@example.com');
	let password = $state('password123');
	let error = $state('');
	let loading = $state(false);

	async function submit() {
		error = '';
		loading = true;
		try {
			await clientJSON<{ user: unknown }>('/api/auth/login', {
				method: 'POST',
				body: JSON.stringify({ email, password })
			});
			await goto(resolve('/chat'));
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to sign in.';
		} finally {
			loading = false;
		}
	}

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		void submit();
	}
</script>

<svelte:head>
	<title>Login | Encanto</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center p-6">
	<div class="surface w-full max-w-4xl overflow-hidden rounded-[32px] border border-white/70 shadow-2xl">
		<div class="grid md:grid-cols-[1.1fr_0.9fr]">
			<section class="bg-[#14213d] px-8 py-10 text-[#f7f3eb] md:px-10">
				<p class="mb-3 text-sm uppercase tracking-[0.28em] text-[#f4a261]">Phase 3-4 Baseline</p>
				<h1 class="text-4xl font-semibold leading-tight">Secure access, scoped visibility, and UI parity in one workspace.</h1>
				<p class="mt-5 max-w-lg text-sm leading-7 text-[#dbe3ea]">
					This build only exposes the first operational surfaces that are actually implemented:
					authentication, context switching, permission-aware chat visibility, roles, users, and profile.
				</p>
				<div class="mt-8 rounded-3xl border border-white/10 bg-white/8 p-5 text-sm text-[#dbe3ea]">
					<p class="font-semibold text-white">Seeded sign-in</p>
					<p class="mt-2">`admin@example.com` / `password123`</p>
					<p class="mt-1">`internal@example.com` / `password123`</p>
					<p class="mt-1">`readonly@example.com` / `password123`</p>
					<p class="mt-1">`scoped@example.com` / `password123`</p>
				</div>
			</section>

		<section class="px-8 py-10 md:px-10">
			<p class="text-sm uppercase tracking-[0.28em] text-[#2a9d8f]">Sign In</p>
			<h2 class="mt-3 text-2xl font-semibold text-[#14213d]">Open your current organization context</h2>

			<form class="mt-8 space-y-5" onsubmit={handleSubmit}>
					<label class="block">
						<span class="mb-2 block text-sm font-medium text-[#3b4d63]">Email</span>
						<input class="w-full rounded-2xl border border-[#d9d4c6] bg-white px-4 py-3 outline-none transition focus:border-[#2a9d8f]" bind:value={email} type="email" />
					</label>

					<label class="block">
						<span class="mb-2 block text-sm font-medium text-[#3b4d63]">Password</span>
						<input class="w-full rounded-2xl border border-[#d9d4c6] bg-white px-4 py-3 outline-none transition focus:border-[#2a9d8f]" bind:value={password} type="password" />
					</label>

					{#if error}
						<p class="rounded-2xl border border-[#e76f51]/30 bg-[#fbe9e4] px-4 py-3 text-sm text-[#9b3d2a]">{error}</p>
					{/if}

					<button
						class="w-full rounded-full bg-[#2a9d8f] px-4 py-3 text-sm font-semibold text-white transition hover:bg-[#23867a] disabled:cursor-wait disabled:opacity-70"
						disabled={loading}
						type="submit"
					>
						{loading ? 'Signing in…' : 'Sign in'}
					</button>
				</form>
			</section>
		</div>
	</div>
</div>
