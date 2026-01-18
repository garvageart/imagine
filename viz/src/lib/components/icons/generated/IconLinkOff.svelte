<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m770-302-60-62q40-11 65-42.5t25-73.5q0-50-35-85t-85-35H520v-80h160q83 0 141.5 58.5T880-480q0 57-29.5 105T770-302M634-440l-80-80h86v80zM792-56 56-792l56-56 736 736zM440-280H280q-83 0-141.5-58.5T80-480q0-69 42-123t108-71l74 74h-24q-50 0-85 35t-35 85 35 85 85 35h160zM320-440v-80h65l79 80z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m750-291-49-49q51-10 85-48t34-89q0-58-41-99t-99-41H525v-60h155q83 0 141.5 58.5T880-477q0 62-36 112t-94 74M594-447l-60-60h101v60zM814-56 63-807l43-43L857-99zM450-280H280q-83 0-141.5-58.5T80-480q0-72 44.5-127T238-676l56 56h-14q-58 0-99 41t-41 99 41 99 99 41h170zM325-450v-60h79l60 60z\" />"
};
    let { size = "1.5em", className = "", title = "IconLinkOff", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
