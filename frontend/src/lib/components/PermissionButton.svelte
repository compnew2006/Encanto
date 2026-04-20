<script lang="ts">
	let {
		allowed,
		reason = '',
		label,
		type = 'button',
		variant = 'primary',
		onclick
	}: {
		allowed: boolean;
		reason?: string;
		label: string;
		type?: 'button' | 'submit';
		variant?: 'primary' | 'secondary';
		onclick?: () => void | Promise<void>;
	} = $props();

	const variants = {
		primary: 'bg-[#2a9d8f] text-white hover:bg-[#23867a]',
		secondary: 'bg-white/80 text-[#14213d] hover:bg-white'
	};
</script>

<button
	type={type}
	class={`inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold shadow-sm transition disabled:cursor-not-allowed disabled:opacity-55 ${variants[variant]}`}
	disabled={!allowed}
	title={!allowed ? reason : undefined}
	onclick={onclick}
>
	{label}
</button>

