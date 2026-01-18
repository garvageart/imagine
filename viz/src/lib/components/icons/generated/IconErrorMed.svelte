<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m443-362 74-101 73 101 130-177-64-47-66 89-74-101-73 101-74-101-129 177 64 47 65-89zm37 282q-83 0-156-31.5T197-197t-85.5-127T80-480t31.5-156T197-763t127-85.5T480-880t156 31.5T763-763t85.5 127T880-480t-31.5 156T763-197t-127 85.5T480-80m0-80q134 0 227-93t93-227-93-227-227-93-227 93-93 227 93 227 227 93m0-320\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m443-379 74-101 73 101 116-158-48-35-68 92-74-101-73 101-74-101-115 158 48 35 67-92zm37 299q-82 0-155-31.5t-127.5-86-86-127.5T80-480q0-83 31.5-156t86-127T325-848.5 480-880q83 0 156 31.5T763-763t85.5 127T880-480q0 82-31.5 155T763-197.5t-127 86T480-80\" />"
};
    let { size = "1.5em", className = "", title = "IconErrorMed", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
