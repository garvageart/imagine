<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-280q17 0 28.5-11.5T520-320t-11.5-28.5T480-360t-28.5 11.5T440-320t11.5 28.5T480-280m-40-160h80v-240h-80zm40 360q-83 0-156-31.5T197-197t-85.5-127T80-480t31.5-156T197-763t127-85.5T480-880t156 31.5T763-763t85.5 127T880-480t-31.5 156T763-197t-127 85.5T480-80m0-80q134 0 227-93t93-227-93-227-227-93-227 93-93 227 93 227 227 93m0-320\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M479.98-280q14.02 0 23.52-9.48t9.5-23.5-9.48-23.52-23.5-9.5-23.52 9.48-9.5 23.5 9.48 23.52 23.5 9.5M453-433h60v-253h-60zm27.27 353q-82.74 0-155.5-31.5Q252-143 197.5-197.5t-86-127.34T80-480.5t31.5-155.66 86-126.84 127.34-85.5T480.5-880t155.66 31.5T763-763t85.5 127T880-480.27q0 82.74-31.5 155.5Q817-252 763-197.68q-54 54.31-127 86Q563-80 480.27-80\" />"
};
    let { size = "1.5em", className = "", title = "IconError", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
