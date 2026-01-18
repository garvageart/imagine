<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M140-140v-223.85h307.31V-140zm372.69 0v-391.15H820V-140zM140-429.23V-820h307.31v390.77zm372.69-167.31V-820H820v223.46zm-327.3 411.15h216.53v-133.07H185.39zm372.69 0h216.53v-300.38H558.08zM185.39-474.61h216.53v-300H185.39zm372.69-167.31h216.53v-132.69H558.08z\" />",
    "400": "<path d=\"M120-120v-240h320v240zm400 0v-400h320v400zM120-440v-400h320v400zm400-160v-240h320v240zM200-200h160v-80H200zm400 0h160v-240H600zM200-520h160v-240H200zm400-160h160v-80H600z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M140-140v-223.85h307.31V-140zm372.69 0v-391.15H820V-140zM140-429.23V-820h307.31v390.77zm372.69-167.31V-820H820v223.46z\" />",
    "400": "<path d=\"M120-120v-245h330v245zm390 0v-415h330v415zM120-425v-415h330v415zm390-170v-245h330v245z\" />"
};
    let { size = "1.5em", className = "", title = "IconBrowse", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
