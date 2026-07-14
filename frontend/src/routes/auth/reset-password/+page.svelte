<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import { authApi, type ErrorModel } from "$lib/services/api";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";

    let token = $state("");
    let password = $state("");
    let validate_password = $state("");
    let loading = $state(false);
    let success = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");

    onMount(() => {
        const t = $page.url.searchParams.get("token");
        if (t) token = t;
    });

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!token.trim()) errs.token = "Reset token is required";
        if (!password) errs.password = "Password is required";
        else if (password.length < 8)
            errs.password = "Password must be at least 8 characters";
        fieldErrors = errs;
        if (password !== validate_password)
            errs.validate_password = "Password not match";
        return Object.keys(errs).length === 0;
    }

    async function handleReset() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.resetPassword(token.trim(), password);
            if (result.success) {
                success = true;
            } else {
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Reset failed";
            dialogOpen = true;
        } finally {
            loading = false;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title class="text-center">Reset password</Card.Title>
        <Card.Description class="text-center"
            >Enter your new password</Card.Description
        >
    </Card.Header>
    <Card.Content class="space-y-4">
        {#if success}
            <p class="text-center text-sm text-green-600">
                Password reset successfully!
            </p>
            <Button class="w-full" onclick={() => goto("/auth/login")}
                >Go to login</Button
            >
        {:else}
            <div class="space-y-2" hidden>
                <Label for="token">Reset token</Label>
                <Input
                    id="token"
                    bind:value={token}
                    placeholder="Paste your reset token"
                />
                {#if fieldErrors.token}
                    <p class="text-destructive text-xs">{fieldErrors.token}</p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="password">New password</Label>
                <Input id="password" type="password" bind:value={password} />
                {#if fieldErrors.password}
                    <p class="text-destructive text-xs">
                        {fieldErrors.password}
                    </p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="validate_password">Validate password</Label>
                <Input
                    id="validate_password"
                    type="password"
                    bind:value={validate_password}
                />
                {#if fieldErrors.validate_password}
                    <p class="text-destructive text-xs">
                        {fieldErrors.validate_password}
                    </p>
                {/if}
            </div>

            <Button
                class="w-full"
                onclick={handleReset}
                disabled={loading || !token.trim() || !password}
            >
                {loading ? "Resetting..." : "Reset password"}
            </Button>
        {/if}
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title="Password Reset Failed"
    message={dialogMessage}
/>
