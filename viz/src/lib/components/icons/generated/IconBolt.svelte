<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m422-232 207-248H469l29-227-185 267h139zM320-80l40-280H160l360-520h80l-40 320h240L400-80zm151-390\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m320-80 40-280H160l360-520h80l-40 320h240L400-80z\" />"
};
    let { size = "1.5em", className = "", title = "IconBolt", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
