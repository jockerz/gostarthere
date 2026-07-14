<script lang="ts">
    import { Sun, Moon } from "@lucide/svelte";
    import { onMount } from "svelte";

    import AppSidebar from "$lib/components/custom/dashboard/AppSidebar.svelte";
    import * as Sidebar from "$lib/components/ui/sidebar";
    import { clearTokens } from "$lib/services/api";
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";

    let sidebarOpen = $state(false);
    let dark = $state(false);
    let userMenuOpen = $state(false);

    $effect(() => {
        dark = document.documentElement.classList.contains("dark");
    });

    function toggleDark() {
        dark = !dark;
        document.documentElement.classList.toggle("dark", dark);
        localStorage.setItem("theme", dark ? "dark" : "light");
    }

    function handleLogout() {
        clearTokens();
        goto("/");
    }

    let { children } = $props();
</script>

<Sidebar.Provider>
    <AppSidebar open={sidebarOpen} onClose={() => (sidebarOpen = false)} />

    <Sidebar.Inset>
        <header
            class="bg-background sticky top-0 z-30 flex h-10 items-center border-b-0 px-6"
        >
            <Sidebar.Trigger />

            <div class="ml-auto flex items-center gap-3">
                <Button variant="ghost" size="icon" onclick={toggleDark}>
                    {#if dark}
                        <Sun class="size-5" />
                    {:else}
                        <Moon class="size-5" />
                    {/if}
                </Button>

                <!-- <div class="relative">
                    <Button
                        variant="ghost"
                        size="icon"
                        onclick={() => (userMenuOpen = !userMenuOpen)}
                    >
                        <CircleUser class="size-5" />
                    </Button>
                    {#if userMenuOpen}
                        <div
                            class="bg-popover text-popover-foreground border-border absolute right-0 top-full z-50 mt-1 w-40 overflow-hidden rounded-md border shadow-md"
                        >
                            <a
                                href="/dashboard/profile"
                                class="hover:bg-muted flex items-center gap-2 px-3 py-2 text-sm"
                                onclick={() => (userMenuOpen = false)}
                            >
                                <User class="size-4" /> Profile
                            </a>
                            <button
                                onclick={() => {
                                    userMenuOpen = false;
                                    handleLogout();
                                }}
                                class="hover:bg-muted flex w-full items-center gap-2 px-3 py-2 text-sm"
                            >
                                <LogOut class="size-4" /> Log out
                            </button>
                        </div>
                        <div
                            class="fixed inset-0 z-[-1]"
                            role="presentation"
                            onclick={() => (userMenuOpen = false)}
                        ></div>
                    {/if}
                </div> -->
            </div>
        </header>

        <main class="flex-1 p-6">
            {@render children()}
        </main>
    </Sidebar.Inset>
</Sidebar.Provider>
