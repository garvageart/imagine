<script lang="ts">
	import { DEFAULT_THEME } from "$lib/constants";
	import { panels } from "$lib/layouts/test";
	import { Splitpanes as Panel, type ITree } from "$lib/third-party/svelte-splitpanes";
	import { layoutState, layoutTree } from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { arrayHasDuplicates, debugEvent, generateKeyId, VizLocalStorage } from "$lib/utils";
	import { onMount } from "svelte";
	import SubPanel, { type VizSubPanel } from "./SubPanel.svelte";

	let { id }: { id: string } = $props();
	const theme = DEFAULT_THEME;
	const saveLayout = new VizLocalStorage<VizSubPanel[]>("layout");
	const treeLayout = new VizLocalStorage<ITree>("tree");
	let storedLayout = saveLayout.get();

	function resetLayoutToDefault() {
		saveLayout.set(panels);
		storedLayout = panels;
	}

	if (storedLayout && storedLayout?.length === 0) {
		console.warn("No layout found in localStorage, using default layout");
		resetLayoutToDefault();
	}

	const duplicateAnswer = arrayHasDuplicates(
		panels
			.flatMap((panel) => panel.views.map((tab) => tab.id))
			.concat(
				panels.flatMap((panel) =>
					panel.childs?.subPanel ? panel.childs.subPanel.flatMap((subPanel) => subPanel.views.map((tab) => tab.id)) : []
				)
			)
	);

	if (duplicateAnswer.hasDuplicates) {
		console.error("The following tabs have duplicate IDs. Please check the panels loaded", duplicateAnswer.duplicates);
	}

	onMount(() => {
		layoutState.tree ??= [];
		layoutState.tree = storedLayout ?? panels;
		layoutTree.childs = layoutState.tree;
		const storedTree = treeLayout.get();

		if (!storedTree) {
			return;
		}

		layoutTree.class = storedTree.class;
		layoutTree.style = storedTree.style;
		layoutTree.theme = storedTree.theme;
		layoutTree.rtl = storedTree.rtl;
		layoutTree.keyId = storedTree.keyId;
		layoutTree.id = storedTree.id;
		layoutTree.dblClickSplitter = storedTree.dblClickSplitter;
		layoutTree.pushOtherPanes = storedTree.pushOtherPanes;
		layoutTree.horizontal = storedTree.horizontal;
		layoutTree.firstSplitter = storedTree.firstSplitter;
	});
	// This derived value was initially used to do
	// further layout calculations like checking if a single pane
	// needs to be used but that seems to have just fixed itself?
	// So for now all it does is save the layout every time it is adjusted
	const internalLayoutState = $derived.by(() => {
		return layoutState.tree;
	});

	if (window.debug === true) {
		$inspect("global state", layoutState.tree);
	}

	$effect(() => {
		saveLayout.set(layoutState.tree);
		layoutTree.childs = layoutState.tree;
		treeLayout.set(layoutTree);
	});
</script>

<Panel
	{id}
	{theme}
	keyId={generateKeyId(16)}
	style="height: 100%"
	pushOtherPanes={false}
	on:resized={(event) => {
		debugEvent(event);
		layoutState.tree = event.detail;
	}}
>
	<!--
Okay so, here's the problem,

I need a single source of (reactive) truth for the layout.
There are a couple of sources for the layout, that being:
- A described layout like the one in `test.ts`, that includes the Component that
  needs to be rendered by the Pane/SubPanel
- A serialized layout that is stored like the one stored
  in localStorage, that doesn't include any component

I need to find a way to be able to create a single source of truth (likely in the shape of VizSubPanel[])
that can be used to generate the layout (like it is currently below)

I can maybe use a schema of some sort that describes the layout based on a sort of input structure,
similar to what ProseMirror does, but without all of the complicated shape stuff and regex etc etc.
Just something that sort of describes the shape

Update - 28/06/2025:
I think in V1, if I'm not able to find a decent way to describe a model,
the user will have to create their own tabs/layout on their first launch
and I can give them some pre-configured layouts to choose if they wish.
However, those will have to be custom made and typed out manually :)

We can soooort of generate layouts at the moment (like those stored in localStorage) but those don't contain the
component yet which is a bit of a problem I guess
	-->
	{#each internalLayoutState as panel, i}
		{#if !panel.childs}
			<SubPanel {...panel}></SubPanel>
		{:else}
			{@const subpanel = panel.childs.internalSubPanelContainer}
			{@const internalParentKeyId = generateKeyId(16)}
			{@const internalSubPanelKeyId = generateKeyId(16)}
			<SubPanel {...subpanel} paneKeyId={internalSubPanelKeyId} header={false} views={[]}>
				<Panel
					{...panel.childs.internalPanelContainer}
					keyId={internalParentKeyId}
					on:resized={(event) => {
						debugEvent(event);
					}}
				>
					<!-- TODO: Document and explain what the hell is going on -->
					<!-- DO NOT MOVE THIS {#key}: THIS ONLY RE-RENDERS ANY CHILD SUBPANELS THAT HAVE NEW VIEWS -->
					<!-- MOVING THIS ANYWHERE ELSE FURTHER UP THE LAYOUT HIERACHY, USING ANY OTHER VALUE, RE-RENDERS EVERYTHING WHICH IS UNNCESSARILY EXPENSIVE OR IT DOESN'T RENDER THE TABS/HEADER OF SOME SUBPANELS AT ALL -->
					<!-- ONLY, AND ONLY CHANGE THIS IF YOU CAN PROVE IT IS BETTER TO DO SO THAN THIS, THIS TOOK ME AGES AND DROVE ME CRAZY FOR 2 DAYS STRAIGHT -->
					{#key panel.childs.subPanel.length}
						<SubPanel {...panel} />
						{#each panel.childs.subPanel as subPanel}
							<SubPanel {...subPanel} />
						{/each}
					{/key}
				</Panel>
			</SubPanel>
		{/if}
	{/each}
</Panel>
