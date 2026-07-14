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
  } from "$lib/services/api";

  let user = $state<{ has_password?: boolean } | null>(null);
  let loading = $state(true);

  let password = $state("");
  let passwordConfirm = $state("");
  let savingPassword = $state(false);
  let passwordError = $state("");
  let passwordSuccess = $state("");

  onMount(async () => {
    try {
      const profileRes = await profileApi.get();
      user = profileRes.data as { has_password?: boolean } | null;
    } catch {
      // ignore
    }
    loading = false;
  });

  async function handleLinkProvider(provider: string) {
    const { generatePKCEChallenge, storeVerifier } = await import("$lib/services/oauth");
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
  }

  async function handleSetPassword() {
    passwordError = "";
    passwordSuccess = "";
    if (password !== passwordConfirm) {
      passwordError = "Passwords do not match";
      return;
    }
    if (password.length < 8) {
      passwordError = "Password must be at least 8 characters";
      return;
    }
    savingPassword = true;
    try {
      await authApi.setPassword(password);
      password = "";
      passwordConfirm = "";
      passwordSuccess = "Password set successfully";
      if (user) user.has_password = true;
    } catch (e) {
      passwordError = (e as ErrorModel).detail || "Failed to set password";
    }
    savingPassword = false;
  }
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>Settings</Card.Title>
    <Card.Description>Manage your account settings</Card.Description>
  </Card.Header>
  <Card.Content class="space-y-6">
    {#if !loading && user && !user.has_password}
      <div class="space-y-4">
        <h3 class="font-medium">Set Password</h3>
        <p class="text-muted-foreground text-sm">
          Set a password to enable email+password login.
        </p>
        <div class="space-y-2">
          <Label for="password">New Password</Label>
          <Input id="password" type="password" bind:value={password} />
        </div>
        <div class="space-y-2">
          <Label for="password-confirm">Confirm Password</Label>
          <Input id="password-confirm" type="password" bind:value={passwordConfirm} />
        </div>
        {#if passwordError}
          <p class="text-destructive text-xs">{passwordError}</p>
        {/if}
        {#if passwordSuccess}
          <p class="text-green-600 text-xs">{passwordSuccess}</p>
        {/if}
        <Button onclick={handleSetPassword} disabled={savingPassword}>
          {savingPassword ? "Saving..." : "Set Password"}
        </Button>
      </div>
      <hr class="border-border" />
    {/if}

    <div class="space-y-4">
      <h3 class="font-medium">Linked Accounts</h3>
      <p class="text-muted-foreground text-sm">
        Link your Google or GitHub account for easy login.
      </p>
      <div class="flex gap-2">
        <Button variant="outline" onclick={() => handleLinkProvider("google")}>
          Link Google
        </Button>
        <Button variant="outline" onclick={() => handleLinkProvider("github")}>
          Link GitHub
        </Button>
      </div>
    </div>
  </Card.Content>
</Card.Root>
