<script lang="ts">
	import { scale } from "svelte/transition";
	import MaterialIcon from "./MaterialIcon.svelte";
	
	let {
		scrollTop = $bindable(0),
		totalHeight = 0,
		viewportHeight = 0,
		dateLabel = "",
		isDragging = $bindable(false)
	} = $props();

	let isHovering = $state(false);
	let trackEl: HTMLDivElement | undefined = $state();

	// Logic for a fixed-size knob scrubber (Google Photos style)
	// The knob travels along a track of length: viewportHeight - margins
	// We use a safe margin to ensure the knob doesn't overlap header/footer excessively if needed
	const marginY = 8;

	let scrollableHeight = $derived(Math.max(0, totalHeight - viewportHeight));

	// Only show if there is enough content to scroll
	let isVisible = $derived(scrollableHeight > 500 && viewportHeight > 300);

	// The visual track length available for the center of the thumb
	// 40px is approx thumb height/hitbox
	let trackLength = $derived(Math.max(0, viewportHeight - marginY * 2 - 40));

	// Calculate thumb position (center Y relative to viewport top) based on scrollTop
	let thumbOffsetY = $derived.by(() => {
		if (scrollableHeight <= 0) return marginY + 20;
		const ratio = scrollTop / scrollableHeight;
		const clampedRatio = Math.max(0, Math.min(1, ratio));
		// Start at margin + half thumb height
		return marginY + 20 + clampedRatio * trackLength;
	});

	function handleTrackClick(e: MouseEvent) {
		if (!isVisible || !trackEl) return;

		// Ignore if clicked on the thumb itself (handled by pointerdown)
		if ((e.target as HTMLElement).closest(".scrubber-thumb")) return;

		const rect = trackEl.getBoundingClientRect();
		// Relative Y within the track container
		const relativeY = e.clientY - rect.top;

		// Calculate target ratio
		// relativeY = marginY + 20 + (ratio * trackLength)
		// ratio = (relativeY - marginY - 20) / trackLength
		let ratio = (relativeY - marginY - 20) / trackLength;
		ratio = Math.max(0, Math.min(1, ratio));

		isDragging = true;
		scrollTop = ratio * scrollableHeight;

		// Small delay to allow sync to complete before releasing
		setTimeout(() => {
			isDragging = false;
		}, 20);
	}

	function handleThumbDown(e: PointerEvent) {
		if (!isVisible) {
			return;
		}

		e.preventDefault();
		isDragging = true;

		const target = e.currentTarget as HTMLElement;
		target.setPointerCapture?.(e.pointerId);

		const startY = e.clientY;
		const startScrollTop = scrollTop;

		const handlePointerMove = (ev: PointerEvent) => {
			const deltaY = ev.clientY - startY;

			// Map pixel delta to scroll delta
			// deltaRatio = deltaY / trackLength
			// deltaScroll = deltaRatio * scrollableHeight
			if (trackLength <= 0) {
				return;
			}

			const deltaScroll = (deltaY / trackLength) * scrollableHeight;
			scrollTop = Math.max(
				0,
				Math.min(scrollableHeight, startScrollTop + deltaScroll)
			);
		};

		const handlePointerUp = (ev: PointerEvent) => {
			isDragging = false;
			target.releasePointerCapture?.(ev.pointerId);
			window.removeEventListener("pointermove", handlePointerMove);
			window.removeEventListener("pointerup", handlePointerUp);
		};

		window.addEventListener("pointermove", handlePointerMove);
		window.addEventListener("pointerup", handlePointerUp);
	}
</script>

{#if isVisible}
	<!-- 
		The track container fills the scrubber area.
		We use it to capture clicks on the "track".
	-->
	<div
		class="timeline-scrubber"
		role="scrollbar"
		aria-controls="photo-grid"
		aria-valuenow={scrollTop}
		aria-orientation="vertical"
		bind:this={trackEl}
		onmouseenter={() => (isHovering = true)}
		onmouseleave={() => (isHovering = false)}
		onclick={handleTrackClick}
		onkeydown={() => {}}
		tabindex="-1"
	>
		<!-- Date Bubble -->
		<!-- Positioned relative to the thumb center -->
		{#if (isDragging || isHovering) && dateLabel}
			<div
				class="scrubber-bubble"
				transition:scale={{ duration: 150, start: 0.8, opacity: 0 }}
				style="top: {thumbOffsetY}px;"
			>
				<span class="bubble-text">{dateLabel}</span>
			</div>
		{/if}

		<!-- Thumb / Knob -->
		<div
			class="scrubber-thumb"
			class:active={isDragging || isHovering}
			style="transform: translateY({thumbOffsetY}px) translateY(-50%);"
			onpointerdown={handleThumbDown}
			role="button"
			tabindex="-1"
		>
			<div class="thumb-inner">
				<MaterialIcon
					iconName="unfold_more"
					iconStyle="rounded"
					class="thumb-icon"
				/>
			</div>
		</div>
	</div>
{/if}

<style lang="scss">
	.timeline-scrubber {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		width: 3rem; /* Hit area width */
		z-index: 100;
		user-select: none;
		touch-action: none;
		pointer-events: auto;
	}

	.scrubber-thumb {
		position: absolute;
		right: 0.25rem;
		width: 3rem;
		height: 3rem;
		cursor: grab;
		display: flex;
		align-items: center;
		justify-content: center;

		&:active {
			cursor: grabbing;
		}
	}

	.thumb-inner {
		width: 3rem;
		height: 3rem;
		background-color: var(--imag-secondary);
		color: var(--imag-10-dark);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow:
			0 1px 3px rgba(0, 0, 0, 0.2),
			0 4px 8px rgba(0, 0, 0, 0.15);
		transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
		border: 1px solid rgba(255, 255, 255, 0.1);

		:global(.thumb-icon) {
			font-size: 1.25rem;
		}
	}

	.scrubber-thumb:hover .thumb-inner,
	.scrubber-thumb.active .thumb-inner {
		width: 3rem;
		height: 3rem;
		box-shadow:
			0 2px 6px rgba(0, 0, 0, 0.25),
			0 8px 16px rgba(0, 0, 0, 0.2);

		:global(.thumb-icon) {
			opacity: 1;
			transform: scale(1.1);
		}
	}

	.scrubber-bubble {
		position: absolute;
		right: 3.5rem;
		top: 0; /* Set via inline style */
		transform: translateY(-50%);
		background-color: var(--imag-10);
		color: var(--imag-100);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
		padding: 0.5rem 1rem;
		border-radius: 0.5rem;
		font-size: 0.95rem;
		font-weight: 600;
		white-space: nowrap;
		pointer-events: none;
		z-index: 101;
		// border: 1px solid rgba(255, 255, 255, 0.08);
		transform-origin: right center;
	}
</style>
