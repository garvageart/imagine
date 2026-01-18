<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M560-280h200v-200h-80v120H560zM200-480h80v-120h120v-80H200zM80-160v-640h800v640zm80-80h640v-480H160zm0 0v-480z\" />",
    "600": "<path d=\"M571.91-279.09h191v-194h-60v134h-131zM198.09-486.91h60v-134h131v-60h-191zM65.87-145.87V-814.7H894.7v668.83zm79.22-79.22h669.82v-509.82H145.09zm0 0v-509.82z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M577-274h191v-194h-60v134H577zM193-492h60v-134h131v-60H193zM80-160v-640h800v640z\" />",
    "600": "<path d=\"M571.91-279.09h191v-194h-60v134h-131zM198.09-486.91h60v-134h131v-60h-191zM65.87-145.87V-814.7H894.7v668.83z\" />"
};
    let { size = "1.5em", className = "", title = "IconAspectRatio", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
