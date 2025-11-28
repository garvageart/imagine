import { dev } from "$app/environment";

export function checkDOMForID(id: string) {
    const el = document.getElementById(id);

    if (el) {
        return true;
    }

    return false;
}

export function debugEvent(event: CustomEvent, printAsString: boolean = false) {
    if (!dev) {
        return;
    }

    console.log("Event:", event.type, new Date().toLocaleTimeString());

    if (printAsString) {
        console.log("Detail:", JSON.stringify(event.detail, null, 2));
        return;
    }

    console.log("Detail:", event.detail);
}

// Taken from here: https://stackoverflow.com/a/29956714
export function isElementScrollable(element: HTMLElement) {
    return element.scrollHeight > element.clientHeight || element.scrollWidth > element.clientWidth;
}

// Taken from here: https://stackoverflow.com/a/58393617
export function getHTMLGridColumsAndRows(element: HTMLElement) {
    // calc computed style
    const gridComputedStyle = window.getComputedStyle(element);

    return {
        // get number of grid rows
        gridRowCount: gridComputedStyle.getPropertyValue("grid-template-rows").split(" ").length,
        // get number of grid columns
        gridColumnCount: gridComputedStyle.getPropertyValue("grid-template-columns").split(" ").length,
        // get grid row sizes
        gridRowSizes: gridComputedStyle.getPropertyValue("grid-template-rows").split(" ").map(parseFloat),
        // get grid column sizes
        gridColumnSizes: gridComputedStyle.getPropertyValue("grid-template-columns").split(" ").map(parseFloat)
    };
}

export function buildGridArray(element: HTMLElement) {
    const { gridRowCount, gridColumnCount, gridRowSizes, gridColumnSizes } = getHTMLGridColumsAndRows(element);

    return Array.from({ length: gridRowCount }, (_, i) =>
        Array.from({ length: gridColumnCount }, (_, j) => {
            const elementAtPosition = element.children[i * gridColumnCount + j] as HTMLElement;
            return {
                element: elementAtPosition,
                row: i,
                column: j,
                rowSize: gridRowSizes[i],
                columnSize: gridColumnSizes[j],
                size: gridRowSizes[i] * gridColumnSizes[j]
            };
        })
    );
}

export function blurOnEsc(event: KeyboardEvent) {
    if (event.key === "Escape" || event.key === "Esc") {
        (event.currentTarget as HTMLElement)?.blur();
    }
}

export async function loadImage(url: string, element: HTMLImageElement, crossOrigin: string | null = null): Promise<boolean> {
    return new Promise((resolve, reject) => {
        element.crossOrigin = crossOrigin;
        element.src = url;
        element.onload = () => {
            resolve(true);
        };
        element.onerror = reject;
    });
}