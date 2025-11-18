<script lang="ts">
	import type { MaterialSymbol } from "material-symbols";
	import type { SvelteHTMLElements } from "svelte/elements";

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

<span
	{...props}
	class="material-symbols-{iconStyle.toLowerCase()} {props.class}"
	style="{props.style} font-variation-settings: {`'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' ${grade}, 'opsz' ${opticalSize}`};"
	>{iconName}
</span>

<style lang="scss">
	.material-symbols-sharp,
	.material-symbols-outlined,
	.material-symbols-rounded {
		transition: all 150ms linear;
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
		user-select: none;
	}
</style>
