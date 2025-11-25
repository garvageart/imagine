<script lang="ts">
    const variants = {
    "200": "<path d=\"M480-387.69 327.69-539.77h304.62z\" />",
    "300": "<path d=\"M480-373.85 303.85-549.38h352.3z\" />",
    "400": "<path d=\"M480-360 280-559h400z\" />",
    "600": "<path d=\"M480-346.43 256.82-568.61h446.36z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowDropDown", viewBox = "0 0 24 24", weight = "400" } = $props();
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
