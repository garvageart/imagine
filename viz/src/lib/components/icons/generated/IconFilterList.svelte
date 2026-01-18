<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M400-240v-80h160v80zM240-440v-80h480v80zM120-640v-80h720v80z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M400-240v-60h160v60zM240-450v-60h480v60zM120-660v-60h720v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconFilterList", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
