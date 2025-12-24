import { Histogram as PHHistogram } from "../third-party/photo-histogram/js/histogram";
import type { HistogramData, HistogramStat } from "../third-party/photo-histogram/js/histogram";
import * as HistogramUtils from "../third-party/photo-histogram/js/util";

export type { HistogramData, HistogramStat };
export { HistogramUtils };
export { PHHistogram as PhotoHistogram };

/**
 * Ensures an HTMLImageElement is loaded before using it as a source for drawing.
 */
export function ensureImageLoaded(img: HTMLImageElement): Promise<void> {
    if (img.complete && img.naturalWidth !== 0) {
        return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
        const onLoad = () => {
            cleanup();
            resolve();
        };

        const onError = (e: Event) => {
            cleanup();
            reject(new Error("Image failed to load"));
        };

        function cleanup() {
            img.removeEventListener("load", onLoad);
            img.removeEventListener("error", onError);
        }

        img.addEventListener("load", onLoad);
        img.addEventListener("error", onError);
    });
}

/**
 * Compute histogram data for an image or canvas source using the existing photo-histogram core.
 * Waits for image load if necessary.
 */
export async function computeHistogram(
    source: HTMLImageElement | HTMLCanvasElement,
    luminanceWeights?: number[]
): Promise<HistogramData> {
    if (source instanceof HTMLImageElement) {
        await ensureImageLoaded(source);
    }

    const h = new PHHistogram(source as HTMLImageElement | HTMLCanvasElement, luminanceWeights);
    return h.data;
}

/**
 * Lightweight wrapper that exposes a persistent histogram instance for callers that want to refresh/reuse.
 */
export class HistogramWrapper {
    private _hist: PHHistogram;

    constructor(source: HTMLImageElement | HTMLCanvasElement, luminanceWeights?: number[]) {
        // caller is responsible for ensuring image is loaded when necessary
        this._hist = new PHHistogram(source as HTMLImageElement | HTMLCanvasElement, luminanceWeights);
    }

    get data(): HistogramData {
        return this._hist.data;
    }

    refresh(): void {
        this._hist.refresh();
    }
}
