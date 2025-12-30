// Adjust from: https://mykolas-mankevicius.medium.com/zoom-pan-clamp-image-preview-with-vanilla-javascript-090215211fc9
import { Renderer } from './renderer';

type InstanceState = 'idle' | 'singleGesture' | 'multiGesture' | 'mouse';

const MIN_SCALE = 1;
const MAX_SCALE = 4;
const DOUBLE_TAP_TIME = 185; // milliseconds

export default class ZoomPan {
    private container: HTMLElement;
    private renderer: Renderer;
    private state: InstanceState = 'idle';
    private scaleValue = 1;
    private lastTapTime = 0;
    private deviceHasTouch = false;
    private wheelTimeout: ReturnType<typeof window.setTimeout> | undefined;
    private start = { x: 0, y: 0, distance: 0, touches: [] as Touch[] };
    private onZoomChange: ((percentage: number) => void) | undefined;

    minScale: number = MIN_SCALE ;
    maxScale: number = MAX_SCALE ;

    constructor(container: HTMLElement, image: HTMLImageElement, minScale = MIN_SCALE, maxScale = MAX_SCALE) {
        this.container = container;
        this.minScale = minScale;
        this.maxScale = maxScale;
        this.renderer = new Renderer({
            container,
            minScale: this.minScale,
            maxScale: this.maxScale,
            element: image,
            scaleSensitivity: 20
        });

        this.attachEventListeners();
    }

    private stateIs = (...states: InstanceState[]) => states.includes(this.state);

    private getPinchDistance = (event: TouchEvent): number =>
        Math.hypot(event.touches[0].pageX - event.touches[1].pageX, event.touches[0].pageY - event.touches[1].pageY);

    private getMidPoint = (event: TouchEvent): { x: number; y: number; } => ({
        x: (event.touches[0].pageX + event.touches[1].pageX) / 2,
        y: (event.touches[0].pageY + event.touches[1].pageY) / 2
    });

    private onDoubleTap = ({ x, y }: { x: number; y: number; }): number => {
        if (this.scaleValue < this.maxScale) {
            this.renderer.zoomTo({ newScale: this.maxScale, x, y });
            return this.maxScale;
        } else {
            this.renderer.reset();
            return this.minScale;
        }
    };

    private setCurrentScale = (value: number) => {
        this.scaleValue = value;
        this.container.style.cursor = value === this.minScale ? 'zoom-in' : 'move';
        if (this.onZoomChange) {
            this.onZoomChange(this.renderer.getScalePercentage());
        }
    };

    private onStart = (event: TouchEvent) => {
        this.deviceHasTouch = true;
        if (this.stateIs('multiGesture')) {
            return;
        }

        const touchCount = event.touches.length;

        if (touchCount === 2 && this.stateIs('idle', 'singleGesture')) {
            const { x, y } = this.getMidPoint(event);

            this.start.x = x;
            this.start.y = y;
            this.start.distance = this.getPinchDistance(event) / this.scaleValue;
            this.start.touches = [event.touches[0], event.touches[1]];

            this.lastTapTime = 0; // Reset to prevent misinterpretation as a double tap
            this.state = 'multiGesture';
            return;
        }

        if (touchCount !== 1) {
            this.state = 'idle';
            return;
        }

        this.state = 'singleGesture';

        const [touch] = event.touches;

        this.start.x = touch.pageX;
        this.start.y = touch.pageY;
        this.start.distance = 0;
        this.start.touches = [touch];
    };

    private onMove = (event: TouchEvent) => {
        if (this.stateIs('idle')) {
            return;
        }


        const touchCount = event.touches.length;

        if (this.stateIs('multiGesture') && touchCount === 2) {
            event.preventDefault();
            const scale = this.getPinchDistance(event) / this.start.distance;
            const { x, y } = this.getMidPoint(event);

            this.renderer.zoomPan({ scale, x, y, deltaX: x - this.start.x, deltaY: y - this.start.y });

            this.setCurrentScale(this.renderer.getScale());

            this.start.x = x;
            this.start.y = y;
            return;
        }

        if (
            this.scaleValue === this.minScale ||
            !this.stateIs('singleGesture') ||
            touchCount !== 1 ||
            event.touches[0]?.identifier !== this.start.touches[0]?.identifier
        ) {
            return;
        }
        event.preventDefault();

        const [touch] = event.touches;
        const deltaX = touch.pageX - this.start.x;
        const deltaY = touch.pageY - this.start.y;

        this.renderer.panBy({ originX: deltaX, originY: deltaY });

        this.start.x = touch.pageX;
        this.start.y = touch.pageY;
    };

