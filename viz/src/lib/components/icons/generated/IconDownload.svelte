<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58zM160-160v-200h80v120h480v-120h80v200z\" />",
    "600": "<path d=\"M480-318.09 266.65-531.44l57.13-56 116.61 116.61V-814.7h79.22v343.87l116.61-116.61 57.13 56zM145.87-145.87v-222.22h79.22v143h509.82v-143h79.79v222.22z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M480-313 287-506l43-43 120 120v-371h60v371l120-120 43 43zM160-160v-203h60v143h520v-143h60v203z\" />",
    "600": "<path d=\"M480-318.09 266.65-531.44l57.13-56 116.61 116.61V-814.7h79.22v343.87l116.61-116.61 57.13 56zM145.87-145.87v-222.22h79.22v143h509.82v-143h79.79v222.22z\" />"
};
    let { size = "1.5em", className = "", title = "IconDownload", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
