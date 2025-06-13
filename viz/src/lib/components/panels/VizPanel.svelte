<script>
	import { Splitpanes as Panel } from "svelte-splitpanes";
	import SubPanel from "./SubPanel.svelte";

	const theme = "viz-theme";
</script>

<Panel {theme} style="height: 100%" pushOtherPanes={false}>
	<SubPanel name="Metadata" class="viz-panel" minSize={10}>
		<span class="specs">1</span>
	</SubPanel>
	<SubPanel name="Preview" class="viz-panel" minSize={15}>
		<span class="specs">2</span>
	</SubPanel>
	<SubPanel name="Export" class="viz-panel" minSize={20}>
		<span class="specs">3</span>
	</SubPanel>
</Panel>

<style>
	/* :global(.viz-panel) {
		padding: 0em 1em;
	} */

	:global(.specs) {
		margin-left: 1em;
	}
</style>
