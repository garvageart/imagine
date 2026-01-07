<script lang="ts">
	import { SplitNode, TabGroup } from "$lib/layouts/model.svelte";
	import { Splitpanes, Pane } from "$lib/third-party/svelte-splitpanes";
	import TabGroupPanel from "./TabGroupPanel.svelte";
	import { DEFAULT_THEME } from "$lib/constants";
	import LayoutNode from "./LayoutNode.svelte";
	import { debugEvent } from "$lib/utils/dom";
	import { workspaceState } from "$lib/states/workspace.svelte";

	interface Props {
		node: SplitNode | TabGroup;
	}

	let { node }: Props = $props();

	// We use a derived here for the template logic
	let maximizedId = $derived(workspaceState.workspace?.maximizedGroupId);

	// Calculate all sizes at once to ensure consistency
	let paneSizes = $derived.by(() => {
		if (!(node instanceof SplitNode)) return [];

		// If nothing maximized, return model sizes
		if (!maximizedId) {
			return node.children.map((c) => c.size);
		}

		// Check if the maximised node is within this node's subtree
		const targetInSubtree = node.containsNode(maximizedId);

		if (targetInSubtree) {
			return node.children.map((child) => {
				// If this specific child is the target or contains it, give it 100%
				if (
					child.id === maximizedId ||
					(child instanceof SplitNode && child.containsNode(maximizedId))
				) {
					return 100;
				}
				// Otherwise give siblings 0%
				return 0;
			});
		}

		// If target is elsewhere, preserve sizes
		return node.children.map((c) => c.size);
	});

	function handleResized(event: CustomEvent<any[]>) {
		// debugEvent(event);
		if (!(node instanceof SplitNode)) {
			return;
		}

		// IMPORTANT: Check the store directly.
		// If we are maximized, we must NOT update the model with the forced 100/0 sizes.
		if (workspaceState.workspace?.maximizedGroupId) {
			return;
		}

		const sizes = event.detail;

		// Safety check: if sizes array length mismatches children length,
		// it means Splitpanes has filtered out some panes (likely due to 0 size).
		// We MUST NOT update the model in this case as indices will be misaligned.
		if (sizes.length !== node.children.length) {
			return;
		}

		// Validate that the reported sizes sum to ~100%.
		// During maximization transitions or internal Splitpanes quirks, we might receive
		// invalid states (e.g. [100, 75]) which would corrupt our model if saved.
		// We use a generous tolerance (1%) to account for floating point math.
		const totalSize = sizes.reduce((acc, s) => acc + s.size, 0);
		if (Math.abs(totalSize - 100) > 1) {
			console.warn(
				"[LayoutNode] Ignoring invalid resize event (total size != 100):",
				totalSize,
				sizes
			);
			return;
		}

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
		{#each node.children as child, i (child.id)}
			<Pane
				id={child.id}
				size={paneSizes[i]}
				minSize={maximizedId ? 0 : undefined}
			>
				<LayoutNode node={child} />
			</Pane>
		{/each}
	</Splitpanes>
{:else if node instanceof TabGroup}
	<TabGroupPanel group={node} />
{/if}
