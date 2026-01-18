<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M197-197q-54-55-85.5-127.5T80-480q0-84 31.5-156.5T197-763l57 57q-44 44-69 102t-25 124q0 67 25 125t69 101zm113-113q-32-33-51-76.5T240-480q0-51 19-94.5t51-75.5l57 57q-22 22-34.5 51T320-480t12.5 62 34.5 51zm170-90q-33 0-56.5-23.5T400-480t23.5-56.5T480-560t56.5 23.5T560-480t-23.5 56.5T480-400m170 90-57-57q22-22 34.5-51t12.5-62-12.5-62-34.5-51l57-57q32 32 51 75.5t19 94.5q0 50-19 93.5T650-310m113 113-57-57q44-44 69-102t25-124q0-67-25-125t-69-101l57-57q54 54 85.5 126.5T880-480q0 83-31.5 155.5T763-197\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M197-197q-54-54-85.5-126.5T80-480t31.5-156.5T197-763l42 43q-46 46-72.5 107.5T140-480t26.5 132.5T240-239zm113-113q-32-32-51-75.5T240-480t19-94.5 51-75.5l42 43q-24 24-38 56.5T300-480t14 70.5 39 57.5zm170-90q-33 0-56.5-23.5T400-480t23.5-56.5T480-560t56.5 23.5T560-480t-23.5 56.5T480-400m170 90-42-43q24-24 38-56.5t14-70.5-14-70.5-39-57.5l43-42q32 32 51 75.5t19 94.5q0 50-19 93.5T650-310m113 113-42-43q46-46 72.5-107.5T820-480t-26.5-132.5T720-721l43-42q54 55 85.5 127.5T880-480t-31.5 155.5T763-197\" />"
};
    let { size = "1.5em", className = "", title = "IconSensors", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
