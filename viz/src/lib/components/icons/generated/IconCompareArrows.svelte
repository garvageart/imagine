<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m320-160-56-57 103-103H80v-80h287L264-503l56-57 200 200zm320-240L440-600l200-200 56 57-103 103h287v80H593l103 103z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m317-160-42-42 121-121H80v-60h316L275-504l42-42 193 193zm326-254L450-607l193-193 42 42-121 121h316v60H564l121 121z\" />"
};
    let { size = "1.5em", className = "", title = "IconCompareArrows", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
