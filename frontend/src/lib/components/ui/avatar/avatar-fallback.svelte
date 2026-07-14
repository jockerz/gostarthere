<script lang="ts">
	import { cn } from "$lib/utils";
	import type { Snippet } from "svelte";

	interface Props {
		children?: Snippet;
		class?: string;
	}

	let { children, class: className = "" }: Props = $props();
</script>

<div class={cn("bg-muted flex h-full w-full items-center justify-center rounded-full", className)}>
	{@render children?.()}
</div>
