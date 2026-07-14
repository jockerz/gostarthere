<script lang="ts">
	import { X } from "@lucide/svelte";
	import { cn } from "$lib/utils";

	interface Props {
		class?: string;
		onclick?: () => void;
	}

	let { class: className = "", onclick }: Props = $props();
</script>

<button
	class={cn(
		"ring-offset-background focus:ring-ring data-[state=open]:bg-accent data-[state=open]:text-muted-foreground absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-offset-2",
		className,
	)}
	{onclick}
>
	<X class="h-4 w-4" />
	<span class="sr-only">Close</span>
</button>
