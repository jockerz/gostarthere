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
    let loading = $state(false);
    let success = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");

    onMount(() => {
        const t = $page.url.searchParams.get("token");
        if (t) {
            token = t;
            handleActivate();
        }
    });

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!token.trim()) errs.token = "Activation token is required";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleActivate() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.activate(token.trim());
            if (result.success) {
                success = true;
            } else {
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Activation failed";
            dialogOpen = true;
        } finally {
            loading = false;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title class="text-center">Activate account</Card.Title>
        <Card.Description class="text-center"
            >Enter your activation token</Card.Description
        >
    </Card.Header>
    <Card.Content class="space-y-4">
        {#if success}
            <p class="text-center text-sm text-green-600">
                Account activated successfully!
            </p>
            <Button class="w-full" onclick={() => goto("/auth/login")}
                >Go to login</Button
            >
        {:else}
            <div class="space-y-2">
                <Label for="token">Activation token</Label>
                <Input
                    id="token"
                    bind:value={token}
                    placeholder="Paste your activation token"
                />
                {#if fieldErrors.token}
                    <p class="text-destructive text-xs">{fieldErrors.token}</p>
                {/if}
            </div>

            <Button
                class="w-full"
                onclick={handleActivate}
                disabled={loading || !token.trim()}
            >
                {loading ? "Activating..." : "Activate"}
            </Button>

            <p class="text-muted-foreground text-center text-sm">
                Need a new token?
                <a
                    href="/auth/resend-activation"
                    class="hover:text-primary hover:underline"
                    >Request new one</a
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
