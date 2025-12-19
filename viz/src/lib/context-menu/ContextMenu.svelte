<!--
	Context Menu (accessible, typed, keyboard navigable)
	Inspiration: https://www.w3.org/WAI/ARIA/apg/patterns/menu/
-->
<script module lang="ts">
	// Exported types must live in the module script so other TS files can import them
	import type { MaterialSymbol } from "material-symbols";

	export type MenuItem = {
		id: string;
		label: string;
		action?: (event: MouseEvent | KeyboardEvent) => void;
		disabled?: boolean;
		danger?: boolean;
		separator?: boolean;
		icon?: MaterialSymbol; // optional icon name/class
		shortcut?: string; // optional keyboard shortcut label
		children?: MenuItem[]; // optional submenu
	};
</script>

<script lang="ts">
	import { tick } from "svelte";
	import ContextMenuItem from "./ContextMenuItem.svelte";
	import type { SvelteHTMLElements } from "svelte/elements";

	type Anchor = { x: number; y: number } | HTMLElement | null;

	let contextMenu: HTMLDivElement | undefined = $state();
	let activeIndex = $state(0);
	// Start off-screen to avoid any layout shift / scrollbars until we compute proper coords
	let position = $state<{ top: number; left: number }>({
		top: -9999,
		left: -9999
	});
	// Default to fixed so initial render doesn't participate in layout
	let cssPosition = $state("fixed");
	let positioned = $state(false);
	// Portal container: we'll move the menu node to document.body to avoid
	// being affected by transformed/overflowing ancestors which can cause
	// page scrollbars. Implemented via a simple Svelte action below.
	let portalTarget: HTMLElement | null = null;

	interface Props {
		// Backward-compatible prop; bind:showMenu still works
		showMenu?: boolean;
		// Explicit items list for the menu
		items?: MenuItem[];
		// Optional anchor (viewport coords or an element)
		anchor?: Anchor;
		// Optional offset adjustments applied after anchor measurement
		offsetX?: number;
		offsetY?: number;
		// Horizontal alignment relative to the anchor element when anchor is an
		// HTMLElement. 'left' aligns menu's left to the anchor's left. 'right'
		// aligns the menu's right to the anchor's right.
		align?: "left" | "right";
		// Reserved for future floating/portal logic
		floating?: boolean;
		// Svelte 5 event callbacks
		onopen?: () => void;
		onclose?: () => void;
		onselect?: (detail: { item: MenuItem; index: number }) => void;
		htmlProps?: SvelteHTMLElements["div"];
		// Debug mode: when true, draw red/blue overlays for anchor and menu
		// and log computed rectangles to the console. Temporary helper.
		debug?: boolean;
	}

	let {
		showMenu = $bindable(false),
		items = $bindable([] as MenuItem[]),
		anchor = $bindable(null),
		offsetX = 0,
		offsetY = 0,
		align = "left",
		onopen,
		onclose,
		onselect,
		htmlProps,
		debug = false
	}: Props = $props();

	function setInitialFocus() {
		// Focus the first enabled item
		const firstEnabled = items.findIndex((i) => !i.disabled && !i.separator);
		activeIndex = firstEnabled === -1 ? 0 : firstEnabled;
		tick().then(() => {
			const current = contextMenu?.querySelector<HTMLButtonElement>(
				`[data-index="${activeIndex}"]`
			);
			current?.focus();
		});
	}

	function computePosition() {
		// If no anchor is provided, render in-place (no coordinates)
		if (!contextMenu) return;

		if (!anchor) {
			position = { top: 0, left: 0 };
			return;
		}

		cssPosition = "fixed";
		const safeMargin = 8;

		// Measure menu size reliably by temporarily placing it at 0,0 hidden
		// so measurements don't depend on previous off-screen coords.
		const prevStyle = {
			left: contextMenu.style.left,
			top: contextMenu.style.top,
			visibility: contextMenu.style.visibility,
			position: contextMenu.style.position
		};
		contextMenu.style.position = "fixed";
		contextMenu.style.left = `0px`;
		contextMenu.style.top = `0px`;
		contextMenu.style.visibility = "hidden";

		const menuRect = contextMenu.getBoundingClientRect();

		let left = safeMargin;
		let top = safeMargin;

		if (anchor instanceof HTMLElement) {
			const rect = anchor.getBoundingClientRect();
			// Horizontal: exact left or exact right alignment
			if (align === "left") {
				left = Math.round(rect.left + offsetX);
			} else {
				left = Math.round(rect.right - menuRect.width - offsetX);
			}

			// Vertical: directly below the anchor
			top = Math.round(rect.bottom + offsetY);

			// Flip above if necessary
			const spaceBelow = window.innerHeight - rect.bottom;
			const spaceAbove = rect.top;
			if (
				menuRect.height + safeMargin > spaceBelow &&
				menuRect.height + safeMargin <= spaceAbove
			) {
				// flip above using same offset magnitude
				top = Math.round(rect.top - menuRect.height - offsetY);
			}
		} else if (anchor && typeof (anchor as any).x === "number") {
			const a = anchor as { x: number; y: number };
			left = Math.round(a.x + offsetX);
			top = Math.round(a.y + offsetY);
		}

		// Clamp into viewport
		const maxLeft = Math.max(
			safeMargin,
			window.innerWidth - menuRect.width - safeMargin
		);
		const maxTop = Math.max(
			safeMargin,
			window.innerHeight - menuRect.height - safeMargin
		);
		left = Math.round(Math.max(safeMargin, Math.min(left, maxLeft)));
		top = Math.round(Math.max(safeMargin, Math.min(top, maxTop)));

		// Restore inline styles (the Svelte binding will apply the real left/top once we set position)
		contextMenu.style.left = prevStyle.left;
		contextMenu.style.top = prevStyle.top;
		contextMenu.style.visibility = prevStyle.visibility;
		contextMenu.style.position = prevStyle.position;

		position = { left, top };
	}

	$effect(() => {
		if (showMenu) {
			// Wait for DOM to render, then compute a safe position and reveal the menu
			tick().then(() => {
				computePosition();
				setInitialFocus();
				positioned = true;
				onopen?.();

				if (debug) {
					// Log rects and computed coords for debugging
					const aRect =
						anchor instanceof HTMLElement
							? anchor.getBoundingClientRect()
							: null;
					const mRect = contextMenu?.getBoundingClientRect();
					console.log("ContextMenu debug:", {
						anchorRect: aRect,
						menuRect: mRect,
						position
					});
					renderDebugOverlays(aRect, mRect);
				}
			});
		} else {
			positioned = false;
			onclose?.();

			if (debug) {
				removeDebugOverlays();
			}
		}
	});

	let debugAnchorEl: HTMLDivElement | null = null;
	let debugMenuEl: HTMLDivElement | null = null;

	function renderDebugOverlays(
		aRect: DOMRect | null,
		mRect: DOMRect | undefined | null
	) {
		removeDebugOverlays();
		if (!aRect || !mRect) return;
		debugAnchorEl = document.createElement("div");
		debugMenuEl = document.createElement("div");

		Object.assign(debugAnchorEl.style, {
			position: "fixed",
			top: `${Math.round(aRect.top)}px`,
			left: `${Math.round(aRect.left)}px`,
			width: `${Math.round(aRect.width)}px`,
			height: `${Math.round(aRect.height)}px`,
			border: "2px dashed red",
			pointerEvents: "none",
			zIndex: "11000"
		});

		Object.assign(debugMenuEl.style, {
			position: "fixed",
			top: `${Math.round(mRect.top)}px`,
			left: `${Math.round(mRect.left)}px`,
			width: `${Math.round(mRect.width)}px`,
			height: `${Math.round(mRect.height)}px`,
			border: "2px dashed blue",
			pointerEvents: "none",
			zIndex: "11000"
		});

		document.body.appendChild(debugAnchorEl);
		document.body.appendChild(debugMenuEl);
	}

	function removeDebugOverlays() {
		if (debugAnchorEl && debugAnchorEl.parentElement)
			debugAnchorEl.parentElement.removeChild(debugAnchorEl);
		if (debugMenuEl && debugMenuEl.parentElement)
			debugMenuEl.parentElement.removeChild(debugMenuEl);
		debugAnchorEl = null;
		debugMenuEl = null;
	}

	// Simple Svelte action to portal a node into document.body. This avoids
	// the menu being clipped or causing scrollbars when inside transformed
	// or overflowed ancestors.
	function portal(node: HTMLElement) {
		portalTarget = document.body;
		portalTarget.appendChild(node);
		return {
			destroy() {
				if (portalTarget && node.parentElement === portalTarget) {
					portalTarget.removeChild(node);
				}
			}
		};
	}

	function isEventInside(target: EventTarget | null): boolean {
		if (!contextMenu) {
			return false;
		}

		if (target instanceof Node) {
			return target === contextMenu || contextMenu.contains(target);
		}

		return false;
	}

	function onWindowPointerDown(e: PointerEvent) {
		if (!showMenu) {
			return;
		}

		const path = (e.composedPath && e.composedPath()) || [];
		if (contextMenu && path.includes(contextMenu)) {
			return;
		}

		showMenu = false;
	}

	function onWindowKeyDown(e: KeyboardEvent) {
		if (!showMenu) {
			return;
		}

		const enabled = items.filter((i) => !i.disabled && !i.separator);
		if (enabled.length === 0) {
			return;
		}

		e.preventDefault();
		switch (e.key) {
			case "Escape":
				showMenu = false;
				break;
			case "ArrowDown": {
				let i = activeIndex;
				for (let step = 0; step < items.length; step++) {
					i = (i + 1) % items.length;
					if (!items[i].disabled && !items[i].separator) {
						activeIndex = i;
						focusActive();
						break;
					}
				}
				break;
			}
			case "ArrowUp": {
				let i = activeIndex;
				for (let step = 0; step < items.length; step++) {
					i = (i - 1 + items.length) % items.length;
					if (!items[i].disabled && !items[i].separator) {
						activeIndex = i;
						focusActive();
						break;
					}
				}
				break;
			}
			case "Home":
				activeIndex = items.findIndex((i) => !i.disabled && !i.separator);
				focusActive();
				break;
			case "End":
				for (let i = items.length - 1; i >= 0; i--) {
					if (!items[i].disabled && !items[i].separator) {
						activeIndex = i;
						break;
					}
				}
				focusActive();
				break;
			case "Enter":
			case " ": // Space
				activate(activeIndex, e);
				break;
		}
	}

	function focusActive() {
		tick().then(() => {
			const el = contextMenu?.querySelector<HTMLButtonElement>(
				`[data-index="${activeIndex}"]`
			);
			el?.focus();
		});
	}

	function activate(index: number, event: MouseEvent | KeyboardEvent) {
		const item = items[index];
		if (!item || item.disabled || item.separator) {
			return;
		}

		item.action?.(event);
		onselect?.({ item, index });
		showMenu = false;
	}
