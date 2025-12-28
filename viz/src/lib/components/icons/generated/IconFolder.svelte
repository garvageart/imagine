<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M157.69-180q-23.53 0-40.61-17.35Q100-214.69 100-237.69v-484.62q0-23 17.08-40.34Q134.16-780 157.69-780h249.46l65.39 65.39h329.77q23 0 40.34 17.34Q860-679.92 860-656.92v419.23q0 23-17.35 40.34Q825.31-180 802.31-180zm0-45.39h644.62q5.38 0 8.84-3.46t3.46-8.84v-419.23q0-5.39-3.46-8.85t-8.84-3.46H454.08l-65.39-65.38h-231q-5.38 0-8.84 3.46t-3.46 8.84v484.62q0 5.38 3.46 8.84t8.84 3.46m-12.3 0v-509.22z\" />",
    "400": "<path d=\"M140-160q-24 0-42-18.5T80-220v-520q0-23 18-41.5t42-18.5h281l60 60h339q23 0 41.5 18.5T880-680v460q0 23-18.5 41.5T820-160zm0-60h680v-460H456l-60-60H140zm0 0v-520z\" />",
    "600": "<path d=\"M145.09-145.87q-32.51 0-55.87-23.85-23.35-23.86-23.35-55.37v-509.82q0-31.74 23.35-55.76 23.36-24.03 55.87-24.03h269.69L481-747.91h333.91q31.74 0 55.76 23.85 24.03 23.86 24.03 55.36v443.61q0 31.51-24.03 55.37-24.02 23.85-55.76 23.85zm0-79.22h669.82V-668.7H448.09l-66.22-66.21H145.09zm0 0v-509.82z\" />"
};
    let { size = "1.5em", className = "", title = "IconFolder", viewBox = "0 0 24 24", weight = "400" } = $props();
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
