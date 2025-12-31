import { onMount } from "svelte";

export function measureComponentRenderTimes(message?: string) {
    let startTime: number | undefined;

    $effect.pre(() => {
        startTime = window.performance.now();
    });

    onMount(() => {
        if (!startTime) {
            return;
        }

        const endTime = window.performance.now();

        if (message) {
            console.log(message, startTime - endTime, "ms");
        }

        console.log(`Render time for component: ${endTime - startTime} ms`);
    });
}

export function resetAndReloadLayout() {

    localStorage.removeItem("viz:workspaceLayout");

    location.reload();

}