</script>

<svelte:window
	onpointerdown={onWindowPointerDown}
	onkeydown={onWindowKeyDown}
/>

{#if showMenu}
	<div
		{...htmlProps}
		class="context-menu {htmlProps?.class ?? ''}"
		role="menu"
		bind:this={contextMenu}
		use:portal
		style={`position: ${cssPosition || "absolute"}; top:${position.top}px; left:${position.left}px; z-index: 10000; visibility: ${positioned ? "visible" : "hidden"}; ${htmlProps?.style ?? ""}`}
	>
		<div class="context-menu-options">
			<ul role="menu" aria-orientation="vertical">
				{#each items as item, i}
					{#if item.separator}
						<li class="separator" role="separator" aria-hidden="true"></li>
					{:else}
						<ContextMenuItem
							{item}
							index={i}
							active={i === activeIndex}
							onselect={(detail) => activate(i, detail.event)}
						/>
					{/if}
				{/each}
			</ul>
		</div>
	</div>
{/if}

<style>
	.context-menu {
		min-width: 10rem;
		list-style: none;
		box-shadow:
			0 10px 30px rgba(0, 0, 0, 0.35),
			0 2px 8px rgba(0, 0, 0, 0.3);
		border-radius: 0.5rem;
		max-width: calc(100vw - 1em);
		overflow: visible;
		/* Prevent horizontal scrollbar on ancestor containers by avoiding sub-pixel overflow */
		box-sizing: border-box;
	}

	.context-menu-options {
		display: inline-flex;
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		flex-direction: column;
		border-radius: 0.5rem;
		width: 100%;
		max-width: inherit;
	}

	ul {
		margin: 0;
		padding: 0px;
		list-style: none;
	}

	ul > :global(li:first-child > button) {
		border-top-left-radius: 0.5rem;
		border-top-right-radius: 0.5rem;
	}

	ul > :global(li:last-child > button) {
		border-bottom-left-radius: 0.5rem;
		border-bottom-right-radius: 0.5rem;
	}

	.separator {
		height: 1px;
		margin: 4px 6px;
		background: var(--imag-40);
	}
</style>
