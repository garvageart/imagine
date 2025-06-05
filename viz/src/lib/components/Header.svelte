<script lang="ts">
	import { sidebar } from "$lib/states/index.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
</script>

<header>
	<a id="viz-title" href="/">viz</a>
	<button
		id="viz-menu-icon"
		onclick={() => {
			sidebar.open = !sidebar.open;
		}}
	>
		<MaterialIcon iconName="menu" />
	</button>
</header>

<style>
	header {
		background-color: var(--imag-blue-100);
		height: 2.5em;
		padding: 0.5em 1.5em;
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--imag-blue-20);
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 2em;
	}

	#viz-menu-icon {
		display: inline-flex;
		margin-left: auto;
		cursor: pointer;
	}
</style>
