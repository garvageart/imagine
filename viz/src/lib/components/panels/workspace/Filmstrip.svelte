<script lang="ts">
	import { selectionManager } from "$lib/states/selection.svelte";
	import { type Image } from "$lib/api";
	import ImageCard from "$lib/components/ImageCard.svelte";

	let activeScope = $derived(selectionManager.activeScope);
	let activeItem = $derived(activeScope?.active as Image | undefined);

	// activeScope.source contains the list of items
	let filmstripImages = $derived((activeScope?.source as Image[]) ?? []);

	let activeItemIndex = $derived(
		filmstripImages.findIndex((img) => img.uid === activeItem?.uid)
	);

	let selectedItems = $derived(activeScope?.selected ?? new Set<Image>());

	function handleImageClick(image: Image, e: MouseEvent) {
		if (!activeScope) return;

		// Ensure we have focus/active scope (though filmstrip usually reflects active scope)
		// If we wanted to enforce this scope being active, we'd do selectionManager.setActive(...)
		// but filmstrip seems to just reflect whatever is active.

		if (e.shiftKey) {
			activeScope.selected.clear();

			const ids = filmstripImages.map((i) => i.uid);
			let startIndex = 0;
			const endIndex = ids.indexOf(image.uid);

			if (activeScope.active) {
				startIndex = ids.indexOf(activeScope.active.uid);
			}

			// Fallback if active item not in current view/source
			if (startIndex === -1) startIndex = 0;
			if (endIndex === -1) return; // Should not happen if clicked image is in list

			const start = Math.min(startIndex, endIndex);
			const end = Math.max(startIndex, endIndex);

			for (let i = start; i <= end; i++) {
				activeScope.add(filmstripImages[i]);
			}
		} else if (e.ctrlKey || e.metaKey) {
			activeScope.toggle(image);
		} else {
			activeScope.select(image);
		}
	}

	function handleItemKeydown(e: KeyboardEvent, image: Image) {
		if (e.key === "Enter" || e.key === " ") {
			e.preventDefault();
			// For keyboard "Enter/Space", we treat it as a standard click (select)
			// unless we want to support modifiers there too.
			// Passing a mock event or just calling select directly.
			// Usually keyboard selection on Enter is just "select this one".
			activeScope?.select(image);
		} else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
			e.preventDefault();
			e.stopPropagation();
			activeScope?.selectPrevious();
		} else if (e.key === "ArrowRight" || e.key === "ArrowDown") {
			e.preventDefault();
			e.stopPropagation();
			activeScope?.selectNext();
		}
	}

	let containerRef = $state<HTMLElement>();
	let orientation = $state<"horizontal" | "vertical">("horizontal");

	function handleWheel(e: WheelEvent) {
		if (orientation === "horizontal" && e.deltaY !== 0) {
			e.preventDefault();
			if (containerRef) {
				containerRef.scrollLeft += e.deltaY;
			}
		}
	}

	let itemRefs: (HTMLElement | null)[] = $state([]);

	$effect(() => {
		if (containerRef) {
			const observer = new ResizeObserver((entries) => {
				for (const entry of entries) {
					const { width, height } = entry.contentRect;
					orientation = width > height ? "horizontal" : "vertical";
				}
			});
			observer.observe(containerRef);
			return () => observer.disconnect();
		}
	});

	$effect(() => {
		if (activeItemIndex !== -1 && itemRefs[activeItemIndex]) {
			const el = itemRefs[activeItemIndex];
			if (el) {
				el.scrollIntoView({
					behavior: "instant",
					block: "nearest",
					inline: "center"
				});
			}
		}
	});
</script>

<nav
	class="filmstrip-container {orientation}"
	aria-label="Filmstrip"
	onwheel={handleWheel}
	bind:this={containerRef}
>
	{#each filmstripImages as image, i (image.uid)}
		{@const isActive = activeItem?.uid === image.uid}
		{@const isSelected = selectedItems.has(image)}
		<div
			class="filmstrip-item"
			class:active={isActive}
			class:selected={isSelected}
			onclick={(e) => handleImageClick(image, e)}
			onkeydown={(e) => handleItemKeydown(e, image)}
			role="button"
			tabindex="0"
			aria-pressed={isActive}
			aria-label={`Select image ${image.name}`}
			bind:this={itemRefs[i]}
		>
			<ImageCard asset={image} variant="mini" />
		</div>
	{/each}
</nav>

<style lang="scss">
	.filmstrip-container {
		display: flex;
		gap: 0.5rem;
		padding: 0.5rem;
		align-items: center;
		box-sizing: border-box;
		width: 100%;
		height: 100%;

		&:focus {
			outline: none;
		}

		&.horizontal {
			flex-direction: row;
			overflow-x: auto;
			overflow-y: hidden;

			.filmstrip-item {
				height: 100%;
				min-width: 7em;
				max-width: 10em;
			}
		}

		&.vertical {
			flex-direction: column;
			overflow-x: hidden;
			overflow-y: auto;

			.filmstrip-item {
				width: 100%;
				min-height: 7em;
				max-height: 10em;
			}
		}
	}

	.filmstrip-item {
		position: relative;
		display: flex;
		flex-direction: column;
		border-radius: 2px;
		overflow: hidden;
		cursor: pointer;
		background-color: #0d0d0d;
		border: 1px solid var(--imag-80);
		transition: all 0.1s ease;
		flex-shrink: 0;
		box-sizing: border-box;

		&:focus {
			outline: none;
		}

		&:hover {
			background-color: #1a1a1a;
			border-color: var(--imag-70);
		}

		&.active {
			border-color: var(--imag-primary);
			background-color: #1a1a1a;
			outline: 1px solid var(--imag-primary);
			z-index: 1;
		}

		&.selected:not(.active) {
			border-color: var(--imag-secondary);
		}
	}
</style>
