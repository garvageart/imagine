<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"m252.85-230.85-22-22L458-480 230.85-707.15l22-22L480-502l227.15-227.15 22 22L502-480l227.15 227.15-22 22L480-458z\" />",
    "300": "<path d=\"m250.92-218.92-32-32L448-480 218.92-709.08l32-32L480-512l229.08-229.08 32 32L512-480l229.08 229.08-32 32L480-448z\" />",
    "400": "<path d=\"m249-207-42-42 231-231-231-231 42-42 231 231 231-231 42 42-231 231 231 231-42 42-231-231z\" />",
    "600": "<path d=\"M249-193.43 193.43-249l231-231-231-231L249-766.57l231 231 231-231L766.57-711l-231 231 231 231L711-193.43l-231-231z\" />"
};
    let { size = "1.5em", className = "", title = "IconClose", viewBox = "0 0 24 24", weight = "400" } = $props();
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
