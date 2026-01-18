<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M440-120v-240h80v80h320v80H520v80zm-320-80v-80h240v80zm160-160v-80H120v-80h160v-80h80v240zm160-80v-80h400v80zm160-160v-240h80v80h160v80H680v80zm-480-80v-80h400v80z\" />",
    "600": "<path d=\"M423.61-107v-238h73v83h353v73h-353v82zm-313.22-82v-73h260v73zm187-172.78v-82h-187v-72.44h187v-84h73v238.44zm126.22-82v-72.44h426v72.44zm166-171.22v-238h73v82h187v73h-187v83zm-479.22-83v-73h426v73z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M427-120v-225h60v83h353v60H487v82zm-307-82v-60h247v60zm187-166v-82H120v-60h187v-84h60v226zm120-82v-60h413v60zm166-165v-225h60v82h187v60H653v83zm-473-83v-60h413v60z\" />",
    "600": "<path d=\"M423.61-107v-238h73v83h353v73h-353v82zm-313.22-82v-73h260v73zm187-172.78v-82h-187v-72.44h187v-84h73v238.44zm126.22-82v-72.44h426v72.44zm166-171.22v-238h73v82h187v73h-187v83zm-479.22-83v-73h426v73z\" />"
};
    let { size = "1.5em", className = "", title = "IconTune", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
