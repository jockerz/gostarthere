<script lang="ts">
	import * as Dialog from "$lib/components/ui/dialog";
	import { Button } from "$lib/components/ui/button";
	import type { Snippet } from "svelte";

	interface Props {
		open?: boolean;
		title?: string;
		message?: string;
		children?: Snippet;
	}

	let { open = $bindable(false), title = "Error", message = "", children }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{title}</Dialog.Title>
		</Dialog.Header>
		<p class="text-muted-foreground text-sm">{message}</p>
		<Dialog.Footer>
			{#if children}
				{@render children()}
			{/if}
			<Button onclick={() => (open = false)}>Got it</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
