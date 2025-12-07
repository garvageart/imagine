<script lang="ts">
	interface Props {
		heading: string;
		description: string;
		actions?: import("svelte").Snippet;
		children: import("svelte").Snippet;
	}

	let { heading, description, actions, children }: Props = $props();
</script>

<div class="admin-page-route">
	<header class="page-header">
		<div class="route-info">
			<h1>{heading}</h1>
			<p class="description">{description}</p>
		</div>
		{#if actions}
			<div class="page-actions">
				{@render actions()}
			</div>
		{/if}
	</header>

	<div class="page-content">
		{@render children()}
	</div>
</div>

<style lang="scss">
	.admin-page-route {
		display: flex;
		flex-direction: column;
		width: 100%;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 2rem;
		border-bottom: 1px solid var(--imag-80);
		background-color: var(--imag-bg-color);
		position: sticky;
		top: 0;
		z-index: 10;

		h1 {
			margin: 0;
			font-size: 1.75rem;
			font-weight: 600;
		}

		.description {
			margin: 0.25rem 0 0 0;
			color: var(--imag-40);
			font-size: 0.95rem;
		}
	}

	.page-actions {
		display: flex;
		gap: 0.75rem;
	}

	.page-content {
		padding: 2rem;
	}
</style>
