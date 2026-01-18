<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M440-440H200v-80h240v-240h80v240h240v80H520v240h-80z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M450-450H200v-60h250v-250h60v250h250v60H510v250h-60z\" />"
};
    let { size = "1.5em", className = "", title = "IconAdd", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
