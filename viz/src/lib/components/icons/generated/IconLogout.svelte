<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M197.69-140q-23.53 0-40.61-17.08T140-197.69v-564.62q0-23.53 17.08-40.61T197.69-820h282.08v45.39H197.69q-4.61 0-8.46 3.84-3.84 3.85-3.84 8.46v564.62q0 4.61 3.84 8.46 3.85 3.84 8.46 3.84h282.08V-140zm470.62-187.69-33-32.23 97.38-97.39H363.85v-45.38h367.61l-97.38-97.39 32.61-32.61L820-479.19z\" />",
    "400": "<path d=\"M180-120q-24 0-42-18t-18-42v-600q0-24 18-42t42-18h299v60H180v600h299v60zm486-185-43-43 102-102H360v-60h363L621-612l43-43 176 176z\" />"
};
    let { size = "1.5em", className = "", title = "IconLogout", viewBox = "0 0 24 24", weight = "400" } = $props();
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
