import { generateKeyId } from "$lib/utils/layout";
import VizView, { type SerializedVizView } from "$lib/views/views.svelte";
import { views as viewRegistry } from "$lib/layouts/views";

// Types
export type Orientation = "horizontal" | "vertical";

export interface SerializedSplitNode {
	type: "split";
	id: string;
	orientation: Orientation;
	size: number;
	locked: boolean;
	children: (SerializedSplitNode | SerializedTabGroup)[];
}

export interface SerializedTabGroup {
	type: "tab-group";
	id: string;
	size: number;
	locked: boolean;
	activeViewId?: number;
	views: SerializedVizView[];
}

export type SerializedNode = SerializedSplitNode | SerializedTabGroup;

export interface SerializedWorkspace {
	root: SerializedNode;
	activeGroupId?: string;
}

/**
 * Helper: match svelte-kit style dynamic routes like "/collections/[uid]" to concrete paths
 */
function pathMatches(
	pattern: string | undefined,
	actual: string | undefined
): boolean {
	if (!pattern || !actual) {
		return false;
	}

	if (pattern === actual) {
		return true;
	}

	// Escape regex specials, then turn dynamic segments \[param\] into [^/]+
	const escaped = pattern
		.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
		.replace(/\\\[[^\]]+\\\]/g, "[^/]+");
	const re = new RegExp("^" + escaped + "$");
	return re.test(actual);
}

// --- Classes ---

/**
 * Represents a group of tabs (Leaf Node)
 */
export class TabGroup {
	readonly type = "tab-group";
	id: string;
	size: number = $state(100);
	locked: boolean = $state(false);
	views: VizView[] = $state([]);
	activeViewId: number | undefined = $state();
	parent: SplitNode | null = null;

	constructor(opts: {
		id?: string;
		size?: number;
		views?: VizView[];
		activeViewId?: number;
		locked?: boolean;
	}) {
		this.id = opts.id ?? generateKeyId(10);
		this.size = opts.size ?? 100;
		this.locked = opts.locked ?? false;
		if (opts.views) {
			this.views = opts.views;
			this.activeViewId = opts.activeViewId;

			if (
				this.activeViewId &&
				!this.views.find((v) => v.id === this.activeViewId)
			) {
				this.activeViewId = undefined;
			}
			if (!this.activeViewId && this.views.length > 0) {
				this.setActive(this.views[0].id);
			}
		}
	}

	get activeView() {
		return this.views.find((v) => v.id === this.activeViewId);
	}

	addTab(view: VizView, index?: number) {
		if (index !== undefined && index >= 0 && index <= this.views.length) {
			this.views.splice(index, 0, view);
		} else {
			this.views.push(view);
		}
		this.setActive(view.id);
	}

	removeTab(viewId: number) {
		const idx = this.views.findIndex((v) => v.id === viewId);
		if (idx === -1) return;

		this.views.splice(idx, 1);

		if (this.activeViewId === viewId) {
			if (this.views.length > 0) {
				const newIdx = Math.max(0, idx - 1);
				this.setActive(this.views[newIdx].id);
			} else {
				this.activeViewId = undefined;
			}
		}
	}

	setActive(viewId: number) {
		const view = this.views.find((v) => v.id === viewId);
		if (view) {
			this.views.forEach((v) => v.setActive(false));
			view.setActive(true);
			this.activeViewId = viewId;
		}
	}

	toJSON(): SerializedTabGroup {
		return {
			type: "tab-group",
			id: this.id,
			size: this.size,
			locked: this.locked,
			activeViewId: this.activeViewId,
			views: this.views.map((v) => v.toJSON())
		};
	}

