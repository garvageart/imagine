<script lang="ts">
	import { getContext, onMount, onDestroy, hasContext } from "svelte";
	import { KEY } from "./Splitpanes.svelte";
	import type {
		ClientCallbacks,
		IPane,
		PaneInitFunction,
		SplitContext
	} from "./index.js";
	import { browser } from "./internal/env.js";
	import { gatheringKey } from "./internal/GatheringRound.svelte";
	import { getDimensionName } from "./internal/utils/sizing.js";
	import { carefullCallbackSource } from "./internal/utils/functions";
	import { generateRandomString } from "$lib/utils/misc";

	const {
		ssrRegisterPaneSize,
		onPaneInit,
		clientOnly: clientOnlyContext,
		isHorizontal,
		showFirstSplitter,
		veryFirstPaneKey,
		keyId
	} = getContext<SplitContext>(KEY);

	interface Props {
		size?: number | null;
		minSize?: number;
		maxSize?: number;
		snapSize?: number;
		class?: string;
		smoothExpand?: boolean;
		id?: string;
		paneKeyId?: string;
		children?: import("svelte").Snippet;
	}

	let {
		size = $bindable(null),
		minSize = $bindable(0),
		maxSize = $bindable(100),
		snapSize = $bindable(0),
		class: clazz = "",
		smoothExpand = false,
		id,
		paneKeyId = "",
		children
	}: Props = $props();

	// VARIABLES
	let usedKeyId = $derived(paneKeyId ?? generateRandomString(10));
	// svelte-ignore state_referenced_locally
	let usedId = id;
	let isActive = $state(false);

	// VARIABLES
	const key = {};
	const gathering = !browser && hasContext(gatheringKey);
	const { undefinedPaneInitSize } = (
		!gathering ? onPaneInit(key) : {}
	) as ReturnType<PaneInitFunction>;

	let element: HTMLElement | undefined = $state();
	let sz: number = $state(size ?? undefinedPaneInitSize);
	let isSplitterActive = $state(false);

	// CALLBACKS

	let clientCallbacks: ClientCallbacks | undefined = undefined;

	/**
	 * This is an object of callbacks that are safe to be called on browser even when the object `clientCallbacks`
	 *  isn't initialized yet (i.e. before `onPaneAdd()`).
	 *
	 * In the case of the object isn't initialized yet, calling this callbacks will do nothing.
	 */

	const carefullClientCallbacks = browser
		? carefullCallbackSource(() => clientCallbacks)
		: undefined;

	const reportGivenSizeChangeSafe = (size: number) => {
		// We put an extra check of `size != sz` here and not in the reactive statement, since we don't want a change
		//  of `sz` to trigger report.
		if (clientCallbacks && size != sz) {
			carefullClientCallbacks?.("reportGivenSizeChange")(size);
		}
	};

	// REACTIVE
	$effect(() => {
		if (browser && typeof size === "number") {
			reportGivenSizeChangeSafe(size);
		}
	});

	let transitionClass = $derived(
		smoothExpand
			? `transition: ${$isHorizontal ? "height" : "width"} 0.2s ease-out;`
			: ""
	);
	let dimension = $derived(getDimensionName($isHorizontal));
	let style = $derived(`${dimension}: ${sz}%; ${transitionClass}`);

	if (gathering && ssrRegisterPaneSize) {
		ssrRegisterPaneSize(size);
	} else if (browser) {
		onMount(() => {
			const inst: IPane = {
				index: 0,
				key,
				element: element!,
				childs: [],
				parent: keyId,
				id: usedId,
				keyId: usedKeyId,
				givenSize: size,
				sz: () => sz,
				setSz: (v) => {
					sz = v;
					if (typeof size === "number" && size !== sz) {
						size = sz;
					}
				},
				min: () => minSize,
				max: () => maxSize,
				snap: () => snapSize,
				setSplitterActive: (isActive: boolean) => {
					isSplitterActive = isActive;
				},
				isReady: false,
				isActive
			};

			clientCallbacks = clientOnlyContext?.onPaneAdd(inst);
		});

		onDestroy(() => {
			clientOnlyContext?.onPaneRemove(key);
		});
	}
</script>

<svelte:document
	on:click={(event) => {
		const target = event.target as HTMLElement;

		if (!element) {
			return;
		}

		if (!element.contains(target)) {
			isActive = false;
		}
	}}
/>

{#if !gathering}
	<!-- Splitter -->
	{#if $veryFirstPaneKey !== key || $showFirstSplitter}
		<!-- this a11y issue is known, and will be taken care of as part of the a11y feature issue in #11 -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			role="separator"
			class="splitpanes__splitter {isSplitterActive
				? 'splitpanes__splitter__active'
				: ''}"
			onmousedown={carefullClientCallbacks?.("onSplitterDown")}
			ontouchstart={carefullClientCallbacks?.("onSplitterDown")}
			onclick={carefullClientCallbacks?.("onSplitterClick")}
			ondblclick={carefullClientCallbacks?.("onSplitterDblClick")}
		></div>
	{/if}

	<!-- Pane -->
	<!-- this a11y issue is known, and will be taken care of as part of the a11y feature issue in #11 -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		id={usedId}
		data-viz-sp-id={usedKeyId}
		class={`splitpanes__pane ${clazz || ""}`}
		bind:this={element}
		onclick={(event) => {
			carefullClientCallbacks?.("onPaneClick")(event);
			isActive = true;
		}}
		{style}
	>
		{@render children?.()}
	</div>
{/if}

<style global lang="scss">
	.transition-vertical {
		transition: width 0.2s ease-out;
	}

	.transition-horizontal {
		transition: height 0.2s ease-out;
	}
</style>
