<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M140-512.69V-820h307.31v307.31zM140-140v-307.31h307.31V-140zm372.69-372.69V-820H820v307.31zm0 372.69v-307.31H820V-140zm-327.3-418.08h216.53v-216.53H185.39zm372.69 0h216.53v-216.53H558.08zm0 372.69h216.53v-216.53H558.08zm-372.69 0h216.53v-216.53H185.39zm216.53-372.69\" />",
    "400": "<path d=\"M120-510v-330h330v330zm0 390v-330h330v330zm390-390v-330h330v330zm0 390v-330h330v330zM180-570h210v-210H180zm390 0h210v-210H570zm0 390h210v-210H570zm-390 0h210v-210H180zm210-390\" />"
};
    let { size = "1.5em", className = "", title = "IconGridView", viewBox = "0 0 24 24", weight = "400" } = $props();
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
