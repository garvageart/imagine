<script lang="ts">
    const variants: Record<string, string> = {
    "300": "<path d=\"M352.31-389.31h367.15L603.23-545.69l-103 132.31L433-496.15zM220.77-230.77V-860H850v629.23zm45.38-45.39h538.46v-538.45H266.15zM110-120v-616.92h45.39v571.53h571.53V-120zm156.15-694.61v538.45z\" />",
    "400": "<path d=\"M360-400h400L622-580l-92 120-62-80zM240-240v-640h640v640zm80-80h480v-480H320zM80-80v-640h80v560h560v80zm240-720v480z\" />"
};
    const filledVariants: Record<string, string> = {
    "300": "<path d=\"M352.31-389.31h367.15L603.23-545.69l-103 132.31L433-496.15zm-73.85 158.54q-23.53 0-40.61-17.08t-17.08-40.61v-513.85q0-23.53 17.08-40.61T278.46-860h513.85q23.52 0 40.61 17.08Q850-825.84 850-802.31v513.85q0 23.53-17.08 40.61-17.09 17.08-40.61 17.08zm0-45.39h513.85q4.61 0 8.46-3.84 3.84-3.85 3.84-8.46v-513.85q0-4.61-3.84-8.46-3.85-3.84-8.46-3.84H278.46q-4.61 0-8.46 3.84-3.85 3.85-3.85 8.46v513.85q0 4.61 3.85 8.46 3.85 3.84 8.46 3.84M167.69-120q-23.52 0-40.61-17.08Q110-154.17 110-177.7v-559.22h45.39v559.22q0 4.62 3.84 8.47 3.85 3.84 8.46 3.84h559.23V-120zm106.93-694.61h512.3q3.08 0 5.77 3.65t2.69 8.65v513.85q0 5-2.69 8.65t-5.77 3.65h-512.3q-3.08 0-5.77-3.65-2.7-3.65-2.7-8.65v-513.85q0-5 2.7-8.65 2.69-3.65 5.77-3.65\" />",
    "400": "<path d=\"M345-377h391L609-548 506-413l-68-87zm-85 177q-24 0-42-18t-18-42v-560q0-24 18-42t42-18h560q24 0 42 18t18 42v560q0 24-18 42t-42 18zm0-60h560v-560H260zM140-80q-24 0-42-18t-18-42v-620h60v620h620v60zm120-740h560v560H260z\" />"
};
    let { size = "1.5em", className = "", title = "IconPhotoLibrary", viewBox = "0 -960 960 960", weight = "400", fill = false, ...rest } = $props();
    
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