	static fromJSON(json: SerializedTabGroup): TabGroup {
		const hydratedViews = json.views
			.map((serializedView) => {
				// Try to find by path (supports dynamic segments), then by name
				const matchedView = viewRegistry.find((registeredView) => {
					if (
						serializedView.path &&
						registeredView.path &&
						pathMatches(registeredView.path, serializedView.path)
					) {
						return true;
					}
					return registeredView.name === serializedView.name;
				});

				if (!matchedView?.component) {
					console.warn(
						`[Workspace] Could not find component for view: ${serializedView.name} (path: ${serializedView.path})`
					);
					return null;
				}

				return VizView.fromJSON(serializedView, matchedView.component);
			})
			.filter((v) => v !== null) as VizView[];

		return new TabGroup({
			id: json.id,
			size: json.size,
			locked: json.locked ?? false,
			activeViewId: json.activeViewId,
			views: hydratedViews
		});
	}
}

/**
 * Represents a split container (Internal Node)
 */
export class SplitNode {
	readonly type = "split";
	id: string;
	orientation: Orientation = $state("horizontal");
	size: number = $state(100);
	locked: boolean = $state(false);
	children: (SplitNode | TabGroup)[] = $state([]);
	parent: SplitNode | null = null;

	constructor(opts: {
		id?: string;
		orientation?: Orientation;
		size?: number;
		children?: (SplitNode | TabGroup)[];
		locked?: boolean;
	}) {
		this.id = opts.id ?? generateKeyId(10);
		this.orientation = opts.orientation ?? "horizontal";
		this.size = opts.size ?? 100;
		this.locked = opts.locked ?? false;
		if (opts.children) {
			this.children = opts.children;
			this.children.forEach((c) => (c.parent = this));
		}
	}

	addChild(node: SplitNode | TabGroup, index?: number) {
		node.parent = this;
		if (index !== undefined && index >= 0 && index <= this.children.length) {
			this.children.splice(index, 0, node);
		} else {
			this.children.push(node);
		}
		this.normalizeSizes();
	}

	removeChild(node: SplitNode | TabGroup) {
		const idx = this.children.indexOf(node);
		if (idx !== -1) {
			this.children.splice(idx, 1);
			node.parent = null;
			this.normalizeSizes();
		}
	}

	replaceChild(oldNode: SplitNode | TabGroup, newNode: SplitNode | TabGroup) {
		const idx = this.children.indexOf(oldNode);
		if (idx !== -1) {
			newNode.size = oldNode.size; // Inherit size
			newNode.parent = this;
			this.children[idx] = newNode;
			oldNode.parent = null;
		}
	}

	normalizeSizes() {
		if (this.children.length === 0) return;
		const sizePerChild = 100 / this.children.length;
		this.children.forEach(child => {
			child.size = sizePerChild;
		});
	}

	toJSON(): SerializedSplitNode {
		return {
			type: "split",
			id: this.id,
			orientation: this.orientation,
			size: this.size,
			locked: this.locked,
			children: this.children.map((c) => c.toJSON())
		};
	}

	static fromJSON(json: SerializedSplitNode): SplitNode {
		const node = new SplitNode({
			id: json.id,
			orientation: json.orientation,
			size: json.size,
			locked: json.locked ?? false
		});

		node.children = json.children.map((c) => {
			if (c.type === "split") {
				return SplitNode.fromJSON(c as SerializedSplitNode);
			}
			return TabGroup.fromJSON(c as SerializedTabGroup);
		});

		node.children.forEach((c) => (c.parent = node));
		return node;
	}
}

/**
 * The Root Workspace Manager
 */
export class Workspace {
	root: SplitNode | TabGroup = $state(new TabGroup({}));
	activeGroupId: string | undefined = $state();

	constructor(initialRoot?: SplitNode | TabGroup) {
		if (initialRoot) {
			this.root = initialRoot;
		}
		// Auto-select first group as active if none set
		if (!this.activeGroupId) {
			const firstGroup = this.findFirstGroup(this.root);
			if (firstGroup) {
				this.activeGroupId = firstGroup.id;
			}
		}
	}

