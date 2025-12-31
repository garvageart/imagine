// viz/src/tests/layouts/model.spec.ts

import { describe, it, expect, beforeEach } from 'vitest';
import { Workspace, TabGroup, SplitNode } from '$lib/layouts/model.svelte';
import VizView, { type SerializedVizView } from '$lib/views/views.svelte';
import { views as viewRegistry } from '$lib/layouts/views';

// Helper function to create a dummy Svelte component for testing
const DummyComponent = { render: () => ({}) } as any; // Simplified for testing

describe('Workspace Layout Model', () => {

    // Helper to create a basic view
    const createTestView = (name: string, path?: string, locked: boolean = false) => {
        return new VizView({
            name: name,
            component: DummyComponent,
            path: path,
            locked: locked
        });
    };

    let workspace: Workspace;

    beforeEach(() => {
        // Reset workspace before each test
        workspace = new Workspace();
    });

    it('should initialize with a default empty tab group if no root is provided', () => {
        expect(workspace.root).toBeInstanceOf(TabGroup);
        expect((workspace.root as TabGroup).views).toHaveLength(0);
        expect(workspace.activeGroupId).toBeDefined();
    });

    it('should correctly serialize and deserialize a simple tab group', () => {
        const view1 = createTestView('Test View 1', '/test/view1');
        const tabGroup = new TabGroup({ views: [view1], activeViewId: view1.id });
        workspace = new Workspace(tabGroup);

        const serialized = workspace.toJSON();
        const deserializedWorkspace = Workspace.fromJSON(serialized);

        expect(deserializedWorkspace.root).toBeInstanceOf(TabGroup);
        const deserializedTabGroup = deserializedWorkspace.root as TabGroup;
        expect(deserializedTabGroup.views).toHaveLength(1);
        expect(deserializedTabGroup.views[0].name).toBe('Test View 1');
        expect(deserializedTabGroup.views[0].path).toBe('/test/view1');
        expect(deserializedTabGroup.activeViewId).toBe(view1.id);
        expect(deserializedTabGroup.activeView?.name).toBe('Test View 1');
    });

    it('should correctly serialize and deserialize a split node with tab groups', () => {
        const view1 = createTestView('View A', '/path/a');
        const view2 = createTestView('View B', '/path/b');
        const view3 = createTestView('View C', '/path/c');

        const tabGroup1 = new TabGroup({ id: 'tg1', views: [view1], activeViewId: view1.id });
        const tabGroup2 = new TabGroup({ id: 'tg2', views: [view2], activeViewId: view2.id });
        const tabGroup3 = new TabGroup({ id: 'tg3', views: [view3], activeViewId: view3.id });

        const splitNodeInner = new SplitNode({
            orientation: 'vertical',
            children: [tabGroup1, tabGroup2]
        });

        const splitNodeOuter = new SplitNode({
            orientation: 'horizontal',
            children: [splitNodeInner, tabGroup3]
        });

        workspace = new Workspace(splitNodeOuter);
        workspace.setActiveGroup('tg3');

        const serialized = workspace.toJSON();
        const deserializedWorkspace = Workspace.fromJSON(serialized);

        expect(deserializedWorkspace.root).toBeInstanceOf(SplitNode);
        const rootSplit = deserializedWorkspace.root as SplitNode;
        expect(rootSplit.children).toHaveLength(2);
        expect(rootSplit.children[0]).toBeInstanceOf(SplitNode);
        expect(rootSplit.children[1]).toBeInstanceOf(TabGroup);

        const innerSplit = rootSplit.children[0] as SplitNode;
        expect(innerSplit.children).toHaveLength(2);
        expect((innerSplit.children[0] as TabGroup).views[0].name).toBe('View A');
        expect((innerSplit.children[1] as TabGroup).views[0].name).toBe('View B');
        expect((rootSplit.children[1] as TabGroup).views[0].name).toBe('View C');
        expect(deserializedWorkspace.activeGroupId).toBe('tg3');
    });

    it('should correctly deserialize dynamic collection views using path matching', () => {
        // Mock a registry entry for a dynamic collection page
        viewRegistry.push(new VizView({
            name: "Collection", // This is the component name in the registry
            component: DummyComponent,
            path: "/collections/[uid]"
        }));

        const serializedCollectionView: SerializedVizView = {
            id: 100,
            name: 'My Special Collection Tab', // This is the dynamic display name
            path: '/collections/abcdef123',
            opticalCenterFix: 0.5,
            isActive: true,
            locked: false
        };

        const tabGroup = new TabGroup({ id: 'tg-dyn', views: [] });
        const deserializedView = VizView.fromJSON(serializedCollectionView, viewRegistry.find(v => v.path === '/collections/[uid]')!.component);
        tabGroup.addTab(deserializedView);
        workspace = new Workspace(tabGroup);

        const serialized = workspace.toJSON();
        const deserializedWorkspace = Workspace.fromJSON(serialized);

        const deserializedTabGroup = deserializedWorkspace.root as TabGroup;
        expect(deserializedTabGroup.views).toHaveLength(1);
        const restoredView = deserializedTabGroup.views[0];

        expect(restoredView.name).toBe('My Special Collection Tab'); // Should retain dynamic name
        expect(restoredView.path).toBe('/collections/abcdef123');
        expect(restoredView.component).toBeDefined();

        // Clean up mock registry entry
        const index = viewRegistry.findIndex(v => v.path === "/collections/[uid]");
        if (index > -1) viewRegistry.splice(index, 1);
    });

    describe('TabGroup operations', () => {
        let tabGroup: TabGroup;
        let view1: VizView, view2: VizView, view3: VizView, view4: VizView;

        beforeEach(() => {
            view1 = createTestView('View 1');
            view2 = createTestView('View 2');
            view3 = createTestView('View 3', undefined, true); // Locked view
            view4 = createTestView('View 4');
            tabGroup = new TabGroup({ views: [view1, view2], activeViewId: view1.id });
            workspace = new Workspace(tabGroup);
        });

        it('should add a tab and set it as active', () => {
            tabGroup.addTab(view3);
            expect(tabGroup.views).toHaveLength(3);
            expect(tabGroup.activeView?.id).toBe(view3.id);
            expect(view3.isActive).toBe(true);
            expect(view1.isActive).toBe(false);
        });

        it('should add a tab at a specific index', () => {
            tabGroup.addTab(view4, 1); // Add view4 at index 1
            expect(tabGroup.views).toHaveLength(3);
            expect(tabGroup.views[1]).toBe(view4);
            expect(tabGroup.activeView?.id).toBe(view4.id);
        });

        it('should remove a tab and activate a sibling', () => {
            tabGroup.removeTab(view1.id); // Remove active view
            expect(tabGroup.views).toHaveLength(1);
            expect(tabGroup.views[0].id).toBe(view2.id);
            expect(tabGroup.activeView?.id).toBe(view2.id);
            expect(view2.isActive).toBe(true);

            tabGroup.addTab(view1); // Add it back
            tabGroup.addTab(view4); // Now: view2, view1, view4 (active)
            tabGroup.removeTab(view1.id); // Remove a non-active view
            expect(tabGroup.views).toHaveLength(2); // Should be view2, view4
            expect(tabGroup.activeView?.id).toBe(view4.id);
            expect(tabGroup.views[0].id).toBe(view2.id);
            expect(tabGroup.views[1].id).toBe(view4.id);
        });

        it('should not remove a locked tab if not explicitly handled (current implementation does remove)', () => {
            // Note: Current removeTab implementation does not prevent removal of locked views.
            // If locking is meant to prevent removal, this test would fail and require adjustment to removeTab.
            const initialLength = tabGroup.views.length;
            tabGroup.addTab(view3); // Add locked view
            tabGroup.removeTab(view3.id);
            expect(tabGroup.views).toHaveLength(initialLength); // It was removed, as per current logic
            expect(tabGroup.views).not.toContain(view3);
        });


        it('should set active tab correctly', () => {
            tabGroup.setActive(view2.id);
            expect(tabGroup.activeView?.id).toBe(view2.id);
            expect(view2.isActive).toBe(true);
            expect(view1.isActive).toBe(false);

            tabGroup.setActive(999); // Non-existent view
            expect(tabGroup.activeView?.id).toBe(view2.id); // Active view should not change
        });
    });

    describe('SplitNode operations', () => {
        let splitNode: SplitNode;
        let tabGroup1: TabGroup, tabGroup2: TabGroup;
        let viewA: VizView, viewB: VizView;

        beforeEach(() => {
            viewA = createTestView('View A');
            viewB = createTestView('View B');
            tabGroup1 = new TabGroup({ id: 'tg1', views: [viewA], activeViewId: viewA.id });
            tabGroup2 = new TabGroup({ id: 'tg2', views: [viewB], activeViewId: viewB.id });
            splitNode = new SplitNode({ orientation: 'horizontal', children: [tabGroup1] });
            workspace = new Workspace(splitNode);
        });

        it('should add a child', () => {
            splitNode.addChild(tabGroup2);
            expect(splitNode.children).toHaveLength(2);
            expect(splitNode.children[1]).toBe(tabGroup2);
            expect(tabGroup2.parent).toBe(splitNode);
            // Sizes should be normalized
            expect(splitNode.children[0].size).toBe(50);
            expect(splitNode.children[1].size).toBe(50);
        });

        it('should add a child at a specific index', () => {
            const tabGroup3 = new TabGroup({ id: 'tg3', views: [createTestView('View C')] });
            splitNode.addChild(tabGroup2); // Add tg2 at end
            splitNode.addChild(tabGroup3, 1); // Add tg3 in middle
            expect(splitNode.children).toHaveLength(3);
            expect(splitNode.children[1]).toBe(tabGroup3);
            expect(tabGroup3.parent).toBe(splitNode);
            expect(splitNode.children[0].size).toBeCloseTo(33.33);
            expect(splitNode.children[1].size).toBeCloseTo(33.33);
            expect(splitNode.children[2].size).toBeCloseTo(33.33);
        });

        it('should remove a child', () => {
            splitNode.addChild(tabGroup2);
            splitNode.removeChild(tabGroup1);
            expect(splitNode.children).toHaveLength(1);
            expect(splitNode.children[0]).toBe(tabGroup2);
            expect(tabGroup1.parent).toBeNull();
            expect(splitNode.children[0].size).toBe(100); // Remaining child takes full size
        });

        it('should replace a child', () => {
            splitNode.addChild(tabGroup2); // Now [tg1, tg2]
            const newTabGroup = new TabGroup({ id: 'tg3', views: [createTestView('View C')] });
            splitNode.replaceChild(tabGroup1, newTabGroup);

            expect(splitNode.children).toHaveLength(2);
            expect(splitNode.children[0]).toBe(newTabGroup);
            expect(newTabGroup.parent).toBe(splitNode);
            expect(tabGroup1.parent).toBeNull();
            expect(newTabGroup.size).toBe(50); // Should inherit size from old node
        });

        it('should normalize sizes after adding/removing children', () => {
            splitNode.addChild(tabGroup2); // 2 children, each 50%
            expect(tabGroup1.size).toBe(50);
            expect(tabGroup2.size).toBe(50);

            splitNode.removeChild(tabGroup2); // 1 child, 100%
            expect(tabGroup1.size).toBe(100);
        });
    });

    describe('Workspace find methods', () => {
        let viewA: VizView, viewB: VizView, viewC: VizView;
        let tabGroup1: TabGroup, tabGroup2: TabGroup, tabGroup3: TabGroup;
        let splitNode1: SplitNode, splitNode2: SplitNode;

        beforeEach(() => {
            viewA = createTestView('View A', '/path/a');
            viewB = createTestView('View B', '/path/b');
            viewC = createTestView('View C', '/path/c');

            tabGroup1 = new TabGroup({ id: 'tg1', views: [viewA], activeViewId: viewA.id });
            tabGroup2 = new TabGroup({ id: 'tg2', views: [viewB], activeViewId: viewB.id });
            tabGroup3 = new TabGroup({ id: 'tg3', views: [viewC], activeViewId: viewC.id });

            splitNode1 = new SplitNode({ id: 'sn1', children: [tabGroup1, tabGroup2], orientation: 'horizontal' });
            splitNode2 = new SplitNode({ id: 'sn2', children: [splitNode1, tabGroup3], orientation: 'vertical' });
            workspace = new Workspace(splitNode2);
            workspace.setActiveGroup(tabGroup3.id); // Set an active group
        });

        it('should find groups with a specific view ID', () => {
            expect(workspace.findGroupWithView(viewA.id)).toBe(tabGroup1);
            expect(workspace.findGroupWithView(viewB.id)).toBe(tabGroup2);
            expect(workspace.findGroupWithView(viewC.id)).toBe(tabGroup3);
            expect(workspace.findGroupWithView(999)).toBeNull();
        });

        it('should find groups with a specific view path', () => {
            expect(workspace.findGroupWithPath('/path/a')).toBe(tabGroup1);
            expect(workspace.findGroupWithPath('/path/b')).toBe(tabGroup2);
            expect(workspace.findGroupWithPath('/path/c')).toBe(tabGroup3);
            expect(workspace.findGroupWithPath('/non/existent')).toBeNull();
        });

        it('should find a view by its path', () => {
            expect(workspace.findViewWithPath('/path/a')).toBe(viewA);
            expect(workspace.findViewWithPath('/path/b')).toBe(viewB);
            expect(workspace.findViewWithPath('/path/c')).toBe(viewC);
            expect(workspace.findViewWithPath('/non/existent')).toBeNull();
        });

        it('should find all tab groups', () => {
            const allGroups = workspace.getAllTabGroups();
            expect(allGroups).toHaveLength(3);
            expect(allGroups).toContain(tabGroup1);
            expect(allGroups).toContain(tabGroup2);
            expect(allGroups).toContain(tabGroup3);
        });

        it('should return the active group', () => {
            expect(workspace.activeGroup).toBe(tabGroup3);
        });

        it('should return null for active group if none is set', () => {
            workspace.activeGroupId = undefined;
            expect(workspace.activeGroup).toBeNull();
        });
    });

    describe('Workspace manipulation methods', () => {
        let view1: VizView, view2: VizView, view3: VizView, lockedView: VizView;
        let tabGroup1: TabGroup, tabGroup2: TabGroup, lockedTabGroup: TabGroup;
        let splitNode: SplitNode;

        beforeEach(() => {
            view1 = createTestView('View 1');
            view2 = createTestView('View 2');
            view3 = createTestView('View 3'); // Unlocked
            lockedView = createTestView('Locked View', undefined, true); // Locked

            tabGroup1 = new TabGroup({ id: 'tg1', views: [view1], activeViewId: view1.id });
            tabGroup2 = new TabGroup({ id: 'tg2', views: [view2], activeViewId: view2.id });
            lockedTabGroup = new TabGroup({ id: 'ltg', views: [lockedView], locked: true, activeViewId: lockedView.id });

            splitNode = new SplitNode({ id: 'sn1', children: [tabGroup1, tabGroup2], orientation: 'horizontal' });
            workspace = new Workspace(splitNode);
        });

        it('should move a tab between groups', () => {
            workspace.moveTab(view1.id, tabGroup2.id);
            expect(tabGroup1.views).toHaveLength(0);
            expect(tabGroup2.views).toHaveLength(2);
            expect(tabGroup2.views).toContain(view1);
            expect(tabGroup2.activeView?.id).toBe(view1.id);
            expect(tabGroup1.activeView).toBeUndefined(); // Source group should have no active view
        });

        it('should not move a locked tab', () => {
            workspace.root = new SplitNode({ children: [lockedTabGroup, tabGroup1] });
            workspace.moveTab(lockedView.id, tabGroup1.id);
            expect(lockedTabGroup.views).toHaveLength(1); // Should still be in locked group
            expect(tabGroup1.views).toHaveLength(1); // Should not have gained the locked view
        });

        it('should not move a tab from/to a locked group', () => {
            workspace.root = new SplitNode({ children: [lockedTabGroup, tabGroup1] });
            // Try to move view1 (unlocked) to lockedTabGroup
            workspace.moveTab(view1.id, lockedTabGroup.id);
            expect(tabGroup1.views).toHaveLength(1); // View1 should still be here
            expect(lockedTabGroup.views).toHaveLength(1);

            // Try to move lockedView (locked) to tabGroup1 (unlocked)
            workspace.moveTab(lockedView.id, tabGroup1.id);
            expect(lockedTabGroup.views).toHaveLength(1); // Locked view should still be in locked group
            expect(tabGroup1.views).toHaveLength(1);
        });

        it('should split a group horizontally and move a view', () => {
            const initialTabGroups = workspace.getAllTabGroups();
            expect(initialTabGroups).toHaveLength(2); // tg1, tg2

            workspace.splitGroup(tabGroup1.id, view1, 'right');

            const newTabGroups = workspace.getAllTabGroups();
            expect(newTabGroups).toHaveLength(3); // tg1 (empty), new group (view1), tg2

            // Root should now be a split node with 3 children
            const rootSplit = workspace.root as SplitNode;
            expect(rootSplit).toBeInstanceOf(SplitNode);
            expect(rootSplit.children).toHaveLength(3);

            // Verify order and content
            expect(rootSplit.children[0]).toBe(tabGroup1); // Original group, now empty
            expect(rootSplit.children[1]).toBeInstanceOf(TabGroup); // New group with view1
            expect((rootSplit.children[1] as TabGroup).views[0]).toBe(view1);
            expect(rootSplit.children[2]).toBe(tabGroup2); // Remaining original group

            // New group should be active
            expect((rootSplit.children[1] as TabGroup).id).toBe(workspace.activeGroupId);
            expect((rootSplit.children[1] as TabGroup).activeView?.id).toBe(view1.id);
            expect(tabGroup1.views).toHaveLength(0); // Original tabGroup1 is empty
        });

        it('should split a group vertically and move a view', () => {
            workspace.splitGroup(tabGroup1.id, view1, 'bottom'); // Split tg1 vertically

            // Expect the original tg1 to be replaced by a vertical split node
            const rootSplit = workspace.root as SplitNode;
            expect(rootSplit).toBeInstanceOf(SplitNode);
            expect(rootSplit.children).toHaveLength(2); // The new split and tabGroup2

            const newVerticalSplit = rootSplit.children[0] as SplitNode;
            expect(newVerticalSplit).toBeInstanceOf(SplitNode);
            expect(newVerticalSplit.orientation).toBe('vertical');
            expect(newVerticalSplit.children).toHaveLength(2);

            // The children of the new vertical split should be an empty tabGroup1 and a new tabGroup with view1
            const emptyTg1 = newVerticalSplit.children[0] as TabGroup;
            const newTg = newVerticalSplit.children[1] as TabGroup;

            expect(emptyTg1.views).toHaveLength(0);
            expect(newTg.views).toHaveLength(1);
            expect(newTg.views[0]).toBe(view1);
            expect(newTg.id).toBe(workspace.activeGroupId); // New group should be active
        });


        it('should not split a locked group or move a locked view during split', () => {
            workspace.root = lockedTabGroup; // Make locked group the root
            const initialRootId = lockedTabGroup.id;
            workspace.splitGroup(lockedTabGroup.id, lockedView, 'right'); // Try to split locked group
            expect(workspace.root).toBe(lockedTabGroup); // Should not have split
            expect(workspace.root.id).toBe(initialRootId); // Root should still be the locked group

            // Try to split an unlocked group with a locked view
            const viewToMove = createTestView('Another View');
            tabGroup1.addTab(viewToMove);
            workspace.root = new SplitNode({ children: [tabGroup1, lockedTabGroup], orientation: 'horizontal' });
            workspace.splitGroup(tabGroup1.id, lockedView, 'right');
            expect(tabGroup1.views).toHaveLength(2); // ViewToMove and lockedView
            expect(tabGroup1.views).toContain(lockedView); // Locked view should still be in the source group
            expect(tabGroup1.views).toContain(viewToMove);

            const newTabGroups = workspace.getAllTabGroups();
            // Should still have tg1 (with both views) and lockedTabGroup
            expect(newTabGroups).toHaveLength(2);
        });

        it('should cleanup empty parent split nodes when child tab group becomes empty', () => {
            const viewA = createTestView('View A');
            const tgA = new TabGroup({ id: 'tgA', views: [viewA] });
            const tgB = new TabGroup({ id: 'tgB', views: [createTestView('View B')] });
            const splitH = new SplitNode({ id: 'splitH', orientation: 'horizontal', children: [tgA, tgB] });
            workspace = new Workspace(splitH); // Workspace is splitH -> [tgA, tgB]

            tgA.removeTab(viewA.id); // tgA becomes empty

            // After tgA is empty, splitH should collapse. Workspace root should become tgB.
            expect(workspace.root).toBe(tgB);
            expect(tgB.parent).toBeNull();
        });

        it('should cleanup nested empty split nodes', () => {
            const viewA = createTestView('View A');
            const tgA = new TabGroup({ id: 'tgA', views: [viewA] });
            const tgB = new TabGroup({ id: 'tgB', views: [createTestView('View B')] });
            const splitV = new SplitNode({ id: 'splitV', orientation: 'vertical', children: [tgA] }); // splitV -> tgA
            const splitH = new SplitNode({ id: 'splitH', orientation: 'horizontal', children: [splitV, tgB] }); // splitH -> [splitV, tgB]
            workspace = new Workspace(splitH); // Root: splitH

            tgA.removeTab(viewA.id); // tgA becomes empty

            // tgA removed from splitV, splitV becomes empty and is removed.
            // splitH now has only tgB, so splitH collapses, and workspace root becomes tgB.
            expect(workspace.root).toBe(tgB);
            expect(tgB.parent).toBeNull();
        });
    });
});