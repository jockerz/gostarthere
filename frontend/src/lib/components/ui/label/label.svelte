<script lang="ts">
	import { cn } from "$lib/utils";
	import type { Snippet } from "svelte";

	interface Props {
		children?: Snippet;
		class?: string;
		for?: string;
	}

	let {
		children,
		class: className = "",
		for: htmlFor,
	}: Props = $props();
</script>

<label
	class={cn(
		"text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
		className,
	)}
	for={htmlFor}
>
	{@render children?.()}
</label>
