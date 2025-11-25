<script lang="ts">
    const variants = {
    "200": "<path d=\"M258.62-464.62 502-222l-22 22-280-280 280-280 22 22-243.38 242.62H760v30.76z\" />",
    "300": "<path d=\"M266.31-457.31 512-212l-32 32-300-300 300-300 32 32-245.69 245.31H780v45.38z\" />",
    "400": "<path d=\"m274-450 248 248-42 42-320-320 320-320 42 42-248 248h526v60z\" />",
    "600": "<path d=\"m297.18-440.39 238.95 238.96L480-145.87 145.87-480 480-814.7l56.13 56.13-238.95 238.96H814.7v79.22z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowBack", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = variants[String(weight)] || variants["400"] || Object.values(variants)[0];
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
