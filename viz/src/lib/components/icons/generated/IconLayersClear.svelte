<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m644-448-56-58 122-94-230-178-94 72-56-58 150-116 360 280zm115 114-58-58 73-56 66 50zm33 258L632-236 480-118 120-398l66-50 294 228 94-73-57-56-37 29-360-280 83-65L55-811l57-57 736 736zM487-606\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M650-415 335-730l145-112 360 280zm86 86-43-43 97-75 50 37zm94 264L641-255 480-130 120-410l50-37 310 241 118-92-43-43-75 59-360-280 120-94L73-822l43-43 757 757z\" />"
};
    let { size = "1.5em", className = "", title = "IconLayersClear", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
