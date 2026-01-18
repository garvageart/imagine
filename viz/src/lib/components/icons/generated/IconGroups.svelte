<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M0-240v-63q0-43 44-70t116-27q13 0 25 .5t23 2.5q-14 21-21 44t-7 48v65zm240 0v-65q0-32 17.5-58.5T307-410t76.5-30 96.5-10q53 0 97.5 10t76.5 30 49 46.5 17 58.5v65zm540 0v-65q0-26-6.5-49T754-397q11-2 22.5-2.5t23.5-.5q72 0 116 26.5t44 70.5v63zm-455-80h311q-10-20-55.5-35T480-370t-100.5 15-54.5 35M160-440q-33 0-56.5-23.5T80-520q0-34 23.5-57t56.5-23q34 0 57 23t23 57q0 33-23 56.5T160-440m640 0q-33 0-56.5-23.5T720-520q0-34 23.5-57t56.5-23q34 0 57 23t23 57q0 33-23 56.5T800-440m-320-40q-50 0-85-35t-35-85q0-51 35-85.5t85-34.5q51 0 85.5 34.5T600-600q0 50-34.5 85T480-480m0-80q17 0 28.5-11.5T520-600t-11.5-28.5T480-640t-28.5 11.5T440-600t11.5 28.5T480-560m0-40\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M0-240v-53q0-38.57 41.5-62.78Q83-380 150.38-380q12.16 0 23.39.5t22.23 2.15q-8 17.35-12 35.17-4 17.81-4 37.18v65zm240 0v-65q0-32 17.5-58.5T307-410t76.5-30 96.5-10q53 0 97.5 10t76.5 30 49 46.5 17 58.5v65zm540 0v-65q0-19.86-3.5-37.43T765-377.27q11-1.73 22.17-2.23T810-380q67.5 0 108.75 23.77T960-293v53zM149.57-410q-28.57 0-49.07-20.56Q80-451.13 80-480q0-29 20.56-49.5Q121.13-550 150-550q29 0 49.5 20.5t20.5 49.93q0 28.57-20.5 49.07T149.57-410m660 0q-28.57 0-49.07-20.56Q740-451.13 740-480q0-29 20.56-49.5Q781.13-550 810-550q29 0 49.5 20.5t20.5 49.93q0 28.57-20.5 49.07T809.57-410M480-480q-50 0-85-35t-35-85q0-51 35-85.5t85-34.5q51 0 85.5 34.5T600-600q0 50-34.5 85T480-480\" />"
};
    let { size = "1.5em", className = "", title = "IconGroups", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
