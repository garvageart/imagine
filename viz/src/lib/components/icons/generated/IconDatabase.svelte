<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-120q-151 0-255.5-46.5T120-280v-400q0-66 105.5-113T480-840t254.5 47T840-680v400q0 67-104.5 113.5T480-120m0-479q89 0 179-25.5T760-679q-11-29-100.5-55T480-760q-91 0-178.5 25.5T200-679q14 30 101.5 55T480-599m0 199q42 0 81-4t74.5-11.5 67-18.5 57.5-25v-120q-26 14-57.5 25t-67 18.5T561-524t-81 4-82-4-75.5-11.5T256-554t-56-25v120q25 14 56 25t66.5 18.5T398-404t82 4m0 200q46 0 93.5-7t87.5-18.5 67-26 32-29.5v-98q-26 14-57.5 25t-67 18.5T561-324t-81 4-82-4-75.5-11.5T256-354t-56-25v99q5 15 31.5 29t66.5 25.5 88 18.5 94 7\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M480-520q161 0 260.5-49T840-680q0-63-99.5-111.5T480-840t-260.5 48.5T120-680q0 62 99.5 111T480-520m0 100q50 0 112-8.5T709.5-455t93-46.5T840-570v100q0 40-37.5 68.5t-93 46.5T592-328.5 480-320q-49 0-111.5-8.5t-118-27-93-47T120-470v-100q0 39 37.5 67.5t93 47 118 27T480-420m0 200q50 0 112-8.5T709.5-255t93-46.5T840-370v100q0 40-37.5 68.5t-93 46.5T592-128.5 480-120q-49 0-111.5-8.5t-118-27-93-47T120-270v-100q0 39 37.5 67.5t93 47 118 27T480-220\" />"
};
    let { size = "1.5em", className = "", title = "IconDatabase", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
