import { JustifiedLayout } from "@immich/justified-layout-wasm";
import type { ImageAsset } from "$lib/api";
import type { ConsolidatedGroup, ImageWithDateLabel } from "$lib/photo-layout";
import { getTakenAt } from "$lib/utils/images";
import { DateTime } from "luxon";

export type GridItem = {
    asset: ImageWithDateLabel;
    width: number;
    height: number;
    left: number;
};

export type GridRowHeader = {
    type: "header";
    id: string;
    label: string;
    top: number;
    height: number;
    date: Date;
    groupId: string;
};

export type GridRowImages = {
    type: "images";
    id: string;
    items: GridItem[];
    top: number;
    height: number;
    groupId: string;
};

export type GridRow = GridRowHeader | GridRowImages;

// Cached row stores relativeTop (y-pos relative to the group start)
type CachedRow =
    | (Omit<GridRowHeader, "top"> & { relativeTop: number; })
    | (Omit<GridRowImages, "top"> & { relativeTop: number; });

type GroupCacheEntry = {
    rows: CachedRow[];
    height: number; // Total height of this group (header + images)
};

export type PhotoGridConfig = {
    targetRowHeight?: number;
    gridGap?: number;
    headerHeight?: number;
    bufferPx?: number;
};

export class PhotoGridVirtualizer {
    // Reactive State
    rows = $state<GridRow[]>([]);
    visibleRows = $state<GridRow[]>([]);
    totalHeight = $state(0);

    // Configuration
    containerWidth = $state(0);
    targetRowHeight = $state(280);
    gridGap = $state(8);
    headerHeight = $state(60);

    // Internal
    private groupCache = new Map<string, GroupCacheEntry>();
    private previousGroups: ConsolidatedGroup[] = [];

    // Scroll State
    scrollTop = $state(0);
    viewportHeight = $state(1000);
    bufferPx = $state(2000);

    constructor(config: PhotoGridConfig = {}) {
        this.updateConfig(config);
    }

    updateConfig(config: PhotoGridConfig) {
        if (config.targetRowHeight) {
            this.targetRowHeight = config.targetRowHeight;
        }
        if (config.gridGap !== undefined) {
            this.gridGap = config.gridGap;
        }
        if (config.headerHeight) {
            this.headerHeight = config.headerHeight;
        }
        if (config.bufferPx) {
            this.bufferPx = config.bufferPx;
        }

        // Clear cache if critical metrics change
        this.groupCache.clear();
    }

    /**
     * Updates the layout based on groups and container width.
     */
    update(groups: ConsolidatedGroup[], width: number) {
        if (width <= 0) {
            return;
        }
        this.containerWidth = width;
        this.previousGroups = groups; // Store for simple refreshes if needed

        let currentTop = 0;
        const allRows: GridRow[] = [];
        let smallGroupBatch: ConsolidatedGroup[] = [];
        const SMALL_GROUP_THRESHOLD = 6;

        const flushBatch = () => {
            if (smallGroupBatch.length === 0) return;

            // Compute Layout for this batch of small groups
            // We create a mixed list of items: [Header, ...Images, Header, ...Images]
            const mixedItems: ImageWithDateLabel[] = [];

            for (const g of smallGroupBatch) {
                // Inject Header Item
                const headerItem: any = {
                    uid: `header-item-${g.label}`,
                    isHeaderItem: true,
                    headerLabel: g.label,
                    // Dummy dimensions for aspect ratio calculation if needed (handled in computeImages)
                    width: 100,
                    height: 100,
                    // Add date for robust handling
                    taken_at: g.startDate.toISO()
                };
                mixedItems.push(headerItem);
                mixedItems.push(...g.allImages);
            }

            // We treat this batch as a single "mixed" group
            const batchId = `batch-${smallGroupBatch[0].label}`;

            // We calculate layout. Header items need a specific aspect ratio.
            const { rows, height } = this.computeImages(mixedItems, width, batchId);

            for (const r of rows) {
                allRows.push({
                    ...r,
                    top: currentTop + r.relativeTop
                } as GridRow);
            }

            currentTop += height;
            smallGroupBatch = [];
        };

        for (const group of groups) {
            // Check if group is small
            if (group.allImages.length <= SMALL_GROUP_THRESHOLD) {
                smallGroupBatch.push(group);
            } else {
                // Flush previous batch
                flushBatch();

                // Check Cache for large group
                const cacheKey = `${group.label}-${group.allImages.length}-${width}`;
                let cached = this.groupCache.get(cacheKey);

                if (!cached) {
                    // Compute Layout for this group
                    cached = this.computeGroup(group, width);
                    this.groupCache.set(cacheKey, cached);
                }

                // Apply absolute positions
                for (const r of cached.rows) {
                    allRows.push({
                        ...r,
                        top: currentTop + r.relativeTop
                    } as GridRow);
                }

                currentTop += cached.height;
            }
        }

        flushBatch();

        this.rows = allRows;
        this.totalHeight = currentTop;

        // Immediate visible update
        this.updateVisible();
    }

