<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { authApi, setToken, setRefreshToken } from "$lib/services/api";
  import { consumeVerifier } from "$lib/services/oauth";

  let status = $state("Processing...");

  onMount(async () => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("code");
    const state = params.get("state");
    const provider = params.get("provider") || "google";

    if (!code || !state) {
      status = "Invalid OAuth response";
      return;
    }

    try {
      const verifier = provider === "google" ? consumeVerifier() : undefined;
      const result = await authApi.oauthCallback(provider, code, state, verifier ?? undefined);
      if (result.success && result.data) {
        setToken(result.data.access_token);
        if (result.data.refresh_token) setRefreshToken(result.data.refresh_token);
        goto("/app");
      } else {
        status = result.message || "Authentication failed";
      }
    } catch {
      status = "Authentication failed. Please try again.";
    }
  });
</script>

<main class="bg-muted/30 flex min-h-screen items-center justify-center p-4">
  <p class="text-muted-foreground">{status}</p>
</main>
