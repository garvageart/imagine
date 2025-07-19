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