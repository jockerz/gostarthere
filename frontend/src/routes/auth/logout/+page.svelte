<script lang="ts">
    import { onMount } from "svelte";
    import { authApi, getToken, clearTokens, type ErrorModel } from "$lib/services/api";
    import { goto } from "$app/navigation";

    let error = $state("");

    onMount(async () => {
        const token = getToken();
        if (token) {
            try {
                await authApi.logout(token);
            } catch (e) {
                const err = e as ErrorModel;
                error = err.detail || "Logout failed";
            }
        }
        clearTokens();
        goto("/");
    });
</script>

{#if error}
    <p class="text-destructive text-center text-sm">{error}</p>
{/if}
