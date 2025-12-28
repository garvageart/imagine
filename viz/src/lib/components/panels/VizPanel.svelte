<script lang="ts">
	import { DEFAULT_THEME } from "$lib/constants";
	import { testLayout } from "$lib/layouts/test";
	import {
		Splitpanes as Panel,
		type ITree
	} from "$lib/third-party/svelte-splitpanes";
	import {
		layoutState,
		layoutTree
	} from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { VizLocalStorage } from "$lib/utils/misc";
	import { onMount } from "svelte";
	import SubPanel from "./SubPanel.svelte";
	import VizSubPanelData, { Content } from "$lib/layouts/subpanel.svelte";
	import { debugEvent } from "$lib/utils/dom";
	import { generateKeyId } from "$lib/utils/layout";
	import { debugMode } from "$lib/states/index.svelte";

	let { id }: { id: string } = $props();
	const theme = DEFAULT_THEME;
	const saveLayout = new VizLocalStorage<VizSubPanelData[]>("layout");
	const treeLayout = new VizLocalStorage<ITree>("tree");
	let storedLayout = saveLayout.get();

	function resetLayoutToDefault() {
		saveLayout.set(testLayout);
		storedLayout = testLayout;
	}

	function intializeLayoutStructures() {
		layoutState.tree ??= [];
		let layout = storedLayout;
		if (layout) {
			layout = layout.map((panel) => {
				return new VizSubPanelData({
					...panel,
					content: panel.childs.content.map((content) => {
						return new Content({
							...content,
							size: content.size === null ? undefined : content.size
						});
					})
				});
			});
		} else {
			layout = testLayout;
		}

		layoutState.tree = layout;
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
		if (storedTree.activeContentId) {
			layoutTree.activeContentId = storedTree.activeContentId;
		}
	}

	if (storedLayout && storedLayout?.length === 0) {
		console.warn("No layout found in localStorage, using default layout");
		resetLayoutToDefault();
	}

	onMount(() => {
		intializeLayoutStructures();
	});

	// This derived value was initially used to do
	// further layout calculations like checking if a single pane
	// needs to be used but that seems to have just fixed itself?
	// So for now all it does is save the layout every time it is adjusted
	const internalLayoutState = $derived.by(() => {
		return layoutState.tree;
	});

	if (debugMode) {
		$inspect("global state", layoutState.tree);
	}

	$effect(() => {
		// Properly serialize the layout tree including views using their toJSON method
		const layoutToSave = layoutState.tree.map((panel) => {
			// Access deep properties to ensure reactivity tracks them
			const childsCopy = {
				...panel.childs,
				content: panel.childs.content.map((content) => ({
					...content,
					views: content.views.map((view) => {
						// Use toJSON method if available (VizView instances), otherwise spread
						if (view && typeof view.toJSON === "function") {
							return view.toJSON();
						}
						return {
							...view,
							isActive: view.isActive
						};
					})
				}))
			};

			return Object.assign(
				{},
				{
					id: panel.id,
					keyId: panel.paneKeyId,
					locked: (panel as any).locked,
					size: panel.size,
					minSize: panel.minSize,
					maxSize: panel.maxSize,
					class: panel.class,
					childs: childsCopy,
					views: panel.views
				}
			);
		}) as unknown as VizSubPanelData[];
		saveLayout.set(layoutToSave);
		layoutTree.childs = layoutState.tree;
		const layoutTreeSave = { ...layoutTree } as unknown as ITree;
		layoutTreeSave.childs = layoutToSave;
		layoutTreeSave.activeContentId = layoutTree.activeContentId;
		treeLayout.set(layoutTreeSave);
	});

	function handleRootResize(event: CustomEvent<VizSubPanelData[]>) {
		debugEvent(event);
		// Update sizes of root panels while preserving the existing objects and their state (views)
		const newPanels = event.detail;
		if (!Array.isArray(newPanels)) {
			return;
		}

		layoutState.tree = layoutState.tree.map((existingPanel) => {
			const newPanel = newPanels.find(
				(p) => p.paneKeyId === existingPanel.paneKeyId
			);
			if (newPanel) {
				existingPanel.size = newPanel.size;
			}
			return existingPanel;
		});
	}

	function handleInnerResize(
		event: CustomEvent<Content[]>,
		parentPanelIndex: number
	) {
		debugEvent(event);
		// Update sizes of content in the specific parent panel
		const newContents = event.detail;
		if (!Array.isArray(newContents)) {
			return;
		}

		const parentPanel = layoutState.tree[parentPanelIndex];
		if (
			!parentPanel ||
			!parentPanel.childs ||
			!Array.isArray(parentPanel.childs.content)
		) {
			return;
		}

		parentPanel.childs.content = parentPanel.childs.content.map(
			(existingContent) => {
				const newContent = newContents.find(
					(c) => c.paneKeyId === existingContent.paneKeyId
				);
				if (newContent) {
					existingContent.size = newContent.size;
				}
				return existingContent;
			}
		);
	}
</script>

<Panel
	{id}
	{theme}
	keyId={generateKeyId(16)}
	style="max-height: 100%;"
	pushOtherPanes={false}
	on:resized={handleRootResize}
>
	<!--
TODO: Rewrite the ENTIRE layout code and create a new framework that isn't this hodgepodge of reworked code
	-->
	{#each internalLayoutState as panel, i}
		{#key panel.childs.content.length}
			<!-- empty array for views to supress typescript errors about required views -->
			<SubPanel
				{...panel.childs.internalSubPanelContainer}
				class="viz-internal-subpanel"
				header={false}
				maxSize={100}
				views={[]}
			>
				<Panel
					{...panel.childs.internalPanelContainer}
					class="viz-internal-panel"
					on:resized={(e) => handleInnerResize(e, i)}
				>
					<!-- TODO: Document and explain what the hell is going on -->
					<!-- ---------------------------------------------------- -->
					<!-- DO NOT MOVE THIS {#key}: THIS ONLY RE-RENDERS ANY CHILD SUBPANELS THAT HAVE NEW VIEWS -->
					<!-- MOVING THIS ANYWHERE ELSE FURTHER UP THE LAYOUT HIERACHY, USING ANY OTHER VALUE, RE-RENDERS EVERYTHING WHICH IS UNNCESSARILY EXPENSIVE OR IT DOESN'T RENDER THE TABS/HEADER OF SOME SUBPANELS AT ALL -->
					<!-- ONLY, AND ONLY CHANGE THIS IF YOU CAN PROVE IT IS BETTER TO DO SO THAN THIS, THIS TOOK ME AGES AND DROVE ME CRAZY FOR 2 DAYS STRAIGHT -->
					{#each panel.childs.content as subPanel}
						{#key subPanel.paneKeyId}
							<SubPanel {...subPanel} id={subPanel.id ?? ""} />
						{/key}
					{/each}
				</Panel>
			</SubPanel>
		{/key}
	{/each}
</Panel>
