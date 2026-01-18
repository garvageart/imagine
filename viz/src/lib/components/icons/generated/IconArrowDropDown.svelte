<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M480-373.85 303.85-549.38h352.3z\" />",
    "400": "<path d=\"M480-360 280-560h400z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M480-373.85 303.85-549.38h352.3z\" />",
    "400": "<path d=\"M480-360 280-559h400z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowDropDown", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
