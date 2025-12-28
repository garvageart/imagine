<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M74.62-140 480-840l405.38 700zm78.61-45.39h653.54L480-749.23zm329.8-60.07q10.7 0 17.76-7.24t7.06-17.94q0-10.71-7.24-17.76-7.24-7.06-17.95-7.06-10.7 0-17.76 7.24-7.05 7.24-7.05 17.94t7.24 17.76 17.94 7.06m-22.88-97.92h45.39v-217.85h-45.39zM480-467.31\" />",
    "400": "<path d=\"m40-120 440-760 440 760zm104-60h672L480-760zm340.18-57q12.82 0 21.32-8.68 8.5-8.67 8.5-21.5 0-12.82-8.68-21.32-8.67-8.5-21.5-8.5-12.82 0-21.32 8.68-8.5 8.67-8.5 21.5 0 12.82 8.68 21.32 8.67 8.5 21.5 8.5M454-348h60v-224h-60zm26-122\" />",
    "600": "<path d=\"M27.56-112.65 480-894.7l452.44 782.05zm127.18-73.57h650.52L480-748.13zm329.44-51.91q13.15 0 22.61-9.64 9.47-9.65 9.47-22.8t-9.64-22.33q-9.65-9.19-22.8-9.19t-22.61 9.36q-9.47 9.36-9.47 22.51t9.64 22.62q9.65 9.47 22.8 9.47M454-348h60v-219.48h-60zm26-119.17\" />"
};
    let { size = "1.5em", className = "", title = "IconWarning", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = $derived(variants[String(weight)] || variants["400"] || Object.values(variants)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox="0 -960 960 960"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
