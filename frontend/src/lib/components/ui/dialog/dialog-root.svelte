<script lang="ts">
	import type { Snippet } from "svelte";

	interface Props {
		children?: Snippet;
		open?: boolean;
		onopenchange?: (open: boolean) => void;
	}

	let { children, open = $bindable(false), onopenchange }: Props = $props();
</script>

{#if open}
	{@render children?.()}
{/if}
