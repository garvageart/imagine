<script lang="ts">
	import { type Component, type Snippet } from "svelte";
	import type VizView from "$lib/views/views.svelte";
	import LoadingContainer from "../LoadingContainer.svelte";
	import type TabOps from "$lib/views/tabs.svelte.ts";
	import { isElementScrollable } from "$lib/utils/dom";
	import { debugMode } from "$lib/states/index.svelte";

	interface Props {
		keyId: string;
		panelViews: VizView[];
		activeView: VizView;
		subPanelContentFocused: boolean;
		tabDropper: TabOps;
		onFocus: () => void;
		componentToRender?: Component<any>;
		panelData:
			| {
					type: "loaded";
					status: number;
					data: any;
			  }
			| Promise<void | {
					type: "loaded";
					status: number;
					data: any;
			  }>;
	}

	let {
		keyId,
		panelViews = $bindable(),
		activeView = $bindable(),
		panelData = $bindable(),
		componentToRender,
		subPanelContentFocused = $bindable(),
		tabDropper,
		onFocus
	}: Props = $props();

	let Comp = $derived(componentToRender ?? activeView?.component);
	let subPanelContentElement: HTMLDivElement | undefined = $state();

	$effect(() => {
		if (activeView) {
			panelData = activeView.viewData ?? activeView.getComponentData();
		}
	});

	if (debugMode === true) {
		$inspect("active view", keyId, activeView);
		$effect(() => {
			(async () => {
				const data = await panelData;
				console.log("panel data", keyId, $state.snapshot(data));
			})();
		});
		$inspect("panel views", keyId, panelViews);
	}

	$effect(() => {
		if (panelViews.length) {
			const element = subPanelContentElement;
			if (!element) {
				return;
			}

			const lastChild = element.lastElementChild as HTMLElement;
			if (!lastChild) {
				return;
			}

			if (subPanelContentFocused) {
				if (isElementScrollable(lastChild)) {
					element.classList.add("with__scrollbar");
				}
				element.classList.add("splitpanes__pane__active");
			} else {
				element.classList.remove("with__scrollbar");
				element.classList.remove("splitpanes__pane__active");
			}
		}
	});

	function subPanelDrop(node: HTMLElement, data: any) {
		return tabDropper.subPanelDropInside(node, data);
	}
</script>

<div
	role="none"
	class="viz-sub_panel-content"
	style="width: 100%;"
	onclick={onFocus}
	onkeydown={onFocus}
	bind:this={subPanelContentElement}
	use:subPanelDrop={{
		index: panelViews?.findIndex((view) => view.id === activeView?.id) ?? -1,
		view: activeView
	}}
>
	{#await panelData}
		<LoadingContainer />
	{:then loadedData}
		{#if Comp}
			{#if loadedData}
				<Comp data={loadedData.data} />
			{:else}
				<Comp />
			{/if}
		{/if}
	{:catch error}
		<h2>Something has gone wrong:</h2>
		<p style="color: red;">{error}</p>
	{/await}
</div>

<style lang="scss">
	/* Dim and disable adjacent split resizers for locked panes (UI-only enforcement).
   The svelte-splitpanes implementation uses sibling resizer elements; target the
   resizer that is immediately before or after a locked pane. */
	:global(.splitpanes__pane.locked) + :global(.splitpanes__resizer),
	:global(.splitpanes__resizer) + :global(.splitpanes__pane.locked) {
		pointer-events: none;
		opacity: 0.45;
		cursor: default;
	}

	.viz-sub_panel-content {
		text-overflow: clip;
		position: relative;
		display: flex;
		flex-direction: column;
		height: 100%;
		max-height: 100%;
	}

	:global(
			.splitpanes__pane > *:last-child,
			.viz-sub_panel-content
				> :last-child:not(
					.splitpanes--horizontal,
					.splitpanes--vertical,
					.viz-view-container
				)
		) {
		padding: 0.5em;
	}

	:global(
			.splitpanes__pane
				> :is(
					.splitpanes,
					.splitpanes__pane,
					.viz-sub_panel-content,
					.splitpanes--horizontal,
					.splitpanes--vertical
				)
		) {
		padding: 0em;
	}

	:global(.drop-hover-above) {
		outline: 1.5px solid var(--imag-primary);
	}
</style>
