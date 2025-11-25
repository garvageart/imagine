<script lang="ts">
    const variants = {
    "200": "<path d=\"M480-567.85 289.92-377.77 267.69-400 480-612.31 692.31-400l-22.23 22.23z\" />",
    "300": "<path d=\"M480-560.92 286.46-367.39 253.85-400 480-626.15 706.15-400l-32.61 32.61z\" />",
    "400": "<path d=\"M480-554 283-357l-43-43 240-240 240 240-43 43z\" />",
    "600": "<path d=\"m480-540.43-197 197L226.43-400 480-653.57 733.57-400 677-343.43z\" />"
};
    let { size = "1.5em", className = "", title = "IconKeyboardArrowUp", viewBox = "0 0 24 24", weight = "400" } = $props();
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
