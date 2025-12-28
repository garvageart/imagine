<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"m317.77-180.39-32-32 126.77-126.38H100v-45.38h312.54L285.77-510.54l32-32 181.08 181.08zm324.46-237.46L461.15-598.92 642.23-780l32 31.62-126.77 126.77H860v45.38H547.46l126.77 126.38z\" />",
    "400": "<path d=\"m317-160-42-42 121-121H80v-60h316L275-504l42-42 193 193zm326-254L450-607l193-193 42 42-121 121h316v60H564l121 121z\" />",
    "600": "<path d=\"m304-145.87-55.57-55.56 111.39-110.83H65.87v-79.78h293.95L248.43-503.43 304-559l207.13 206.57zm352-255.69L449.43-608.13 656-814.7l55.57 55.57-110.83 111.39H894.7v79.78H600.74l110.83 110.83z\" />"
};
    let { size = "1.5em", className = "", title = "IconCompareArrows", viewBox = "0 0 24 24", weight = "400" } = $props();
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
