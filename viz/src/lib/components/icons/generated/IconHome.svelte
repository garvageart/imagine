<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M225.39-185.39h155.76v-244.99h197.7v244.99h155.76v-381.92L480-759.23 225.39-567.44zM180-140v-450l300-225.77L780-590v450H533.46v-245H426.54v245zm300-332.62\" />",
    "400": "<path d=\"M220-180h150v-250h220v250h150v-390L480-765 220-570zm-60 60v-480l320-240 320 240v480H530v-250H430v250zm320-353\" />",
    "600": "<path d=\"M225.09-185.09h135.3v-254.52h239.22v254.52h135.3v-382.65L480-758.78 225.09-567.55zm-79.22 79.22v-501.48L480-858.09l334.7 250.74v501.48H526.61v-260.74h-93.22v260.74zM480-472.43\" />"
};
    let { size = "1.5em", className = "", title = "IconHome", viewBox = "0 0 24 24", weight = "400" } = $props();
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
