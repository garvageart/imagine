<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"M160-269.23V-300h640v30.77zm0-195.39v-30.76h640v30.76zM160-660v-30.77h640V-660z\" />",
    "400": "<path d=\"M120-240v-60h720v60zm0-210v-60h720v60zm0-210v-60h720v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconMenu", viewBox = "0 0 24 24", weight = "400" } = $props();
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
