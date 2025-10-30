<script module>
    export { searchForData };

    function searchForData(searchValue: string, images: Image[]) {
        if (searchValue.trim() === "") {
            return [];
        }
        // eventually this should also look through keywords/tags
        // and labels idk. fuzzy search???
        return images.filter((i) =>
            i.name.toLowerCase().includes(searchValue.toLowerCase()),
        );
    }
</script>

<script lang="ts">
    import { invalidateAll } from "$app/navigation";
    import { page } from "$app/state";
    import AssetGrid from "$lib/components/AssetGrid.svelte";
    import AssetsShell from "$lib/components/AssetsShell.svelte";
    import Lightbox from "$lib/components/Lightbox.svelte";
    import LoadingContainer from "$lib/components/LoadingContainer.svelte";
    import VizViewContainer from "$lib/components/panels/VizViewContainer.svelte";
    import SearchInput from "$lib/components/SearchInput.svelte";
    import { lightbox, modal, sort } from "$lib/states/index.svelte";
    import type { AssetGridArray } from "$lib/types/asset.js";
    import {
        SUPPORTED_IMAGE_TYPES,
        SUPPORTED_RAW_FILES,
        type SupportedImageTypes,
    } from "$lib/types/images";
    import { blurOnEsc, loadImage } from "$lib/utils/dom.js";
    import hotkeys from "hotkeys-js";
    import { DateTime } from "luxon";
    import { SvelteSet } from "svelte/reactivity";
    import { onMount, type ComponentProps } from "svelte";
    import { sortCollectionImages } from "$lib/sort/sort.js";
    import ImageCard from "$lib/components/ImageCard.svelte";
    import Button from "$lib/components/Button.svelte";
    import MaterialIcon from "$lib/components/MaterialIcon.svelte";
    import UploadManager from "$lib/upload/manager.svelte.js";
    import {
        addCollectionImages,
        getFullImagePath,
        updateCollection,
        type CollectionUpdate,
        type Image,
    } from "$lib/api";
    import { thumbHashToDataURL } from "thumbhash";
    import { fade } from "svelte/transition";
    import { toastState } from "$lib/toast-notifcations/notif-state.svelte.js";
    import CollectionModal from "$lib/components/CollectionModal.svelte";
    import InputText from "$lib/components/dom/InputText.svelte";
    import { load } from "./+page";

    let { data } = $props();
    // Keyboard events
    const permittedKeys: string[] = [];
    const selectKeys = ["Enter", "Space", " "];
    const moveKeys = ["ArrowRight", "ArrowLeft", "ArrowUp", "ArrowDown"];
    permittedKeys.push(...selectKeys, ...moveKeys);

    // Data
    let loadedData = $derived(data);
    // Track local edits separately to avoid clobbering them on refresh
    let localDataUpdates = $derived({
        name: loadedData.name,
        description: loadedData.description,
        private: loadedData.private as boolean | undefined,
    });

    let loadedImages = $derived(
        loadedData.images?.items?.map((img) => img.image) ?? [],
    );

    // Lightbox
    let lightboxImage: Image | undefined = $state();
    let currentImageEl: HTMLImageElement | undefined = $derived(
        lightboxImage ? document.createElement("img") : undefined,
    );

    $effect(() => {
        if (lightboxImage) {
            lightbox.show = true;
        }
    });

    // Search stuff
    let searchValue = $state("");
    let searchData = $derived(searchForData(searchValue, loadedImages));

    // Pagination
    // NOTE: This might be moved to a settings thing and this could just be default
    const pagination = $derived({
        limit: loadedData.images.limit ?? 25,
        offset: loadedData.images.offset ?? 0,
    });

    // the searchValue hides the loading indicator when searching since we're
    // already searching through *all* the data that is available on the client
    let shouldUpdate = $derived(
        loadedImages.length > pagination.limit * pagination.offset &&
            searchValue.trim() === "",
    );

    // Selection
    let selectedAssets = $state<SvelteSet<Image>>(new SvelteSet());
    let singleSelectedAsset: Image | undefined = $state();

    let imageGridArray: AssetGridArray<Image> | undefined = $state();

    // Toolbar stuff
    let toolbarOpacity = $state(0);

    // Thumbhash placeholder
    let placeholderDataURL = $derived.by(() => {
        if (lightboxImage?.image_metadata?.thumbhash) {
            try {
                const binaryString = atob(
                    lightboxImage.image_metadata.thumbhash,
                );
                const bytes = new Uint8Array(binaryString.length);
                for (let i = 0; i < binaryString.length; i++) {
                    bytes[i] = binaryString.charCodeAt(i);
                }
                return thumbHashToDataURL(bytes);
            } catch (error) {
                console.warn("Failed to decode thumbhash:", error);
            }
        }
    });

    // Display Data
    let displayData = $derived(
        searchValue.trim()
            ? sortCollectionImages(searchData, sort)
            : sortCollectionImages(
                  loadedData.images?.items?.map((img) => img.image) ?? [],
                  sort,
              ),
    );

    // Grid props
    let grid: ComponentProps<typeof AssetGrid<Image>> = $derived({
        assetSnippet: imageCard,
        assetGridArray: imageGridArray,
        selectedAssets,
        singleSelectedAsset,
        data: displayData,
        assetDblClick: (_, asset) => {
            lightboxImage = asset;
        },
    });

    async function handleCollectionUpload() {
        // allowed image types will come from the config but for now just hardcode
        const controller = new UploadManager([
            ...SUPPORTED_RAW_FILES,
            ...SUPPORTED_IMAGE_TYPES,
        ] as SupportedImageTypes[]);
        controller.openFileHolder();
        const uploadedImages = await controller.uploadImage();

        const response = await addCollectionImages(loadedData.uid, {
            uids: uploadedImages.map((img) => img.uid),
        });

        if (response.data.added) {
            toastState.addToast({
                message: `Added photos to collection`,
                type: "success",
                timeout: 3000,
            });

            await invalidateAll();
        }
    }

    async function updateCollectionDetails(data?: CollectionUpdate) {
        const response = await updateCollection(
            loadedData.uid,
            data ?? {
                ...localDataUpdates,
            },
        );

        modal.show = false;
        if (response.status !== 200) {
            toastState.addToast({
                type: "error",
                message: `Failed to update collection: ${response.data || "Unknown error"}`,
            });

            return;
        }

        loadedData = {
            ...loadedData,
            updated_at: response.data.updated_at,
            description: response.data.description,
            name: response.data.name,
            private: response.data.private,
            image_count: response.data.image_count,
            created_by: response.data.created_by,
        };

        toastState.addToast({
            type: "success",
            message: `Succesfully updated collection ${response.data.name}`,
        });
    }

    hotkeys("esc", (e) => {
        lightboxImage = undefined;
    });
