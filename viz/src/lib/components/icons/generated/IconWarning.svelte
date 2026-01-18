<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m40-120 440-760 440 760zm138-80h604L480-720zm302-40q17 0 28.5-11.5T520-280t-11.5-28.5T480-320t-28.5 11.5T440-280t11.5 28.5T480-240m-40-120h80v-200h-80zm40-100\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m40-120 440-760 440 760zm444.18-117q12.82 0 21.32-8.68 8.5-8.67 8.5-21.5 0-12.82-8.68-21.32-8.67-8.5-21.5-8.5-12.82 0-21.32 8.68-8.5 8.67-8.5 21.5 0 12.82 8.68 21.32 8.67 8.5 21.5 8.5M454-348h60v-224h-60z\" />"
};
    let { size = "1.5em", className = "", title = "IconWarning", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
