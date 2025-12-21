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
		subPanelContentElement?: HTMLDivElement;
	}

	let {
		keyId,
		panelViews = $bindable(),
		activeView = $bindable(),
		panelData = $bindable(),
		componentToRender,
		subPanelContentFocused = $bindable(),
		tabDropper,
		onFocus,
		subPanelContentElement = $bindable()
	}: Props = $props();

	let Comp = $derived(componentToRender ?? activeView?.component);

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

	function subPanelDrop(node: HTMLElement, data: any) {
		return tabDropper.subPanelDropInside(node, data);
	}
</script>

<div
	role="none"
	class="viz-content-panel"
	style="width: 100%;"
	onclick={onFocus}
	onkeydown={onFocus}
	bind:this={subPanelContentElement}
	use:subPanelDrop={{
		index: panelViews?.findIndex((view) => view.id === activeView?.id) ?? -1,
		view: activeView
	}}
>
	{#if subPanelContentFocused}
		<div class="viz-panel-active-overlay"></div>
	{/if}
	{#await panelData}
		<LoadingContainer />
	{:then loadedData}
		{#if Comp}
			{#if loadedData}
				<Comp data={loadedData.data} view={activeView} />
			{:else}
				<Comp view={activeView} />
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

	.viz-content-panel {
		text-overflow: clip;
		position: relative;
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
	}

	.viz-panel-active-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 900;
		box-shadow:
			0 1.5px 0 var(--imag-primary) inset,
			1.5px 0 0 var(--imag-primary) inset,
			-1.5px 0 0 var(--imag-primary) inset,
			0 -1.5px 0 var(--imag-primary) inset;
	}

	:global(
			.splitpanes__pane > *:last-child,
			.viz-sub_panel-content,
			.viz-content-panel
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
					.viz-content-panel,
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
