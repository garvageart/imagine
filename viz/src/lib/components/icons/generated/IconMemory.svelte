<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M360-360v-240h240v240zm80-80h80v-80h-80zm-80 320v-80H200v-160h-80v-80h80v-80h-80v-80h80v-160h160v-80h80v80h80v-80h80v80h160v160h80v80h-80v80h80v80h-80v160H600v80h-80v-80h-80v80zm320-160v-400H280v400zM480-480\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M377-377v-205h205v205zm-17 257v-80H200v-160h-80v-60h80v-124h-80v-60h80v-160h160v-76h60v76h124v-76h60v76h160v160h76v60h-76v124h76v60h-76v160H604v80h-60v-80H420v80zm344-140v-444H260v444z\" />"
};
    let { size = "1.5em", className = "", title = "IconMemory", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
