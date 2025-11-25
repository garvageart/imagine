<script lang="ts">
    const variants = {
    "200": "<path d=\"M160-515.38V-800h284.62v284.62zM160-160v-284.62h284.62V-160zm355.38-355.38V-800H800v284.62zm0 355.38v-284.62H800V-160zM190.77-546.15h223.08v-223.08H190.77zm355.38 0h223.08v-223.08H546.15zm0 355.38h223.08v-223.08H546.15zm-355.38 0h223.08v-223.08H190.77zm223.08-355.38\" />",
    "400": "<path d=\"M120-510v-330h330v330zm0 390v-330h330v330zm390-390v-330h330v330zm0 390v-330h330v330zM180-570h210v-210H180zm390 0h210v-210H570zm0 390h210v-210H570zm-390 0h210v-210H180zm210-390\" />"
};
    let { size = "1.5em", className = "", title = "IconGridView", viewBox = "0 0 24 24", weight = "400" } = $props();
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
