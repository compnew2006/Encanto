<script lang="ts">
	let { data }: { data: any } = $props();
</script>

<svelte:head>
	<title>Chat Workspace</title>
</svelte:head>

<div class="h-full flex flex-col items-center justify-center py-20 text-center">
	<h2 class="text-3xl font-bold text-gray-800 mb-2">Welcome to your Inbox</h2>
	<p class="text-gray-600 mb-8">You are securely logged in as <span class="font-semibold">{data.user?.email}</span></p>
	<div class="bg-blue-50 border border-blue-200 rounded p-6 shadow-sm max-w-md w-full text-blue-800">
		<p class="text-sm">This page is protected by server-side hooks validating your HTTPOnly session cookie. Realtime features will drop securely into this workspace.</p>
	</div>
</div>
