<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M140-254.62V-300h680v45.38zm0-202.69v-45.38h680v45.38zM140-660v-45.38h680V-660z\" />",
    "400": "<path d=\"M120-240v-80h720v80zm0-200v-80h720v80zm0-200v-80h720v80z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M140-254.62V-300h680v45.38zm0-202.69v-45.38h680v45.38zM140-660v-45.38h680V-660z\" />",
    "400": "<path d=\"M120-240v-60h720v60zm0-210v-60h720v60zm0-210v-60h720v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconMenu", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
