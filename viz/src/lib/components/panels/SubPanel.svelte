<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";
	import { Pane } from "svelte-splitpanes";
	import MaterialIcon from "../MaterialIcon.svelte";

	let { children, name, ...props }: { name: string } & ComponentProps<typeof Pane> = $props();
</script>

<Pane {...props}>
	<div class="viz-sub_panel-header">
		<div class="viz-sub_panel-name-container">
			<span class="viz-sub_panel-name">{name}</span>
			<MaterialIcon style="margin-top: 3px;" iconName="menu" />
		</div>
	</div>
	{@render children()}
</Pane>

<style>
	.viz-sub_panel-header {
		height: 1.5em;
		border-bottom: 1px solid var(--imag-blue-20);
		background-color: var(--imag-blue-80);
		margin-bottom: 0.5em;
		font-size: 0.9em;
		padding: 0.3em 0.7em;
	}

    .viz-sub_panel-name-container {
        display: flex;
        align-items: center;
        position: relative;
    }
</style>
