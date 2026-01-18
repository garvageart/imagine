<script lang="ts">
    const variants: Record<string, string> = {
    "400": "<path d=\"M40-160v-112q0-34 17.5-62.5T104-378q62-31 126-46.5T360-440t130 15.5T616-378q29 15 46.5 43.5T680-272v112zm720 0v-120q0-44-24.5-84.5T666-434q51 6 96 20.5t84 35.5q36 20 55 44.5t19 53.5v120zM360-480q-66 0-113-47t-47-113 47-113 113-47 113 47 47 113-47 113-113 47m400-160q0 66-47 113t-113 47q-11 0-28-2.5t-28-5.5q27-32 41.5-71t14.5-81-14.5-81-41.5-71q14-5 28-6.5t28-1.5q66 0 113 47t47 113M120-240h480v-32q0-11-5.5-20T580-306q-54-27-109-40.5T360-360t-111 13.5T140-306q-9 5-14.5 14t-5.5 20zm240-320q33 0 56.5-23.5T440-640t-23.5-56.5T360-720t-56.5 23.5T280-640t23.5 56.5T360-560m0-80\" />"
};
    const filledVariants: Record<string, string> = {
    "400": "<path d=\"M38-160v-94q0-35 18-63.5t50-42.5q73-32 131.5-46T358-420t120 14 131 46q32 14 50.5 42.5T678-254v94zm700 0v-94q0-63-32-103.5T622-423q69 8 130 23.5t99 35.5q33 19 52 47t19 63v94zM358-481q-66 0-108-42t-42-108 42-108 108-42 108 42 42 108-42 108-108 42m360-150q0 66-42 108t-108 42q-11 0-24.5-1.5T519-488q24-25 36.5-61.5T568-631t-12.5-79.5T519-774q11-3 24.5-5t24.5-2q66 0 108 42t42 108\" />"
};
    let { size = "1.5em", className = "", title = "IconGroup", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
