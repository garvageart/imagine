<script module lang="ts">
	export { collectionCard };
</script>

<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import type { PageProps } from "./$types";
	import CollectionCard, { openCollection } from "$lib/components/CollectionCard.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import Button from "$lib/components/Button.svelte";
	import { sendAPIRequest } from "$lib/utils/http";
	import type { Collection } from "$lib/types/images";
	import ModalOverlay from "$lib/components/modal/ModalOverlay.svelte";
	import { modal, sort } from "$lib/states/index.svelte";
	import { goto } from "$app/navigation";
	import SliderToggle from "$lib/components/SliderToggle.svelte";
	import { page } from "$app/state";
	import { SvelteSet } from "svelte/reactivity";
	import AssetGrid from "$lib/components/AssetGrid.svelte";
	import type { Content } from "$lib/layouts/subpanel.svelte";
	import { getContext, type ComponentProps } from "svelte";
	import type { AssetGridArray } from "$lib/types/asset";
	import AssetsShell from "$lib/components/AssetsShell.svelte";
	import { sortCollections } from "$lib/sort/sort";

	let { data }: PageProps = $props();

	const pagination = $state({
		limit: 10,
		offset: 0
	});
	let shouldUpdate = $derived(data.response.length > pagination.limit * pagination.offset);

	let displayData = $derived(sortCollections(data.response, sort));

	let fadeOpacity = false;
	let toolbarOpacity = $state(1);

	// Selection
	let selectedAssets = $state<SvelteSet<any>>(new SvelteSet());
	let singleSelectedAsset: any | undefined = $state();

	let collectionGridArray: AssetGridArray<Collection> | undefined = $state();
	let grid: ComponentProps<typeof AssetGrid<Collection>> = $derived({
		assetDblClick: (_, asset) => {
			openCollection(asset, currentPanelContent);
		},
		assetSnippet: collectionCard,
		data: displayData,
		assetGridArray: collectionGridArray,
		selectedAssets,
		singleSelectedAsset
	});

	const currentPanelContent = getContext<Content>("content");
</script>

<ModalOverlay>
	<div id="viz-collection-modal">
		<h1>Create Collection</h1>
		<form
			id="create_collection-form"
			onsubmit={async (event) => {
				event.preventDefault();

				const data = new FormData(event.currentTarget);
				const formObject = Object.fromEntries(data.entries());
				const toggleSwitch = event.currentTarget.querySelector("#create_collection-private")
					?.lastElementChild as HTMLButtonElement;
				formObject["private"] = `${toggleSwitch.getAttribute("data-checked") === "true"}`;

				const response = (await sendAPIRequest<Collection>(
					"/collections",
					{
						method: "POST",
						body: JSON.stringify(formObject)
					},
					true
				)) as Response;

				if (response) {
					const data = (await response.json()) as Collection;

					modal.show = false;
					goto(`/collections/${data.id}`);
				}
			}}
		>
			<input id="create_collection-name" name="name" placeholder="Name" type="text" required />
			<!-- svelte-ignore element_invalid_self_closing_tag -->
			<textarea id="create_collection-description" name="description" rows="1" placeholder="Description (optional)" />
			<SliderToggle id="create_collection-private" style="margin-bottom: 1rem;" label="Private" value="off" />
			<Button style="margin-top: 1rem;">
				<input id="create_collection-submit" type="submit" value="Create" />
			</Button>
		</form>
	</div>
</ModalOverlay>

