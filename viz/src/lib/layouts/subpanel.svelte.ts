import { dev } from "$app/environment";
import type { Content as IContent, SubPanelChilds } from "$lib/components/panels/SubPanel.svelte";
import { DEFAULT_THEME } from "$lib/constants";
import type { Pane } from "$lib/third-party/svelte-splitpanes";
import { findSubPanel, generateKeyId } from "$lib/utils/layout";
import type VizView from "$lib/views/views.svelte";
import type { ComponentProps } from "svelte";

interface VizSubPanelDataOptions {
    id?: string;
    keyId?: string;
    content: Content[];
    size?: number;
    minSize?: number;
    maxSize?: number;
    class?: string;
    locked?: boolean;
}

interface ContentOptions {
    id?: string;
    views: VizView[];
    paneKeyId?: string;
    size?: number;
    minSize?: number;
    maxSize?: number;
    locked?: boolean;
    isActive?: boolean;
}

class Content implements IContent {
    id?: string;
    views: VizView[];
    paneKeyId?: string;
    size?: number | null = null;
    minSize?: number;
    maxSize?: number;
    locked?: boolean;
    isActive?: boolean = $state(false);


    constructor(opts: ContentOptions) {
        this.id = opts.id;
        this.views = opts.views;
        this.paneKeyId = opts.paneKeyId ?? generateKeyId(10);
        if (!this.id) {
            this.id = `viz-content-${this.paneKeyId}`;
        }
        this.size = opts.size;
        this.minSize = opts.minSize ?? 10;
        this.maxSize = opts.maxSize ?? 100;
        this.locked = opts.locked ?? false;
        this.isActive = opts.isActive ?? false;

        if (!this.views.length) {
            throw new Error("Viz: No views provided in subpanel content. Please provide at least one view");
        }
    }
}

const theme = DEFAULT_THEME;
/**
 * Manages sub-panel data and layout configuration for visualization components in a split-pane layout system.
 * This class is responsible for creating and managing configurable panel sections that can contain multiple views.
 * It works in conjunction with the svelte-splitpanes library to create resizable panel layouts.
 * 
 * @class
 * @implements {Omit<ComponentProps<typeof Pane>, "children" | "snapSize">}
 * 
 * @property {string} id - Unique identifier for the sub-panel
 * @property {string} paneKeyId - Generated or provided key identifier for the pane
 * @property {SubPanelChilds} childs - State-managed child components containing:
 *   - internalSubPanelContainer: Configuration for the internal sub-panel
 *   - internalPanelContainer: Configuration for the panel container including theme
 *   - content: Array of content configurations with views
 * @property {VizView[]} views - State-managed array of visualization views
 * @property {number | undefined} size - Optional size of the panel
 * @property {number} minSize - Minimum size constraint (defaults to 10)
 * @property {number} maxSize - Maximum size constraint (defaults to 100)
 * @property {string | undefined} class - Optional CSS class name
 * 
 * @throws {Error} When no views are provided in the subpanel initialization
 * 
 * @example
 * // Creating a layout with multiple sub-panels and views
 * const panels = [
 *     new VizSubPanelData({
 *         id: "viz-panel-1",
 *         minSize: 10,
 *         maxSize: 100,
 *         content: [
 *             {
 *                 id: "viz-content-1",
 *                 views: views.filter(view => view.id === 2),
 *             },
 *             {
 *                 id: "viz-content-2",
 *                 views: views.filter(view => view.id === 3)
 *             }
 *         ]
 *     }),
 *     new VizSubPanelData({
 *         id: "viz-panel-2",
 *         content: [
 *             {
 *                 id: "viz-clock",
 *                 views: views.filter(view => view.id === 1),
 *             }
 *         ]
 *     })
 * ];
 * 
 * // Making a view active
 * panels[0].makeViewActive(someView);
 * 
 * @remarks
 * - Uses the svelte-splitpanes library for split pane functionality
 * - Supports state management for views and child components
 * - Implements theme customization through CSS variables
 * - Handles view activation and updates through internal methods
 * - Creates a hierarchical structure of panels and sub-panels
 */
class VizSubPanelData implements Omit<ComponentProps<typeof Pane>, "children" | "snapSize"> {
    id: string;
    paneKeyId: string;
    // @ts-ignore
    public childs: SubPanelChilds = $state();
    public views: VizView[] = $state([]);
    size: number | undefined;
    minSize: number;
    maxSize: number;
    class?: string | undefined;
    locked: boolean = $state(false);

    constructor(opts: VizSubPanelDataOptions) {
        this.paneKeyId = opts.keyId ?? generateKeyId(16);
        this.id = opts.id ?? this.paneKeyId;
        this.size = opts.size;
        this.minSize = opts.minSize ?? 10;
        this.maxSize = opts.maxSize ?? 100;
        this.class = opts.class;
        this.locked = opts.locked ?? false;
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
                const paneKeyId = sub.paneKeyId ?? generateKeyId(10);
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

    toJSON() {
        return {
            id: this.id,
            keyId: this.paneKeyId,
            size: this.size,
            minSize: this.minSize,
            maxSize: this.maxSize,
            class: this.class,
            content: this.childs.content,
            locked: this.locked
        };
    }

    static fromJSON(json: any) {
        const panel = new VizSubPanelData({
            id: json.id,
            keyId: json.keyId,
            content: json.content,
            size: json.size,
            minSize: json.minSize,
            maxSize: json.maxSize,
            class: json.class,
            locked: json.locked
        });
        return panel;
    }

    /**
     * Updates the active view of a subpanel based on the given key ID.
     *
     * Finds the subpanel associated with the provided key ID
     * and sets the specified view as the active view. If there is a current
     * active view, it is deactivated before activating the new view.
     * The views array in the subpanel is updated to ensure that the new
     * active view is correctly reflected.
     *
     * @param view - The view to be set as the active view.
     */
    public makeViewActive(view: VizView) {
        this.updateSubPanelActiveView(view);
    }

    private updateSubPanelActiveView(view: VizView) {
        const subPanel = findSubPanel("paneKeyId", this.paneKeyId)?.subPanel;

        if (!subPanel) {
            if (dev) {
                // Something has gone so badly wrong that we can't find the subpanel
                throw new Error("Viz: Subpanel not found");
            }

            console.error("Viz: Subpanel not found");
            return;
        }

        subPanel.views.splice(
            subPanel.views.findIndex((spview) => spview.id === view.id),
            1,
            view
        );
    }
}

export { Content };
export default VizSubPanelData;