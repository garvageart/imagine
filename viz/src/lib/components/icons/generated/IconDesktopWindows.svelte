<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M350.16-140v-45.39h84.46V-280H100v-540h760v540H525.38v94.61h84.46V-140zM145.39-325.38h669.22v-449.23H145.39zm0 0v-449.23z\" />",
    "400": "<path d=\"M320-120v-80h80v-80H80v-560h800v560H560v80h80v80zM160-360h640v-400H160zm0 0v-400z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M350.16-140v-45.39h84.46V-280H100v-540h760v540H525.38v94.61h84.46V-140z\" />",
    "400": "<path d=\"M334-120v-60h86v-100H80v-560h800v560H540v100h86v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconDesktopWindows", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
