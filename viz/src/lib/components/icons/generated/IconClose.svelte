<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m256-200-56-56 224-224-224-224 56-56 224 224 224-224 56 56-224 224 224 224-56 56-224-224z\" />",
    "600": "<path d=\"M249-193.43 193.43-249l231-231-231-231L249-766.57l231 231 231-231L766.57-711l-231 231 231 231L711-193.43l-231-231z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m249-207-42-42 231-231-231-231 42-42 231 231 231-231 42 42-231 231 231 231-42 42-231-231z\" />",
    "600": "<path d=\"M249-193.43 193.43-249l231-231-231-231L249-766.57l231 231 231-231L766.57-711l-231 231 231 231L711-193.43l-231-231z\" />"
};
    let { size = "1.5em", className = "", title = "IconClose", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
