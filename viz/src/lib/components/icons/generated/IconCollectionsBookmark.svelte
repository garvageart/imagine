<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M320-320h480v-480h-80v280l-100-60-100 60v-280H320zm-80 80v-640h640v640zM80-80v-640h80v560h560v80zm240-720v480zm200 280 100-60 100 60-100-60z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M200-200v-680h680v680zM80-80v-680h60v620h620v60zm494-740v278l93-57 93 57v-278z\" />"
};
    let { size = "1.5em", className = "", title = "IconCollectionsBookmark", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
