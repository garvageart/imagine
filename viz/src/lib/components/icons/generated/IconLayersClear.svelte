<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m650-415-42-42 134-105-262-204-102 79-43-43 145-112 360 280zm86 86-43-43 97-75 50 37zm94 264L641-255 480-130 120-410l50-37 310 241 118-92-43-43-75 59-360-280 120-94L73-822l43-43 757 757zM493-572\" />"
};
    let { size = "1.5em", className = "", title = "IconLayersClear", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = $derived(variants[String(weight)] || variants["400"] || Object.values(variants)[0]);
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
