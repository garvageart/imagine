<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M418-340q24 24 62 23.5t56-27.5l224-336-336 224q-27 18-28.5 55t22.5 61m62-460q59 0 113.5 16.5T696-734l-76 48q-33-17-68.5-25.5T480-720q-133 0-226.5 93.5T160-400q0 42 11.5 83t32.5 77h552q23-38 33.5-79t10.5-85q0-36-8.5-70T766-540l48-76q30 47 47.5 100T880-406t-13 109-41 99q-11 18-30 28t-40 10H204q-21 0-40-10t-30-28q-26-45-40-95.5T80-400q0-83 31.5-155.5t86-127 127.5-86T480-800m7 313\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M418-340q25 25 63 23.5t55-27.5l221-333-333 221q-26 18-28.5 54.5T418-340M192-160q-18 0-34-8.5T134-193q-26-48-40-100T80-399q0-83 31.5-156T197-682.5t126.5-86T478-800q83 0 156.5 31.5t128 86 86 127.5T880-399q0 54-13 106.5T827-193q-9 16-25 24.5t-34 8.5z\" />"
};
    let { size = "1.5em", className = "", title = "IconSpeed", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
