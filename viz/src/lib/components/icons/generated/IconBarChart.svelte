<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M640-160v-280h160v280zm-240 0v-640h160v640zm-240 0v-440h160v440z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M660-160v-280h140v280zm-250 0v-640h140v640zm-250 0v-440h140v440z\" />"
};
    let { size = "1.5em", className = "", title = "IconBarChart", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
