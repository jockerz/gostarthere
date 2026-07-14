<script lang="ts">
    import { onMount } from "svelte";
    import { Button } from "$lib/components/ui/button";
    import { checkApi, getToken, profileApi } from "$lib/services/api";
    import { Sun, Moon } from "@lucide/svelte";

    let status = $state<"loading" | "up" | "down">("loading");
    let dark = $state(false);
    let isLoggedIn = $state(false);

    $effect(() => {
        dark = document.documentElement.classList.contains("dark");
    });

    $effect(() => {
        checkApi.health().then(
            (data) => (status = data.success ? "up" : "down"),
            () => (status = "down"),
        );
    });

    onMount(async () => {
        if (!getToken()) {
            isLoggedIn = false;
            return;
        }

        profileApi
            .get()
            .then(() => (isLoggedIn = true))
            .catch(() => (isLoggedIn = false));
    });

    function toggleDark() {
        dark = !dark;
        document.documentElement.classList.toggle("dark", dark);
        localStorage.setItem("theme", dark ? "dark" : "light");
    }
</script>

<header class="border-border/40 border-b">
    <nav class="mx-auto flex h-14 max-w-4xl items-center gap-6 px-4">
        <a href="/" class="font-bold">FE</a>
        <div class="ml-auto flex items-center gap-4">
            {#if !isLoggedIn}
                <a
                    href="/auth/login"
                    class="hover:text-muted-foreground text-sm hover:underline"
                    >Log in</a
                >
                <a
                    href="/auth/register"
                    class="hover:text-muted-foreground text-sm hover:underline"
                    >Register</a
                >
            {:else}
                <a
                    href="/app"
                    class="hover:text-muted-foreground text-sm hover:underline"
                    >Dashboard</a
                >
            {/if}
            <button
                onclick={toggleDark}
                class="hover:text-muted-foreground cursor-pointer"
            >
                {#if dark}
                    <Sun class="size-5" />
                {:else}
                    <Moon class="size-5" />
                {/if}
            </button>
        </div>
    </nav>
</header>

<main class="mx-auto max-w-4xl px-4 py-16">
    <section class="py-16 text-center">
        <h1 class="mb-4 text-4xl font-bold">Welcome to FE</h1>
        <p class="text-muted-foreground mb-8 text-lg">
            A modern web application
        </p>
        <div class="flex justify-center gap-4">
            {#if !isLoggedIn}
                <a href="/auth/login"><Button>Log in</Button></a>
                <a href="/auth/register"
                    ><Button variant="outline">Register</Button></a
                >
            {:else}
                <a href="/app"><Button>Dashboard</Button></a>
            {/if}
        </div>
    </section>

    <section class="text-muted-foreground text-center text-sm">
        System status:
        {#if status === "loading"}
            <span class="text-yellow-600">Checking...</span>
        {:else if status === "up"}
            <span class="text-green-600">Online</span>
        {:else}
            <span class="text-red-600">Offline</span>
        {/if}
    </section>
</main>
