<script lang="ts">
    import { page } from "$app/stores";
    import { Button } from "$lib/components/ui/button";
    import { Home, ArrowLeft, Frown, TriangleAlert } from "@lucide/svelte";

    let is404 = $derived($page.status === 404);

    let heading = $derived(is404 ? "Page Not Found" : "Something went wrong");
    let description = $derived(
        is404
            ? "The page you're looking for doesn't exist or has been moved."
            : "An unexpected error occurred. Please try again later.",
    );
</script>

<main class="bg-muted/30 flex min-h-screen items-center justify-center p-4">
    <div class="w-full max-w-sm">
        <div class="text-muted-foreground mb-6 text-center">
            <Button variant="ghost">
                {#if is404}
                    <Frown class="size-24" />
                {:else}
                    <TriangleAlert class="size-24" />
                {/if}
            </Button>
        </div>

        <h1 class="mb-2 text-6xl font-bold text-center">
            {is404 ? "404" : "Error"}
        </h1>
        <p class="text-muted-foreground mb-8 max-w-md text-center text-lg">
            {description}
        </p>

        <div class="flex gap-4 justify-center">
            <Button variant="outline" onclick={() => history.back()}>
                <ArrowLeft class="size-4" />
                Go Back
            </Button>
            <a href="/"><Button><Home class="size-4" /> Home</Button></a>
        </div>
    </div>
</main>
