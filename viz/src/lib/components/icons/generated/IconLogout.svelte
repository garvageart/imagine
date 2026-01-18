<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h360v80H200v560h280v80zm520-160-55-58 102-102H360v-80h327L585-622l55-58 200 200z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h359v60H180v600h299v60zm546-185-43-43 102-102H360v-60h363L621-612l43-43 176 176z\" />"
};
    let { size = "1.5em", className = "", title = "IconLogout", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
