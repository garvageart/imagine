<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"M479.77-267.69 266.46-480.23l22-22.23 176.16 173.92v-410h30.76v411.23l174.93-174.92 22 22z\" />",
    "300": "<path d=\"M479.38-253.85 252.23-480.62l32-32.61 173.08 170.46v-405h45.38v406.62l171.46-171.46 32 31.99z\" />",
    "400": "<path d=\"M479-240 238-481l42-43 170 167v-400h60v402l168-168 42 42z\" />",
    "600": "<path d=\"M479-226.43 224.43-481 280-537.57l160.39 157.39v-386.43h79.22v388.43L678-536.57 733.57-481z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowDownwardAlt", viewBox = "0 0 24 24", weight = "400" } = $props();
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
