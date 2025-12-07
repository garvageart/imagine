<script lang="ts">
	import Sidebar from "$lib/components/Sidebar.svelte";
	import { page } from "$app/state";

	const items = [
		{ label: "Dashboard", href: "/admin", exact: true },
		{ label: "Users", href: "/admin/users" },
		{ label: "Jobs", href: "/admin/jobs" },
		{ label: "Events", href: "/admin/events" },
		{ label: "Cache", href: "/admin/cache" }
	];

	function isActive(href: string, exact = false) {
		if (exact) {
			return page.url.pathname === href;
		}

		return page.url.pathname.startsWith(href);
	}
</script>

<Sidebar>
	<nav class="admin-nav" data-sveltekit-preload-data="hover" data-sveltekit-preload-code="hover">
		<ul>
			{#each items as item}
				<li>
					<a href={item.href} class="nav-link" class:active={isActive(item.href, item.exact)}>
						{item.label}
					</a>
				</li>
			{/each}
		</ul>
	</nav>
</Sidebar>

<style lang="scss">
	.admin-nav {
		padding: 1rem 0.5rem;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.nav-link {
		display: block;
		padding: 0.5rem 0.75rem;
		color: var(--imag-text-color);
		text-decoration: none;
		border-radius: 0.375rem;
		font-size: 0.95rem;
		transition: background-color 0.2s;

		&:hover {
			background-color: var(--imag-80);
		}

		&.active {
			background-color: var(--imag-90);
			font-weight: 500;
			color: var(--imag-10);
		}
	}
</style>
