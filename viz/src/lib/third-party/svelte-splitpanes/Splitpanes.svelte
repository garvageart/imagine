<script module lang="ts">
	import "./internal/default-theme.scss";
	export const KEY = {};

	type ResultType = boolean | { passive: boolean };
	/**
	 * the third argument for event bundler
	 * @see https://github.com/WICG/EventListenerOptions/blob/gh-pages/explainer.md
	 */
	const thirdEventArg = (() => {
		let result: ResultType = false;

		try {
			const arg = Object.defineProperty({}, "passive", {
				get() {
					result = { passive: true };
					return true;
				}
			});

			// @ts-expect-error no overload matches
			window.addEventListener("testpassive", arg, arg);

			// @ts-expect-error no overload matches
			window.remove("testpassive", arg, arg);
		} catch (_e) {
			/* */
		}

		return result;
	})();
</script>

<script lang="ts">
	import { onMount, onDestroy, setContext, createEventDispatcher, tick } from "svelte";
	import { get, writable } from "svelte/store";
	import type { IPane, IPaneSizingEvent, SplitContext, PaneInitFunction, ClientCallbacks, ITree } from "./index.js";
	import GatheringRound from "./internal/GatheringRound.svelte";
	import { browser } from "./internal/env.js";
	import { getDimensionName } from "./internal/utils/sizing.js";
	import {
		type Position,
		elementRectWithoutBorder,
		getGlobalMousePosition,
		positionDiff,
		getElementRect
	} from "./internal/utils/position.js";
	import { forEachPartial, sumPartial } from "./internal/utils/array.js";
	import { calcComputedStyle } from "./internal/utils/styling.js";
	import { generateKeyId, VizLocalStorage } from "$lib/utils";
	import { allSplitpanes, layoutState } from "./state.svelte";
	import type { VizSubPanel } from "$lib/components/panels/SubPanel.svelte";

	// TYPE DECLARATIONS ----------------

	// type events

	// TODO: Pass objects as details for all the events. This is a breaking change, but gives us:
	// * Better typing capabilities of the parameters.
	// * API stability: Adding more parameters will be possible without breaking the API.

	// TODO: The readme file isn't correctly describe the `IPaneSizingEvent` stuff - it doesn't return the pane itself.
	//  I suggest to change the event to give the panes themselves or at least their `IPanes`, but this is another breaking change.

	// Suggestions for the `TODO` stuff - collect all the breaking changes and perform them after the sizing in pixels,
	//  since this is another (possibly) breaking change already?

	// eslint-disable-next-line
	interface $$Events {
		/**
		 * When clicking (or touching) a pane.
		 *
		 * Returns the clicked pane object with its dimensions.
		 */
		"pane-click": CustomEvent<IPane>;

		/**
		 * Fires when you click a splitter.
		 *
		 * Returns the next pane object (with its dimensions) directly after the clicked splitter.
		 *
		 * This event is only emitted if dragging did not occur between mousedown and mouseup.
		 */
		"splitter-click": CustomEvent<IPane>;

		/**
		 * Fires when the pane is maximized (ie. typically by double clicking the splitter).
		 *
		 * Returns the maximized pane object with its.
		 */
		"pane-maximize": CustomEvent<IPane>;

		/**
		 * Fires when a pane is reset to its default size.
		 *
		 * Returns the reset pane object with its dimensions.
		 *
		 */
		"pane-reset": CustomEvent<IPane>;

		/**
		 * Fires when splitpanes is ready.
		 */
		ready: CustomEvent<VizSubPanel[]>;

		/**
		 * Fires while resizing (on mousemove/touchmove).
		 *
		 * Returns the clicked pane object with its dimensions.
		 */
		resize: CustomEvent<IPaneSizingEvent[]>;

		/**
		 * Fires once when the resizing stops after user drag (on mouseup/touchend) or when adding or removing a pane.
		 *
		 * Returns the clicked pane object with its dimensions.
		 */
		resized: CustomEvent<VizSubPanel[]>;

		/**
		 * Fires when a pane is added.
		 */
		"pane-add": CustomEvent<{
			/**
			 * The index of the added pane.
			 */
			index: number;

			/**
			 * The new array of panes after resize.
			 */
			panes: IPaneSizingEvent[];
		}>;

		/**
		 * Fires when a pane is removed.
		 */
		"pane-remove": CustomEvent<{
			/**
			 * The removed pane.
			 */
			removed: IPane;

			/**
			 * An array of all the remaining pane objects with their dimensions (after resize).
			 */
			panes: IPaneSizingEvent[];
		}>;
	}

	// used to size panes
	interface Sums {
		prevPanesSize: number;
		nextPanesSize: number;
		prevReachedMinPanes: number;
		nextReachedMinPanes: number;
	}

	// css class

	// FOR VIZ ONLY ----------------
	const storedLayout = new VizLocalStorage<VizSubPanel[]>("layout").get();

	interface Props {
		// horiz or verti?
		horizontal?: boolean;
		/**
		 * NOTE: The library default is true, but a Viz `SubPanel` size should only
		 * be able to be changed by the user, not by other panes
		 *
		 * when true, moving a splitter can push other panes
		 */
		pushOtherPanes?: boolean;
		// open/close on double click
		dblClickSplitter?: boolean;
		// true if RTL
		rtl?: boolean | "auto";
		// true to display the first splitter
		firstSplitter?: boolean;
		// css style
		style?: string | null;
		// the splitter theme to use
		theme?: string;
		class?: string;
		id: string;
		keyId?: string;
		children?: import("svelte").Snippet;
	}

	let {
		horizontal = false,
		pushOtherPanes = false,
		dblClickSplitter = true,
		rtl = "auto",
		firstSplitter = false,
		style = null,
		theme = "default-theme",
		class: clazz = "",
		id = $bindable(),
		keyId = "",
		children
	}: Props = $props();

	if (!id || id.trim() === "") {
		throw new Error("Splitpanes: id is required");
	}

	let currentFocusedPane: HTMLElement;
	let splitpanesKeyId = storedLayout?.flat().find((sp) => sp.id === id)?.paneKeyId;
	const usedKeyId = splitpanesKeyId?.trim() ?? keyId ?? generateKeyId(16);
	id += `-${usedKeyId}`;

	// VARIABLES ----------------

	//used to bubble events up
	const dispatch = createEventDispatcher();

	// the splitpane component
	let container: HTMLElement | undefined = $state(undefined);
	// true when component is ready, prevents emitting console warnings on hot reloading.
	let isReady = false;
	// true when pane reset is awaiting for the next tick, to avoid double call to pane reset.
	let isAwaitingPaneReset = false;
	// true after the initial timeout 0 waiting, prevents CSS transitions until then.
	let isAfterInitialTimeoutZero = $state(false);
	// true when mouse is down
	let isMouseDown = $state(false);
	// true when a splitter is being dragged
	let isDragging = $state(false);
	// that's the splitter than is being dragged
	let activeSplitter = -1;
	// that's well the clicked splitter!
	let clickedSplitter = -1;
	// used to detect double clicks
	let timeoutId: NodeJS.Timeout | null;
	// panes per insertion order (pane.index is the order index)
	let panes = new Array<IPane>();
	// passed to the children via the context - writable to ensure proper reactivity
	const isHorizontal = writable<boolean>(horizontal);
	const showFirstSplitter = writable<boolean>(firstSplitter);
	// tells the key of the very first pane, or undefined if not recieved yet
	const veryFirstPaneKey = writable<unknown>(undefined);

	let activeSplitterElement: HTMLElement;
	let activeSplitterDrag: number;
	let ssrPaneDefinedSizeSum = 0;
	let ssrPaneUndefinedSizeCount = 0;

	// REACTIVE ----------------

	$effect(() => {
		$isHorizontal = horizontal;
	});
	$effect(() => {
		$showFirstSplitter = firstSplitter;
	});

	function isSplitpanes(element: Element): boolean {
		return element.classList.contains("splitpanes");
	}

	function isPane(element: Element): boolean {
		return element.classList.contains("splitpanes__pane");
	}

	// Determines if this is the very first splitpanes element in the DOM,
	// the root element in the tree. Anything else is nested
	function isSplitpanesRoot(element: Element): boolean {
		const parent = element.parentElement;
		if (!parent) {
			return true;
		}

		const parentIsPane = isPane(parent);
		const parentIsSplitpanes = isSplitpanes(parent);
		const isSplitpanesElement = isSplitpanes(element);

		return isSplitpanesElement && parentIsPane === false && parentIsSplitpanes === false;
	}

	// used to complete rendering service side (SSR mode)
	function ssrRegisterPaneSize(size: number | null) {
		if (size === null) {
			++ssrPaneUndefinedSizeCount;
		} else {
			ssrPaneDefinedSizeSum += size;
		}
	}

	const onPaneInit: PaneInitFunction = (_key) => {
		if ($veryFirstPaneKey === undefined) {
			$veryFirstPaneKey = _key;
		}

		return {
			undefinedPaneInitSize: browser ? 0 : (100 - ssrPaneDefinedSizeSum) / ssrPaneUndefinedSizeCount
		};
	};

	setContext<SplitContext>(KEY, {
		keyId: usedKeyId,
		showFirstSplitter,
		veryFirstPaneKey,
		isHorizontal,
		ssrRegisterPaneSize: browser ? undefined : ssrRegisterPaneSize,
		onPaneInit,
		clientOnly: browser
			? {
					onPaneAdd,
					onPaneRemove
				}
			: undefined
	});

	function onPaneAdd(pane: IPane): ClientCallbacks {
		// 1. Add pane to array at the same index it was inserted in the <splitpanes> tag.
		let index = -1;
		if (pane.element.parentNode) {
			Array.from(pane.element.parentNode.children).some((el: Element) => {
				if (el.className.includes("splitpanes__pane")) {
					index++;
				}

				return el === pane.element;
			});
		}

		if (index === 0) {
			// Need to update the first pane key, because the first pane can be changed in runtime.
			$veryFirstPaneKey = pane.key;
		}

		// inserts pane at proper array index
		panes.splice(index, 0, pane);

		// reindex panes
		for (let i = 0; i < panes.length; i++) {
			panes[i].index = i;
		}

		if (isReady) {
			// 2. tick and resize the panes.
			tickAndResetPaneSizes().then(() => {
				// 3. Set the pane as ready
				pane.isReady = true;

				// 4. Fire `pane-add` event.
				dispatch("pane-add", {
					index,
					panes: prepareResizeEvent()
				});
			});
		}

		const paneForward =
			<T,>(cb: (value: T, pane: IPane) => void, includingFirst = true) =>
			(value: T) => {
				if (includingFirst || pane.index > 0) {
					cb(value, pane);
				}
			};

		return {
			onSplitterDown: paneForward(onMouseDown, false),
			onSplitterClick: paneForward(onSplitterClick, false),
			onSplitterDblClick: paneForward(onSplitterDblClick),
			onPaneClick: paneForward(onPaneClick),
			reportGivenSizeChange: paneForward(reportGivenSizeChange)
		};
	}

	async function onPaneRemove(key: unknown) {
		// 1. Remove the pane from array and redo indexes.
		const index = panes.findIndex((p) => p.key === key);

		// race condition - typically happens when the dev server restarts
		if (index >= 0) {
			const removed = panes.splice(index, 1)[0];

			// reindex panes
			for (let i = 0; i < panes.length; i++) {
				panes[i].index = i;
			}

			if (index === 0) {
				$veryFirstPaneKey = panes.length > 0 ? panes[0].key : undefined;
			}

			if (isReady) {
				// 3. tick and resize the panes.
				await tickAndResetPaneSizes();
				normalizePaneSizes();

				// 4. Fire `pane-remove` event.
				dispatch("pane-remove", {
					removed,
					panes: prepareResizeEvent()
				});
			}
		}
	}

	// called by sub-panes
	function onPaneClick(_event: MouseEvent, pane: IPane) {
		const element = pane.element;
		if (currentFocusedPane && currentFocusedPane !== element) {
			currentFocusedPane.classList.remove("splitpanes__pane__active");
		}

		currentFocusedPane = element as HTMLElement;

		if (!element.classList.contains("splitpanes__pane__active")) {
			element.classList.add("splitpanes__pane__active");
		} else {
			element.classList.remove("splitpanes__pane__active");
		}

		pane.isActive.set(!get(pane.isActive));
		dispatch("pane-click", pane);
	}

	function reportGivenSizeChange(newGivenSize: number, pane: IPane) {
		pane.setSz(newGivenSize);
		tickAndResetPaneSizes();
	}

	onMount(() => {
		verifyAndUpdatePanesOrder();
		resetPaneSizes();

		for (let i = 0; i < panes.length; i++) {
			const pane = panes[i];
			pane.isReady = true;
			pane.parent = usedKeyId;

			panes[i] = pane;
		}

		isReady = true;
		dispatch("ready", calculateTree());

		setTimeout(() => {
			isAfterInitialTimeoutZero = true;
		}, 0);
	});

	if (browser) {
		onDestroy(() => {
			if (isReady) {
				// this is to solve an edge case:
				// when the user starts dragging and the component is destroyed, leaving behind hanging events
				unbindEvents();
			}

			// Prevent emitting console warnings on hot reloading.
			isReady = false;
			$allSplitpanes.delete(usedKeyId);
		});
	}

	// Tells in the current DOM state if we are in RTL direction or not.
	function isRTL(containerComputedStyle?: CSSStyleDeclaration) {
		if (rtl === "auto") {
			// the try catch is to support old browser, flag is preset to false
			try {
				return (containerComputedStyle ?? calcComputedStyle(container!)).direction === "rtl";
			} catch (_err) {
				// We want application to not crush, but don't care about the message
			}
		}
		// otherwise

		return rtl === true;
	}

	function bindEvents() {
		document.body.style.cursor = isHorizontal ? "ns-resize" : "ew-resize";

		document.addEventListener("mousemove", onMouseMove, thirdEventArg);
		document.addEventListener("mouseup", onMouseUp);

		if ("ontouchstart" in window) {
			document.addEventListener("touchmove", onMouseMove, thirdEventArg);
			document.addEventListener("touchend", onMouseUp);
		}
	}

	function unbindEvents() {
		document.body.style.cursor = "";

		document.removeEventListener("mousemove", onMouseMove);
		document.removeEventListener("mouseup", onMouseUp);

		if ("ontouchstart" in window) {
			document.removeEventListener("touchmove", onMouseMove);
			document.removeEventListener("touchend", onMouseUp);
		}
	}

	const isSplitterElement = (node: Node) =>
		node.nodeType === Node.ELEMENT_NODE && (node as HTMLElement).classList.contains("splitpanes__splitter");

	function getOrientedDiff(drag: Position, elementSize: number, isRTL: boolean): number {
		let tdrag = drag[horizontal ? "top" : "left"];
		if (isRTL && !horizontal) tdrag = elementSize - tdrag;

		return tdrag;
	}

	const getCurrentDimensionName = () => getDimensionName(horizontal);

	function onMouseDown(event: TouchEvent | MouseEvent, splitterPane: IPane) {
		isMouseDown = true;
		activeSplitter = splitterPane.index;

		splitterPane.setSplitterActive(true);
		const paneElement = splitterPane.element;

		let activeSplitterNode: Node | null = paneElement;
		while (activeSplitterNode) {
			activeSplitterNode = activeSplitterNode.previousSibling;
			if (activeSplitterNode && isSplitterElement(activeSplitterNode)) {
				break;
			}
		}
		if (activeSplitterNode === null) {
			console.error("Splitpane Error: Active splitter wasn't found!");
			return; // Don't bind move event on error
		}

		activeSplitterElement = activeSplitterNode as HTMLElement;

		const globalMousePosition = getGlobalMousePosition(event);
		const splitterRect = getElementRect(activeSplitterElement);
		activeSplitterDrag = getOrientedDiff(
			positionDiff(globalMousePosition, splitterRect),
			splitterRect[getCurrentDimensionName()],
			isRTL()
		);

		bindEvents();
	}

	function onMouseMove(event: MouseEvent | TouchEvent) {
		if (isMouseDown) {
			isDragging = true;

			const globalMousePosition = getGlobalMousePosition(event);
			const containerComputedStyle = calcComputedStyle(container!);
			const containerRectWithoutBorder = elementRectWithoutBorder(container!, containerComputedStyle);
			const containerSizeWithoutBorder: number = containerRectWithoutBorder[getCurrentDimensionName()];
			const _isRTL = isRTL(containerComputedStyle);

			const currentMouseDrag = positionDiff(globalMousePosition, containerRectWithoutBorder);
			const tdrag = getOrientedDiff(currentMouseDrag, containerSizeWithoutBorder, _isRTL);
			calculatePanesSize(tdrag, containerSizeWithoutBorder);

			dispatch("resize", prepareResizeEvent());
		}
	}

	function onMouseUp() {
		if (isDragging) {
			dispatch("resized", calculateTree());
		}
		isMouseDown = false;

		const pane = panes[activeSplitter];
		pane.setSplitterActive(false);

		// Keep dragging flag until click event is finished (click happens immediately after mouseup)
		// in order to prevent emitting `splitter-click` event if splitter was dragged.
		setTimeout(() => {
			isDragging = false;
			unbindEvents();
		}, 100);
	}

	// If touch device, detect double tap manually (2 taps separated by less than 500ms).
	function onSplitterClick(event: MouseEvent, splitterPane: IPane) {
		if ("ontouchstart" in window) {
			const splitterIndex = splitterPane.index;

			// Detect splitter double taps if the option is on.
			if (dblClickSplitter) {
				if (clickedSplitter === splitterIndex) {
					if (timeoutId) clearTimeout(timeoutId);
					timeoutId = null;
					onSplitterDblClick(event, splitterPane);
					clickedSplitter = -1; // Reset for the next tap check.
				} else {
					clickedSplitter = splitterIndex;
					timeoutId = setTimeout(() => {
						clickedSplitter = -1;
					}, 500);
				}
			}
		}

		if (!isDragging) dispatch("splitter-click", splitterPane);
	}

	// On splitter dbl click or dbl tap maximize this pane.
	function onSplitterDblClick(_event: MouseEvent, splitterPane: IPane) {
		if (dblClickSplitter) {
			const splitterIndex = splitterPane.index;

			let totalMinSizes = 0;
			for (let i = 0; i < panes.length; i++) {
				const pane = panes[i];
				if (i !== splitterIndex) {
					totalMinSizes += pane.min();
				}
			}

			const maxExtendedSize = Math.min(Math.max(0, 100 - totalMinSizes), splitterPane.max());

			// Resets to default size if the splitter pane is already maximized
			if (splitterPane.sz() >= maxExtendedSize) {
				// Reset all panes to their default (given) size or equal share if not defined
				const defaultTotal = panes.reduce((acc, p) => acc + (typeof p.givenSize === "number" ? p.givenSize : 0), 0);
				const undefinedCount = panes.filter((p) => typeof p.givenSize !== "number").length;
				const fallbackSize = undefinedCount > 0 ? (100 - defaultTotal) / undefinedCount : 0;

				for (let i = 0; i < panes.length; i++) {
					const pane = panes[i];
					if (typeof pane.givenSize === "number") {
						pane.setSz(pane.givenSize);
					} else {
						pane.setSz(fallbackSize);
					}
				}
				dispatch("pane-reset", splitterPane);
				dispatch("resized", calculateTree());

				isMouseDown = false;
				return;
			}

			const totalMaxExtendedPlusMinSizes = totalMinSizes + maxExtendedSize;
			if (totalMaxExtendedPlusMinSizes >= 100) {
				// put everything to the minimum, and in the splitterPane put the rest of the size
				for (let i = 0; i < panes.length; i++) {
					const pane = panes[i];
					if (pane !== splitterPane) {
						pane.setSz(pane.min());
					} else {
						pane.setSz(100 - totalMinSizes);
					}
				}
			} else {
				// notice that in this case, we can conclude that `panes.length >= 2`

				// put splitterPane to the maximum (the actual one), and the normal panes to the minimum,
				//  and give the spare to left pane (or to the right pane, if the splitterPane is the first pane)
				// if this spare size is beyond the pane maximum, need to pass it along to the other panes

				let leftSpare = 100 - totalMaxExtendedPlusMinSizes;

				splitterPane.setSz(maxExtendedSize);

				const giveBest = (pane: IPane) => {
					const min = pane.min();
					const max = pane.max();

					const szExtra = Math.min(Math.max(0, leftSpare), max - min);
					pane.setSz(min + szExtra);
					leftSpare -= szExtra;
				};

				// go backward and give the most size as we can
				for (let i = splitterIndex - 1; i >= 0; i--) giveBest(panes[i]);

				// go forward and give the most size as we can
				for (let i = splitterIndex + 1; i < panes.length; i++) giveBest(panes[i]);

				// at the end of the process, we must have that `leftSpare` is 0
				if (leftSpare !== 0) {
					console.warn(
						"Splitpanes: there is a left spare size after computation of splitter double click, which means there are issues on the size constains of the panes."
					);
				}
			}

			dispatch("pane-maximize", splitterPane);
			dispatch("resized", calculateTree());
		}
		// onMouseUp might not be called on the second click, so update the mouse state.
		// TODO: Should also check and unbind events, but better IMO to not bind&unbind on every click, so ignored for now.
		isMouseDown = false;
	}

	function findPanesChildren() {
		for (const pane of panes) {
			const childrenElements = Array.from(pane.element.children).filter((el) => isSplitpanes(el) || isPane(el));
			for (const child of childrenElements) {
				pane.childs.push(child.attributes.getNamedItem("data-viz-sp-id")?.value as string);
			}
		}
	}

	const prepareResizeEvent = (): IPaneSizingEvent[] =>
		panes.map((pane) => ({
			min: pane.min(),
			max: pane.max(),
			size: pane.sz(),
			snap: pane.snap()
		}));

	// Note: Partially written by Co-Pilot
	function normalizePaneSizes() {
		// Remove panes with size 0 or that are not visible
		panes = panes.filter((pane) => pane.sz() > 0);

		// Reindex panes
		for (let i = 0; i < panes.length; i++) {
			panes[i].index = i;
		}

		// Recalculate total size
		const total = panes.reduce((sum, pane) => sum + pane.sz(), 0);

		if (total !== 100 && total > 0) {
			for (let i = 0; i < panes.length; i++) {
				const pane = panes[i];
				pane.setSz((pane.sz() / total) * 100);
			}
		}

		// Enforce min/max constraints
		for (let i = 0; i < panes.length; i++) {
			const pane = panes[i];
			if (pane.sz() < pane.min()) {
				pane.setSz(pane.min());
			}

			if (pane.sz() > pane.max()) {
				pane.setSz(pane.max());
			}
		}
	}

	function calculateTree() {
		findPanesChildren();
		const currentLayout = layoutState.tree;

		function updatePanelSize(panel: VizSubPanel) {
			const pane = panes.find((p) => p.keyId === panel.paneKeyId);
			let updatedPanel = { ...panel };

			if (pane) {
				updatedPanel.size = pane.sz();
				updatedPanel.minSize = pane.min();
				updatedPanel.maxSize = pane.max();
				updatedPanel.snapSize = pane.snap();
			}

			if (panel.childs && panel.childs.subPanel && Array.isArray(panel.childs.subPanel)) {
				updatedPanel.childs = {
					...panel.childs,
					subPanel: panel.childs.subPanel.map(updatePanelSize)
				};
			}

			if (panel.childs && panel.childs.parentPanel) {
				updatedPanel.childs = {
					...updatedPanel.childs,
					parentPanel:
						panel.childs.parentPanel && "paneKeyId" in panel.childs.parentPanel
							? updatePanelSize(panel.childs.parentPanel as VizSubPanel)
							: panel.childs.parentPanel,
					parentSubPanel: updatedPanel.childs?.parentSubPanel ?? ({} as any),
					subPanel: updatedPanel.childs?.subPanel ?? []
				};
			}

			if (panel.childs && panel.childs.parentSubPanel) {
				updatedPanel.childs = {
					...updatedPanel.childs,
					parentSubPanel: updatePanelSize(panel.childs.parentSubPanel as VizSubPanel),
					parentPanel: updatedPanel.childs?.parentPanel ?? ({} as any),
					subPanel: updatedPanel.childs?.subPanel ?? []
				};
			}
			return updatedPanel;
		}

		const updatedLayout = currentLayout.map(updatePanelSize);

		return updatedLayout;
	}

	/**
	 * Returns the drag percentage of the splitter relative to the 2 parts it's inbetween, meaning the ratio between
	 *  the size that all the panes before the splitter consumes (ignoring other splitters size) and the total size of the container.
	 */
	function getCurrentDragPercentage(tdrag: number, containerSizeWithoutBorder: number) {
		// Here we want the splitter size **including the borders**.
		// We need to use `Element.getBoundingClientRect()` and not `Element.clientWidth` and `Element.clientHeight`,
		//  bacause the latter round the number of pixels to integer, and additionally, they don't include the borders.
		const splitterSize = (node: Node) => getElementRect(node as HTMLElement)[getCurrentDimensionName()];

		const activeSplitterSize = splitterSize(activeSplitterElement);

		let splittersTotalSizeBefore = 0;
		let currentBeforeNode = activeSplitterElement.previousSibling;
		while (currentBeforeNode) {
			if (isSplitterElement(currentBeforeNode)) {
				splittersTotalSizeBefore += splitterSize(currentBeforeNode);
			}
			currentBeforeNode = currentBeforeNode.previousSibling;
		}

		let splittersTotalSizeAfter = 0;
		let currentAfterNode = activeSplitterElement.nextSibling;
		while (currentAfterNode) {
			if (isSplitterElement(currentAfterNode)) {
				splittersTotalSizeAfter += splitterSize(currentAfterNode);
			}
			currentAfterNode = currentAfterNode.nextSibling;
		}

		const totalSplitterBefore = splittersTotalSizeBefore + activeSplitterDrag;
		const totalSplitter = splittersTotalSizeBefore + activeSplitterSize + splittersTotalSizeAfter;

		// An explanation to the mathematical computation:
		//
		// Let's start with the case of only two panes. If we mark the first pane size in prec
		//  (thinking about it as a number between 0 to 1) as `x`, we'll get that the size of the left pane in pixels will be:
		// `x*containerSizeWithoutBorder - x*totalSplitter = x*(containerSizeWithoutBorder - totalSplitter)`
		// Since we want that the total size in pixels before the user mouse pointer will be `tdrag`, and we need to add the
		//  size of the splitter itself that is before the mouse pointer, we get the equation:
		// `x*(containerSizeWithoutBorder - totalSplitter) + activeSplitterDrag = tdrag`
		//
		// Now in the general case when we have many panes before the splitter, mark their precentages
		//  (again, thinking about it as a number between 0 to 1) by x1,x2,...,xn we'll get the equation:
		// `(x1 + ... + xn)*(containerSizeWithoutBorder - totalSplitter) + totalSplitterBefore = tdrag`
		// And solving it yeild the answer:
		// `x1 + ... + xn = (tdrag - totalSplitterBefore) / (containerSizeWithoutBorder - totalSplitter)`

		return ((tdrag - totalSplitterBefore) / (containerSizeWithoutBorder - totalSplitter)) * 100;
	}

	/**
	 * Called when slitters are moving to adjust pane sizes
	 */
	function calculatePanesSize(tdrag: number, containerSizeWithoutBorder: number) {
		let paneBeforeIndex = activeSplitter - 1;
		let paneBefore = panes[paneBeforeIndex];

		let paneAfterIndex = activeSplitter;
		let paneAfter = panes[paneAfterIndex];

		let sums: Sums = {
			prevPanesSize: sumPrevPanesSize(paneBeforeIndex),
			nextPanesSize: sumNextPanesSize(paneAfterIndex),
			prevReachedMinPanes: 0,
			nextReachedMinPanes: 0
		};

		// If not pushing other panes, panes to resize are right before and right after splitter.
		const minDrag = 0 + (pushOtherPanes ? 0 : sums.prevPanesSize);
		const maxDrag = 100 - (pushOtherPanes ? 0 : sums.nextPanesSize);

		// Calculate drag percentage
		const mouseDragPercentage = Math.max(Math.min(getCurrentDragPercentage(tdrag, containerSizeWithoutBorder), maxDrag), minDrag);

		// Handle snap
		const paneBeforeSnap = sums.prevPanesSize + paneBefore.min() + paneBefore.snap();

		const paneAfterSnap = 100 - (sums.nextPanesSize + paneAfter.min() + paneAfter.snap());

		let dragPercentage = mouseDragPercentage;
		let snapped = false;

		if (mouseDragPercentage <= paneBeforeSnap) {
			if (mouseDragPercentage > sums.prevPanesSize + paneBefore.min()) {
				dragPercentage = Math.max(paneBefore.min() + sums.prevPanesSize, 100 - (paneAfter.max() + sums.nextPanesSize));
				snapped = true;
			}
		} else if (mouseDragPercentage >= paneAfterSnap) {
			if (mouseDragPercentage < 100 - sums.nextPanesSize - paneAfter.min()) {
				dragPercentage = Math.min(100 - (paneAfter.min() + sums.nextPanesSize), paneBefore.max() + sums.prevPanesSize);
				snapped = true;
			}
		}

		const paneBeforeMaxReached = paneBefore.max() < 100 && dragPercentage >= paneBefore.max() + sums.prevPanesSize;
		const paneAfterMaxReached = paneAfter.max() < 100 && dragPercentage <= 100 - (paneAfter.max() + sums.nextPanesSize);
		// Prevent dragging beyond pane max.
		if (paneBeforeMaxReached || paneAfterMaxReached) {
			if (paneBeforeMaxReached) {
				paneBefore.setSz(paneBefore.max());
				paneAfter.setSz(Math.max(100 - paneBefore.max() - sums.prevPanesSize - sums.nextPanesSize, 0));
			} else {
				paneBefore.setSz(Math.max(100 - paneAfter.max() - sums.prevPanesSize - sums.nextPanesSize, 0));
				paneAfter.setSz(paneAfter.max());
			}
		} else {
			// When pushOtherPanes = true, find the closest expanded pane on each side of the splitter.
			// TODO: Bug: This should work also when removing `!snapped` condition, but it's not!
			//   To reproduce, reload the example page and see the example "Min & max with snap".
			//   It gets wrongly pushed when try to snap on the initial dragging of the first splitter to the right.
			if (pushOtherPanes && !snapped) {
				const vars = doPushOtherPanes(sums, dragPercentage);
				if (!vars) {
					//		setAllPaneDimensions();
					return; // Prevent other calculation.
				}
				({ sums, paneBeforeIndex, paneAfterIndex } = vars);
				paneBefore = panes[paneBeforeIndex];
				paneAfter = panes[paneAfterIndex];
			}

			if (typeof paneBeforeIndex === "number") {
				paneBefore.setSz(
					Math.min(Math.max(dragPercentage - sums.prevPanesSize - sums.prevReachedMinPanes, paneBefore.min()), paneBefore.max())
				);
			}
			if (typeof paneAfterIndex === "number") {
				paneAfter.setSz(
					Math.min(
						Math.max(100 - dragPercentage - sums.nextPanesSize - sums.nextReachedMinPanes, paneAfter.min()),
						paneAfter.max()
					)
				);
			}
		}
	}

	function doPushOtherPanes(sums: Sums, dragPercentage: number) {
		const splitterIndex = activeSplitter - 1;
		let paneBeforeIndex: number | undefined = splitterIndex;
		let paneAfterIndex: number | undefined = splitterIndex + 1;
		// Pushing Down.
		// Going smaller than the current pane min size: take the previous expanded pane.
		if (dragPercentage < sums.prevPanesSize + panes[paneBeforeIndex].min()) {
			paneBeforeIndex = findPrevExpandedPane(splitterIndex)?.index;

			sums.prevReachedMinPanes = 0;

			if (paneBeforeIndex === undefined) {
				// If nothing else to push down, cancel dragging.
				sums.prevPanesSize = 0;
				panes[0].setSz(panes[0].min());
				forEachPartial(panes, 1, splitterIndex + 1, (pane: IPane) => {
					pane.setSz(pane.min());
					sums.prevReachedMinPanes += pane.min();
				});

				panes[paneAfterIndex].setSz(100 - sums.prevReachedMinPanes - panes[0].min() - sums.prevPanesSize - sums.nextPanesSize);
				return null;
			} else {
				// If pushing a n-2 or less pane, from splitter, then make sure all in between is at min size.
				if (paneBeforeIndex < splitterIndex) {
					forEachPartial(panes, paneBeforeIndex + 1, splitterIndex + 1, (pane: IPane) => {
						pane.setSz(pane.min());
						sums.prevReachedMinPanes += pane.min();
					});
				}
				sums.prevPanesSize = sumPrevPanesSize(paneBeforeIndex);
			}
		}
		// Pushing Up.
		// Pushing up beyond min size is reached: take the next expanded pane.
		if (dragPercentage > 100 - sums.nextPanesSize - panes[paneAfterIndex].min()) {
			paneAfterIndex = findNextExpandedPane(splitterIndex)?.index;
			if (paneBeforeIndex === undefined) console.log("Yep undefined paneAfterIndex");
			sums.nextReachedMinPanes = 0;

			if (paneAfterIndex === undefined) {
				// If nothing else to push up, cancel dragging.
				sums.nextPanesSize = 0;

				const panesCount = panes.length;
				panes[panesCount - 1].setSz(panes[panesCount - 1].min());

				forEachPartial(panes, splitterIndex + 1, panesCount - 1, (pane) => {
					pane.setSz(pane.min());
					sums.nextReachedMinPanes += pane.min();
				});

				panes[paneBeforeIndex].setSz(
					100 - sums.prevPanesSize - sums.nextReachedMinPanes - panes[panesCount - 1].min() - sums.nextPanesSize
				);
				return null;
			} else {
				// If pushing a n+2 or more pane, from splitter, then make sure all in between is at min size.
				if (paneAfterIndex > splitterIndex + 1) {
					forEachPartial(panes, splitterIndex + 1, paneAfterIndex, (pane: IPane) => {
						pane.setSz(pane.min());
						sums.nextReachedMinPanes += pane.min();
					});
				}
				sums.nextPanesSize = sumNextPanesSize(paneAfterIndex);
			}
		}
		return { sums, paneBeforeIndex, paneAfterIndex };
	}

	const getSizeOfPane = (pane: IPane) => pane.sz();
	const sumPrevPanesSize = (splitterIndex: number) => sumPartial(panes, 0, splitterIndex, getSizeOfPane);
	const sumNextPanesSize = (splitterIndex: number) => sumPartial(panes, splitterIndex + 1, panes.length, getSizeOfPane);

	// Return the previous pane from siblings which has a size (width for vert or height for horz) of more than 0.
	const findPrevExpandedPane = (splitterIndex: number): IPane | undefined =>
		[...panes].reverse().find((p) => p.index < splitterIndex && p.sz() > p.min());

	// Return the next pane from siblings which has a size (width for vert or height for horz) of more than 0.
	const findNextExpandedPane = (splitterIndex: number): IPane | undefined =>
		panes.find((p) => p.index > splitterIndex + 1 && p.sz() > p.min());

	async function tickAndResetPaneSizes() {
		isAwaitingPaneReset = true;
		await tick();

		if (isAwaitingPaneReset) {
			resetPaneSizes();
			isAwaitingPaneReset = false;
		}
	}

	/**
	 *
	 * @param addedPane
	 * @param removedPane
	 */
	function resetPaneSizes() {
		equalize();
		normalizePaneSizes();

		if (isReady) dispatch("resized", calculateTree());
	}

	function equalize() {
		// Escape the function on the edge case that there is not even a single pane
		if (panes.length === 0) {
			return;
		}
		// otherwise

		const panesCount = panes.length;
		let leftToAllocate = 100;
		let definedSizesCount = 0;
		let undefinedSizesNotReadyCount = 0;
		let undefinedSizesSum = 0;
		const ungrowable: IPane[] = [];
		const unshrinkable: IPane[] = [];

		for (let i = 0; i < panesCount; i++) {
			const pane = panes[i];
			const sz = pane.sz();
			if (pane.givenSize === null) {
				if (pane.isReady) {
					undefinedSizesSum += sz;
					if (sz >= pane.max()) ungrowable.push(pane);
					if (sz <= pane.min()) unshrinkable.push(pane);
				} else {
					undefinedSizesNotReadyCount += 1;
				}
			} else {
				// if the size is defined, we don't modify its size at all
				leftToAllocate -= sz;
				definedSizesCount++;
				ungrowable.push(pane);
				unshrinkable.push(pane);
			}
		}

		const undefinedSizesCount = panesCount - definedSizesCount;
		const undefinedSizesReadyCount = undefinedSizesCount - undefinedSizesNotReadyCount;

		// the proportion of the newly added panes
		let undefinedSizesNotReadySz: number;
		let undefinedScaleFactor: number;
		if (undefinedSizesReadyCount > 0) {
			// if has undefined sizes panes that are ready:
			undefinedSizesNotReadySz = undefinedSizesSum / undefinedSizesReadyCount;
			if (undefinedSizesNotReadySz > 0.1 && leftToAllocate > 0.1) {
				undefinedSizesSum += undefinedSizesNotReadyCount * undefinedSizesNotReadySz;
				undefinedScaleFactor = leftToAllocate / undefinedSizesSum;
			} else {
				// when the size of the ready undefined panes shares are negligible, need to set the not ready
				//  undefined one to size 0, for being "proportional" to negligible sizes
				undefinedSizesNotReadySz = 0;
				undefinedScaleFactor = 1;
			}
		} else {
			// otherwise, divide the space of the undefined sizes panes equally:
			undefinedSizesNotReadySz = leftToAllocate / undefinedSizesCount;
			undefinedScaleFactor = 1;
		}

		// whenever `leftToAllocate` or `undefinedSizesSum` aren't negligible, need to adjact the sizes
		if (leftToAllocate + undefinedSizesSum > 0.1) {
			leftToAllocate = 100; // reset the space calculation

			for (let i = 0; i < panesCount; i++) {
				const pane = panes[i];
				if (!(typeof pane.givenSize === "number")) {
					// add the proportion of the newly added pane if has undefined size
					const currentSz = pane.isReady ? pane.sz() : undefinedSizesNotReadySz;
					const sz = Math.max(Math.min(currentSz * undefinedScaleFactor, pane.max()), pane.min());
					pane.setSz(sz);
				}
				leftToAllocate -= pane.sz();
			}

			// since we multiply by scaling, there might be left space that is needed to be saturated
			if (Math.abs(leftToAllocate) > 0.1) {
				leftToAllocate = readjustSizes(leftToAllocate, ungrowable, unshrinkable);
			}
		}

		if (!isFinite(leftToAllocate)) {
			console.warn("Splitpanes: Internal error, sizes might be NaN as a result.");
		} else if (Math.abs(leftToAllocate) > 0.1) {
			console.warn("Splitpanes: Could not resize panes correctly due to their constraints.", Math.abs(leftToAllocate));
		}
	}

	// Second loop to adjust sizes now that we know more about the panes constraints.
	function readjustSizes(leftToAllocate: number, ungrowable: Array<IPane>, unshrinkable: Array<IPane>): number {
		const panesCount = panes.length;
		const panesSizableCount = panesCount - (leftToAllocate > 0 ? ungrowable.length : unshrinkable.length);
		if (panesSizableCount <= 0) {
			return leftToAllocate;
		}

		const equalSpaceToAllocate = leftToAllocate / panesSizableCount;

		if (panes.length === 1) {
			panes[0].setSz(100);
			leftToAllocate = 0;
		} else
			for (let i = 0; i < panes.length; i++) {
				const pane = panes[i];
				const sz = pane.sz();
				if (leftToAllocate > 0 && !ungrowable.includes(pane)) {
					// Need to diff the size before and after to get the exact allocated space.
					const newPaneSize = Math.max(Math.min(sz + equalSpaceToAllocate, pane.max()), pane.min());
					const allocated = newPaneSize - sz;
					leftToAllocate -= allocated;
					pane.setSz(newPaneSize);
				} else if (!unshrinkable.includes(pane)) {
					// Need to diff the size before and after to get the exact allocated space.
					const newPaneSize = Math.max(Math.min(sz + equalSpaceToAllocate, pane.max()), pane.min());
					const allocated = newPaneSize - sz;
					leftToAllocate -= allocated;
					pane.setSz(newPaneSize);
				}
			}

		return leftToAllocate;
	}

	/**
	 * Checks that Splitpanes is composed of Pane, and verify that the panes are still in the right order,
	 * and if not update the internal order.
	 */
	function verifyAndUpdatePanesOrder() {
		if (!container) {
			return;
		}
		const { children } = container;
		let currentPaneIndex = 0;
		let needReorder = false;

		for (let i = 0; i < children.length; i++) {
			const child = children.item(i) as Element;
			const isPane = child.classList.contains("splitpanes__pane");
			const isSplitter = child.classList.contains("splitpanes__splitter");

			// Node is not a Pane or a splitter: remove it.
			if (!isPane && !isSplitter) {
				child.parentNode?.removeChild(child); // el.remove() doesn't work on IE11.
				console.warn(
					"Splitpanes: Only <Pane> elements are allowed at the root of <Splitpanes>. One of your DOM nodes was removed."
				);
				return;
			} else if (isPane) {
				if (!needReorder && panes[currentPaneIndex].element !== child) {
					needReorder = true;
				}

				currentPaneIndex++;
			}
		}

		if (needReorder) {
			const newPanes = [];

			for (let i = 0; i < children.length; i++) {
				const child = children.item(i);
				const isPane = child?.classList.contains("splitpanes__pane");
				if (isPane) {
					const pane = panes.find((pane) => pane.element === child);
					if (pane) {
						pane.index = newPanes.length;
						newPanes.push(pane);
					} else {
						console.warn("Splitpanes: Internal error - found a <Pane> elements which isn't tracked.");
					}
				}
			}

			panes = newPanes;
			$veryFirstPaneKey = panes.length > 0 ? panes[0].key : undefined;
		}
	}
</script>

<div
	{id}
	data-viz-sp-id={usedKeyId}
	bind:this={container}
	class={`splitpanes ${theme || ""} ${clazz || ""}`}
	class:splitpanes--horizontal={horizontal}
	class:splitpanes--vertical={!horizontal}
	class:splitpanes--dragging={isMouseDown || isDragging}
	class:splitpanes--freeze={!isAfterInitialTimeoutZero}
	{style}
>
	{#if !browser}
		<GatheringRound>{@render children?.()}</GatheringRound>
	{/if}
	{@render children?.()}
</div>
