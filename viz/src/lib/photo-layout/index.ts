import type { Image } from "$lib/api";
import { getTakenAt, compareByTakenAtDesc } from "$lib/utils/images";
import { DateTime } from "luxon";
import { SvelteMap } from "svelte/reactivity";

export type ImageWithDateLabel = Image & {
    dateLabel?: string;
    isFirstOfDate?: boolean;
};

// Consolidated groups: merge small consecutive date groups into visual sections
export type ConsolidatedGroup = {
    label: string; // Combined label like "8 Mar - 26 Aug 2025"
    totalCount: number;
    allImages: ImageWithDateLabel[]; // All images merged together with date labels
    isConsolidated: boolean; // true if multiple date groups merged
    startDate: DateTime; // newest date in this consolidated block
    endDate: DateTime; // oldest date in this consolidated block
};

export interface DateGroup { 
    key: string; 
    date: DateTime; 
    label: string; 
    items: Image[];
}

export function groupImagesByDate(list: Image[]) {
    const map = new SvelteMap<string, Image[]>();

    for (const img of list) {
        const taken = getTakenAt(img);
        const dt = DateTime.fromJSDate(taken);
        const key = dt.toISODate()!;
        if (!map.has(key)) {
            map.set(key, []);
        }

        map.get(key)!.push(img);
    }

    // Convert to array and sort by date desc. Ensure items within each
    // date group are ordered by taken_at descending (most recent first).
    const arr = Array.from(map.entries()).map(([key, items]) => {
        const date = DateTime.fromISO(key);
        items.sort(compareByTakenAtDesc);
        return { key, date, items };
    });

    arr.sort((a, b) => b.date.toMillis() - a.date.toMillis());

    // create display label (Today / Yesterday / date)
    const labelled = arr.map((g) => {
        const today = DateTime.now().startOf("day");
        const diff = today.diff(g.date.startOf("day"), "days").days;
        let label = g.date.toLocaleString(DateTime.DATE_MED);

        if (diff === 0) {
            label = "Today";
        } else if (diff === 1) {
            label = "Yesterday";
        }

        return { key: g.key, date: g.date, label, items: g.items };
    });

    return labelled;
}

export function getConsolidatedGroups(groups: DateGroup[]) {
    const consolidated: ConsolidatedGroup[] = [];
    const SMALL_GROUP_THRESHOLD = 4;
    let currentConsolidation: ConsolidatedGroup | null = null;

    for (let i = 0; i < groups.length; i++) {
        const group = groups[i];
        const isSmall = group.items.length <= SMALL_GROUP_THRESHOLD;
        const isLastGroup = i === groups.length - 1;

        if (isSmall) {
            if (!currentConsolidation) {
                const imagesWithLabels: ImageWithDateLabel[] = group.items.map(
                    (img, idx) => ({
                        ...img,
                        dateLabel: group.label,
                        isFirstOfDate: idx === 0
                    })
                );

                currentConsolidation = {
                    label: group.label,
                    totalCount: group.items.length,
                    allImages: imagesWithLabels,
                    isConsolidated: false,
                    startDate: group.date,
                    endDate: group.date
                };
            } else {
                const imagesWithLabels: ImageWithDateLabel[] = group.items.map(
                    (img, idx) => ({
                        ...img,
                        dateLabel: group.label,
                        isFirstOfDate: idx === 0
                    })
                );

                currentConsolidation.allImages.push(...imagesWithLabels);
                currentConsolidation.totalCount += group.items.length;
                currentConsolidation.isConsolidated = true;

                currentConsolidation.endDate = group.date;
                const start = currentConsolidation.startDate;
                const end = currentConsolidation.endDate;

                if (start.year === end.year && start.month === end.month) {
                    currentConsolidation.label = `${start.day}\u2013${end.day} ${start.toFormat("LLL yyyy")}`;
                } else if (start.year === end.year) {
                    currentConsolidation.label = `${start.toFormat("d LLL")} - ${end.toFormat("d LLL yyyy")}`;
                } else {
                    currentConsolidation.label = `${start.toFormat("d LLL yyyy")} - ${end.toFormat("d LLL yyyy")}`;
                }
            }

            if (isLastGroup && currentConsolidation) {
                consolidated.push(currentConsolidation);
                currentConsolidation = null;
            }
        } else {
            if (currentConsolidation) {
                consolidated.push(currentConsolidation);
                currentConsolidation = null;
            }

            const imagesWithLabels: ImageWithDateLabel[] = group.items.map(
                (img) => ({ ...img })
            );
            consolidated.push({
                label: group.label,
                totalCount: group.items.length,
                allImages: imagesWithLabels,
                isConsolidated: false,
                startDate: group.date,
                endDate: group.date
            });
        }
    }

    return consolidated;
} 