<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"M649.23-200v-192.31H760V-200zm-224.61 0v-560h110.76v560zM200-200v-367.69h110.77V-200z\" />",
    "300": "<path d=\"M654.61-180v-236.15H780V-180zm-237.3 0v-600h125.38v600zM180-180v-403.84h125.39V-180z\" />",
    "400": "<path d=\"M660-160v-280h140v280zm-250 0v-640h140v640zm-250 0v-440h140v440z\" />",
    "600": "<path d=\"M661.13-145.87v-300.35H814.7v300.35zm-257.35 0V-814.7h152.44v668.83zm-257.91 0V-614.7h153v468.83z\" />"
};
    let { size = "1.5em", className = "", title = "IconBarChart", viewBox = "0 0 24 24", weight = "400" } = $props();
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
