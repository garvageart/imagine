import type { InternalPanelContainer, InternalSubPanelContainer, Content, SubPanelChilds } from "$lib/components/panels/SubPanel.svelte";
import { DEFAULT_THEME } from "$lib/constants";
import { generateKeyId } from "$lib/utils";
import type VizView from "$lib/views/views.svelte";

interface IVizSubPanelDataOptions {
    id?: string;
    subPanels: Content[];
    size?: number;
    minSize?: number;
    maxSize?: number;
}

const theme = DEFAULT_THEME;
class VizSubPanelData {
    id: string;
    paneKeyId: string;
    childs: SubPanelChilds;
    views: VizView[];
    subPanels: Content[];
    internalSubPanelContainer: InternalSubPanelContainer;
    internalPanelContainer: InternalPanelContainer;

    constructor(opts: IVizSubPanelDataOptions) {
        this.paneKeyId = generateKeyId(16);
        this.id = opts.id ?? this.paneKeyId;

        this.internalSubPanelContainer = {
            id: "viz-internal-subpanel-" + this.paneKeyId,
            paneKeyId: this.paneKeyId,
            smoothExpand: false,
            minSize: opts.minSize ?? 10,
            size: opts.size,
            maxSize: opts.maxSize ?? 100
        };
        this.internalPanelContainer = {
            id: "viz-internal-panel-" + this.paneKeyId,
            horizontal: true,
            keyId: generateKeyId(16),
            theme,
            style: "height: 100%",
            pushOtherPanes: true,
        };
        this.subPanels = opts.subPanels;
        this.childs = {
            internalSubPanelContainer: this.internalSubPanelContainer,
            internalPanelContainer: this.internalPanelContainer,
            content: this.subPanels.map((sub) => {
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


        this.views = this.subPanels.flatMap((sub) => sub.views);
        if (!this.views) {
            throw new Error("Viz: No views provided in subpanel. Please provide at least one view");
        }
    }
}

export default VizSubPanelData;