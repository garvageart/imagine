<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M240-200h120v-240h240v240h120v-360L480-740 240-560zm-80 80v-480l320-240 320 240v480H520v-240h-80v240zm320-350\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M160-120v-480l320-240 320 240v480H560v-280H400v280z\" />"
};
    let { size = "1.5em", className = "", title = "IconHome", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
