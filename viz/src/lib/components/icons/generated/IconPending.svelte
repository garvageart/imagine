<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M280-420q25 0 42.5-17.5T340-480t-17.5-42.5T280-540t-42.5 17.5T220-480t17.5 42.5T280-420m200 0q25 0 42.5-17.5T540-480t-17.5-42.5T480-540t-42.5 17.5T420-480t17.5 42.5T480-420m200 0q25 0 42.5-17.5T740-480t-17.5-42.5T680-540t-42.5 17.5T620-480t17.5 42.5T680-420M480-80q-83 0-156-31.5T197-197t-85.5-127T80-480t31.5-156T197-763t127-85.5T480-880t156 31.5T763-763t85.5 127T880-480t-31.5 156T763-197t-127 85.5T480-80m0-80q134 0 227-93t93-227-93-227-227-93-227 93-93 227 93 227 227 93m0-320\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M266.12-430q20.88 0 35.38-14.62t14.5-35.5-14.62-35.38-35.5-14.5-35.38 14.62-14.5 35.5 14.62 35.38 35.5 14.5m214 0q20.88 0 35.38-14.62t14.5-35.5-14.62-35.38-35.5-14.5-35.38 14.62-14.5 35.5 14.62 35.38 35.5 14.5m213 0q20.88 0 35.38-14.62t14.5-35.5-14.62-35.38-35.5-14.5-35.38 14.62-14.5 35.5 14.62 35.38 35.5 14.5M480.27-80q-82.74 0-155.5-31.5Q252-143 197.5-197.5t-86-127.34T80-480.5t31.5-155.66 86-126.84 127.34-85.5T480.5-880t155.66 31.5T763-763t85.5 127T880-480.27q0 82.74-31.5 155.5Q817-252 763-197.68q-54 54.31-127 86Q563-80 480.27-80\" />"
};
    let { size = "1.5em", className = "", title = "IconPending", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
