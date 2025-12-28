<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M300-200q-24 0-42-18t-18-42v-560q0-24 18-42t42-18h440q24 0 42 18t18 42v560q0 24-18 42t-42 18zm0-60h440v-560H300zM180-80q-24 0-42-18t-18-42v-620h60v620h500v60zm120-180v-560z\" />"
};
    let { size = "1.5em", className = "", title = "IconContentCopy", viewBox = "0 0 24 24", weight = "400" } = $props();
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
