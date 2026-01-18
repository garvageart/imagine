<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M200-200h57l391-391-57-57-391 391zm-80 80v-170l585-583 167 171-582 582zm640-584-56-56zm-141 85-28-29 57 57z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M121-120v-128l616-616 128 128-616 616zm618-577 40-40-41-41-40 40z\" />"
};
    let { size = "1.5em", className = "", title = "IconEdit", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
