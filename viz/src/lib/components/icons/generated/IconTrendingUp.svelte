<script lang="ts">
    const variants = {
    "200": "<path d=\"M142.23-276.15 120-298.38l250.46-251 163.16 163.15L790-639.54H646.08v-30.77H840v193.93h-30.54v-140.7L533.38-340.23 370.23-503.38z\" />",
    "300": "<path d=\"M132.62-258.08 100-290.69l271.23-271 165.08 165.08L785-643.77H649.54v-45.38H860v210.46h-44.77v-131.85L535.69-330.62 370.62-495.69z\" />",
    "400": "<path d=\"m123-240-43-43 292-291 167 167 241-241H653v-60h227v227h-59v-123L538-321 371-488z\" />",
    "600": "<path d=\"m122.43-227.56-56.56-56.57 308.39-306.83 167 167 219.52-220.65H638.87v-79.78H894.7v255.83h-78.79v-118.48L540.26-310.83l-167-167z\" />"
};
    let { size = "1.5em", className = "", title = "IconTrendingUp", viewBox = "0 0 24 24", weight = "400" } = $props();
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
