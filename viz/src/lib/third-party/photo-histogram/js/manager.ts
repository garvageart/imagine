import * as util from "./util";

export interface SvgCreateAttributes {
    id?: string;
    class?: string;
    viewBox?: string;
    width?: string;
    height?: string;
    style?: string;
    preserveAspectRatio?: string;
}

/**
 * Manage creating and removing elements from an svg element
 */
export class Svg {
    static readonly xmlns = "http://www.w3.org/2000/svg";
    element: SVGSVGElement;

    constructor(element: SVGSVGElement) {
        this.element = element;
    }

    /**
     * Create Svg element and append to parent
     */
    static create(parent: Element, attributes?: SvgCreateAttributes): Svg {
        const element = util.createElementNS("svg", Svg.xmlns, attributes, parent) as SVGSVGElement;
        return new Svg(element);
    }

    /**
     * Remove all children from the svg element
     */
    clear(): void {
        while (this.element.firstChild) {
            this.element.removeChild(this.element.firstChild);
        }
    }

    /**
     * Create and append a rect element
     */
    rect(x: number, y: number, width: number, height: number, attributes?: Record<string, any>): SVGRectElement {
        const attrs = attributes ?? {};
        attrs.x = x.toString(10);
        attrs.y = y.toString(10);
        attrs.width = width.toString(10);
        attrs.height = height.toString(10);
        return util.createElementNS("rect", Svg.xmlns, attrs, this.element) as SVGRectElement;
    }

    /**
     * Create and append a line element
     */
    line(x1: number, y1: number, x2: number, y2: number, attributes?: Record<string, any>): SVGLineElement {
        const attrs = attributes ?? {};
        attrs.x1 = x1.toString(10);
        attrs.y1 = y1.toString(10);
        attrs.x2 = x2.toString(10);
        attrs.y2 = y2.toString(10);
        return util.createElementNS("line", Svg.xmlns, attrs, this.element) as SVGLineElement;
    }

    /**
     * Create and append a path element
     */
    path(d: string, attributes?: Record<string, any>): SVGPathElement {
        const attrs = attributes ?? {};
        attrs.d = d;
        return util.createElementNS("path", Svg.xmlns, attrs, this.element) as SVGPathElement;
    }
}

/**
 * Builds d attribute to construct a path element
 */
export class SvgPathBuilder {
    private _data = "";
    private _manager: Svg;

    constructor(svgManager: Svg) {
        this._manager = svgManager;
    }

    clear(): void {
        this._data = "";
    }

    build(attributes?: Record<string, any>): SVGPathElement {
        return this._manager.path(this._data, attributes);
    }

    private appendCommand(command: string, x: number, y: number): SvgPathBuilder {
        const coordinate = `${x} ${y}`;
        this._data = this._data === "" ? `${command} ${coordinate}` : `${this._data} ${command} ${coordinate}`;
        return this;
    }

    moveTo(x: number, y: number): SvgPathBuilder {
        return this.appendCommand("M", x, y);
    }

    lineTo(x: number, y: number): SvgPathBuilder {
        return this.appendCommand("L", x, y);
    }
}


