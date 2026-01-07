import type { MenuItem } from "$lib/context-menu/types";
import { TabGroup } from "$lib/layouts/model.svelte";
import { workspaceState } from "$lib/states/workspace.svelte";
import type VizView from "$lib/views/views.svelte";

export type TabHandlers = {
	closeTab: (v: VizView) => void;
	closeOtherTabs: (v: VizView) => void;
	closeTabsToRight: (v: VizView) => void;
	closeAllTabs: () => void;
	closePanel: () => void;
	toggleTabLock: (v: VizView) => void;
	splitRight: (v: VizView) => void;
	splitDown: (v: VizView) => void;
};

export function buildTabContextMenu(
	view: VizView,
	group: TabGroup,
	handlers: TabHandlers
): MenuItem[] {
	const viewIndex = group.views.findIndex((v) => v.id === view.id);
	const isLastTab = viewIndex === group.views.length - 1;
	const isOnlyTab = group.views.length === 1;

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
		{ id: "separator3", label: "", separator: true },
		{
			id: "close-all",
			label: "Close All Tabs in this Group",
			action: () => handlers.closeAllTabs(),
			icon: "cancel_presentation",
			danger: true,
			disabled: group.locked
		}
	];

	return items;
}

export function buildLayoutContextMenu(): MenuItem[] {
	const workspace = workspaceState.workspace;
	if (!workspace) return [];
	const isLayoutLocked = workspace.root.locked;

	return [
		{
			id: isLayoutLocked ? "unlock-layout" : "lock-layout",
			label: isLayoutLocked ? "Unlock Layout" : "Lock Layout",
			action: () => {
				if (workspace) {
					workspace.root.locked = !isLayoutLocked;
					// Optionally lock all children nodes as well
				}
			},
			icon: isLayoutLocked ? "lock_open" : "lock"
		}
	];
}

export function buildPanelContextMenu(group: TabGroup, handlers?: TabHandlers): MenuItem[] {
	const allLocked = group.views.length > 0 && group.views.every((v) => v.locked);
	const nextLockedState = !allLocked;
	const isMaximized = workspaceState.workspace?.maximizedGroupId === group.id;

	const items: MenuItem[] = [
		{
			id: `maximize-panel-${group.id}`,
			label: isMaximized ? "Restore Panel" : "Maximize Panel",
			action: () => {
				workspaceState.workspace?.toggleMaximize(group.id);
			},
			icon: isMaximized ? "close_fullscreen" : "open_in_full"
		},
		{
			id: `lock-panel-${group.id}`,
			label: group.locked ? "Unlock Panel" : "Lock Panel (Prevent Splits)",
			action: () => {
				group.locked = !group.locked;
			},
			icon: group.locked ? "lock_open" : "lock"
		},
		{
			id: `lock-all-tabs-${group.id}`,
			label: allLocked ? "Unlock All Tabs" : "Lock All Tabs",
			action: () => {
				for (const view of group.views) {
					view.locked = nextLockedState;
				}
			},
			icon: allLocked ? "checklist" : "checklist_rtl"
		}
	];

	if (handlers?.closePanel) {
		items.push({ id: "sep-panel-close", label: "", separator: true });
		items.push({
			id: `close-panel-${group.id}`,
			label: "Close Panel",
			action: () => handlers.closePanel(),
			icon: "cancel_presentation",
			danger: true
		});
	}

	return items;
}