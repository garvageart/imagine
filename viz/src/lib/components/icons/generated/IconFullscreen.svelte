<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M120-120v-200h80v120h120v80zm520 0v-80h120v-120h80v200zM120-640v-200h200v80H200v120zm640 0v-120H640v-80h200v200z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M120-120v-193h60v133h133v60zm527 0v-60h133v-133h60v193zM120-647v-193h193v60H180v133zm660 0v-133H647v-60h193v193z\" />"
};
    let { size = "1.5em", className = "", title = "IconFullscreen", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