    /**
     * Updates layout for flat list of images (no groups).
     * We treat this as a single unnamed group without a header.
     */
    updateFlat(images: ImageAsset[], width: number) {
        if (width <= 0) {
            return;
        }
        this.containerWidth = width;

        // Compute straight images, no header
        const { rows, height } = this.computeImages(
            images as ImageWithDateLabel[],
            width,
            "flat"
        );

        const finalRows: GridRow[] = rows.map((r) => ({
            ...r,
            top: r.relativeTop
        })) as unknown as GridRow[];

        this.rows = finalRows;
        this.totalHeight = height;
        this.updateVisible();
    }

    updateScroll(scrollTop: number, viewportHeight: number) {
        let changed = false;
        // Check tolerance to avoid micro-updates
        if (Math.abs(this.scrollTop - scrollTop) > 1) {
            this.scrollTop = scrollTop;
            changed = true;
        }
        if (Math.abs(this.viewportHeight - viewportHeight) > 1) {
            this.viewportHeight = viewportHeight;
            changed = true;
        }

        if (changed) {
            this.updateVisible();
        }
    }

    private updateVisible() {
        if (this.rows.length === 0) {
            this.visibleRows = [];
            return;
        }

        const minY = Math.max(0, this.scrollTop - this.bufferPx);
        const maxY = this.scrollTop + this.viewportHeight + this.bufferPx;

        const startIndex = this.findStartIndex(minY);
        let endIndex = startIndex;

        while (endIndex < this.rows.length && this.rows[endIndex].top < maxY) {
            endIndex++;
        }

        this.visibleRows = this.rows.slice(startIndex, endIndex);
    }

    /**
     * Binary search to find the first row that *might* be visible
     * (Row bottom >= minY)
     */
    private findStartIndex(y: number): number {
        let low = 0;
        let high = this.rows.length - 1;
        while (low <= high) {
            const mid = (low + high) >>> 1;
            const row = this.rows[mid];
            if (row.top + row.height < y) {
                low = mid + 1;
            } else {
                high = mid - 1;
            }
        }
        return Math.max(0, low);
    }

    private computeGroup(
        group: ConsolidatedGroup,
        width: number
    ): GroupCacheEntry {
        const rows: CachedRow[] = [];
        let cursorY = 0;

        // 1. Add Header Row
        // We use the group label and date
        const headerHeight = this.headerHeight; // Configurable
        const headerRow: CachedRow = {
            type: "header",
            id: `header-${group.label}`,
            label: group.label,
            date: group.startDate.toJSDate(),
            height: headerHeight,
            relativeTop: cursorY,
            groupId: group.label
        };
        rows.push(headerRow);
        cursorY += headerHeight;

        // 2. Compute Image Rows
        const imageResult = this.computeImages(group.allImages, width, group.label);

        // Offset the image rows by the header height
        for (const r of imageResult.rows) {
            rows.push({
                ...r,
                relativeTop: r.relativeTop + cursorY
            });
        }

        cursorY += imageResult.height;

        return {
            rows,
            height: cursorY
        };
    }

