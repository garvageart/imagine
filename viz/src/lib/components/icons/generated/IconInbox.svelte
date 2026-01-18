<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h720v720zm80-80h560v-120H640q-30 38-71.5 59T480-240t-88.5-21-71.5-59H200zm280-120q38 0 69-22t43-58h168v-360H200v360h168q12 36 43 58t69 22M200-200h560z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M120-120v-720h720v720zm360.25-173Q521-293 554-316.5t56-59.5h170v-404H180v404h170q23 36 56.25 59.5 33.24 23.5 74 23.5\" />"
};
    let { size = "1.5em", className = "", title = "IconInbox", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
