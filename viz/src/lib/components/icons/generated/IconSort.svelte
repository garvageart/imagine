<script lang="ts">
    const variants = {
    "200": "<path d=\"M160-280v-30.77h190V-280zm0-184.62v-30.76h414.62v30.76zm0-184.61V-680h640v30.77z\" />",
    "400": "<path d=\"M120-240v-60h240v60zm0-210v-60h480v60zm0-210v-60h720v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconSort", viewBox = "0 0 24 24", weight = "400" } = $props();
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
