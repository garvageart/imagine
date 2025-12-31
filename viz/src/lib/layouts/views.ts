import DifferentContent from "$lib/components/panels/workspace/generic/DifferentContent.svelte";
import SomeContent from "$lib/components/panels/workspace/generic/SomeContent.svelte";
import EvenMoreDifferent from "$lib/components/panels/workspace/generic/EvenMoreDifferent.svelte";
import HistogramPanel from "$lib/components/panels/workspace/HistogramPanel.svelte";
import FilterPanel from "$lib/components/panels/workspace/FilterPanel.svelte";
import ImagePreview from "$lib/components/panels/workspace/ImagePreview.svelte";
import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
import VizView from "$lib/views/views.svelte";
import Collections from "../../routes/(app)/collections/+page.svelte";
import CollectionPage from "../../routes/(app)/collections/[uid]/+page.svelte";

// Only one instance of a view/panel and its component
// can exist in the layout so we declare all of them here
export const views: VizView[] = [
    new VizView({
        name: "Clock",
        component: SomeContent
    }),
    new VizView({
        name: "What Day Is It?",
        component: DifferentContent
        
    }),
    new VizView({
        name: "Timezone",
        component: EvenMoreDifferent
    }),
    new VizView({
        name: "Welcome Text",
        component: DevWelcomeText
    }),
    new VizView({
        name: "Collections",
        component: Collections,
        path: "/collections"
    }),
    new VizView({
        name: "Collection",
        component: CollectionPage,
        path: "/collections/[uid]"
    }),
    new VizView({
        name: "Filter",
        component: FilterPanel,
    }),
    new VizView({
        name: "Preview",
        component: ImagePreview,
    }),
    new VizView({
        name: "Histogram",
        component: HistogramPanel
    })
];