<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-240 240-480l56-56 144 144v-368h80v368l144-144 56 56z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M479-240 238-481l42-43 170 167v-400h60v402l168-168 42 42z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowDownwardAlt", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
