<script lang="ts">
	import type { MaterialSymbol } from "material-symbols";
	import type { SvelteHTMLElements } from "svelte/elements";

	type IconStyle = "sharp" | "outlined" | "rounded";

	interface Props {
		fill?: boolean;
		weight?: number;
		grade?: -25 | 0 | 200;
		opticalSize?: 20 | 24 | 40 | 48;
		iconName: MaterialSymbol;
		asSVG?: boolean;
		iconStyle?: IconStyle;
	}

	let {
		iconName,
		iconStyle = "sharp",
		asSVG = false,
		fill = false,
		weight = 400,
		grade = 0,
		opticalSize = 24,
		...props
	}: Props & SvelteHTMLElements["span"] = $props();

	let svgIcon: HTMLSpanElement | undefined = $state();
	let externalSVG = $state({
		enabled: false,
		src: ""
	});

	if (asSVG) {
		$effect(() => {
			if (!svgIcon) {
				return;
			}

			try {
				const fields = import.meta.glob([`/node_modules/@material-design-icons/svg/*/*.svg`]);
				const icon = fields[`/node_modules/@material-design-icons/svg/${iconStyle}/${iconName}.svg`]();

				const iconMod = icon.then(async (mod) => {
					const iconModule = mod as typeof import("*.svg");
					let iconSvgString = iconModule.default;
					return iconSvgString;
				});

				iconMod.then((str) => {
					svgIcon!.innerHTML = decodeURIComponent(str.replace("data:image/svg+xml,", ""));
				});

				return () => {
					svgIcon!.remove();
				};
			} catch (error) {
				// Default to Google GitHub hosting
				externalSVG.enabled = true;
				const options = [
					weight == 400 ? undefined : `wght${weight}`,
					grade == 0 ? undefined : `grad${grade}`,
					fill == false ? undefined : `fill1`
				].join("");

				const ary = [iconName, options.length > 0 ? options : "", `${opticalSize}px`].filter((x) => x.length > 0);
				const filename = ary.join("_");

				externalSVG.src = `https://raw.githubusercontent.com/google/material-design-icons/master/symbols/web/${iconName}/materialsymbols${iconStyle}/${filename}.svg`;
			}
		});
	}
</script>

{#if asSVG}
	{#if externalSVG.enabled}
		<img {...props} src={externalSVG.src} alt={iconName} class="material-symbols-{iconStyle.toLowerCase()} {props.class}" />
	{:else}
		<span {...props} bind:this={svgIcon} class="material-symbols-{iconStyle.toLowerCase()} {props.class}"></span>
	{/if}
{:else}
	<span
		{...props}
		class="material-symbols-{iconStyle.toLowerCase()} {props.class}"
		style="{props.style} font-variation-settings: {`'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' ${grade}, 'opsz' ${opticalSize}`};"
		>{iconName}
	</span>
{/if}

<style lang="scss">
	.material-symbols-sharp,
	.material-symbols-outlined,
	.material-symbols-rounded {
		transition: all 150ms linear;
		padding: 0.1em;
		display: inline-block;
		border-radius: 100%;
		font-variation-settings:
			"FILL" 0,
			"wght" 400,
			"GRAD" 0,
			"opsz" 48;
		font-size: 1.5em;
		user-select: none;
	}
</style>
