<script lang="ts">
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { user } from "$lib/states/index.svelte";
	import AdminSidebar from "$lib/components/admin/AdminSidebar.svelte";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	let { children } = $props();
	let authed = $state(false);

	onMount(() => {
		authed = !!user.data && (user.data.role === "admin" || user.data.role === "superadmin");
		if (!authed) {
			// Add a small delay before redirecting to allow the toast to be 
			toastState.addToast({
				message: "You do not have permission to access the admin panel.",
				type: "error"
			});
			goto("/");
		}
	});
</script>

{#if authed}
	<div class="admin-layout">
		<AdminSidebar />
		<main class="admin-content">
			{@render children()}
		</main>
	</div>
{/if}

<style lang="scss">
	.admin-layout {
		display: flex;
		width: 100%;
		height: 100%;
		overflow: hidden;
		background-color: var(--imag-100);
	}

	.admin-content {
		flex: 1;
		overflow-y: auto;
		background-color: var(--imag-bg-color);
		display: flex;
		flex-direction: column;
	}
</style>
