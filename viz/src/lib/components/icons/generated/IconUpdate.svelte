<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-120q-75 0-140.5-28.5t-114-77-77-114T120-480t28.5-140.5 77-114 114-77T480-840q82 0 155.5 35T760-706v-94h80v240H600v-80h110q-41-56-101-88t-129-32q-117 0-198.5 81.5T200-480t81.5 198.5T480-200q105 0 183.5-68T756-440h82q-15 137-117.5 228.5T480-120m112-192L440-464v-216h80v184l128 128z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M483-120q-75 0-141-28.5T226.5-226t-78-115T120-482t28.5-140 78-113.5T342-812t141-28q80 0 151.5 35T758-709v-106h60v208H609v-60h105q-44-51-103.5-82T483-780q-125 0-214 85.5T180-485q0 127 88 216t215 89q125 0 211-88t86-213h60q0 150-104 255.5T483-120m122-197L451-469v-214h60v189l137 134z\" />"
};
    let { size = "1.5em", className = "", title = "IconUpdate", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