{#snippet collectionCard(collectionData: Collection)}
	{#if page.url.pathname !== "/"}
		<a
			data-sveltekit-preload-data
			data-asset-id={collectionData.id}
			class="collection-card-link"
			href="/collections/{collectionData.id}"
		>
			<CollectionCard collection={collectionData} />
		</a>
	{:else}
		<CollectionCard style={page.url.pathname === "/" ? "font-size: 0.9rem;" : ""} collection={collectionData} />
	{/if}
{/snippet}

{#snippet toolbarSnippet()}
	<div id="coll-details-toolbar">
		<div id="coll-tools">
			<Button
				id="create-collection"
				class="toolbar-button"
				style="font-size: 0.8rem; background-color: var(--imag-80);"
				title="Create Collection"
				aria-label="Create Collection"
				onclick={() => {
					modal.show = true;
				}}
			>
				Create
				<MaterialIcon iconName="add" />
			</Button>
		</div>
		<span id="coll-details-floating">{data.response.length} {data.response.length === 1 ? "collection" : "collections"}</span>
	</div>
{/snippet}

<VizViewContainer
	bind:data={displayData}
	bind:hasMore={shouldUpdate}
	name="Collections"
	style="padding: 0em {page.url.pathname === '/' ? '1em' : '0em'};"
	paginate={() => {
		pagination.offset++;
	}}
	onscroll={(e) => {
		if (!fadeOpacity) {
			return;
		}

		const assetGrid = document.querySelector(".viz-asset-grid-container")! as HTMLElement;

		const top = Math.max(assetGrid.offsetTop, 1);
		const current = e.currentTarget.scrollTop;
		if (current >= top) {
			toolbarOpacity = 0;
		} else {
			toolbarOpacity = Math.max(0, Math.min(1, 1 - current / top));
		}
	}}
>
	<AssetsShell
		bind:grid
		{pagination}
		{toolbarSnippet}
		toolbarProps={{
			style: `justify-content: space-between;` + (fadeOpacity ? `opacity: ${toolbarOpacity};` : "")
		}}
	/>
</VizViewContainer>

<style lang="scss">
	@use "sass:color";

	.collection-card-link {
		cursor: context-menu;
	}

	:global(.toolbar-button) {
		border-radius: 10em;
		margin: 0.5em 0em;
		font-size: 0.9em;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-text-color);
		cursor: pointer;
	}

	#coll-details-toolbar {
		width: 100%;
		display: flex;
		justify-content: left;
	}

	#coll-details-floating {
		color: var(--imag-40);
		background-color: transparent;
		margin: 0.5rem 0rem;
		font-family: var(--imag-code-font);
	}

	#coll-tools {
		display: flex;
		align-items: center;
		height: 100%;
	}

	form {
		width: 60%;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: flex-start;
		height: 80%;
		max-height: 100%;
		margin-top: 1em;
		position: relative;
	}

	input:not([type="submit"]),
	textarea {
		width: 100%;
		max-width: 100%;
		min-width: 100%;
		min-height: 2rem;
		color: var(--imag-text-color);
		background-color: var(--imag-bg-color);
		outline: none;
		border: none;
		box-shadow: 0 -1.5px 0 var(--imag-60) inset;
		font-size: 2rem;
		font-family: var(--imag-font-family);
		font-weight: bold;
		padding: 0.5rem 1rem;
		margin-bottom: 1em;

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-60);
		}

		&:focus {
			background-color: var(--imag-100);
			box-shadow: 0 -2px 0 var(--imag-primary) inset;
		}

		&:-webkit-autofill,
		&:-webkit-autofill:focus {
			-webkit-text-fill-color: var(--imag-text-color);
			-webkit-box-shadow: 0 0 0px 1000px var(--imag-100) inset;
			-webkit-box-shadow: 0 -5px 0 var(--imag-primary) inset;
			transition:
				background-color 0s 600000s,
				color 0s 600000s !important;
		}
	}

	#create_collection-description {
		font-size: 1.2rem;
		resize: none;
		font-weight: 400;
	}

	#create_collection-submit {
		border: inherit;
		background-color: transparent;
		color: inherit;
		font-family: inherit;
		font-weight: bold;
		font-size: inherit;
		cursor: pointer;
		width: 100%;
		height: 100%;
	}

	#viz-collection-modal {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: center;
	}
</style>
