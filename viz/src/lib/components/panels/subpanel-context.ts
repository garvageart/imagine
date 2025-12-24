import type VizView from "$lib/views/views.svelte";
import { findSubPanel, generateKeyId } from "$lib/utils/layout";
import { layoutState, layoutTree } from "$lib/third-party/svelte-splitpanes/state.svelte";
import type { MenuItem } from "$lib/context-menu/types";

export type TabHandlers = {
    closeTab: (v: VizView) => void;
    closeOtherTabs: (v: VizView) => void;
    closeTabsToRight: (v: VizView) => void;
    splitRight: (v: VizView) => void;
    splitDown: (v: VizView) => void;
    moveToPanel: (v: VizView, dir: string) => void;
    closeAllTabs: () => void;
    toggleTabLock: (v: VizView) => void;
};

/**
 * Build the context menu items for a tab.
 */
export function buildTabContextMenu(view: VizView, panelViews: VizView[], keyId: string, handlers: TabHandlers): MenuItem[] {
    const viewIndex = panelViews.findIndex((v) => v.id === view.id);
    const isLastTab = viewIndex === panelViews.length - 1;
    const isOnlyTab = panelViews.length === 1;

    // Determine move availability by resolving the view's actual parent
    const viewParentId = (view.parent ?? keyId) as any;
    const viewLoc = findSubPanel("paneKeyId", viewParentId);
    let canMoveLeft = false,
        canMoveRight = false,
        canMoveUp = false,
        canMoveDown = false;

    if (viewLoc) {
        const { parentIndex, childIndex, isChild } = viewLoc as any;
        const currentPanel = layoutState.tree[parentIndex];

        const resolvedChildIndex = isChild
            ? childIndex
            : (currentPanel.childs?.content?.findIndex((c: any) => c.views.some((v: any) => v.id === view.id)) ?? -1);

        canMoveLeft = resolvedChildIndex > 0;
        canMoveRight = resolvedChildIndex !== -1 && resolvedChildIndex < (currentPanel.childs?.content?.length ?? 0) - 1;
        if (!canMoveLeft && parentIndex > 0) canMoveLeft = true;
        if (!canMoveRight && parentIndex < layoutState.tree.length - 1) canMoveRight = true;
        const panelsAreStacked = !!layoutTree.horizontal;
        canMoveUp = panelsAreStacked && parentIndex > 0;
        canMoveDown = panelsAreStacked && parentIndex < layoutState.tree.length - 1;
    }

    const items: MenuItem[] = [
        {
            id: view.locked ? "unlock-tab" : "lock-tab",
            label: view.locked ? "Unlock Tab" : "Lock Tab",
            action: () => handlers.toggleTabLock(view),
            icon: view.locked ? "lock_open" : "lock",
            danger: false
        },
        { id: "separator-lock", label: "", separator: true },
        {
            id: "close",
            label: "Close Tab",
            action: () => handlers.closeTab(view),
            icon: "close",
            shortcut: "Ctrl+W",
            disabled: view.locked
        },
        {
            id: "close-others",
            label: "Close Other Tabs",
            action: () => handlers.closeOtherTabs(view),
            icon: "tab_close",
            disabled: view.locked || isOnlyTab
        },
        {
            id: "close-right",
            label: "Close Tabs to the Right",
            action: () => handlers.closeTabsToRight(view),
            icon: "close_fullscreen",
            disabled: view.locked || isLastTab || isOnlyTab
        },
        { id: "separator1", label: "", separator: true },
        {
            id: "split-right",
            label: "Split Right",
            action: () => handlers.splitRight(view),
            icon: "vertical_split",
            disabled: view.locked
        },
        {
            id: "split-down",
            label: "Split Down",
            action: () => handlers.splitDown(view),
            icon: "horizontal_split",
            disabled: view.locked
        },
        { id: "separator2", label: "", separator: true },
        {
            id: "move-to-group",
            label: "Move to Group",
            icon: "open_with",
            children: [
                {
                    id: "move-left",
                    label: "Move Left",
                    action: () => handlers.moveToPanel(view, "left"),
                    icon: "arrow_back",
                    disabled: view.locked || !canMoveLeft
                },
                {
                    id: "move-right",
                    label: "Move Right",
                    action: () => handlers.moveToPanel(view, "right"),
                    icon: "arrow_forward",
                    disabled: view.locked || !canMoveRight
                },
                {
                    id: "move-up",
                    label: "Move Above",
                    action: () => handlers.moveToPanel(view, "up"),
                    icon: "arrow_upward",
                    disabled: view.locked || !canMoveUp
                },
                {
                    id: "move-down",
                    label: "Move Below",
                    action: () => handlers.moveToPanel(view, "down"),
                    icon: "arrow_downward",
                    disabled: view.locked || !canMoveDown
                }
            ]
        },
        { id: "separator3", label: "", separator: true },
        {
            id: "close-all",
            label: "Close All Tabs",
            action: () => handlers.closeAllTabs(),
            icon: "cancel_presentation",
            danger: true
        }
    ];

    return items;
}

/**
 * Build the layout-level context menu for locking/unlocking the entire layout.
 */
export function buildLayoutContextMenu(): MenuItem[] {
    const items: MenuItem[] = [
        {
            id: layoutTree.locked ? "unlock-layout" : "lock-layout",
            label: layoutTree.locked ? "Unlock Layout" : "Lock Layout",
            action: () => {
                layoutTree.locked = !layoutTree.locked;
                for (const p of layoutState.tree) {
                    p.locked = layoutTree.locked as any;
                }
            },
            icon: layoutTree.locked ? "lock_open" : "lock",
            danger: false
        }
    ];

    return items;
}

/**
 * Build a panel-level context menu for actions that target a single panel.
 * This will add a toggle action to lock/unlock every tab in the specified panel.
 */
export function buildPanelContextMenu(keyId: string, panelViews: VizView[]): MenuItem[] {
    const allLocked = panelViews.length > 0 && panelViews.every((v) => v.locked === true);

    const items: MenuItem[] = [
        {
            id: allLocked ? "unlock-all-tabs-panel" : "lock-all-tabs-panel",
            label: allLocked ? "Unlock All Tabs in Panel" : "Lock All Tabs in Panel",
            action: () => {
                // Toggle all views' locked state for this panel
                const result = findSubPanel("paneKeyId", keyId);
                const nextLocked = !allLocked;
                if (result && result.subPanel) {
                    const sp = result.subPanel as any;
                    if (sp.views && Array.isArray(sp.views)) {
                        for (const v of sp.views) {
                            if (v) v.locked = nextLocked;
                        }
                    }

                    // Persist to layoutState backing structure so the change is reflected
                    const { parentIndex, childIndex, isChild } = result as any;
                    if (typeof parentIndex === "number") {
                        if (!isChild) {
                            layoutState.tree[parentIndex].views = sp.views;
                        } else {
                            layoutState.tree[parentIndex].childs.content[childIndex].views = sp.views;
                        }
                    }
                }
            },
            icon: allLocked ? "lock_open" : "lock",
            danger: false
        }
    ];

    return items;
}

export default { buildTabContextMenu, buildLayoutContextMenu };
