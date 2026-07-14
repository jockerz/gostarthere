<script lang="ts">
	import { cn } from "$lib/utils";
	import type { Snippet } from "svelte";

	interface Props {
		children?: Snippet;
		class?: string;
	}

	let { children, class: className = "" }: Props = $props();
</script>

<div class={cn("relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full", className)}>
	{@render children?.()}
</div>