	private findFirstGroup(node: SplitNode | TabGroup): TabGroup | null {
		if (node instanceof TabGroup) return node;
		// Use a simple depth-first search to find the first TabGroup
		for (const child of node.children) {
			const found = this.findFirstGroup(child);
			if (found) return found;
		}
		return null;
	}

	setActiveGroup(groupId: string) {
		this.activeGroupId = groupId;
	}

	get activeGroup(): TabGroup | null {
		if (!this.activeGroupId) return null;
		const node = this.findNode(this.activeGroupId);
		return node instanceof TabGroup ? node : null;
	}

	/**
	 * Finds a node by ID in the tree
	 */
	findNode(
		id: string,
		startNode: SplitNode | TabGroup = this.root
	): SplitNode | TabGroup | null {
		if (startNode.id === id) {
			return startNode;
		}

		if (startNode instanceof SplitNode) {
			for (const child of startNode.children) {
				const found = this.findNode(id, child);
				if (found) {
					return found;
				}
			}
		}
		return null;
	}

	/**
	 * Finds the TabGroup containing a specific view ID
	 */
	findGroupWithView(
		viewId: number,
		startNode: SplitNode | TabGroup = this.root
	): TabGroup | null {
		if (startNode instanceof TabGroup) {
			if (startNode.views.some((v) => v.id === viewId)) {
				return startNode;
			}
		} else if (startNode instanceof SplitNode) {
			for (const child of startNode.children) {
				const found = this.findGroupWithView(viewId, child);
				if (found) {
					return found;
				}
			}
		}
		return null;
	}

	/**
	 * Finds the TabGroup containing a specific view path
	 */
	findGroupWithPath(
		path: string,
		startNode: SplitNode | TabGroup = this.root
	): TabGroup | null {
		if (startNode instanceof TabGroup) {
			if (startNode.views.some((v) => v.path === path)) {
				return startNode;
			}
		} else if (startNode instanceof SplitNode) {
			for (const child of startNode.children) {
				const found = this.findGroupWithPath(path, child);
				if (found) {
					return found;
				}
			}
		}
		return null;
	}

	/**
	 * Finds a view by its path
	 */
	findViewWithPath(
		path: string,
		startNode: SplitNode | TabGroup = this.root
	): VizView | null {
		if (startNode instanceof TabGroup) {
			const view = startNode.views.find((v) => v.path === path);
			if (view) {
				return view;
			}
		} else if (startNode instanceof SplitNode) {
			for (const child of startNode.children) {
				const found = this.findViewWithPath(path, child);
				if (found) {
					return found;
				}
			}
		}
		return null;
	}

	/**
	 * Moves a tab from its current group to a target group
	 */
	moveTab(viewId: number, targetGroupId: string, index?: number) {
		const sourceGroup = this.findGroupWithView(viewId);
		const targetGroup = this.findNode(targetGroupId);

		if (!sourceGroup || !(targetGroup instanceof TabGroup) || sourceGroup.locked || targetGroup.locked) {
			return; // Cannot move from/to a locked group
		}

		const view = sourceGroup.views.find((v) => v.id === viewId);
		if (!view || view.locked) return; // Cannot move a locked view

		// Check if the tab is already in the target group at the desired position
		if (targetGroup === sourceGroup) {
			const currentIdx = targetGroup.views.indexOf(view);
			if (index !== undefined && currentIdx === index) {
				return; // Already in place
			}
		}

		sourceGroup.removeTab(viewId);
		targetGroup.addTab(view, index);

		this.cleanupNode(sourceGroup); // Cleanup source group if it becomes empty
	}

