<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M160-280h640v-240H160zm520-60q25 0 42.5-17.5T740-400t-17.5-42.5T680-460t-42.5 17.5T620-400t17.5 42.5T680-340m200-260H767l-80-80H273l-80 80H80l160-160h480zM80-200v-400h800v400z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M680.12-330q20.88 0 35.38-14.62t14.5-35.5-14.62-35.38-35.5-14.5-35.38 14.62-14.5 35.5 14.62 35.38 35.5 14.5M880-570H80l157-190h486zM80-200v-310h800v310z\" />"
};
    let { size = "1.5em", className = "", title = "IconHardDrive", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
