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
			handleConfirm();
		}
	});

	function validate(): boolean {
		const errs: Record<string, string> = {};
		if (!token.trim()) errs.token = "Confirmation token is required";
		fieldErrors = errs;
		return Object.keys(errs).length === 0;
	}

	async function handleConfirm() {
		if (!validate()) return;
		loading = true;
		try {
			const result = await authApi.confirmEmail(token.trim());
			if (result.success) {
				success = true;
			} else {
				dialogMessage = result.message;
				dialogOpen = true;
			}
		} catch (e) {
			const err = e as ErrorModel;
			dialogMessage = err.detail || "Email confirmation failed";
			dialogOpen = true;
		} finally {
			loading = false;
		}
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Confirm email</Card.Title>
		<Card.Description>Confirm your new email address</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		{#if success}
			<p class="text-center text-sm text-green-600">Email changed successfully!</p>
			<Button class="w-full" onclick={() => goto("/auth/login")}>Go to login</Button>
		{:else}
			<div class="space-y-2">
				<Label for="token">Confirmation token</Label>
				<Input id="token" bind:value={token} placeholder="Paste your confirmation token" />
				{#if fieldErrors.token}
					<p class="text-destructive text-xs">{fieldErrors.token}</p>
				{/if}
			</div>

			<Button class="w-full" onclick={handleConfirm} disabled={loading || !token.trim()}>
				{loading ? "Confirming..." : "Confirm email"}
			</Button>

			<p class="text-muted-foreground text-center text-sm">
				<a href="/auth/login" class="hover:text-primary hover:underline">Back to login</a>
			</p>
		{/if}
	</Card.Content>
</Card.Root>

<NotificationDialog bind:open={dialogOpen} title="Confirmation Error" message={dialogMessage} />
