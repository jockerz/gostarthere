<script lang="ts">
	import { cn } from "$lib/utils";

	interface Props {
		class?: string;
		src?: string;
		alt?: string;
	}

	let { class: className = "", src, alt }: Props = $props();
</script>

<img class={cn("aspect-square h-full w-full", className)} {src} {alt} />
