<!-- 
 @component
 A robust Material Icon component that intelligently switches between 
 generated SVG components (for performance and consistency) and 
 font ligatures (as a fallback).
-->
<script lang="ts">
	import type { MaterialSymbol } from "$lib/types/MaterialSymbol";
	import type { SvelteHTMLElements } from "svelte/elements";
	import { registerReady } from "$lib/stores/appReady";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";
	import type { Component } from "svelte";
	import { building, dev } from "$app/environment";

	// Global Font Loading State
	// We keep this global so we don't try to load the same font multiple times.
	const fontLoadMap = new SvelteMap<string, Promise<any>>();
	const warnedMissing = new SvelteSet<string>();

	// Icon Modules (Vite Glob Imports)
	// Eager: Available synchronously during build/prerender.
	const ICON_MODULES_EAGER = import.meta.glob(
		"$lib/components/icons/generated/**/*.svelte",
		{ eager: true }
	);
	// Lazy: Loaded on demand in the browser/dev to save bundle size.
	const ICON_MODULES = import.meta.glob(
		"$lib/components/icons/generated/**/*.svelte"
	);

	// Props
	type IconStyle = "sharp" | "outlined" | "rounded" | "filled";

	export interface IconProps {
		/** The Material Symbol name */
		iconName: MaterialSymbol;
		/** The style variant of the icon */
		iconStyle?: IconStyle;
		/** Fill the icon (1) or outline (0) */
		fill?: boolean;
		/** Font weight / stroke width (100-700) */
		weight?: number;
		/** Grade (-25, 0, 200) - affects thickness */
		grade?: -25 | 0 | 200;
		/** Optical size (20, 24, 40, 48) */
		opticalSize?: 20 | 24 | 40 | 48;
		/** Custom size (e.g. "1rem", "24px") */
		size?: string;
	}

	let {
		iconName,
		iconStyle = "sharp",
		fill = false,
		weight = 400,
		grade = 0,
		opticalSize = 24,
		size = "1.5em",
		...props
	}: IconProps & SvelteHTMLElements["span"] = $props();

	// State
	let GeneratedComponent: Component | null = $state(null);

	// Helpers
	function familyForStyle(style: string) {
		switch (style) {
			case "outlined":
				return "Material Symbols Outlined";
			case "sharp":
				return "Material Symbols Sharp";
			case "rounded":
				return "Material Symbols Rounded";
			case "filled":
				return "Material Symbols Filled";
			default:
				return "Material Icons";
		}
	}

	function normalizeName(n: string) {
		return String(n)
			.replace(/[^a-z0-9]+/gi, " ")
			.trim()
			.split(/\s+/)
			.map((p) => p[0].toUpperCase() + p.slice(1))
			.join("");
	}

	function ensureFontLoaded(family: string) {
		if (typeof document === "undefined" || !("fonts" in document)) {
			return;
		}

		if (!fontLoadMap.has(family)) {
			const p = document.fonts.load(`1em "${family}"`).catch(() => null);
			fontLoadMap.set(family, p);
			registerReady(p);
		}
		return fontLoadMap.get(family);
	}

	async function loadGeneratedIcon(name: string, style: IconStyle) {
		const base = normalizeName(name);
		const styleSuffix = style === 'sharp' ? '' : normalizeName(style);
		const modulePath = `/src/lib/components/icons/generated/Icon${base}${styleSuffix}.svelte`;

		// 1. Try Eager (Build/Prerender)
		if (building && modulePath in ICON_MODULES_EAGER) {
			return (ICON_MODULES_EAGER[modulePath] as any).default;
		}

		// 2. Try Lazy (Runtime/Dev)
		if (modulePath in ICON_MODULES) {
			try {
				const mod = await (ICON_MODULES[modulePath] as any)();
				return mod.default;
			} catch (err) {
				// Silent failure expected for icons that haven't been generated
			}
		}

		return null;
	}

	// Effects
	// Load generated component when iconName or iconStyle changes
	$effect(() => {
		if (!iconName) {
			GeneratedComponent = null;
			return;
		}

		loadGeneratedIcon(iconName, iconStyle).then((comp) => {
			GeneratedComponent = comp;

			// Warn only in dev if missing and not already warned
			if (!comp && dev && !warnedMissing.has(`${iconName}-${iconStyle}`)) {
				// Only warn if we EXPECTED it to be there (i.e. we don't have it in our glob).
				// Actually, if it's not in the glob, it wasn't generated.
				// So we warn that we are falling back to font.
				console.warn(
					`[MaterialIcon] No generated component found for "${iconName}" (${iconStyle}) — falling back to font ligature.`
				);
				warnedMissing.add(`${iconName}-${iconStyle}`);
			}
		});
	});

	// Ensure font is loaded as fallback (or primary if no generated icon)
	$effect(() => {
		if (!GeneratedComponent) {
			ensureFontLoaded(familyForStyle(iconStyle));
		}
	});

	// Styles (for fallback)
	let fontSettings = $derived(
		`'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' ${grade}, 'opsz' ${opticalSize}`
	);

	const fallbackClass = $derived(
		(props.class ? props.class + " " : "") +
			"material-symbols-" +
			iconStyle.toLowerCase()
	);

	const fallbackStyle = $derived(
		`${props.style || ""}; font-variation-settings: ${fontSettings}; ${size ? `font-size: ${size};` : ""}`
	);
</script>

<span 
	class="viz-material-icon"
	style:width={size}
	style:height={size}
	style:flex-shrink="0"
>
	{#if GeneratedComponent}
		<GeneratedComponent
			{...props}
			className={props.class || ""}
			{weight}
			{fill}
			{size}
		/>
	{:else}
		<span {...props} class={fallbackClass} style={fallbackStyle}>
			{iconName}
		</span>
	{/if}
</span>

<style lang="scss">
	.viz-material-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		line-height: 0; /* Prevents line-height from messing with SVG size */

		:global(svg) {
			display: block;
			width: 100%;
			height: 100%;
		}
	}

	/* Fallback Font Styles */
	.material-symbols-sharp,
	.material-symbols-outlined,
	.material-symbols-rounded {
		display: inline-block;
		vertical-align: middle;
		line-height: 1;
		font-variation-settings:
			"FILL" 0,
			"wght" 400,
			"GRAD" 0,
			"opsz" 48;
		font-size: 1.5em;
		min-width: 1em;
		text-align: center;
		user-select: none;
		white-space: nowrap;
		-webkit-font-smoothing: antialiased;
		text-rendering: optimizeLegibility;
		-moz-osx-font-smoothing: grayscale;
		font-feature-settings: "liga";
	}
</style>
