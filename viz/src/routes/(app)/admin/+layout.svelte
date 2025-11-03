<script lang="ts">
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { user } from "$lib/states/index.svelte";

	let { children } = $props();
	let authed = $state(false);

	onMount(() => {
		authed = !!user.data && (user.data.role === "admin" || user.data.role === "superadmin");
		if (!authed) {
			goto("/");
		}
	});
</script>

{#if authed}
	{@render children()}
{/if}
