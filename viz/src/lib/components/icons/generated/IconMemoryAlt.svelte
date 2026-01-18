<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M240-360h80v-240h-80zm200 0h80v-240h-80zm200 0h80v-240h-80zm-480 80h640v-400H160zm0 0v-400zm40 160v-80H80v-560h120v-80h80v80h160v-80h80v80h160v-80h80v80h120v560H760v80h-80v-80H520v80h-80v-80H280v80z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M250-370h60v-220h-60zm200 0h60v-220h-60zm200 0h60v-220h-60zM210-120v-80H80v-560h130v-80h60v80h180v-80h60v80h180v-80h60v80h130v560H750v80h-60v-80H510v80h-60v-80H270v80z\" />"
};
    let { size = "1.5em", className = "", title = "IconMemoryAlt", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
