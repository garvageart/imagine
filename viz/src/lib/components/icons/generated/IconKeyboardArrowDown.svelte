<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"M480-371.69 267.69-584l22.23-22.23L480-416.15l190.08-190.08L692.31-584z\" />",
    "300": "<path d=\"M480-357.85 253.85-584l32.61-32.61L480-423.08l193.54-193.53L706.15-584z\" />",
    "400": "<path d=\"M480-344 240-584l43-43 197 197 197-197 43 43z\" />",
    "600": "<path d=\"M480-330.43 226.43-584 283-640.57l197 197 197-197L733.57-584z\" />"
};
    let { size = "1.5em", className = "", title = "IconKeyboardArrowDown", viewBox = "0 0 24 24", weight = "400" } = $props();
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
