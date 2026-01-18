<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M266.31-457.31 512-212l-32 32-300-300 300-300 32 32-245.69 245.31H780v45.38z\" />",
    "400": "<path d=\"m313-440 224 224-57 56-320-320 320-320 57 56-224 224h487v80z\" />",
    "600": "<path d=\"m297.18-440.39 238.95 238.96L480-145.87 145.87-480 480-814.7l56.13 56.13-238.95 238.96H814.7v79.22z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M266.31-457.31 512-212l-32 32-300-300 300-300 32 32-245.69 245.31H780v45.38z\" />",
    "400": "<path d=\"m274-450 248 248-42 42-320-320 320-320 42 42-248 248h526v60z\" />",
    "600": "<path d=\"m297.18-440.39 238.95 238.96L480-145.87 145.87-480 480-814.7l56.13 56.13-238.95 238.96H814.7v79.22z\" />"
};
    let { size = "1.5em", className = "", title = "IconArrowBack", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
