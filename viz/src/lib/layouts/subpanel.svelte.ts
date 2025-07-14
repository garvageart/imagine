import type { Content, SubPanelChilds } from "$lib/components/panels/SubPanel.svelte";
import { DEFAULT_THEME } from "$lib/constants";
import type { Pane } from "$lib/third-party/svelte-splitpanes";
import { generateKeyId } from "$lib/utils";
import type VizView from "$lib/views/views.svelte";
import type { ComponentProps } from "svelte";

interface VizSubPanelDataOptions {
    id?: string;
    content: Content[];
    size?: number;
    minSize?: number;
    maxSize?: number;
    class?: string;
}

const theme = DEFAULT_THEME;
class VizSubPanelData implements Omit<ComponentProps<typeof Pane>, "children" | "snapSize"> {
    id: string;
    paneKeyId: string;
    // @ts-ignore
    childs: SubPanelChilds = $state();
    views: VizView[] = $state([]);
    size: number | undefined;
    minSize: number;
    maxSize: number;
    class?: string | undefined;

    constructor(opts: VizSubPanelDataOptions) {
        this.paneKeyId = generateKeyId(16);
        this.id = opts.id ?? this.paneKeyId;
        this.size = opts.size;
        this.minSize = opts.minSize ?? 10;
        this.maxSize = opts.maxSize ?? 100;
        this.class = opts.class;

        const internalPanelKeyId = generateKeyId(16);

        this.childs = {
            internalSubPanelContainer: {
                id: "viz-internal-subpanel-" + this.paneKeyId,
                paneKeyId: this.paneKeyId,
                smoothExpand: false,
                minSize: opts.minSize ?? 10,
                size: opts.size,
                maxSize: opts.maxSize ?? 100
            },
            internalPanelContainer: {
                id: "viz-internal-panel-" + internalPanelKeyId,
                horizontal: true,
                keyId: internalPanelKeyId,
                theme,
                style: "height: 100%",
                pushOtherPanes: true,
            },
            content: opts.content.map((sub) => {
                const paneKeyId = generateKeyId(10);
                const id = sub.id ?? `viz-subpanel-${paneKeyId}`;
                return {
                    ...sub,
                    id,
                    paneKeyId: paneKeyId,
                    size: sub.size,
                    minSize: sub.minSize ?? 10,
                    maxSize: sub.maxSize ?? 100
                };
            }),
        };

        this.views = this.childs.content.flatMap((sub) => sub.views);

        if (!this.views.length) {
            throw new Error("Viz: No views provided in subpanel. Please provide at least one view");
        }
    }
}

export default VizSubPanelData;