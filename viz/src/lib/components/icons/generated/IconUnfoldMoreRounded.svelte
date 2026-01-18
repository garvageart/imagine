<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m480-236 93-93q12-12 29-12t29 12 12 29-12 29L508-148q-6 6-13 8.5t-15 2.5-15-2.5-13-8.5L329-271q-12-12-12-29t12-29 29-12 29 12zm0-484-93 93q-12 12-29 12t-29-12-12-29 12-29l123-123q6-6 13-8.5t15-2.5 15 2.5 13 8.5l123 123q12 12 12 29t-12 29-29 12-29-12z\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m480-208 114-114q9.13-9 22.07-9 12.93 0 21.93 9.1 9 9.11 9 22 0 12.9-9 21.9L501-141q-5 5-10.13 7-5.14 2-11 2-5.87 0-10.87-2t-10-7L322-278q-9-9.13-9-22.07 0-12.93 9.1-21.93 9.11-9 22-9 12.9 0 21.9 9zm0-540L366-634q-9.13 9-22.07 9-12.93 0-21.93-9.1-9-9.11-9-22 0-12.9 9-21.9l137-137q5-5 10.13-7 5.14-2 11-2 5.87 0 10.87 2t10 7l137 137q9 9.13 9 22.07 0 12.93-9.1 21.93-9.11 9-22 9-12.9 0-21.9-9z\" />"
};
    let { size = "1.5em", className = "", title = "IconUnfoldMoreRounded", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
