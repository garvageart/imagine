import * as util from "./util";

export interface HistogramStat {
    red: number;
    green: number;
    blue: number;
    luminance: number;
    rgb: number;
}

export interface HistogramData {
    hist: {
        red: number[];
        green: number[];
        blue: number[];
        luminance: number[];
        rgb: number[];
    };
    count: HistogramStat;
    max: HistogramStat;
    mean: HistogramStat;
    median: HistogramStat;
    mode: HistogramStat;
    stddev: HistogramStat;
}

/**
 * Calculates the histogram and statistics from a source canvas or image element
 */
export class Histogram {
    _source: HTMLCanvasElement | HTMLImageElement;
    _canvas: HTMLCanvasElement;
    _ctx: CanvasRenderingContext2D;
    _luminanceWeights: number[];
    data: HistogramData;

    /**
     * @param source - source element
     * @param luminanceWeights - array of weights to convert rgb to luminance
     */
    constructor(
        source: HTMLCanvasElement | HTMLImageElement,
        luminanceWeights: number[] = [0.2126, 0.7152, 0.0722]
    ) {
        this._source = source;
        this._canvas = document.createElement("canvas");
        this._ctx = this._canvas.getContext("2d")!;
        if (luminanceWeights.length !== 3) {
            throw new Error("luminance weights must have 3 values that sum to one");
        }
        this._luminanceWeights = luminanceWeights.slice();
        this.data = this.calcData();
    }

    get red(): number[] {
        return this.data.hist.red;
    }

    get green(): number[] {
        return this.data.hist.green;
    }

    get blue(): number[] {
        return this.data.hist.blue;
    }

    get luminance(): number[] {
        return this.data.hist.luminance;
    }

    get rgb(): number[] {
        return this.data.hist.rgb;
    }

    get count(): HistogramStat {
        return this.data.count;
    }

    get mean(): HistogramStat {
        return this.data.mean;
    }

    get median(): HistogramStat {
        return this.data.median;
    }

    get mode(): HistogramStat {
        return this.data.mode;
    }

    get std(): HistogramStat {
        return this.data.stddev;
    }

    get max(): HistogramStat {
        return this.data.max;
    }

    private calcData(): HistogramData {
        // draw image/canvas source to a new canvas and get pixel data
        this._canvas.width = this._source.width;
        this._canvas.height = this._source.height;
        this._ctx.drawImage(this._source, 0, 0);
        const imageData = this._ctx.getImageData(0, 0, this._canvas.width, this._canvas.height);

        // Initialize histogram arrays
        const histograms = {
            red: new Array(256).fill(0),
            green: new Array(256).fill(0),
            blue: new Array(256).fill(0),
            luminance: new Array(256).fill(0),
            rgb: new Array(256).fill(0),
        };

        // Tally each pixel into its bin where index = bin value
        for (let i = 0; i < imageData.data.length; i += 4) {
            const r = imageData.data[i];
            const g = imageData.data[i + 1];
            const b = imageData.data[i + 2];
            const l = Math.floor(
                this._luminanceWeights[0] * r +
                this._luminanceWeights[1] * g +
                this._luminanceWeights[2] * b
            );

            histograms.red[r] += 1;
            histograms.green[g] += 1;
            histograms.blue[b] += 1;
            histograms.luminance[l] += 1;
            histograms.rgb[r] += 1;
            histograms.rgb[g] += 1;
            histograms.rgb[b] += 1;
        }

        // Find max values for each channel
        const channels = ["red", "green", "blue", "luminance", "rgb"] as const;
        const maxValues: HistogramStat = {
            red: 0,
            green: 0,
            blue: 0,
            luminance: 0,
            rgb: 0,
        };

        for (let i = 0; i < 256; i++) {
            for (const channel of channels) {
                maxValues[channel] = Math.max(maxValues[channel], histograms[channel][i]);
            }
        }

        // Calculate statistics for each channel
        const countRed = imageData.data.length / 4;
        const countRGB = countRed * 3;

        const stats = {
            mean: {} as HistogramStat,
            median: {} as HistogramStat,
            mode: {} as HistogramStat,
            stddev: {} as HistogramStat,
            count: {} as HistogramStat,
        };

        for (const channel of channels) {
            const hist = histograms[channel];
            stats.mean[channel] = util.mean(hist);
            stats.median[channel] = util.median(hist);
            stats.mode[channel] = util.mode(hist);
            stats.stddev[channel] = util.std(hist);
            stats.count[channel] = channel === "rgb" ? countRGB : countRed;
        }

        // can use average of averages since sizes are equal
        stats.mean.rgb = (stats.mean.red + stats.mean.green + stats.mean.blue) / 3.0;

        return {
            hist: histograms,
            count: stats.count,
            max: maxValues,
            mean: stats.mean,
            median: stats.median,
            mode: stats.mode,
            stddev: stats.stddev,
        };
    }

    /**
     * Recalculate all data
     */
    refresh(): void {
        this.data = this.calcData();
    }
}



