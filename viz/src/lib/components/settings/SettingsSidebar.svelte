<script lang="ts">
	import Sidebar from "../Sidebar.svelte";

	interface Props {
		groups?: string[];
		activeGroup?: string;
	}

	let { groups = [], activeGroup = "" }: Props = $props();

	// Format group name for display (PascalCase)
	function formatGroupName(group: string): string {
		return group.charAt(0).toUpperCase() + group.slice(1);
	}
</script>

<Sidebar>
	<div class="settings-nav-content">
		<div class="sidebar-header">
			<h3>Settings</h3>
		</div>
		<ul class="nav-list">
			{#each groups as group}
				<li>
					<a
						href="/settings/{group.toLowerCase()}"
						class="nav-link"
						class:active={activeGroup.toLowerCase() === group.toLowerCase()}
					>
						{formatGroupName(group)}
					</a>
				</li>
			{/each}
		</ul>
	</div>
</Sidebar>

<style lang="scss">
	.settings-nav-content {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding-top: 1rem;
	}

	.sidebar-header {
		padding: 1.5rem 1.5rem 1rem;

		h3 {
			font-size: 1rem;
			color: var(--imag-10);
			font-weight: 600;
			margin: 0;
		}
	}

	.nav-list {
		list-style: none;
		padding: 0 0.5rem;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.nav-link {
		display: block;
		padding: 0.5rem 1rem;
		color: var(--imag-40);
		text-decoration: none;
		border-radius: 0.375rem;
		font-size: 1rem;
		transition: all 0.2s;

		&:hover {
			background-color: var(--imag-70);
			color: var(--imag-text-color);
		}

		&.active {
			background-color: var(--imag-80);
			color: var(--imag-10);
		}
	}
</style>
