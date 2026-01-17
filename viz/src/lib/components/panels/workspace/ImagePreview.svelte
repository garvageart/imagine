<script lang="ts">
	import { selectionManager } from "$lib/states/selection.svelte";
	import { getFullImagePath, type Image } from "$lib/api";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import LabelSelector from "$lib/components/LabelSelector.svelte";
	import { getImageLabel } from "$lib/utils/images";
	import { createImageMenu } from "$lib/context-menu/menus/images";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";

	let activeScope = $derived(selectionManager.activeScope);
	let activeItem = $derived(activeScope?.active as Image | undefined);
	let isImage = $derived(!!activeItem?.image_paths);

	let selectionCount = $derived(activeScope?.selected.size ?? 0);
	let imageSrc = $derived(
		activeItem?.image_paths?.preview
			? getFullImagePath(activeItem.image_paths.preview)
			: null
	);

	let showMenu = $state(false);
	let menuAnchor = $state<{ x: number; y: number } | HTMLElement | null>(null);
	let menuItems = $derived.by(() => {
		if (!activeItem || !activeScope) {
			return [];
		}

		return createImageMenu([activeItem], activeScope, {
			onDelete(deletedUIDs) {
				activeScope?.remove(deletedUIDs);
			}
		});
	});

	function handleContextMenu(e: MouseEvent) {
		if (!activeItem) {
			return;
		}

		e.preventDefault();
		menuAnchor = { x: e.clientX, y: e.clientY };
		showMenu = true;
	}
</script>

<div
	class="preview-container"
	role="presentation"
	oncontextmenu={handleContextMenu}
>
	{#if isImage}
		{#if activeItem && imageSrc}
			<div class="image-wrapper">
				<img src={imageSrc} alt={activeItem.name} loading="lazy" />
			</div>
			<div class="info">
				<span class="filename" title={activeItem.name}>{activeItem.name}</span>
				<span class="meta">
					{activeItem.width}x{activeItem.height} • {activeItem.image_metadata?.file_type?.toUpperCase() ??
						"IMG"}
					<LabelSelector
						label={getImageLabel(activeItem)}
						variant="compact"
						enableSelection={false}
					/>
					{#if activeItem.favourited}
						<MaterialIcon
							iconName="favorite"
							style="font-size: 0.8rem;"
							fill={true}
						/>
					{/if}
				</span>
			</div>
		{:else if selectionCount > 0}
			<div class="placeholder">
				<MaterialIcon
					iconName="photo_library"
					opticalSize={48}
					style="font-size: 4rem; opacity: 0.5;"
				/>
				<span class="text">{selectionCount} items selected</span>
			</div>
		{/if}
	{:else}
		<div class="placeholder">
			<MaterialIcon
				iconName="image"
				opticalSize={48}
				style="font-size: 4rem; opacity: 0.5;"
			/>
			<span class="text">No image(s) selected</span>
		</div>
	{/if}

	<ContextMenu bind:showMenu items={menuItems} anchor={menuAnchor} />
</div>

<style lang="scss">
	.preview-container {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding: 0.5rem;
		color: var(--imag-text-color);
		position: relative;
		box-sizing: border-box;
		justify-content: space-between;
		align-items: stretch;
	}

	.image-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		padding: 0.5rem;
		height: 100%;

		img {
			max-width: 100%;
			max-height: 100%;
			object-fit: contain;
		}
	}

	.info {
		padding: 0.75rem 1rem;
		background-color: var(--imag-bg-color);
		border-top: 1px solid var(--imag-80);
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		max-width: 100%;
		box-sizing: border-box;

		.filename {
			font-size: 0.9rem;
			font-weight: 500;
			color: var(--imag-text-color);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.meta {
			display: flex;
			align-items: center;
			gap: 0.5rem;
			font-size: 0.75rem;
			color: var(--imag-40);
		}
	}

	.placeholder {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: var(--imag-60);
		gap: 0.5rem;

		.text {
			font-size: 0.9rem;
		}
	}
</style>
