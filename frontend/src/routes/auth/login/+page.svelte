<script lang="ts">
    import { DASHBOARD_PATH, ENABLE_GITHUB, ENABLE_GOOGLE } from "$lib/const";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import {
        authApi,
        profileApi,
        setToken,
        setRefreshToken,
        clearTokens,
        type ErrorModel,
    } from "$lib/services/api";
    import { generatePKCEChallenge, storeVerifier } from "$lib/services/oauth";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";
    import { goto } from "$app/navigation";

    let email = $state("");
    let password = $state("");
    let remember_me = $state(false);
    let loading = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogTitle = $state("Login Error");
    let dialogMessage = $state("");
    let notActivated = $state(false);
    let resending = $state(false);
    let resendSent = $state(false);

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!email.trim()) errs.email = "Email is required";
        else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()))
            errs.email = "Invalid email format";
        if (!password) errs.password = "Password is required";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleLogin() {
        if (!validate()) return;

        notActivated = false;
        loading = true;

        try {
            const result = await authApi.login(
                email.trim(),
                password,
                remember_me,
            );
            if (result.success) {
                const data = result.data as
                    | { access_token: string; refresh_token?: string }
                    | undefined;
                if (data?.access_token) {
                    setToken(data.access_token);
                    if (data.refresh_token) setRefreshToken(data.refresh_token);
                }

                const profileResult = await profileApi.get();
                const userData = profileResult.data as
                    { active?: boolean } | undefined;
                if (userData && !userData.active) {
                    clearTokens();
                    notActivated = true;
                    dialogTitle = "Account Is Not Activated"
                    dialogMessage =
                        "Please check your email for the activation link.";
                    dialogOpen = true;
                    return;
                }

                goto(DASHBOARD_PATH);
            } else {
              console.log(result)
                dialogTitle = "Login Failed"
                dialogMessage = result?.message || "Invalid credentials";
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogTitle = "Login Failed"
            dialogMessage = err.detail;
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

    async function handleResend() {
        resending = true;
        try {
            await authApi.resendActivation(email.trim());
            resendSent = true;
        } catch (error) {
            console.log(error);
            // Silently handle - don't reveal if email exists
        } finally {
            resending = false;
        }
    }

    async function handleOAuth(provider: string) {
        // Required for error dialog
        notActivated = false;

        try {
            let url: string;
            if (provider === "google") {
                const { verifier, challenge } = await generatePKCEChallenge();
                storeVerifier(verifier);
                const res = await authApi.oauthAuthorize(provider, challenge);
                url = res.url;
            } else {
                const res = await authApi.oauthAuthorize(provider);
                url = res.url;
            }
            window.location.href = url;
        } catch (e) {
            const err = e as ErrorModel;
            console.log(`handleOAuth: ${JSON.stringify(err)}`);
            dialogMessage = err.detail || `${provider} login failed`;
            dialogOpen = true;
        }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title class="text-center">Login</Card.Title>
        <Card.Description class="text-center">Welcome back</Card.Description>
    </Card.Header>
    <Card.Content class="space-y-4">
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
            <div class="flex items-center justify-between">
                <Label for="password">Password</Label>
                <a
                    href="/auth/forgot-password"
                    class="text-muted-foreground hover:text-primary text-xs"
                    >Forgot password?</a
                >
            </div>
            <Input id="password" type="password" bind:value={password} />
            {#if fieldErrors.password}
                <p class="text-destructive text-xs">{fieldErrors.password}</p>
            {/if}
        </div>

        <Button class="w-full" onclick={handleLogin} disabled={loading}>
            {loading ? "Logging in..." : "Login"}
        </Button>

        {#if ENABLE_GOOGLE || ENABLE_GITHUB}
        <div class="relative my-4">
            <div class="absolute inset-0 flex items-center">
                <span class="w-full border-t"></span>
            </div>
            <div class="relative flex justify-center text-xs uppercase">
                <span class="bg-background text-muted-foreground px-2">
                    or continue with
                </span>
            </div>
        </div>

        <div class="grid grid-cols-2 gap-2">
            {#if ENABLE_GOOGLE}
            <Button
                variant="outline"
                class="w-full text-red-600"
                onclick={() => handleOAuth("google")}
            >
                Google
            </Button>
            {/if}
            {#if ENABLE_GITHUB}
            <Button
                variant="outline"
                class="w-full"
                onclick={() => handleOAuth("github")}
            >
                GitHub
            </Button>
            {/if}
        </div>
        {/if}

        <p class="text-muted-foreground text-center text-sm">
            Don't have an account?
            <a href="/auth/register" class="hover:text-primary hover:underline"
                >Register</a
            >
        </p>
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title={dialogTitle}
    message={dialogMessage}
>
    {#snippet children()}
        {#if notActivated}
            {#if resendSent}
                <p class="text-muted-foreground text-xs">
                    A new activation link has been sent if the email exists.
                </p>
            {:else}
                <Button
                    variant="outline"
                    onclick={handleResend}
                    disabled={resending}
                >
                    {resending ? "Sending..." : "Resend activation link"}
                </Button>
            {/if}
        {/if}
    {/snippet}
</NotificationDialog>
