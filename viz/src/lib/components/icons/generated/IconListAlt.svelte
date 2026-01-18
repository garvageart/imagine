<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M279.62-293.69h56.15v-56.16h-56.15zm0-158.23h56.15v-56.16h-56.15zm0-158.23h56.15v-56.16h-56.15zm158.53 311.07h230.54v-45.38H438.15zm0-158.23h230.54v-45.38H438.15zm0-158.23h230.54v-45.38H438.15zM140-140v-680h680v680zm45.39-45.39h589.22v-589.22H185.39zm0 0v-589.22z\" />",
    "400": "<path d=\"M280-280h80v-80h-80zm0-160h80v-80h-80zm0-160h80v-80h-80zm160 320h240v-80H440zm0-160h240v-80H440zm0-160h240v-80H440zM120-120v-720h720v720zm80-80h560v-560H200zm0 0v-560z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M279.62-293.69h56.15v-56.16h-56.15zm0-158.23h56.15v-56.16h-56.15zm0-158.23h56.15v-56.16h-56.15zm158.53 311.07h230.54v-45.38H438.15zm0-158.23h230.54v-45.38H438.15zm0-158.23h230.54v-45.38H438.15zM140-140v-680h680v680z\" />",
    "400": "<path d=\"M270-286h60v-60h-60zm0-164h60v-60h-60zm0-164h60v-60h-60zm162 328h244v-60H432zm0-164h244v-60H432zm0-164h244v-60H432zM120-120v-720h720v720z\" />"
};
    let { size = "1.5em", className = "", title = "IconListAlt", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
