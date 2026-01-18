<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h720v720zm80-80h560v-560H200zm80-80h80v-200h-80zm320 0h80v-400h-80zm-160 0h80v-120h-80zm0-200h80v-80h-80zM200-200v-560z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h720v720zm164-157h60v-205h-60zm332 0h60v-420h-60zm-166 0h60v-118h-60zm0-205h60v-60h-60z\" />"
};
    let { size = "1.5em", className = "", title = "IconAnalytics", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
