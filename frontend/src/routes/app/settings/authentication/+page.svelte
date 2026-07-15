<script lang="ts">
    import { onMount } from "svelte";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import * as Card from "$lib/components/ui/card";
    import * as Separator from "$lib/components/ui/separator";
    import * as Item from "$lib/components/ui/item";
    import Link from "@lucide/svelte/icons/link";
    import Unlink from "@lucide/svelte/icons/unlink";
    import NotificationDialog from "$lib/components/custom/NotificationDialog.svelte";

    import { authApi, type ErrorModel } from "$lib/services/api";
    import { type UserAuthProvider } from "./types";

    let authProviderData = $state<{has_password: boolean; data: UserAuthProvider[]} | null>(null);

    let has_password = $state(true);
    let authDataGithub = $state<UserAuthProvider | null>(null);
    let authDataGoogle = $state<UserAuthProvider | null>(null);
    let currentPassword = $state("");
    let newPassword = $state("");
    let validatePassword = $state("");
    let loading = $state(false);
    let fieldErrors = $state<Record<string, string>>({});
    let dialogOpen = $state(false);
    let dialogMessage = $state("");
    let dialogTitle = $state("Error");

    onMount(async () => {
      try {
        const resp = await authApi.getOAuthProviderData()
        authProviderData = resp.data as {
          has_password: boolean; data: UserAuthProvider[]
        } | null;
        if (authProviderData !== null) {
          has_password = authProviderData.has_password;
          for (const data of authProviderData.data) {
            if (data.provider === "github") {
              authDataGithub = data
            } else if (data.provider === "google") {
              authDataGoogle = data
            }
          }
        }
      } catch(e) {
        const err = e as ErrorModel;
        console.error(err)
      }
    })

    function validate(): boolean {
        const errs: Record<string, string> = {};
        if (!currentPassword)
            errs.currentPassword = "Current password is required";
        if (!newPassword) errs.newPassword = "New password is required";
        else if (newPassword.length < 8)
            errs.newPassword = "Password must be at least 8 characters";
        if (newPassword !== validatePassword)
            errs.validatePassword = "Password do not match";
        fieldErrors = errs;
        return Object.keys(errs).length === 0;
    }

    async function handleChange() {
        if (!validate()) return;
        loading = true;
        try {
            const result = await authApi.changePassword(
                currentPassword,
                newPassword,
            );
            if (result.success) {
                dialogTitle = "Success";
                dialogMessage = "Password changed successfully";
                dialogOpen = true;
                currentPassword = "";
                newPassword = "";
            } else {
                dialogTitle = "Error";
                dialogMessage = result.message;
                dialogOpen = true;
            }
        } catch (e) {
            const err = e as ErrorModel;
            dialogTitle = "Error";
            dialogMessage = err.detail || "Failed to change password";
            dialogOpen = true;
        } finally {
            loading = false;
        }
    }

    async function handleLinkProvider(provider: string) {
      const { generatePKCEChallenge, storeVerifier } = await import("$lib/services/oauth");
      let url: string;

      try {
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
      } catch(e) {
        const error = e as {title: string; detail: string;}
        dialogTitle = "Authentication Failed"
        dialogMessage = error.detail;
        dialogOpen = true;
      }
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>Authentication</Card.Title>
        <Card.Description>Configure password and third party authentication.</Card.Description>
    </Card.Header>

    <Card.Content class="space-y-4">
        <h3 class="font-medium">Password</h3>
        {#if has_password}
        <div class="space-y-2">
            <Label for="currentPassword">Current password</Label>
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
        {:else}
        <div class="space-y-2">
            <Label for="currentPassword">Password is never set</Label>
        </div>
        {/if}

        <Separator.Root />

        <div class="space-y-2">
            <Label for="newPassword">New password</Label>
            <Input id="newPassword" type="password" bind:value={newPassword} />
            {#if fieldErrors.newPassword}
                <p class="text-destructive text-xs">
                    {fieldErrors.newPassword}
                </p>
            {/if}
        </div>

        <div class="space-y-2">
            <Label for="validatePassword">Validate password</Label>
            <Input
                id="validatePassword"
                type="password"
                bind:value={validatePassword}
            />
            {#if fieldErrors.validatePassword}
                <p class="text-destructive text-xs">
                    {fieldErrors.validatePassword}
                </p>
            {/if}
        </div>

        <Button
            onclick={handleChange}
            disabled={loading || !currentPassword || !newPassword}
        >
            {loading ? "Changing..." : "Change password"}
        </Button>

        <Separator.Root />

        <div class="space-y-4">
          <h3 class="font-medium">Linked Accounts</h3>
          <p class="text-muted-foreground text-sm">
            Link your Google or GitHub account for easy login.
          </p>
          <Item.Root variant="outline">
            {#if authDataGithub == null}
                <Item.Media>
                    <Unlink class="size-5" />
                </Item.Media>
                <Item.Content>
                  <Item.Title>Github</Item.Title>
                  <Item.Description>Not linked</Item.Description>
                </Item.Content>
                <Item.Actions>
                    <Button variant="outline" size="sm" onclick={() => handleLinkProvider("github")}>Link</Button>
                </Item.Actions>
            {:else}
                <Item.Media>
                    <Link class="size-5" />
                </Item.Media>
                <Item.Content>
                    <Item.Title>Github</Item.Title>
                    <Item.Description>Linked</Item.Description>
                </Item.Content>
            {/if}
            </Item.Root>
            <Item.Root variant="outline">
            {#if authDataGoogle == null}
                <Item.Media>
                    <Unlink class="size-5" />
                </Item.Media>
                <Item.Content>
                  <Item.Title>Google</Item.Title>
                  <Item.Description>Not linked</Item.Description>
                </Item.Content>
                <Item.Actions>
                    <Button variant="outline" size="sm" onclick={() => handleLinkProvider("google")}>Link</Button>
                </Item.Actions>
            {:else}
                <Item.Media>
                    <Link class="size-5" />
                </Item.Media>
                <Item.Content>
                    <Item.Title>Google</Item.Title>
                    <Item.Description>Linked</Item.Description>
                </Item.Content>
            {/if}
          </Item.Root>
        </div>
    </Card.Content>
</Card.Root>

<NotificationDialog
    bind:open={dialogOpen}
    title={dialogTitle}
    message={dialogMessage}
/>
