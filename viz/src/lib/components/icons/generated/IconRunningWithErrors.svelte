<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-80q-83 0-156-31.5T197-197t-85.5-127T80-480t31.5-156T197-763t127-85.5T480-880q88 0 166.5 36T782-742L480-440v-360q-134 0-227 93t-93 227 93 227 227 93q69 0 132-28.5T720-270v110q-53 38-114 59T480-80m320-160v-320h80v320zm40 160q-17 0-28.5-11.5T800-120t11.5-28.5T840-160t28.5 11.5T880-120t-11.5 28.5T840-80\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M478-80q-83 0-156-31.5T195-197t-85.5-127T78-480t31.5-156T195-763t127-85.5T478-880q88 0 166.5 36T780-742L478-440v-380q-142 0-241 98.81T138-480q0 142.37 98.81 241.19Q335.63-140 478-140q81 0 153.5-36.5T757-276v83q-57 54-129 83.5T478-80m339-147v-327h60v327zm33.02 140Q836-87 826.5-96.48t-9.5-23.5 9.48-23.52 23.5-9.5 23.52 9.48 9.5 23.5-9.48 23.52-23.5 9.5\" />"
};
    let { size = "1.5em", className = "", title = "IconRunningWithErrors", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