    private onEndTouch = (event: TouchEvent) => {
        if (this.stateIs('idle') || event.touches.length !== 0) {
            return;
        }

        const currentTime = new Date().getTime();
        const tapLength = currentTime - this.lastTapTime;

        if (tapLength < DOUBLE_TAP_TIME && tapLength > 0) {
            event.preventDefault();
            const [touch] = event.changedTouches;
            if (!touch) {
                return;
            }

            this.setCurrentScale(this.onDoubleTap({ x: touch.clientX, y: touch.clientY }));
        }

        this.lastTapTime = currentTime;
        this.setCurrentScale(this.renderer.getScale());
        this.state = 'idle';
    };

    private onWheel = (event: WheelEvent) => {
        if (this.deviceHasTouch) {
            return;
        }

        event.preventDefault();
        this.renderer.zoom({
            deltaScale: Math.sign(event.deltaY) > 0 ? -1 : 1,
            x: event.pageX,
            y: event.pageY
        });

        if (this.onZoomChange) {
            this.onZoomChange(this.renderer.getScalePercentage());
        }

        clearTimeout(this.wheelTimeout);
        this.wheelTimeout = setTimeout(() => {
            this.setCurrentScale(this.renderer.getScale());
        }, 100);
    };

    private onMouseMove = (event: MouseEvent) => {
        if (this.deviceHasTouch) {
            return;
        }

        if (event.buttons !== 1 || this.scaleValue === this.minScale) {
            return;
        }
        event.preventDefault();

        if (event.movementX === 0 && event.movementY === 0) {
            return;
        }

        this.state = 'mouse';
        this.renderer.panBy({ originX: event.movementX, originY: event.movementY });
    };

    private onMouseEnd = () => {
        if (this.deviceHasTouch) {
            return;
        }

        this.state = 'idle';
        this.setCurrentScale(this.renderer.getScale());
    };

    private onMouseUp = (event: MouseEvent) => {
        if (this.deviceHasTouch) {
            return;
        }

        // Ignore right-click
        if (event.button === 2) {
            return;
        }

        if (!this.stateIs('mouse')) {
            this.setCurrentScale(this.onDoubleTap({ x: event.pageX, y: event.pageY }));
        }

        this.onMouseEnd();
    };

    private attachEventListeners = () => {
        this.container.addEventListener('touchstart', this.onStart, { passive: false });
        this.container.addEventListener('touchmove', this.onMove, { passive: false });
        this.container.addEventListener('touchend', this.onEndTouch, { passive: false });
        this.container.addEventListener('touchcancel', this.onEndTouch, { passive: false });

        this.container.addEventListener('mousemove', this.onMouseMove, { passive: false });
        this.container.addEventListener('mouseup', this.onMouseUp, { passive: false });
        this.container.addEventListener('mouseleave', this.onMouseEnd, { passive: false });
        this.container.addEventListener('mouseout', this.onMouseEnd, { passive: false });
        this.container.addEventListener('wheel', this.onWheel, { passive: false });
    };

    public getZoomPercentage = (): number => {
        return this.renderer.getScalePercentage();
    };

    public setZoomPercentage = (percentage: number) => {
        this.renderer.setScalePercentage(percentage);
        this.setCurrentScale(this.renderer.getScale());
    };

    public onZoom = (callback: (percentage: number) => void) => {
        this.onZoomChange = callback;
    };

    public reset = () => {
        this.state = 'idle';
        this.setCurrentScale(1);
        this.lastTapTime = 0;
        this.start = { x: 0, y: 0, distance: 0, touches: [] };
        this.renderer.reset();
    };

    public destroy = () => {
        this.container.removeEventListener('touchstart', this.onStart);
        this.container.removeEventListener('touchmove', this.onMove);
        this.container.removeEventListener('touchend', this.onEndTouch);
        this.container.removeEventListener('touchcancel', this.onEndTouch);

        this.container.removeEventListener('mousemove', this.onMouseMove);
        this.container.removeEventListener('mouseup', this.onMouseUp);
        this.container.removeEventListener('mouseleave', this.onMouseEnd);
        this.container.removeEventListener('mouseout', this.onMouseEnd);
        this.container.removeEventListener('wheel', this.onWheel);
    };
}