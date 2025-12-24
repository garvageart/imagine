<script lang="ts">
	import { page } from "$app/state";
	import { onDestroy, onMount, type Snippet } from "svelte";
	import LoadingContainer from "../LoadingContainer.svelte";
	import { dev } from "$app/environment";
	import type { SvelteHTMLElements } from "svelte/elements";
	import { isElementScrollable } from "$lib/utils/dom";
	import { isLayoutPage } from "$lib/states/index.svelte";

	interface Props {
		children: Snippet;
		name: string;
		style?: string;
		data?: typeof page.data;
		hasMore?: boolean;
		paginate?: () => void;
		randomLatency?: boolean;
	}

	let {
		children,
		name,
		style,
		hasMore = $bindable(false), // default to false, don't spam unnecessarily
		data = $bindable(),
		paginate,
		randomLatency = $bindable(false),
		...props
	}: SvelteHTMLElements["div"] & Props = $props();

	let viewContainer: HTMLElement | undefined = $state();
	let isLoading = $state(true);
	const initStyle = $derived(`${isLoading ? "height: 100%;" : ""} ${style}`);
	let pageData = $derived.by(() => {
		if (dev && randomLatency) {
			const randomLatency = Math.floor(Math.random() * 2000) + 500; // Random latency between 1 and 3 seconds in dev mode
			return new Promise((resolve) => {
				setTimeout(() => {
					isLoading = false;
					console.log("random latency:", randomLatency, "ms");
					console.log("pageData resolved:", $state.snapshot(data));
					resolve(data);
				}, randomLatency);
			});
		} else {
			return data;
		}
	});

	// Clean up stuff since this component gets mounted and unmounted often, especially during development
	onDestroy(() => {
		pageData = undefined;
	});

	// Scroll handling
	let scrollThreshold = window.innerHeight / 2; // in pixels. idk maybe this setting somewhere else? customisable?
	let isLoadMore = $state(false);

	function loadMore() {
		paginate?.();
	}

	function handleLoadOnMount(element: HTMLDivElement) {
		onMount(() => {
			// When the component is mounted, check if there's more data to show
			// and if so, load more so that the user can scroll down
			// If not, then there's no data to load and no scrollbar is required
			if (!isLoadMore && hasMore && !isElementScrollable(element)) {
				loadMore();
			}
		});
	}

	function onScroll(e: UIEvent) {
		const element = e.target! as HTMLDivElement;
		const offset =
			element.scrollHeight - element.clientHeight - element.scrollTop;

		if (offset <= scrollThreshold) {
			if (!isLoadMore && hasMore) {
				loadMore();
			}
			isLoadMore = true;
		} else {
			isLoadMore = false;
		}
	}
</script>

<svelte:head>
	{#if page.url.pathname !== "/"}
		<title>{name}</title>
	{/if}
</svelte:head>
<div
	{...props}
	class="viz-view-container no-select {props.class}"
	use:handleLoadOnMount
	onscroll={(e) => {
		onScroll(e);
		props.onscroll?.(e);
	}}
	onresize={(e) => {
		onScroll(e);
		props.onresize?.(e);
	}}
	style="{initStyle} {style}"
	data-view-name={name}
	bind:this={viewContainer}
>
	{#if isLayoutPage()}
		{@render children()}
	{:else}
		{#await pageData}
			<LoadingContainer />
		{:then data}
			{#if data}
				{@render children()}
				{#if hasMore}
					<div style="width: 3em; height: 3em; margin: 5em;">
						<LoadingContainer />
					</div>
				{/if}
			{:else}
				<p>No data available</p>
			{/if}
		{:catch error}
			<p>Error loading data: {error.message}</p>
		{/await}
	{/if}
</div>

<style>
	.viz-view-container {
		white-space: wrap;
		display: flex;
		align-items: center;
		flex-direction: column;
		overflow: auto;
		width: 100%;
		max-width: 100%;
		height: 100%;
	}
</style>
