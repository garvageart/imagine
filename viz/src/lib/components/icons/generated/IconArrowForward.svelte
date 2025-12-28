<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M693.69-457.31H180v-45.38h513.69L448.39-748 480-780l300 300-300 300-31.61-32z\" />",
    "400": "<path d=\"M686-450H160v-60h526L438-758l42-42 320 320-320 320-42-42z\" />",
    "600": "<path d=\"M662.82-440.39H145.87v-79.22h516.95L423.87-758.57 480-814.7 814.7-480 480-145.87l-56.13-55.56z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowForward", viewBox = "0 0 24 24", weight = "400" } = $props();
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
