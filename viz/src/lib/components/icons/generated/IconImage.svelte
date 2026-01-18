<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M240-280h480L570-480 450-320l-90-120zM120-120v-720h720v720zm80-80h560v-560H200zm0 0v-560z\" />",
    "600": "<path d=\"M228.65-271.91h503.7L578-481.48l-132 171-93-127zM105.87-105.87V-854.7H854.7v748.83zm79.22-79.22h589.82v-589.82H185.09zm0 0v-589.82z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M236-277h489L578-473 446-302l-93-127zM120-120v-720h720v720z\" />",
    "600": "<path d=\"M228.65-271.91h503.7L578-481.48l-132 171-93-127zM105.87-105.87V-854.7H854.7v748.83z\" />"
};
    let { size = "1.5em", className = "", title = "IconImage", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
