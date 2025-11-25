<script lang="ts">
	const variants: Record<string, string> = {
		"200": '<path d="M464.62-464.62H240v-30.76h224.62V-720h30.76v224.62H720v30.76H495.38V-240h-30.76z" />',
		"300": '<path d="M457.31-457.31H220v-45.38h237.31V-740h45.38v237.31H740v45.38H502.69V-220h-45.38z" />',
		"400": '<path d="M450-450H200v-60h250v-250h60v250h250v60H510v250h-60z" />',
		"600": '<path d="M440.39-440.39H185.87v-79.22h254.52V-774.7h79.22v255.09H774.7v79.22H519.61v254.52h-79.22z" />'
	};
	let { size = "1.5em", className = "", title = "IconAdd", viewBox = "0 0 24 24", weight = "400" } = $props();
	const inner = variants[String(weight)] || variants["400"] || Object.values(variants)[0];
</script>

<svg
	class={className}
	width={size}
	height={size}
	viewBox="0 -960 960 960"
	xmlns="http://www.w3.org/2000/svg"
	aria-label={title}
	focusable="false"
>
	{@html inner}
</svg>

<style>
	svg {
		display: inline-block;
		vertical-align: middle;
	}
</style>
