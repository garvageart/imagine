import { workspaceState } from "$lib/states/workspace.svelte";
import { DragData } from "$lib/drag-drop/data";
import { VizMimeTypes } from "$lib/constants";
import { TabGroup } from "./model.svelte";

export interface TabDragData {
	viewId: number;
	sourceGroupId: string;
}

export type DropPosition = "left" | "right" | "top" | "bottom" | "center";

export class TabOps {
	draggable(node: HTMLElement, data: TabDragData) {
		node.draggable = true;

		const onDragStart = (e: DragEvent) => {
			if (!e.dataTransfer) {
				return;
			}

			const dragData = new DragData(VizMimeTypes.TAB_VIEW, data);
			dragData.setData(e.dataTransfer);
			e.dataTransfer.effectAllowed = "move";
		};

		node.addEventListener("dragstart", onDragStart);

		return {
			destroy() {
				node.removeEventListener("dragstart", onDragStart);
			}
		};
	}

	dropTarget(node: HTMLElement, targetGroupId: string) {
		const onDragOver = (e: DragEvent) => {
			e.preventDefault();
			if (!e.dataTransfer) return;
			e.dataTransfer.dropEffect = "move";

			// Show visual feedback (overlay)
			this.updateOverlay(node, e);
		};

		const onDragLeave = () => {
			this.removeOverlay(node);
		};

		const onDrop = (e: DragEvent) => {
			e.preventDefault();
			this.removeOverlay(node);

			if (!e.dataTransfer) return;
			const dragData = DragData.getData<TabDragData>(
				e.dataTransfer,
				VizMimeTypes.TAB_VIEW
			);
			if (!dragData) return;

			const { viewId, sourceGroupId } = dragData.payload;
			const position = this.calculateDropPosition(node, e);

			this.handleDrop(viewId, sourceGroupId, targetGroupId, position);
		};

		node.addEventListener("dragover", onDragOver);
		node.addEventListener("dragleave", onDragLeave);
		node.addEventListener("drop", onDrop);

		return {
			destroy() {
				node.removeEventListener("dragover", onDragOver);
				node.removeEventListener("dragleave", onDragLeave);
				node.removeEventListener("drop", onDrop);
			}
		};
	}

	addToGroup(node: HTMLElement, targetGroupId: string) {
		const onDragOver = (e: DragEvent) => {
			e.preventDefault();
			e.stopPropagation(); // Stop bubbling to parent panel
			if (!e.dataTransfer) {
				return;
			}

			e.dataTransfer.dropEffect = "move";
			node.classList.add("drop-active");
		};

		const onDragLeave = () => {
			node.classList.remove("drop-active");
		};

		const onDrop = (e: DragEvent) => {
			e.preventDefault();
			e.stopPropagation();
			node.classList.remove("drop-active");

			if (!e.dataTransfer) {
				return;
			}

			const dragData = DragData.getData<TabDragData>(
				e.dataTransfer,
				VizMimeTypes.TAB_VIEW
			);

			if (!dragData) return;

			const { viewId } = dragData.payload;
			// Force merge by calling moveTab directly
			const workspace = workspaceState.workspace;
			if (workspace) {
				workspace.moveTab(viewId, targetGroupId);
			}
		};

		node.addEventListener("dragover", onDragOver);
		node.addEventListener("dragleave", onDragLeave);
		node.addEventListener("drop", onDrop);

		return {
			destroy() {
				node.removeEventListener("dragover", onDragOver);
				node.removeEventListener("dragleave", onDragLeave);
				node.removeEventListener("drop", onDrop);
			}
		};
	}

	private calculateDropPosition(
		node: HTMLElement,
		e: DragEvent
	): DropPosition {
		const rect = node.getBoundingClientRect();
		const x = e.clientX - rect.left;
		const y = e.clientY - rect.top;

		const xPct = x / rect.width;
		const yPct = y / rect.height;

		const threshold = 0.2;

		if (xPct < threshold) return "left";
		if (xPct > 1 - threshold) return "right";
		if (yPct < threshold) return "top";
		if (yPct > 1 - threshold) return "bottom";

		return "center";
	}

	private updateOverlay(node: HTMLElement, e: DragEvent) {
		let overlay = node.querySelector(".drop-overlay") as HTMLElement;
		if (!overlay) {
			overlay = document.createElement("div");
			overlay.className = "drop-overlay";
			node.appendChild(overlay);
		}

		const pos = this.calculateDropPosition(node, e);
		overlay.dataset.position = pos;
	}

	private removeOverlay(node: HTMLElement) {
		const overlay = node.querySelector(".drop-overlay");
		if (overlay) overlay.remove();
	}

	private handleDrop(
		viewId: number,
		sourceGroupId: string,
		targetGroupId: string,
		position: DropPosition
	) {
		const workspace = workspaceState.workspace;
		if (!workspace) {
			return;
		}

		const sourceGroup = workspace.findNode(sourceGroupId) as TabGroup;
		const view = sourceGroup?.views.find((v) => v.id === viewId);
		if (!view) {
			return;
		}

		if (position === "center") {
			workspace.moveTab(viewId, targetGroupId);
		} else {
			workspace.splitGroup(targetGroupId, view, position);
		}
	}
}

export const tabOps = new TabOps();
