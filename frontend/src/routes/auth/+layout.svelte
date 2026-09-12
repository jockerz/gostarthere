<script lang="ts">
	import { onMount } from "svelte";

	let { children } = $props();
	let checking = $state(true);

	onMount(() => {
		checking = false;
	});
</script>

{#if !checking}
	<main class="bg-muted/30 flex min-h-screen items-center justify-center p-4">
		<div class="w-full max-w-sm">
			{@render children()}
			<div class="my-4">
    			<p class="text-muted-foreground text-center text-sm">
                    <a href="/" class="hover:text-primary hover:underline">Go to Home</a>
                </p>
			</div>
		</div>
	</main>
{/if}
