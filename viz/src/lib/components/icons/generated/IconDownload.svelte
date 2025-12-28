<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M480-323.39 314.31-489.08l32.61-32.23 110.39 110V-780h45.38v368.69l110.39-110 32.61 32.23zM237.69-180q-23.53 0-40.61-17.08T180-237.69V-363h45.39v125.31q0 4.61 3.84 8.46 3.85 3.84 8.46 3.84h484.62q4.61 0 8.46-3.84 3.84-3.85 3.84-8.46V-363H780v125.31q0 23.53-17.08 40.61T722.31-180z\" />",
    "400": "<path d=\"M480-313 287-506l43-43 120 120v-371h60v371l120-120 43 43zM220-160q-24 0-42-18t-18-42v-143h60v143h520v-143h60v143q0 24-18 42t-42 18z\" />",
    "600": "<path d=\"M480-318.09 266.65-531.44l57.13-56 116.61 116.61V-814.7h79.22v343.87l116.61-116.61 57.13 56zM225.09-145.87q-32.51 0-55.87-23.35-23.35-23.36-23.35-55.87v-143h79.22v143h509.82v-143h79.79v143q0 32.48-23.53 55.85-23.52 23.37-56.26 23.37z\" />"
};
    let { size = "1.5em", className = "", title = "IconDownload", viewBox = "0 0 24 24", weight = "400" } = $props();
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
