<script lang="ts">
    import { onMount } from "svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import {
        authApi,
        profileApi,
        type ErrorModel,
        type UserData,
        type SuccessBody,
    } from "$lib/services/api";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";

    let currentPassword = $state("");
    let curEmail = $state("");
    let newEmail = $state("");
    let loading = $state(false);
    let sent = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");
    let dialogTitle = $state("Error");

    onMount(async () => {
        profileApi.get().then(
            (d: SuccessBody) => (curEmail = (d.data as UserData)?.email),
            (e) => console.log(e),
        );
    });

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!currentPassword)
            errs.currentPassword = "Current password is required";
        if (!newEmail.trim()) errs.newEmail = "New email is required";
        else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(newEmail.trim()))
            errs.newEmail = "Invalid email format";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleChange() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.changeEmail(
                currentPassword,
                newEmail.trim(),
            );
            if (result.success) {
                sent = true;
            } else {
                dialogTitle = "Error";
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogTitle = "Error";
            dialogMessage = err.detail || "Failed to update email";
            if (err.errors) {
                for (const fe of err.errors) {
                    const key = fe.location?.replace("body.", "") || "newEmail";
                    fieldErrors = { ...fieldErrors, [key]: fe.message };
                }
            }
            dialogOpen = true;
        } finally {
            loading = false;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>Update Email</Card.Title>
        <Card.Description>Change your email address.</Card.Description>
        <Card.Description>
            <strong>An update confirmation link</strong> will be sent to your email.
        </Card.Description>
    </Card.Header>
    <Card.Content class="space-y-4">
        {#if sent}
            <p class="text-center text-sm text-green-600">
                A confirmation link has been sent to your new email. Please
                check your inbox to confirm the change.
            </p>
            <Button class="w-full" variant="outline" href="/dashboard/profile"
                >Back to profile</Button
            >
        {:else}
            <div class="space-y-2">
                <Label for="curEmail">Current email</Label>
                <Input
                    id="curEmail"
                    class="text-muted-foreground"
                    type="email"
                    bind:value={curEmail}
                    readonly
                />
            </div>

            <div class="space-y-2">
                <Label for="newEmail">New email</Label>
                <Input
                    id="newEmail"
                    type="email"
                    bind:value={newEmail}
                    placeholder="new@example.com"
                />
                {#if fieldErrors.newEmail}
                    <p class="text-destructive text-xs">
                        {fieldErrors.newEmail}
                    </p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="currentPassword">Password</Label>
                <Input
                    id="currentPassword"
                    type="password"
                    bind:value={currentPassword}
                />
                {#if fieldErrors.currentPassword}
                    <p class="text-destructive text-xs">
                        {fieldErrors.currentPassword}
                    </p>
                {/if}
            </div>

            <Button
                onclick={handleChange}
                disabled={loading || !currentPassword || !newEmail.trim()}
            >
                {loading ? "Sending..." : "Send confirmation"}
            </Button>
        {/if}
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title={dialogTitle}
    message={dialogMessage}
/>
