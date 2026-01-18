<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"m480-120-58-52q-101-91-167-157T150-447.5 95.5-544 80-634q0-94 63-157t157-63q52 0 99 22t81 62q34-40 81-62t99-22q94 0 157 63t63 157q0 46-15.5 90T810-447.5 705-329 538-172zm0-108q96-86 158-147.5t98-107 50-81 14-70.5q0-60-40-100t-100-40q-47 0-87 26.5T518-680h-76q-15-41-55-67.5T300-774q-60 0-100 40t-40 100q0 35 14 70.5t50 81 98 107T480-228m0-273\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"m480-121-41-37q-106-97-175-167.5t-110-126T96.5-552 80-643q0-90 60.5-150.5T290-854q57 0 105.5 27t84.5 78q42-54 89-79.5T670-854q89 0 149.5 60.5T880-643q0 46-16.5 91T806-451.5t-110 126T521-158z\" />"
};
    let { size = "1.5em", className = "", title = "IconFavorite", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
