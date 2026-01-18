<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M457.31-323.39v-368.69l-110.39 110-32.61-32.23L480-780l165.69 165.69-32.61 32.23-110.39-110v368.69zM180-180v-183h45.39v137.61h509.22V-363H780v183z\" />",
    "400": "<path d=\"M440-320v-326L336-542l-56-58 200-200 200 200-56 58-104-104v326zM160-160v-200h80v120h480v-120h80v200z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M457.31-323.39v-368.69l-110.39 110-32.61-32.23L480-780l165.69 165.69-32.61 32.23-110.39-110v368.69zM180-180v-183h45.39v137.61h509.22V-363H780v183z\" />",
    "400": "<path d=\"M450-313v-371L330-564l-43-43 193-193 193 193-43 43-120-120v371zM160-160v-203h60v143h520v-143h60v203z\" />"
};
    let { size = "1.5em", className = "", title = "IconUpload", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
