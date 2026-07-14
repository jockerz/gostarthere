<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import {
        authApi,
        setToken,
        setRefreshToken,
        type ErrorModel,
    } from "$lib/services/api";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";
    import { goto } from "$app/navigation";

    let name = $state("");
    let username = $state("");
    let email = $state("");
    let password = $state("");
    let passwordValidate = $state("");
    let loading = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");
    let dialogTitle = $state("Registration Error");

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!name.trim()) errs.name = "Name is required";
        if (!username.trim()) errs.username = "Username is required";
        else if (!/^[a-zA-Z0-9_]+$/.test(username.trim()))
            errs.username = "Only letters, numbers, and underscores";
        if (!email.trim()) errs.email = "Email is required";
        else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()))
            errs.email = "Invalid email format";
        if (!password) errs.password = "Password is required";
        else if (password.length < 8)
            errs.password = "Password must be at least 8 characters";
        if (!passwordValidate)
            errs.passwordValidate = "Please confirm your password";
        else if (password !== passwordValidate)
            errs.passwordValidate = "Passwords do not match";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleRegister() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.register(
                name.trim(),
                username.trim(),
                email.trim(),
                password,
                passwordValidate,
            );
            if (result.success) {
                const data = result.data as
                    | { access_token: string; refresh_token?: string }
                    | undefined;
                if (data?.access_token) {
                    setToken(data.access_token);
                    if (data.refresh_token) setRefreshToken(data.refresh_token);
                }
                dialogTitle = "Registration complete";
                dialogMessage =
                    "Registration complete. Account activation link is being sent to your email";
                dialogOpen = true;

                await new Promise((resolve) => setTimeout(resolve, 3000));
                goto("/dashboard");
            } else {
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogMessage = err.detail || "Registration failed";
            if (err.errors) {
                for (const fe of err.errors) {
                    const key = fe.location?.replace("body.", "") || "email";
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
        <Card.Title class="text-center">Create an account</Card.Title>
        <Card.Description class="text-center">
            Enter your information below to create your account
        </Card.Description>
    </Card.Header>
    <Card.Content class="space-y-4">
        <div class="space-y-2">
            <Label for="name">Name</Label>
            <Input id="name" bind:value={name} placeholder="Your full name" />
            {#if fieldErrors.name}
                <p class="text-destructive text-xs">{fieldErrors.name}</p>
            {/if}
        </div>

        <div class="space-y-2">
            <Label for="username">Username</Label>
            <Input
                id="username"
                bind:value={username}
                placeholder="your_username"
            />
            {#if fieldErrors.username}
                <p class="text-destructive text-xs">{fieldErrors.username}</p>
            {/if}
        </div>

        <div class="space-y-2">
            <Label for="email">Email</Label>
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

        <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input id="password" type="password" bind:value={password} />
            {#if fieldErrors.password}
                <p class="text-destructive text-xs">{fieldErrors.password}</p>
            {/if}
        </div>

        <div class="space-y-2">
            <Label for="password_validate">Confirm password</Label>
            <Input
                id="password_validate"
                type="password"
                bind:value={passwordValidate}
            />
            {#if fieldErrors.passwordValidate}
                <p class="text-destructive text-xs">
                    {fieldErrors.passwordValidate}
                </p>
            {/if}
        </div>

        <Button class="w-full" onclick={handleRegister} disabled={loading}>
            {loading ? "Creating account..." : "Create account"}
        </Button>

        <p class="text-muted-foreground text-center text-sm">
            Already have an account?
            <a href="/auth/login" class="hover:text-primary hover:underline"
                >Log in</a
            >
        </p>
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title={dialogTitle}
    message={dialogMessage}
/>
