<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-120 300-300l58-58 122 122 122-122 58 58zM358-598l-58-58 180-180 180 180-58 58-122-122z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M480-120 300-300l44-44 136 136 136-136 44 44zM344-612l-44-44 180-180 180 180-44 44-136-136z\" />"
};
    let { size = "1.5em", className = "", title = "IconUnfoldMore", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
