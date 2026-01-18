<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M280-80q-83 0-141.5-58.5T80-280t58.5-141.5T280-480t141.5 58.5T480-280t-58.5 141.5T280-80m544-40L568-376q-12-13-25.5-26.5T516-428q38-24 61-64t23-88q0-75-52.5-127.5T420-760t-127.5 52.5T240-580q0 6 .5 11.5T242-557q-18 2-39.5 8T164-535q-2-11-3-22t-1-23q0-109 75.5-184.5T420-840t184.5 75.5T680-580q0 43-13.5 81.5T629-428l251 252zm-615-61 71-71 70 71 29-28-71-71 71-71-28-28-71 71-71-71-28 28 71 71-71 71z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M270-80q-78 0-134-56T80-270t56-134 134-56 134 56 56 134-56 134-134 56m566-40L573-383q-14 11-31.5 21.5T508-344q-5-14-11-28.5T483-399q54-21 91.5-69.5T612-584q0-81-57-138.5T417-780q-82 0-139.5 57.5T220-584q0 17 3.5 35.5T232-517q-13 2-29 6.5T174-500q-7-18-10.5-40t-3.5-44q0-107 75-181.5T417-840q106 0 180.5 75T672-584q0 43-15 85t-41 73l264 262zm-635-56 69-69 68 68 23-23-69-69 71-71-23-23-70 70-70-70-23 23 70 70-70 70z\" />"
};
    let { size = "1.5em", className = "", title = "IconSearchOff", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
