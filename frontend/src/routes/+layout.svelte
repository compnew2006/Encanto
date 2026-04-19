<script lang="ts">
	import '../app.css';
	import type { PageData } from './$types';

	let { data, children }: { data: PageData, children: any } = $props();
</script>

<div class="min-h-screen bg-gray-50 flex flex-col">
	<header class="bg-white shadow">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex justify-between items-center">
			<h1 class="text-xl font-bold text-gray-900">Encanto Workspace</h1>
			{#if data.user}
				<div class="flex items-center gap-4">
					<span class="text-sm text-gray-600">{data.user.email}</span>
					<form action="/logout" method="POST">
						<button class="bg-red-50 text-red-600 px-3 py-1 rounded text-sm hover:bg-red-100">
							Logout
						</button>
					</form>
				</div>
			{/if}
		</div>
	</header>

	<main class="flex-1 w-full relative">
		{@render children()}
	</main>
</div>
