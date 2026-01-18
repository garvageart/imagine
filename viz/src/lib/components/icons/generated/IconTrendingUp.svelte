<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m136-240-56-56 296-298 160 160 208-206H640v-80h240v240h-80v-104L536-320 376-480z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m123-240-43-43 292-291 167 167 241-241H653v-60h227v227h-59v-123L538-321 371-488z\" />"
};
    let { size = "1.5em", className = "", title = "IconTrendingUp", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
