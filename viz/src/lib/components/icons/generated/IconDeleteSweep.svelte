<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M600-240v-80h160v80zm0-320v-80h280v80zm0 160v-80h240v80zM120-640H80v-80h160v-60h160v60h160v80h-40v440H120zm80 0v360h240v-360zm0 0v360z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M600-230v-60h145v60zm0-368v-60h280v60zm0 184v-60h235v60zM125-675H80v-60h170v-45h135v45h170v60h-45v475H125z\" />"
};
    let { size = "1.5em", className = "", title = "IconDeleteSweep", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
