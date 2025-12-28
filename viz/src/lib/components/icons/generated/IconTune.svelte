<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M440.46-130v-209.23h45.39V-257H830v45.38H485.85V-130zM130-211.62V-257h225.08v45.38zm179.69-164.07v-81.62H130v-45.38h179.69v-82.85h45.39v209.85zm130.77-81.62v-45.38H830v45.38zm164.46-163.84V-830h45.39v81.62H830V-703H650.31v81.85zM130-703v-45.38h389.54V-703z\" />",
    "400": "<path d=\"M427-120v-225h60v83h353v60H487v82zm-307-82v-60h247v60zm187-166v-82H120v-60h187v-84h60v226zm120-82v-60h413v60zm166-165v-225h60v82h187v60H653v83zm-473-83v-60h413v60z\" />",
    "600": "<path d=\"M423.61-107v-238h73v83h353v73h-353v82zm-313.22-82v-73h260v73zm187-172.78v-82h-187v-72.44h187v-84h73v238.44zm126.22-82v-72.44h426v72.44zm166-171.22v-238h73v82h187v73h-187v83zm-479.22-83v-73h426v73z\" />"
};
    let { size = "1.5em", className = "", title = "IconTune", viewBox = "0 0 24 24", weight = "400" } = $props();
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
