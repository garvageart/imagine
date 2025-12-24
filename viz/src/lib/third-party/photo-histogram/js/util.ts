/**
 * Count the number of elements in a histogram
 */
/**
 * Count the number of elements in a histogram
 */
export function cnt(histogram: number[]): number {
    let count = 0;
    for (let i = 0; i < histogram.length; i++) {
        count += histogram[i];
    }
    return count;
}

/**
 * Return the middle index of a histogram. Middle index of histogram with count of 5 == 3, and
 * middle index of 6 == 3.
 */
export function middleIndex(histogram: number[]): number {
    const totalCount = cnt(histogram);
    let middleIndex = Math.floor(totalCount / 2);
    if (totalCount > 1 && totalCount % 2 === 1) middleIndex += 1;
    return middleIndex;
}

/**
 * Calculate the mean from a histogram array
 */
export function mean(histogram: number[]): number {
    let count = 0;
    let sum = 0;
    for (let i = 0; i < histogram.length; i++) {
        count += histogram[i];
        sum += i * histogram[i];
    }
    return sum / count;
}

/**
 * Calculate the median from a histogram array
 */
export function median(histogram: number[]): number {
    let count = 0;
    let i = 0;
    const mi = middleIndex(histogram);
    while (count < mi) {
        count += histogram[i++];
    }
    return i - 1;
}

/**
 * Calculate the mode from a histogram array
 */
export function mode(histogram: number[]): number {
    let max = 0;
    let bin = 0;
    for (let i = 0; i < histogram.length; i++) {
        if (histogram[i] > max) {
            max = histogram[i];
            bin = i;
        }
    }
    return bin;
}

/**
 * Calculate the variance of a histogram array
 */
export function variance(histogram: number[]): number {
    const average = mean(histogram);
    let sum = 0;
    let count = 0;
    for (let i = 0; i < histogram.length; i++) {
        count += histogram[i];
        sum += histogram[i] * Math.pow(i - average, 2);
    }

    return sum / count;
}

/**
 * Calculate the standard deviation of a histogram array
 */
export function std(histogram: number[]): number {
    return Math.sqrt(variance(histogram));
}

/**
 * Returns a random integer between min (included) and max (included)
 */
export function integer(min: number, max: number): number {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

/**
 * Returns a random base string concatenated with a random integer
 */
export function id(base: string): string {
    return base + integer(0, 1e10);
}

/**
 * Create a new element and set attributes
 */
export function createElement(
    tagName: string,
    attributes?: Record<string, any>,
    parent?: Element
): HTMLElement {
    const element = document.createElement(tagName);
    if (attributes) {
        Object.entries(attributes).forEach(([key, value]) => {
            element.setAttribute(key, value);
        });
    }
    if (parent) {
        parent.appendChild(element);
    }
    return element;
}

/**
 * Create a new element with namespace and set attributes
 */
export function createElementNS(
    tagName: string,
    namespace: string,
    attributes?: Record<string, any>,
    parent?: Element
): Element {
    const element = document.createElementNS(namespace, tagName);
    if (attributes) {
        Object.entries(attributes).forEach(([key, value]) => {
            element.setAttributeNS(null, key, String(value));
        });
    }
    if (parent) {
        parent.appendChild(element);
    }
    return element;
}

export class EnumEx {
    /**
     * Get names of enumeration
     */
    static getNames(e: any): string[] {
        return Object.keys(e).filter((v) => isNaN(parseInt(v, 10)));
    }

    /**
     * Get values of enumeration
     */
    static getValues(e: any): number[] {
        return Object.keys(e)
            .map((v) => parseInt(v, 10))
            .filter((v) => !isNaN(v));
    }

    /**
     * Get name/value pairs of enumeration
     */
    static getNamesAndValues(e: any): Array<{ name: string; value: number; }> {
        return EnumEx.getValues(e).map((v) => ({ name: e[v], value: v }));
    }
}

/**
 * Convert clientXY from a mouse event to scaled point on an svg element
 */
export function clientXY2SvgPoint(svg: SVGSVGElement, clientX: number, clientY: number): SVGPoint {
    const point = svg.createSVGPoint();
    point.x = clientX;
    point.y = clientY;
    return point.matrixTransform(svg.getScreenCTM()!.inverse());
}




