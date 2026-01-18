<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m354-287 126-76 126 77-33-144 111-96-146-13-58-136-58 135-146 13 111 97zM233-120l65-281L80-590l288-25 112-265 112 265 288 25-218 189 65 281-247-149zm247-350\" />",
    "600": "<path d=\"m330.95-256.33 149.07-89 149.07 90.03-40.31-168.4L720-538.04l-172.96-14.87L480-711.87l-67.04 158.52L240-539.04l131.43 113.79zM212.09-90.61l70.65-305.3L45.52-601.3l312.64-26.58L480-916.18l121.84 288.3 312.64 26.58-237.22 205.39 71.22 305.3L480-253.17zM480-473.87\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m233-120 65-281L80-590l288-25 112-265 112 265 288 25-218 189 65 281-247-149z\" />",
    "600": "<path d=\"m212.09-90.61 70.65-305.3L45.52-601.3 358.39-628 480-916.18 601.61-628l312.87 26.7-237.22 205.39 71.22 305.3L480-253.17z\" />"
};
    let { size = "1.5em", className = "", title = "IconStar", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
    const activeMap = $derived(fill ? filledVariants : variants);
    const inner = $derived(activeMap[String(weight)] || activeMap["400"] || Object.values(activeMap)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox={viewBox}
    fill="currentColor"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
    {...rest}
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
