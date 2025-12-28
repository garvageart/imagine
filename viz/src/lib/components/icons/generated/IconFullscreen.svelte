<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M140-140v-183.39h45.39v138h138V-140zm497 0v-45.39h138v-138h45.38V-140zM140-636.61V-820h183.39v45.39h-138v138zm635 0v-138H637V-820h183.38v183.39z\" />",
    "400": "<path d=\"M120-120v-193h60v133h133v60zm527 0v-60h133v-133h60v193zM120-647v-193h193v60H180v133zm660 0v-133H647v-60h193v193z\" />",
    "600": "<path d=\"M105.87-105.87V-324.3h79.22v139.21H324.3v79.22zm530.39 0v-79.22h138.65V-324.3h79.79v218.43zM105.87-636.26V-854.7H324.3v79.79H185.09v138.65zm669.04 0v-138.65H636.26v-79.79H854.7v218.44z\" />"
};
    let { size = "1.5em", className = "", title = "IconFullscreen", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = $derived(variants[String(weight)] || variants["400"] || Object.values(variants)[0]);
</script>

<svg
    class={className}
    width={size}
    height={size}
    viewBox="0 -960 960 960"
    xmlns="http://www.w3.org/2000/svg"
    aria-label={title}
    focusable="false"
>
    {@html inner}
</svg>

<style>
    svg {
        display: inline-block;
        vertical-align: middle;
    }
</style>
