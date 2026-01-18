<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m480-336 128-184H494l80-280H360v320h120zM400-80v-320H280v-480h400l-80 280h160zm80-400H360z\" />",
    "600": "<path d=\"m466.78-293.48 178.09-256.69H516.04l75.92-263.05H344.52v363.39h122.26zM387.56-39.3v-331.31H265.3V-893h431.66l-75.48 263.04h175.26zm79.22-410.53H344.52z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M400-80v-320H280v-480h400l-80 280h160z\" />",
    "600": "<path d=\"M387.56-39.3v-331.31H265.3V-893h431.66l-75.48 263.04h175.26z\" />"
};
    let { size = "1.5em", className = "", title = "IconFlashOn", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
