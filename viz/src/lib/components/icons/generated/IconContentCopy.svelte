<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M280-240v-640h520v640zm80-80h360v-480H360zM120-80v-640h80v560h440v80zm240-240v-480z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M240-200v-680h560v680zM120-80v-680h60v620h500v60z\" />"
};
    let { size = "1.5em", className = "", title = "IconContentCopy", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
