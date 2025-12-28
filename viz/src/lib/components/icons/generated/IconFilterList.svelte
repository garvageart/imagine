<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M411.15-260v-45.39h137.31V-260zM256.16-457.31v-45.38h447.3v45.38zM140-654.61V-700h680v45.39z\" />",
    "400": "<path d=\"M400-240v-60h160v60zM240-450v-60h480v60zM120-660v-60h720v60z\" />",
    "600": "<path d=\"M389.83-211.74v-79.22h180.34v79.22zM227.56-440.96v-78.65h504.88v78.65zM105.87-669.61v-79.22H854.7v79.22z\" />"
};
    let { size = "1.5em", className = "", title = "IconFilterList", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = $derived(variants[String(weight)] || variants["400"] || Object.values(variants)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox="0 -960 960 960"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
