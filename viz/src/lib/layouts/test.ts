import type { VizSubPanel, VizView } from "$lib/components/panels/SubPanel.svelte";
import { generateKeyId } from "$lib/utils";
import { DEFAULT_THEME } from "$lib/constants";
import DifferentContent from "$lib/components/panels/workspace/generic/DifferentContent.svelte";
import SomeContent from "$lib/components/panels/workspace/generic/SomeContent.svelte";
import EvenMoreDifferent from "$lib/components/panels/workspace/generic/EvenMoreDifferent.svelte";
import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";

// Only one instance of a view/panel and its component
// can exist in the layout so we declare all of them here
export const views: VizView[] = [
    {
        name: "View 1",
        component: SomeContent,
        id: 1,
        opticalCenterFix: 0.2
    },
    {
        name: "View 2",
        component: DifferentContent,
        id: 2,
        opticalCenterFix: 0.2
    },
    {
        name: "View 3",
        component: EvenMoreDifferent,
        id: 3,
        opticalCenterFix: 0.2
    },
    {
        name: "Welcome Text",
        component: DevWelcomeText,
        id: 4,
        opticalCenterFix: 0.2
    },
];

const theme = DEFAULT_THEME;
// this seems so..... complex?? 
// but i guess it's fine
export const panels: VizSubPanel[] = [
    {
        id: "viz-test-content",
        views: views.filter(view => view.id === 2),
        header: true,
        maxSize: 100,
        minSize: 10,
        paneKeyId: generateKeyId(),
        childs: {
            parentSubPanel: {
                id: "viz-internal-subpanel-sp-xBdVNtbU5nx60sZ1",
                paneKeyId: generateKeyId(16),
                smoothExpand: false,
                minSize: 10,
                size: 25,
                maxSize: 100
            },
            parentPanel: {
                id: "viz-internal-panel-sp-3jaetpKK8Tc0F8iB",
                horizontal: true,
                keyId: generateKeyId(16),
                theme,
                style: "height: 100%",
                pushOtherPanes: true,
            },
            subPanel: [
                {
                    id: "viz-test-again",
                    maxSize: 100,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.id === 3)
                }
            ]
        }
    },
    {
        id: "viz-test-something-else",
        views: views.filter(view => view.id === 1),
        header: true,
        maxSize: 100,
        minSize: 10,
        paneKeyId: generateKeyId(),
        childs: {
            parentSubPanel: {
                id: "viz-internal-subpanel-sp-6H9DnoBrjh8fCEGy",
                smoothExpand: false,
                minSize: 10,
                maxSize: 100
            },
            parentPanel: {
                id: "viz-internal-panel-sp-Q4AYXfRIaHmchHmX",
                horizontal: true,
                theme,
                style: "height: 100%",
                pushOtherPanes: true,
            },
            subPanel: [
                {
                    id: "viz-welcome-text-3",
                    maxSize: 100,
                    paneKeyId: generateKeyId(),
                    views: views.filter(view => view.id === 4)
                }
            ]
        }
    },
];