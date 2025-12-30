class Renderer {
    private container: HTMLElement;
    private element: HTMLElement;
    private minScale: number;
    private maxScale: number;
    private scaleSensitivity: number;
    private transform: {
        originOffset: boolean;
        originX: number;
        originY: number;
        translateX: number;
        translateY: number;
        scale: number;
    };

    private static readonly DEFAULT_TRANSFORMATION = {
        originOffset: false,
        originX: 0,
        originY: 0,
        translateX: 0,
        translateY: 0,
        scale: 1
    };

    constructor({
        container,
        element,
        minScale,
        maxScale,
        scaleSensitivity = 10
    }: {
        container: HTMLElement;
        element: HTMLElement;
        minScale: number;
        maxScale: number;
        scaleSensitivity?: number;
    }) {
        this.container = container;
        this.element = element;
        this.minScale = minScale;
        this.maxScale = maxScale;
        this.scaleSensitivity = scaleSensitivity;
        this.transform = { ...Renderer.DEFAULT_TRANSFORMATION };
    }

    private valueInRange = () => this.transform.scale <= this.maxScale && this.transform.scale >= this.minScale;

    private getTranslate = ({ pos, axis }: { pos: number; axis: 'x' | 'y'; }) => {
        const { originX, originY, translateX, translateY, scale } = this.transform;
        const axisIsX = axis === 'x';
        const prevPos = axisIsX ? originX : originY;
        const translate = axisIsX ? translateX : translateY;

        return this.valueInRange() && pos !== prevPos
            ? translate + (pos - prevPos * scale) * (1 - 1 / scale)
            : translate;
    };

    private getMatrix = () =>
        `matrix(${this.transform.scale}, 0, 0, ${this.transform.scale}, ${this.transform.translateX}, ${this.transform.translateY})`;

    private clamp = (value: number, min: number, max: number) => Math.max(Math.min(value, max), min);

    private getNewScale = (deltaScale: number) => {
        const newScale = this.transform.scale + deltaScale / (this.scaleSensitivity / this.transform.scale);
        return this.clamp(newScale, this.minScale, this.maxScale);
    };

    private clampedTranslate = ({ axis, translate }: { axis: 'x' | 'y'; translate: number; }) => {
        const { scale, originX, originY } = this.transform;
        const axisIsX = axis === 'x';
        const origin = axisIsX ? originX : originY;
        const axisKey = axisIsX ? 'offsetWidth' : 'offsetHeight';

        const containerSize = this.container[axisKey];
        const imageSize = this.element[axisKey];
        const bounds = this.element.getBoundingClientRect();

        const imageScaledSize = axisIsX ? bounds.width : bounds.height;

        const defaultOrigin = imageSize / 2;
        const originOffset = (origin - defaultOrigin) * (scale - 1);

        const range = Math.max(0, Math.round(imageScaledSize) - containerSize);

        const max = Math.round(range / 2);
        const min = 0 - max;

        return this.clamp(translate, min + originOffset, max + originOffset);
    };

    private renderClamped = ({ translateX, translateY }: { translateX: number; translateY: number; }) => {
        this.transform.translateX = this.clampedTranslate({ axis: 'x', translate: translateX });
        this.transform.translateY = this.clampedTranslate({ axis: 'y', translate: translateY });

        requestAnimationFrame(() => {
            if (this.transform.originOffset) {
                this.element.style.transformOrigin = `${this.transform.originX}px ${this.transform.originY}px`;
            }
            this.element.style.transform = this.getMatrix();
        });
    };

    private pan = ({ originX, originY }: { originX: number; originY: number; }) => {
        this.renderClamped({
            translateX: this.transform.translateX + originX,
            translateY: this.transform.translateY + originY
        });
    };

    public panBy = (origin: { originX: number; originY: number; }) => this.pan(origin);

    public panTo = ({ originX, originY, scale }: { originX: number; originY: number; scale: number; }) => {
        this.transform.scale = this.clamp(scale, this.minScale, this.maxScale);
        this.pan({
            originX: originX - this.transform.translateX,
            originY: originY - this.transform.translateY
        });
    };

    public zoomPan = ({
        scale: scaleValue,
        x,
        y,
        deltaX,
        deltaY
    }: {
        scale: number;
        x: number;
        y: number;
        deltaX: number;
        deltaY: number;
    }) => {
        const newScale = this.clamp(scaleValue, this.minScale, this.maxScale);
        const { left, top } = this.element.getBoundingClientRect();
        const originX = x - left;
        const originY = y - top;
        const newOriginX = originX / this.transform.scale;
        const newOriginY = originY / this.transform.scale;
        const translateX = this.getTranslate({ pos: originX, axis: 'x' });
        const translateY = this.getTranslate({ pos: originY, axis: 'y' });

        this.transform = {
            originOffset: true,
            originX: newOriginX,
            originY: newOriginY,
            translateX,
            translateY,
            scale: newScale
        };

        this.pan({ originX: deltaX, originY: deltaY });
    };

    public zoom = ({ x, y, deltaScale }: { x: number; y: number; deltaScale: number; }) => {
        const { scale } = this.transform;
        const { left, top } = this.element.getBoundingClientRect();
        const newScale = this.getNewScale(deltaScale);
        const originX = x - left;
        const originY = y - top;
        const newOriginX = originX / scale;
        const newOriginY = originY / scale;

        const translateX = this.getTranslate({ pos: originX, axis: 'x' });
        const translateY = this.getTranslate({ pos: originY, axis: 'y' });

        this.transform = {
            ...this.transform,
            originOffset: true,
            originX: newOriginX,
            originY: newOriginY,
            scale: newScale
        };

        this.renderClamped({ translateX, translateY });
    };

    public zoomTo = ({ newScale, x, y }: { newScale: number; x: number; y: number; }) => {
        const { scale } = this.transform;

        const { left, top } = this.element.getBoundingClientRect();
        const originX = x - left;
        const originY = y - top;
        const newOriginX = originX / scale;
        const newOriginY = originY / scale;

        const translateX = this.getTranslate({ pos: originX, axis: 'x' });
        const translateY = this.getTranslate({ pos: originY, axis: 'y' });

        this.transform = {
            originOffset: true,
            originX: newOriginX,
            originY: newOriginY,
            scale: newScale,
            translateX,
            translateY
        };

        requestAnimationFrame(() => {
            this.element.style.transformOrigin = `${newOriginX}px ${newOriginY}px`;
            this.element.style.transform = `matrix(${newScale}, 0, 0, ${newScale}, ${translateX}, ${translateY})`;
        });
    };

    public getScale = () => this.transform.scale;

    public getScalePercentage = (): number => {
        const { scale } = this.transform;
        const percentage = (scale - this.minScale) / (this.maxScale - this.minScale) * 100;
        return this.clamp(percentage, 0, 100);
    }

    public setScalePercentage = (percentage: number) => {
        const newScale = this.minScale + (this.clamp(percentage, 0, 100) / 100) * (this.maxScale - this.minScale);
        
        const { offsetWidth, offsetHeight } = this.container;
        const x = offsetWidth / 2;
        const y = offsetHeight / 2;
    
        this.zoomTo({ newScale, x, y });
    }

    public reset = () => {
        this.transform = { ...Renderer.DEFAULT_TRANSFORMATION };
        this.pan({ originX: 0, originY: 0 });
    };
}

export { Renderer };