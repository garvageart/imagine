<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M480-80q-33 0-56.5-23.5T400-160h160q0 33-23.5 56.5T480-80M320-200v-80h320v80zm10-120q-69-41-109.5-110T180-580q0-125 87.5-212.5T480-880t212.5 87.5T780-580q0 81-40.5 150T630-320zm24-80h252q45-32 69.5-79T700-580q0-92-64-156t-156-64-156 64-64 156q0 54 24.5 101t69.5 79m126 0\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M480-80q-34 0-57.5-23.5T399-161h162q0 34-23.5 57.5T480-80M318-223v-60h324v60zm5-121q-66-43-104.5-107.5T180-597q0-122 89-211t211-89 211 89 89 211q0 81-38 145.5T637-344z\" />"
};
    let { size = "1.5em", className = "", title = "IconLightbulb", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
