<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import { authApi, type ErrorModel } from "$lib/services/api";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";
    import { goto } from "$app/navigation";

    let email = $state("");
    let loading = $state(false);
    let sent = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!email.trim()) errs.email = "Email/username is required";
        // else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) errs.email = "Invalid email format";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleSubmit() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.forgotPassword(email.trim());
            if (result.success) {
                sent = true;
            } else {
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Request failed";
            dialogOpen = true;
        } finally {
            loading = false;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title class="text-center">Forgot password</Card.Title>
        <Card.Description class="text-center"
            >We'll send you a reset link to your email</Card.Description
        >
    </Card.Header>
    <Card.Content class="space-y-4">
        {#if sent}
            <p class="text-center text-sm text-green-600">
                If an account with that email/username exists, a reset link has
                been sent.
            </p>
            <Button
                class="w-full"
                variant="outline"
                onclick={() => goto("/auth/login")}>Back to login</Button
            >
        {:else}
            <div class="space-y-2">
                <Label for="email">Email or Username</Label>
                <Input
                    id="email"
                    type="email"
                    bind:value={email}
                    placeholder="you@example.com"
                />
                {#if fieldErrors.email}
                    <p class="text-destructive text-xs">{fieldErrors.email}</p>
                {/if}
            </div>

            <Button
                class="w-full"
                onclick={handleSubmit}
                disabled={loading || !email.trim()}
            >
                {loading ? "Sending..." : "Send reset link"}
            </Button>

            <p class="text-muted-foreground text-center text-sm">
                Remember your password?
                <a href="/auth/login" class="hover:text-primary hover:underline"
                    >Log in</a
                >
            </p>
        {/if}
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title="Error"
    message={dialogMessage}
/>