</script>

<CollectionModal
    bind:data={localDataUpdates}
    heading="Edit Collection"
    buttonText="Save"
    modalAction={() => {
        updateCollectionDetails();
    }}
/>

{#if lightboxImage}
    {@const imageToLoad =
        getFullImagePath(lightboxImage.image_paths?.preview) ?? ""}
    <Lightbox
        onclick={() => {
            lightboxImage = undefined;
        }}
    >
        <!-- Awaitng like this is better inline but `currentImageEl` is kinda
	 being created/allocated unncessarily and is never removed or freed until the component is destroyed
	 It's small but annoying enough where I want to find a different way to load an image
	  -->
        {#await loadImage(imageToLoad, currentImageEl!)}
            <img
                src={placeholderDataURL}
                class="lightbox-image"
                style="height: 90%; position: absolute;"
                out:fade={{ duration: 300 }}
                alt="Placeholder image for {lightboxImage.name}"
                aria-hidden="true"
            />
        {:then _}
            <img
                src={imageToLoad}
                class="lightbox-image"
                in:fade={{ duration: 300 }}
                alt={lightboxImage.name}
                title={lightboxImage.name}
                loading="eager"
                data-image-uid={lightboxImage.uid}
            />
        {:catch error}
            <p>Failed to load image</p>
            <p>{error}</p>
        {/await}
    </Lightbox>
{/if}

{#snippet imageCard(asset: Image)}
    <ImageCard {asset} />
{/snippet}

{#snippet searchInputSnippet()}
    <SearchInput style="margin: 0em 1em;" bind:value={searchValue} />
    <div id="coll-tools">
        <Button
            id="upload_to_collection"
            class="toolbar-button"
            style="font-size: 0.8rem; background-color: var(--imag-100);"
            title="Upload to Collection"
            aria-label="Upload to Collection"
            onclick={() => {
                handleCollectionUpload();
            }}
        >
            Upload
            <MaterialIcon iconName="upload" />
        </Button>
        <Button
            id="upload_to_collection"
            class="toolbar-button"
            style="font-size: 0.8rem; background-color: var(--imag-100);"
            title="Edit Collection"
            aria-label="Edit Collection"
            onclick={() => {
                modal.show = true;
            }}
        >
            Edit
            <MaterialIcon iconName="edit" />
        </Button>
    </div>
{/snippet}

{#snippet noAssetsSnippet()}
    <div id="add_to_collection-container">
        <span style="margin: 1em; color: var(--imag-20); font-size: 1.2rem;"
            >Add images to this collection</span
        >
        <Button
            id="add_to_collection-button"
            style="padding: 2em 8em; display: flex; align-items: center; justify-content: center;"
            title="Select Photos"
            aria-label="Select Photos"
            onclick={async () => handleCollectionUpload()}
        >
            Select Photos
            <MaterialIcon iconName="add" style="font-size: 2em;" />
        </Button>
    </div>
{/snippet}

<VizViewContainer
    bind:data={displayData}
    bind:hasMore={shouldUpdate}
    name="{localDataUpdates.name} - Collection"
    style="font-size: {page.url.pathname === '/' ? '0.9em' : 'inherit'};"
    paginate={() => {
        pagination.offset++;
    }}
    onscroll={(e) => {
        const info = document.getElementById("viz-info-container")!;
        const bottom = info.scrollHeight;

        if (e.currentTarget.scrollTop < bottom) {
            toolbarOpacity = e.currentTarget.scrollTop / bottom;
        } else {
            toolbarOpacity = 1;
        }
    }}
>
    <AssetsShell
        bind:grid
        {pagination}
        {noAssetsSnippet}
        toolbarSnippet={searchInputSnippet}
        toolbarProps={{
            style: "justify-content: center; ",
        }}
    >
        <div id="viz-info-container">
            <div id="coll-metadata">
                <span id="coll-name">
                    <InputText
                        id="coll-name-input"
                        style="padding: 0% 0.5rem;"
                        oninput={(e) =>
                            (localDataUpdates.name = e.currentTarget.value)}
                        onblur={() => {
                            if (localDataUpdates.name !== loadedData.name) {
                                updateCollectionDetails({
                                    name: localDataUpdates.name,
                                });
                            }
                        }}
                    ></InputText>
                </span>
                <span
                    id="coll-details"
                    style="padding: 0% 0.5rem;"
                    title="Updated at: {DateTime.fromJSDate(
                        new Date(loadedData.updated_at),
                    ).toFormat('dd.MM.yyyy HH:mm')}"
                    >{DateTime.fromJSDate(
                        new Date(loadedData.created_at),
                    ).toFormat("dd.MM.yyyy")}
                    •
                    {#if searchValue.trim()}
                        {searchData.length}
                        {searchData.length === 1 ? "image" : "images"} of {loadedData.image_count}
                    {:else}
                        {loadedData.image_count}
                        {loadedData.image_count === 1 ? "image" : "images"}
                    {/if}
                </span>
            </div>
        </div></AssetsShell
    >
</VizViewContainer>

<style lang="scss">
    :global(#create-collection) {
        margin: 0em 1rem;
    }

    #add_to_collection-container {
        display: flex;
        flex-direction: column;
        justify-content: left;
    }

    #viz-info-container {
        width: 100%;
        max-width: 100%;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        margin: 1em 0em;
    }

    #coll-name {
        color: var(--imag-20);
        font-weight: bold;
    }

    #coll-metadata {
        padding: 0.5rem 2rem;
        display: flex;
        flex-direction: column;
        width: 25%;
        overflow: hidden;
        color: var(--imag-60);
        font-family: var(--imag-code-font);
    }

    #coll-tools {
        display: flex;
        align-items: center;
        height: 100%;
        position: absolute;
        right: 2rem;
    }

    :global(.lightbox-image) {
        max-width: 80%;
        max-height: 90%;
    }
</style>
