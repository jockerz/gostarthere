<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import * as Avatar from "$lib/components/ui/avatar";
    import { profileApi, getToken, type ErrorModel } from "$lib/services/api";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";
    import { onMount } from "svelte";
    import { PUBLIC_API_URL } from "$env/static/public";
    const mediaBase = PUBLIC_API_URL || "";

    let name = $state("");
    let username = $state("");
    let avatar = $state("");
    let loading = $state(false);
    let saving = $state(false);
    let uploading = $state(false);
    let fileInput: HTMLInputElement | undefined = $state();
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");

    onMount(async () => {
        if (!getToken()) return;
        loading = true;
        try {
            const result = await profileApi.get();
            if (result.success && result.data) {
                const u = result.data as
                    | { name: string; username: string; avatar?: string }
                    | undefined;
                if (u) {
                    name = u.name || "";
                    username = u.username || "";
                    avatar = u.avatar || "";
                }
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Failed to load profile";
            dialogOpen = true;
        } finally {
            loading = false;
        }
    });

    function handleAvatarClick() {
        fileInput?.click();
    }

    async function handleFileSelect(e: Event) {
        const input = e.target as HTMLInputElement;
        const file = input.files?.[0];
        if (!file) return;

        uploading = true;
        try {
            const result = await profileApi.uploadAvatar(file);
            if (result.success && result.data) {
                const data = result.data as { avatar: string };
                avatar = data.avatar;
            } else {
                dialogMessage = result.message || "Failed to upload avatar";
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Failed to upload avatar";
            dialogOpen = true;
        } finally {
            uploading = false;
            input.value = "";
        }
    }

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!name.trim()) errs.name = "Name is required";
        if (!username.trim()) errs.username = "Username is required";
        else if (!/^[a-zA-Z0-9_]+$/.test(username.trim()))
            errs.username = "Only letters, numbers, and underscores";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleSave() {
        if (!validate()) return;
        saving = true;
        try {
            const result = await profileApi.update({
                name: name.trim(),
                username: username.trim(),
                avatar: avatar || undefined,
            });
            if (result.success) {
                dialogMessage = "Profile updated successfully";
                dialogOpen = true;
            } else {
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Failed to update profile";
            dialogOpen = true;
        } finally {
            saving = false;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>Profile</Card.Title>
        <Card.Description>Manage your profile information</Card.Description>
    </Card.Header>
    <Card.Content class="space-y-6">
        {#if loading}
            <p class="text-muted-foreground text-sm">Loading...</p>
        {:else}
            <div class="flex items-center gap-4">
                <button
                    type="button"
                    class="relative rounded-full overflow-hidden focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:opacity-50"
                    disabled={uploading}
                    onclick={handleAvatarClick}
                >
                    <Avatar.Root class="size-32">
                        {#if avatar}
                            <Avatar.Image
                                src={mediaBase + "/media/" + avatar}
                                alt={name}
                            />
                        {/if}
                        <Avatar.Fallback class="text-lg">
                            {name.charAt(0)?.toUpperCase() || "U"}
                        </Avatar.Fallback>
                    </Avatar.Root>
                    {#if uploading}
                        <div
                            class="absolute inset-0 bg-black/40 flex items-center justify-center"
                        >
                            <svg
                                class="size-5 animate-spin text-white"
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                            >
                                <circle
                                    class="opacity-25"
                                    cx="12"
                                    cy="12"
                                    r="10"
                                    stroke="currentColor"
                                    stroke-width="4"
                                />
                                <path
                                    class="opacity-75"
                                    fill="currentColor"
                                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                                />
                            </svg>
                        </div>
                    {/if}
                </button>
                <div>
                    <p class="font-medium">{name || "User"}</p>
                    <p class="text-muted-foreground text-sm">@{username}</p>
                </div>

                <input
                    type="file"
                    accept="image/jpeg,image/png,image/gif,image/webp"
                    class="hidden"
                    bind:this={fileInput}
                    onchange={handleFileSelect}
                />
            </div>

            <div class="space-y-2">
                <Label for="name">Name</Label>
                <Input id="name" bind:value={name} />
                {#if fieldErrors.name}
                    <p class="text-destructive text-xs">{fieldErrors.name}</p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="username">Username</Label>
                <Input id="username" bind:value={username} />
                {#if fieldErrors.username}
                    <p class="text-destructive text-xs">
                        {fieldErrors.username}
                    </p>
                {/if}
            </div>

            <Button onclick={handleSave} disabled={saving}>
                {saving ? "Saving..." : "Save changes"}
            </Button>
        {/if}
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title={dialogMessage.includes("updated") ? "Success" : "Error"}
    message={dialogMessage}
/>
