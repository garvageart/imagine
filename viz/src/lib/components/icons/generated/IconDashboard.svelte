<script lang="ts">
    const variants: Record<string, string> = {
    "200": "<path d=\"M535.38-586.15V-800H800v213.85zM160-484.62V-800h264.62v315.38zM535.38-160v-315.38H800V-160zM160-160v-213.85h264.62V-160zm30.77-355.38h203.08v-253.85H190.77zm375.38 324.61h203.08v-253.85H566.15zm0-426.15h203.08v-152.31H566.15zM190.77-190.77h203.08v-152.31H190.77zm203.08-152.31\" />",
    "300": "<path d=\"M522.69-578.08V-820H820v241.92zM140-467.31V-820h297.31v352.69zM522.69-140v-352.69H820V-140zM140-140v-241.92h297.31V-140zm45.39-372.69h206.53v-261.92H185.39zm382.69 327.3h206.53v-261.92H568.08zm0-438.07h206.53v-151.15H568.08zM185.39-185.39h206.53v-151.15H185.39zm206.53-151.15\" />",
    "400": "<path d=\"M510-570v-270h330v270zM120-450v-390h330v390zm390 330v-390h330v390zm-390 0v-270h330v270zm60-390h210v-270H180zm390 330h210v-270H570zm0-450h210v-150H570zM180-180h210v-150H180zm210-150\" />",
    "600": "<path d=\"M506.61-576.22v-289.22h358.83v289.22zM94.56-439.26v-426.18h358.83v426.18zM506.61-95.69v-426.75h358.83v426.75zm-412.05 0v-289.79h358.83v289.79zm79.79-422.79h199.82v-267.74H174.35zm412.04 343h199.83v-267.17H586.39zm0-479.96h199.83v-130.78H586.39zM174.35-175.48h199.82v-130.21H174.35zm199.82-130.21\" />"
};
    let { size = "1.5em", className = "", title = "IconDashboard", viewBox = "0 0 24 24", weight = "400" } = $props();
    const inner = variants[String(weight)] || variants["400"] || Object.values(variants)[0];
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
