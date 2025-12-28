<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M297.46-287.77h45.39v-193.85h-45.39zm319.69 0h45.39v-404.61h-45.39zm-159.84 0h45.38v-108.77h-45.38zm0-193.85h45.38V-547h-45.38zM197.69-140q-23.53 0-40.61-17.08T140-197.69v-564.62q0-23.53 17.08-40.61T197.69-820h564.62q23.53 0 40.61 17.08T820-762.31v564.62q0 23.53-17.08 40.61T762.31-140zm0-45.39h564.62q4.61 0 8.46-3.84 3.84-3.85 3.84-8.46v-564.62q0-4.61-3.84-8.46-3.85-3.84-8.46-3.84H197.69q-4.61 0-8.46 3.84-3.84 3.85-3.84 8.46v564.62q0 4.61 3.84 8.46 3.85 3.84 8.46 3.84m-12.3-589.22v589.22z\" />",
    "400": "<path d=\"M284-277h60v-205h-60zm332 0h60v-420h-60zm-166 0h60v-118h-60zm0-205h60v-60h-60zM180-120q-24 0-42-18t-18-42v-600q0-24 18-42t42-18h600q24 0 42 18t18 42v600q0 24-18 42t-42 18zm0-60h600v-600H180zm0-600v600z\" />",
    "600": "<path d=\"M286.83-277h60v-205h-60zm326.91 0h60v-420h-60zM450-277h60v-118h-60zm0-205h60v-60h-60zM185.09-105.87q-32.51 0-55.87-23.35-23.35-23.36-23.35-55.87v-589.82q0-32.74 23.35-56.26 23.36-23.53 55.87-23.53h589.82q32.74 0 56.26 23.53 23.53 23.52 23.53 56.26v589.82q0 32.51-23.53 55.87-23.52 23.35-56.26 23.35zm0-79.22h589.82v-589.82H185.09zm0-589.82v589.82z\" />"
};
    let { size = "1.5em", className = "", title = "IconAnalytics", viewBox = "0 0 24 24", weight = "400" } = $props();
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
