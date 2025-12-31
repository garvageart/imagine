<script lang="ts">
	import { SplitNode, TabGroup } from "$lib/layouts/model.svelte";
	import { Splitpanes, Pane } from "$lib/third-party/svelte-splitpanes";
	import TabGroupPanel from "./TabGroupPanel.svelte";
	import { DEFAULT_THEME } from "$lib/constants";
	import LayoutNode from "./LayoutNode.svelte";
	import { debugEvent } from "$lib/utils/dom";

	interface Props {
		node: SplitNode | TabGroup;
	}

	let { node }: Props = $props();

	function handleResized(event: CustomEvent<any[]>) {
		debugEvent(event);
		if (!(node instanceof SplitNode)) {
			return;
		}

		const sizes = event.detail;
		
		// Update sizes in our model
		node.children.forEach((child, i) => {
			if (sizes[i]) {
				child.size = sizes[i].size;
			}
		});
	}
</script>

{#if node instanceof SplitNode}
	<Splitpanes
		id={node.id}
		horizontal={node.orientation === "vertical"}
		theme={DEFAULT_THEME}
		on:resized={handleResized}
		style="height: 100%; width: 100%;"
	>
		{#each node.children as child (child.id)}
			<Pane id={child.id} size={child.size}>
				<LayoutNode node={child} />
			</Pane>
		{/each}
	</Splitpanes>
{:else if node instanceof TabGroup}
	<TabGroupPanel group={node} />
{/if}
