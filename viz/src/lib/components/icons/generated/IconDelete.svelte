<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M200-120v-600h-40v-80h200v-40h240v40h200v80h-40v600zm80-80h400v-520H280zm80-80h80v-360h-80zm160 0h80v-360h-80zM280-720v520z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M367-266h60v-399h-60zm166 0h60v-399h-60zM201-120v-630h-41v-60h188v-30h264v30h188v60h-41v630z\" />"
};
    let { size = "1.5em", className = "", title = "IconDelete", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
