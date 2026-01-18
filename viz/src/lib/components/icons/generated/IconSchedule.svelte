<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m612-292 56-56-148-148v-184h-80v216zM480-80q-83 0-156-31.5T197-197t-85.5-127T80-480t31.5-156T197-763t127-85.5T480-880t156 31.5T763-763t85.5 127T880-480t-31.5 156T763-197t-127 85.5T480-80m0-80q133 0 226.5-93.5T800-480t-93.5-226.5T480-800t-226.5 93.5T160-480t93.5 226.5T480-160\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m627-287 45-45-159-160v-201h-60v225zM480-80q-82 0-155-31.5t-127.5-86-86-127.5T80-480t31.5-155 86-127.5 127.5-86T480-880t155 31.5 127.5 86 86 127.5T880-480t-31.5 155-86 127.5-127.5 86T480-80\" />"
};
    let { size = "1.5em", className = "", title = "IconSchedule", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
    const activeMap = $derived(fill ? filledVariants : variants);
    const inner = $derived(activeMap[String(weight)] || activeMap["400"] || Object.values(activeMap)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox={viewBox}
    fill="currentColor"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
    {...rest}
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