	/**
	 * Removes a node from the tree and handles layout cleanup
	 */	private cleanupNode(node: SplitNode | TabGroup) {
		const parent = node.parent;
		if (!parent) {
			// If it's the root and the only child is also gone, reset to a default empty group
			if (this.root === node && (node instanceof TabGroup && node.views.length === 0)) {
				this.root = new TabGroup({});
			}
			return; // Cannot remove root unless it's empty
		}

		// If the node we're cleaning is an empty tab group, remove it.
		if (node instanceof TabGroup && node.views.length === 0) {
			parent.removeChild(node);
			this.cleanupNode(parent); // Recurse up the tree
		}
		// If the node is a SplitNode with only one child, collapse it.
		else if (node instanceof SplitNode && node.children.length === 1) {
			const grandParent = node.parent;
			const remainingChild = node.children[0];
			if (grandParent) {
				grandParent.replaceChild(node, remainingChild);
			} else {
				// This node was the root
				this.root = remainingChild;
				remainingChild.parent = null;
			}
		} else if (node instanceof SplitNode && node.children.length === 0) {
			// This shouldn't happen, but if it does, remove the empty split node.
			this.cleanupNode(node);
		}
	}

	/**
	 * Splits a group and moves a view into the newly created group.
	 * If the source group becomes empty, it's cleaned up.
	 */
	splitGroup(
		groupId: string,
		viewToMove: VizView,
		direction: "left" | "right" | "top" | "bottom"
	) {
		const targetGroup = this.findNode(groupId);
		if (!(targetGroup instanceof TabGroup) || targetGroup.locked) return; // Cannot split a locked group
		if (viewToMove.locked) return; // Cannot split a locked view

		const parent = targetGroup.parent;
		const orientation: Orientation =
			direction === "left" || direction === "right" ? "horizontal" : "vertical";
		const isAfter = direction === "right" || direction === "bottom";

		const sourceGroup = this.findGroupWithView(viewToMove.id);
		// If the view is already in the targetGroup, just activate it and return.
		if (sourceGroup === targetGroup) {
			targetGroup.setActive(viewToMove.id);
			return;
		}

		// Remove view from source first
		if (sourceGroup) {
			sourceGroup.removeTab(viewToMove.id);
		}

		const newGroup = new TabGroup({ views: [viewToMove] });
		newGroup.setActive(viewToMove.id);
		this.setActiveGroup(newGroup.id);

		if (parent && parent.orientation === orientation) {
			// If the parent is a SplitNode with the same orientation,
			// just add the new group as a sibling.
			const idx = parent.children.indexOf(targetGroup);
			parent.addChild(newGroup, isAfter ? idx + 1 : idx);
		} else {
			// If targetGroup is root, or parent has different orientation,
			// replace the targetGroup with a new SplitNode.
			const newSplit = new SplitNode({
				orientation,
				children: isAfter ? [targetGroup, newGroup] : [newGroup, targetGroup]
			});

			if (parent) {
				parent.replaceChild(targetGroup, newSplit);
			} else {
				this.root = newSplit; // New SplitNode becomes the root
				newSplit.parent = null;
			}
			// Normalize sizes for the newly created split
			newSplit.normalizeSizes();
		}

		if (sourceGroup) {
			this.cleanupNode(sourceGroup); // Cleanup source group if it becomes empty
		}
	}

	/**
	 * Returns all TabGroups in the tree
	 */
	getAllTabGroups(
		startNode: SplitNode | TabGroup = this.root
	): TabGroup[] {
		if (startNode instanceof TabGroup) return [startNode];
		return startNode.children.flatMap((child) => this.getAllTabGroups(child));
	}

	toJSON(): SerializedWorkspace {
		return {
			root: this.root.toJSON(),
			activeGroupId: this.activeGroupId
		};
	}

	static fromJSON(json: SerializedWorkspace): Workspace {
		let root: SplitNode | TabGroup;
		if (json.root.type === "split") {
			root = SplitNode.fromJSON(json.root as SerializedSplitNode);
		} else {
			root = TabGroup.fromJSON(json.root as SerializedTabGroup);
		}
		const workspace = new Workspace(root);
		workspace.activeGroupId = json.activeGroupId;
		return workspace;
	}
}