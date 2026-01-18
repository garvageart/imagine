<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M120-520v-320h320v320zm0 400v-320h320v320zm400-400v-320h320v320zm0 400v-320h320v320zM200-600h160v-160H200zm400 0h160v-160H600zm0 400h160v-160H600zm-400 0h160v-160H200zm160-400\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M120-510v-330h330v330zm0 390v-330h330v330zm390-390v-330h330v330zm0 390v-330h330v330z\" />"
};
    let { size = "1.5em", className = "", title = "IconGridView", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
