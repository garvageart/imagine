<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M140-254.62V-300h680v45.38zm0-202.69v-45.38h680v45.38zM140-660v-45.38h680V-660z\" />",
    "400": "<path d=\"M120-240v-60h720v60zm0-210v-60h720v60zm0-210v-60h720v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconMenu", viewBox = "0 0 24 24", weight = "400" } = $props();
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
