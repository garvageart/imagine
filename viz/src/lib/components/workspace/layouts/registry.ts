import { Workspace, SplitNode, TabGroup } from "$lib/layouts/model.svelte";
import { views as viewRegistry } from "$lib/layouts/views";

export function createDefaultLayout(): Workspace {
    const collectionsView = viewRegistry.find((v) => v.path === "/collections");
    const filterView = viewRegistry.find((v) => v.name === "Filter");
    const clockView = viewRegistry.find((v) => v.name === "Clock");

    const root = new SplitNode({
        orientation: "horizontal",
        children: [
            new SplitNode({
                size: 30,
                orientation: "vertical",
                children: [
                    new TabGroup({
                        size: 50,
                        views: clockView ? [clockView] : []
                    }),
                    new TabGroup({
                        size: 50,
                        views: filterView ? [filterView] : []
                    })
                ]
            }),
            new TabGroup({
                size: 70,
                views: collectionsView ? [collectionsView] : []
            }),
        ]
    });

    return new Workspace(root);
}