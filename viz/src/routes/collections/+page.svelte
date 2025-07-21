<script lang="ts">
	import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
	import type { PageProps } from "./$types";
	import CollectionCard from "$lib/components/CollectionCard.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import Button from "$lib/components/Button.svelte";
	import { sendAPIRequest } from "$lib/utils/http";
	import type { Collection } from "$lib/types/images";
	import ModalOverlay from "$lib/components/modal/ModalOverlay.svelte";
	import { modal } from "$lib/states/index.svelte";
	import { goto } from "$app/navigation";
	import type { ActionResult } from "@sveltejs/kit";
	import { deserialize } from "$app/forms";
	import SliderToggle from "$lib/components/SliderToggle.svelte";
	import { page } from "$app/state";

	let { data }: PageProps = $props();

	const pagination = $state({
		limit: 10,
		offset: 0
	});
	let shouldUpdate = $derived(data.response.length > pagination.limit * pagination.offset);
	let collectionsData = $derived(
		data.response.slice(0, pagination.limit * (pagination.offset === 0 ? 1 : pagination.offset + 1))
	); // initialize with pagination limit
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
					const result: ActionResult<Record<string, Collection>> = deserialize(await response.json());

					// TODO: fix the shape of data
					if (result.type === "success") {
						goto(`/collections/${result.data?.data.id}`, { state: result.data?.data });
						return;
					}
				}
			}}
		>
			<input id="create_collection-name" name="name" placeholder="Name" type="text" required />
			<!-- svelte-ignore element_invalid_self_closing_tag -->
			<textarea id="create_collection-description" name="description" rows="1" placeholder="Description (optional)" />
			<SliderToggle id="create_collection-private" style="margin-bottom: 1rem;" label="Private" value="off" />
			<Button>
				<input id="create_collection-submit" type="submit" value="Create" />
			</Button>
		</form>
	</div>
</ModalOverlay>
<VizViewContainer
	name="Collections"
	bind:data={collectionsData}
	bind:hasMore={shouldUpdate}
	paginate={() => {
		pagination.offset++;
	}}
	style="height: 100%;"
>
	<div id="viz-collections-toolbar">
		<Button
			id="create-collection"
			onclick={() => {
				modal.show = true;
			}}
		>
			Create
			<MaterialIcon iconName="add" />
		</Button>
	</div>
	<div id="viz-card-container" style="padding: 0em {page.url.pathname === '/' ? '1em' : '3em'};">
		{#each collectionsData as collection}
			<CollectionCard {collection} />
		{/each}
	</div>
</VizViewContainer>

<style lang="scss">
	#viz-collections-toolbar {
		width: 100%;
		max-width: 100%;
		padding: 0.5rem 0rem;
		display: flex;
		justify-content: space-between;
		border-bottom: 1px solid var(--imag-60);
		font-size: 0.8em;
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
		padding: 0.5rem 0rem;
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

	:global(#create-collection) {
		margin: 0em 1rem;
	}

	#viz-card-container {
		margin: 2em 0em;
		display: grid;
		gap: 1em;
		max-width: 100%;
		text-overflow: clip;
		justify-content: center;
		grid-template-columns: repeat(auto-fit, minmax(15em, 1fr));
	}
</style>
