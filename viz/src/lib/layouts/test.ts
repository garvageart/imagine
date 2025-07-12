import type { VizSubPanel } from "$lib/components/panels/SubPanel.svelte";
import { generateKeyId } from "$lib/utils";
import { DEFAULT_THEME } from "$lib/constants";
import DifferentContent from "$lib/components/panels/workspace/generic/DifferentContent.svelte";
import SomeContent from "$lib/components/panels/workspace/generic/SomeContent.svelte";
import EvenMoreDifferent from "$lib/components/panels/workspace/generic/EvenMoreDifferent.svelte";
import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
import VizView from "$lib/views/views.svelte";
import Collections from "../../routes/collections/+page.svelte";
import type { Collection } from "$lib/types/images";

// Only one instance of a view/panel and its component
// can exist in the layout so we declare all of them here
export const views: VizView[] = [
    new VizView({
        name: "Clock",
        component: SomeContent as any
    }),
    new VizView({
        name: "What Day Is It?",
        component: DifferentContent as any
    }),
    new VizView({
        name: "Timezone",
        component: EvenMoreDifferent as any
    }),
    new VizView({
        name: "Welcome Text",
        component: DevWelcomeText as any
    }),
    new VizView<Collection[]>({
        name: "Collections",
        component: Collections as any,
        path: "/collections"
    })
];

const theme = DEFAULT_THEME;
// this seems so..... complex?? 
// but i guess it's fine
export const panels: VizSubPanel[] = [
    {
        id: "viz-test-content",
        maxSize: 100,
        minSize: 10,
        paneKeyId: generateKeyId(),
        views: [],
        childs: {
            internalSubPanelContainer: {
                id: "viz-internal-subpanel-sp-xBdVNtbU5nx60sZ1",
                paneKeyId: generateKeyId(16),
                smoothExpand: false,
                minSize: 10,
                size: 25,
                maxSize: 100
            },
            internalPanelContainer: {
                id: "viz-internal-panel-sp-3jaetpKK8Tc0F8iB",
                horizontal: true,
                keyId: generateKeyId(16),
                theme,
                style: "height: 100%",
                pushOtherPanes: true,
            },
            subPanels: [
                {
                    id: "viz-test-4365763",
                    maxSize: 100,
                    minSize: 10,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.id === 2),
                },
                {
                    id: "viz-test-again",
                    maxSize: 100,
                    minSize: 10,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.id === 3)
                }
            ]
        }
    },
    {
        id: "viz-test-something-else",
        maxSize: 100,
        minSize: 10,
        views: [],
        paneKeyId: generateKeyId(),
        childs: {
            internalSubPanelContainer: {
                id: "viz-internal-subpanel-sp-6H9DnoBrjh8fCEGy",
                smoothExpand: false,
                minSize: 10,
                maxSize: 100
            },
            internalPanelContainer: {
                id: "viz-internal-panel-sp-Q4AYXfRIaHmchHmX",
                horizontal: true,
                theme,
                style: "height: 100%",
                pushOtherPanes: true,
            },
            subPanels: [
                {
                    id: "viz-clock",
                    maxSize: 100,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.id === 1),
                },
                {
                    id: "viz-collections",
                    maxSize: 100,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.path === "/collections")
                }
            ]
        }
    },
];