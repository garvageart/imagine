import VizSubPanelData from "./subpanel.svelte";
import VizView from "$lib/views/views.svelte";
import { Content } from "./subpanel.svelte";

/**
 * Normalizes sizes for all content groups within panels
 * Each content group gets an equal share of 100%
 */
export function normalizeContentSizes(panels: VizSubPanelData[]) {
    panels.forEach((panel) => {
        if (panel.childs && Array.isArray(panel.childs.content) && panel.childs.content.length > 0) {
            const subSize = 100 / panel.childs.content.length;
            panel.childs.content.forEach((content) => {
                content.size = subSize;
            });
        }
    });
}

/**
 * Normalizes sizes for all panels in the layout tree
 * Each panel gets an equal share of 100%
 */
export function normalizePanelSizes(panels: VizSubPanelData[]) {
    if (panels.length > 1) {
        const sizePerPanel = 100 / panels.length;
        panels.forEach((panel) => {
            panel.size = sizePerPanel;
            panel.childs.internalSubPanelContainer.size = sizePerPanel;
        });
    } else if (panels.length === 1) {
        panels[0].size = 100;
        panels[0].childs.internalSubPanelContainer.size = 100;
    }
}

/**
 * Removes empty content groups from a panel and normalizes remaining sizes
 */
export function removeEmptyContent(panel: VizSubPanelData, contentIndex: number): boolean {
    const content = panel.childs.content[contentIndex];
    if (!content || content.views.length > 0) {
        return false;
    }

    if (window.debug) {
        console.log(`Removing empty content group ${content.paneKeyId}`);
    }

    panel.childs.content.splice(contentIndex, 1);

    // Normalize sizes for remaining content groups
    if (panel.childs.content.length > 0) {
        const subSize = 100 / panel.childs.content.length;
        panel.childs.content.forEach((c) => {
            c.size = subSize;
        });
    }

    return true;
}

/**
 * Removes an empty panel from the layout tree and normalizes remaining panel sizes
 */
export function removeEmptyPanel(panels: VizSubPanelData[], panelIndex: number): boolean {
    const panel = panels[panelIndex];
    if (!panel || panel.views.length > 0) {
        return false;
    }

    if (window.debug) {
        console.log(`Removing empty panel ${panel.paneKeyId}`);
    }

    panels.splice(panelIndex, 1);
    normalizePanelSizes(panels);

    return true;
}

/**
 * Cleans up empty content groups and panels after a view is moved or removed
 * Returns true if any cleanup was performed
 */
export function cleanupEmptyPanels(panels: VizSubPanelData[]): boolean {
    let didCleanup = false;

    // First pass: remove empty content groups and normalize their parent panels
    for (let i = panels.length - 1; i >= 0; i--) {
        const panel = panels[i];
        if (!panel.childs.content) continue;

        // Remove empty content groups
        for (let j = panel.childs.content.length - 1; j >= 0; j--) {
            if (panel.childs.content[j].views.length === 0) {
                if (removeEmptyContent(panel, j)) {
                    didCleanup = true;
                }
            }
        }

        // Update panel's views array
        panel.views = panel.childs.content.flatMap((c) => c.views);
    }

    // Second pass: remove empty panels
    for (let i = panels.length - 1; i >= 0; i--) {
        if (panels[i].views.length === 0) {
            if (removeEmptyPanel(panels, i)) {
                didCleanup = true;
            }
        }
    }

    return didCleanup;
}

/**
 * Creates a new view instance with the same properties but a new ID
 */
export function duplicateView(view: VizView): VizView {
    const VizViewClass = view.constructor as typeof VizView;
    return new VizViewClass({
        name: view.name,
        component: view.component,
        path: view.path,
        opticalCenterFix: view.opticalCenterFix,
        isActive: true
    });
}

/**
 * Creates a new panel with a view at a specific position
 */
export function createPanelWithView(view: VizView, size?: number): VizSubPanelData {
    const VizSubPanelDataClass = VizSubPanelData;
    const ContentClass = Content;

    const newPanel = new VizSubPanelDataClass({
        content: [
            new ContentClass({
                views: [view]
            })
        ],
        size
    });

    view.parent = newPanel.childs.content[0].paneKeyId;
    return newPanel;
}

/**
 * Splits a panel vertically by creating a new panel after it
 * The current panel's size is halved and shared with the new panel
 */
export function splitPanelVertically(
    panels: VizSubPanelData[],
    panelIndex: number,
    view: VizView
): VizSubPanelData | null {
    const currentPanel = panels[panelIndex];
    if (!currentPanel) return null;

    // Calculate sizes
    const currentSize = currentPanel.size ?? 50;
    const newSize = currentSize / 2;

    // Update current panel size
    currentPanel.size = newSize;
    currentPanel.childs.internalSubPanelContainer.size = newSize;

    // Create new panel
    const newPanel = createPanelWithView(view, newSize);

    // Insert after current panel
    panels.splice(panelIndex + 1, 0, newPanel);

    return newPanel;
}

/**
 * Splits a panel horizontally by adding a new content group to it
 */
export function splitPanelHorizontally(panel: VizSubPanelData, view: VizView): Content | null {
    if (!panel.childs.content) return null;

    const ContentClass = Content as any;
    const newContent = new ContentClass({
        views: [view]
    });

    view.parent = newContent.paneKeyId;

    // Add new content group
    panel.childs.content.push(newContent);

    // Normalize sizes
    const subSize = 100 / panel.childs.content.length;
    panel.childs.content.forEach((content) => {
        content.size = subSize;
    });

    // Update panel's views array
    panel.views = panel.childs.content.flatMap((c) => c.views);

    return newContent;
}