    private computeImages(
        items: ImageWithDateLabel[],
        width: number,
        groupId: string
    ) {
        if (items.length === 0) {
            return { rows: [], height: 0 };
        }

        // Justified Layout expects aspect ratios
        const aspectRatios = new Float32Array(items.length);
        for (let i = 0; i < items.length; i++) {
            const item = items[i];
            if (item.isHeaderItem) {
                // Use a fixed aspect ratio for inline headers (e.g. narrow vertical strip)
                // Row height is typically ~250-300px. A ratio of 0.5 gives ~125-150px width.
                aspectRatios[i] = 0.5;
            } else {
                aspectRatios[i] = (item.width || 4) / (item.height || 3);
            }
        }

        // Filter out bad items if necessary, but simpler to just run it

        const layout = new JustifiedLayout(aspectRatios, {
            rowHeight: this.targetRowHeight,
            rowWidth: width,
            spacing: this.gridGap,
            heightTolerance: 0.15 // Allow some flex to avoid cropping
        });

        const rows: CachedRow[] = [];
        let rowIndex = 0;

        // Convert JustifiedLayout boxes to GridRowImages
        // The library returns a flat list of boxes. We need to chunk them by "top" position.

        let currentRowItems: GridItem[] = [];
        let currentRowTop = -1;
        let currentRowHeight = 0;

        // JustifiedLayout output geometry

        for (let i = 0; i < items.length; i++) {
            const box = layout.getPosition(i);
            const asset = items[i]; // Corresponding asset

            // If new row detected (based on Top position change)
            // Note: Floating point comparison needs epsilon
            if (currentRowTop === -1) {
                currentRowTop = box.top;
                currentRowHeight = box.height;
            } else if (Math.abs(box.top - currentRowTop) > 1) {
                // Check if the NEW row we are about to start contains a header item
                // The current item `asset` is the first item of the new row (conceptually)
                // However, JustifiedLayout has already computed positions.
                // We might need to manually shift things if we want extra gaps?
                // Actually, the `JustifiedLayout` library handles spacing. 
                // If we want EXTRA gap for rows with headers, we might need to post-process?
                // Or: simpler approach - just check if the previous row had a header? 
                // No, we want gap ABOVE the row with header.

                // Push previous row
                rows.push({
                    type: "images",
                    id: `${groupId}-imgrow-${rowIndex++}`,
                    items: currentRowItems,
                    height: currentRowHeight,
                    relativeTop: currentRowTop,
                    groupId
                });

                // Start new row
                currentRowItems = [];
                currentRowTop = box.top;
                currentRowHeight = box.height;
            }

            currentRowItems.push({
                asset,
                width: box.width,
                height: box.height,
                left: box.left
            });
        }

        // Push final row
        if (currentRowItems.length > 0) {
            rows.push({
                type: "images",
                id: `${groupId}-imgrow-${rowIndex++}`,
                items: currentRowItems,
                height: currentRowHeight,
                relativeTop: currentRowTop,
                groupId
            });
        }

        // Post-process rows to add gap above rows with header items
        let cummulativeOffset = 0;
        const GAP_ABOVE_HEADER = this.gridGap; // Pixel gap

        for (let i = 0; i < rows.length; i++) {
            const row = rows[i];

            // Apply accumulated offset to current row
            row.relativeTop += cummulativeOffset;

            // Check if we need to add a gap before this row
            if (row.type === 'images') {
                const hasHeader = row.items.some(item => item.asset.isHeaderItem);

                // Add gap if this row has a header.
                // This ensures that even the first row of a batch gets separation from whatever came before (e.g. a previous large group).
                // We don't need to worry about Large Groups here because they don't use isHeaderItem (they use separate GridRowHeader objects).
                if (hasHeader) {
                    row.relativeTop += GAP_ABOVE_HEADER;
                    cummulativeOffset += GAP_ABOVE_HEADER;
                }
            }
        }

        return { rows, height: layout.containerHeight + cummulativeOffset };
    }

    // Helper to find the "Active" header (sticky)
    getActiveHeader(scrollTop: number): GridRowHeader | null {
        // We want the header of the group that covers the top of the screen.
        // Or essentially, the last header that has (top <= scrollTop).
        // Since rows are sorted by top, we can search.

        let low = 0;
        let high = this.rows.length - 1;
        let candidate: GridRowHeader | null = null;

        // Binary search for the last row with top <= scrollTop
        while (low <= high) {
            const mid = (low + high) >>> 1;
            const row = this.rows[mid];

            if (row.top <= scrollTop) {
                if (row.type === "header") {
                    candidate = row;
                }
                // Even if it's an image row, its group header might be the candidate.
                // But simpler: just find the row at scrollTop, then find that row's group header.
                low = mid + 1;
            } else {
                high = mid - 1;
            }
        }

        // Refined approach:
        // Find the row currently at scrollTop.
        // Identify its group.
        // Return that group's header.

        const idx = Math.max(
            0,
            Math.min(this.findStartIndex(scrollTop), this.rows.length - 1)
        );
        if (this.rows.length === 0) {
            return null;
        }

        // Scan backwards from the visible start to find the nearest header
        for (let i = idx; i >= 0; i--) {
            if (this.rows[i].type === "header") {
                return this.rows[i] as GridRowHeader;
            }
        }

        return null;
    }

    // Helper to get the NEXT header (for pushing effect)
    getNextHeader(scrollTop: number): GridRowHeader | null {
        const idx = this.findStartIndex(scrollTop);
        // We look forward from the current viewport top
        for (let i = idx; i < this.rows.length; i++) {
            const row = this.rows[i];
            if (row.type === "header" && row.top > scrollTop) {
                return row as GridRowHeader;
            }
        }
        return null;
    }

    getDateLabel(scrollTop: number): string {
        const header = this.getActiveHeader(scrollTop);
        if (header) {
            return header.label; // Already formatted by the grouping logic usually
        }

        // Fallback: Check the row at the current scroll position
        if (this.rows.length > 0) {
            const idx = Math.max(0, Math.min(this.findStartIndex(scrollTop), this.rows.length - 1));
            const row = this.rows[idx];
            if (row.type === "images" && row.items.length > 0) {
                const asset = row.items[0].asset;
                if (asset.isHeaderItem && asset.headerLabel) {
                    return asset.headerLabel;
                }
                // Format date generic if no specific label found
                return DateTime.fromJSDate(getTakenAt(asset)).toFormat("LLL yyyy");
            }
        }
        return "";
    }
}