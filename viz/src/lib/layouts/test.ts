import VizSubPanelData from "./subpanel.svelte";
import { views } from "./views";

export const testLayout: VizSubPanelData[] = [
    new VizSubPanelData({
        id: "viz-test",
        size: 30,
        minSize: 10,
        maxSize: 100,
        subPanels: [
            {
                id: "viz-test-4365763",
                views: views.filter(view => view.id === 2),
            },
            {
                id: "viz-test-again",
                views: views.filter(view => view.id === 3)
            }
        ]
    }),
    new VizSubPanelData({
        id: "viz-test-2",
        subPanels: [
            {
                id: "viz-clock",
                views: views.filter(view => view.id === 1),
            },
            {
                id: "viz-collections",
                views: views.filter(view => view.path === "/collections")
            }
        ]
    })

];