<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M80-160v-640h320l80 80h400v80H447l-80-80H160v480l96-320h684L820-160zm164-80h516l72-240H316zm0 0 72-240zm-84-400v-80z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M80-160v-640h341l60 60h399v60H140v460l102-400h698L821-160z\" />"
};
    let { size = "1.5em", className = "", title = "IconFolderOpen", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
