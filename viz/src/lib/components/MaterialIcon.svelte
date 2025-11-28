<!-- 
 @component
 I love an over-engineered solution to a problem that few people experience
  
-->
<script lang="ts">
	import type { MaterialSymbol } from "material-symbols";
	import type { SvelteHTMLElements } from "svelte/elements";
	import { onMount } from "svelte";
	import { registerReady } from "$lib/stores/appReady";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";
	import type { Component } from "svelte";
	import { building, dev } from "$app/environment";

	const fontLoadMap = new SvelteMap<string, Promise<any>>();

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

	function ensureFontLoaded(family: string) {
		if (typeof document === "undefined" || !("fonts" in document)) {
			return null;
		}

		if (!fontLoadMap.has(family)) {
			const p = document.fonts.load(`1em "${family}"`).catch(() => null);
			fontLoadMap.set(family, p);
			registerReady(p);
		}

		return fontLoadMap.get(family);
	}

	// Build-time eager map (Vite) to include generated SVG components into the
	// build output so server-side render / prerender can directly render SVGs.
	// At runtime we keep the lazy `import.meta.glob` fallback for dynamic loading.
	const ICON_MODULES_EAGER: Record<string, any> = import.meta.glob("$lib/components/icons/generated/**/*.svelte", {
		eager: true
	});
	const ICON_MODULES: Record<string, () => Promise<any>> = import.meta.glob("$lib/components/icons/generated/**/*.svelte");
	const warnedMissing = new SvelteSet();

	let GeneratedComponent: Component | null = $state(null);

	function normalizeName(n: string) {
		return String(n)
			.replace(/[^a-z0-9]+/gi, " ")
			.trim()
			.split(/\s+/)
			.map((p) => p[0].toUpperCase() + p.slice(1))
			.join("");
	}

	async function tryLoadGenerated() {
		GeneratedComponent = null;

		const base = normalizeName(String(iconName));
		const styleSuffixPascal = iconStyle ? String(iconStyle)[0].toUpperCase() + String(iconStyle).slice(1) : "";
		const weightSuffix = `W${String(weight).replace(/[^0-9]/g, "")}`;

		// Prefer base component first to avoid many failed imports when only
		// multi-weight/base components exist. From there try weight-specific,
		// then style-specific, then style+weight as a last resort.
		const symbolCandidates = [`Icon${base}`];

		// Module path candidates mirror the symbolCandidates order so we can
		// dynamically import by path if manifest lookup fails.
		const pathCandidates = symbolCandidates.map((sym) => `$lib/components/icons/generated/${sym}.svelte`);

		function anyCandidateExists() {
			for (const p of pathCandidates) {
				if (p in ICON_MODULES_EAGER) {
					return true;
				}

				if (p in ICON_MODULES) {
					return true;
				}
			}
			return false;
		}

		// If we're in the build / prerender phase, prefer the eager map so the
		// generated SVG component is available synchronously and becomes part of
		// the server-rendered output (avoids font-ligature flash on first paint).
		if (building) {
			let matched = null;
			for (const p of pathCandidates) {
				if (p in ICON_MODULES_EAGER) {
					try {
						GeneratedComponent = ICON_MODULES_EAGER[p].default;
						matched = p;
						break;
					} catch (e) {
						// ignore and continue
					}
				}
			}

			if (!GeneratedComponent) {
				// Only warn if there were generated candidates present; if no files
				// exist for this icon at all, the warning is noisy — skip it.
				if (anyCandidateExists()) {
					if (!warnedMissing.has(String(iconName))) {
						console.warn(
							`[MaterialIcon] No generated component found for "${iconName}" (style=${iconStyle}, weight=${weight}) — falling back to font ligature.`
						);
						console.info(`[MaterialIcon] Candidates tried: ${pathCandidates.join(", ")}`);
						warnedMissing.add(String(iconName));
					}
				}
			}

			return;
		}

		if (typeof window === "undefined") {
			return;
		}

		// First try the manifest (HMR-friendly) using symbol names, then
		// fall back to dynamic module imports by path.
		let matched = null;

		if (!GeneratedComponent) {
			for (const p of pathCandidates) {
				if (p in ICON_MODULES) {
					try {
						const mod = await ICON_MODULES[p]();
						GeneratedComponent = mod.default;
						matched = p;
						break;
					} catch (e) {
						// ignore and continue
					}
				}
			}
		}

		if (dev) {
			if (!GeneratedComponent) {
				if (anyCandidateExists()) {
					if (!warnedMissing.has(String(iconName))) {
						console.warn(
							`[MaterialIcon] No generated component found for "${iconName}" (style=${iconStyle}, weight=${weight}) — falling back to font ligature.`
						);
						warnedMissing.add(String(iconName));
					}
				}
			}
		}
	}

	onMount(() => {
		tryLoadGenerated();
		ensureFontLoaded(familyForStyle(iconStyle));
	});

	$effect(() => {
		if (iconName || iconStyle) {
			tryLoadGenerated();
		}
	});

	type IconStyle = "sharp" | "outlined" | "rounded" | "filled";

	interface Props {
		fill?: boolean;
		weight?: number;
		grade?: -25 | 0 | 200;
		opticalSize?: 20 | 24 | 40 | 48;
		iconName: MaterialSymbol;
		iconStyle?: IconStyle;
	}

	let {
		iconName,
		iconStyle = "sharp",
		fill = false,
		weight = 400,
		grade = 0,
		opticalSize = 24,
		...props
	}: Props & SvelteHTMLElements["span"] = $props();
</script>

{#if GeneratedComponent}
	<GeneratedComponent {...props} className={props.class || ""} {weight} />
{:else}
	<span
		{...props}
		class={"material-symbols-" + iconStyle.toLowerCase() + " " + (props.class || "")}
		style={props.style +
			`; font-variation-settings: ${`'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' ${grade}, 'opsz' ${opticalSize}`}`}
		>{iconName}
	</span>
{/if}

<style lang="scss">
	.material-symbols-sharp,
	.material-symbols-outlined,
	.material-symbols-rounded {
		color: var(--imag-text-color);
		fill: var(--imag-text-color);
		padding: 0.1em;
		display: inline-block;
		vertical-align: middle;
		line-height: 1;
		border-radius: 100%;
		font-variation-settings:
			"FILL" 0,
			"wght" 400,
			"GRAD" 0,
			"opsz" 48;
		font-size: 1.5em;
		min-width: 1em;
		text-align: center;
		user-select: none;
	}
</style>
