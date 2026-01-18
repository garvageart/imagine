<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M382-240 154-468l57-57 171 171 367-367 57 57z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M378-246 154-470l43-43 181 181 384-384 43 43z\" />"
};
    let { size = "1.5em", className = "", title = "IconCheck", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
