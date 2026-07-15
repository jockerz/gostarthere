<script lang="ts">
    import { page } from "$app/stores";
    import { LayoutDashboard, User, Lock, Mail, Settings, LogOut } from "@lucide/svelte";
    import * as Sidebar from "$lib/components/ui/sidebar";
    interface Props {
        open: boolean;
        onClose: () => void;
    }

    let { open, onClose }: Props = $props();

    const menu = [
        {
            title: "Dashboard",
            url: "/app",
            // path: "",
            icon: LayoutDashboard,
        },
        {
            title: "Settings",
            icon: Settings,
            path: "settings",
            children: [
                {
                    title: "Profile",
                    path: "profile",
                    // Explicit
                    // url: "/app/settings/profile"
                    icon: User,
                },
                {
                    title: "Email",
                    path: "email",
                    icon: Mail,
                },
                {
                    title: "Authentication",
                    path: "authentication",
                    icon: Lock,
                },
            ]
        }
    ]
</script>

{#if open}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
        class="fixed inset-0 z-40 bg-black/50 lg:hidden"
        role="presentation"
        onclick={onClose}
    ></div>
{/if}

<Sidebar.Root>
    <Sidebar.Header class="border-sidebar-border border-b px-6 font-bold">
        <a href="/">Gostarthere</a>
    </Sidebar.Header>
    <Sidebar.Content>
        {const rootPath = "/app"}
        {#each menu as item (item.path)}
            <Sidebar.Group>
                {const hasChildren = item?.children !== undefined && item?.children?.length > 0}
                {#if item?.children}
                    <Sidebar.GroupLabel>{item.title}</Sidebar.GroupLabel>
                {/if}
                {const parentPath = `${rootPath}/${item.path}`}

                <Sidebar.GroupContent>
                    <Sidebar.Menu>
                        {#if hasChildren}
                            {#each item.children as childItem (`${parentPath}/${childItem.path}`)}
                                {const childPath = `${parentPath}/${childItem.path}`}
                                {const isActive = $derived($page.route.id == childPath)}
                                <Sidebar.MenuItem>
                                    <Sidebar.MenuButton isActive={isActive}>
                                        {#snippet child({ props })}
                                            <a href={childPath} {...props}>
                                                <childItem.icon />
                                                <span>{childItem.title}</span>
                                            </a>
                                        {/snippet}
                                    </Sidebar.MenuButton>
                                </Sidebar.MenuItem>
                            {/each}
                        {:else}
                            <Sidebar.MenuItem>
                                <Sidebar.MenuButton>
                                    {#snippet child({ props })}
                                        <a href={item.url || `/app/${item.path}`} {...props}>
                                            <item.icon />
                                            <span>{item.title}</span>
                                        </a>
                                    {/snippet}
                                </Sidebar.MenuButton>
                            </Sidebar.MenuItem>
                        {/if}
                    </Sidebar.Menu>
                </Sidebar.GroupContent>
            </Sidebar.Group>
        {/each}
    </Sidebar.Content>
    <Sidebar.Footer class="border-sidebar-border border-t">
        <a
            href="/auth/logout"
            class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-3 rounded-md px-3 py-2 text-sm"
        >
            <LogOut class="size-4" />
            Log out
        </a>
    </Sidebar.Footer>
</Sidebar.Root>
