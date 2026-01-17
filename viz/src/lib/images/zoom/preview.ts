// Adjust from: https://mykolas-mankevicius.medium.com/zoom-pan-clamp-image-preview-with-vanilla-javascript-090215211fc9
import { Renderer } from './renderer';

type InstanceState = 'idle' | 'singleGesture' | 'multiGesture' | 'mouse';

const MIN_SCALE = 1;
const MAX_SCALE = 4;
const DOUBLE_TAP_TIME = 185; // milliseconds
const INERTIA_FRICTION = 0.95; // 0-1, lower is more friction
const MIN_VELOCITY = 0.1;

export default class ZoomPan {
    protected container: HTMLElement;
    protected renderer: Renderer;
    protected state: InstanceState = 'idle';
    protected scaleValue = 1;
    protected lastTapTime = 0;
    protected deviceHasTouch = false;
    protected wheelTimeout: ReturnType<typeof window.setTimeout> | undefined;
    protected start = { x: 0, y: 0, distance: 0, touches: [] as Touch[] };
    protected pinchStartCenter = { x: 0, y: 0 }; // Store the center of the pinch at start
    protected onZoomChange: ((percentage: number) => void) | undefined;
    protected onTransformChangeCallback: ((transform: any) => void) | undefined;

    // Inertia state
    protected velocity = { x: 0, y: 0 };
    protected lastMoveTime = 0;
    protected animationFrameId: number | null = null;

    public isPanningDisabled = false;

    minScale: number = MIN_SCALE ;
    maxScale: number = MAX_SCALE ;

    constructor(container: HTMLElement, image: HTMLElement, minScale = MIN_SCALE, maxScale = MAX_SCALE) {
        this.container = container;
        this.minScale = minScale;
        this.maxScale = maxScale;
        this.renderer = new Renderer({
            container,
            minScale: this.minScale,
            maxScale: this.maxScale,
            element: image,
            scaleSensitivity: 20,
            onChange: (transform) => {
                if (this.onTransformChangeCallback) {
                    this.onTransformChangeCallback(transform);
                }
            }
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

    private stopInertia() {
        if (this.animationFrameId) {
            cancelAnimationFrame(this.animationFrameId);
            this.animationFrameId = null;
        }
        this.velocity = { x: 0, y: 0 };
    }

    private startInertia() {
        this.stopInertia();
        
        const loop = () => {
            if (Math.abs(this.velocity.x) < MIN_VELOCITY && Math.abs(this.velocity.y) < MIN_VELOCITY) {
                this.stopInertia();
                return;
            }

            this.renderer.panBy({ originX: this.velocity.x, originY: this.velocity.y });
            this.velocity.x *= INERTIA_FRICTION;
            this.velocity.y *= INERTIA_FRICTION;

            this.animationFrameId = requestAnimationFrame(loop);
        };

        this.animationFrameId = requestAnimationFrame(loop);
    }

    private onStart = (event: TouchEvent) => {
        if (this.isPanningDisabled) return;

        this.deviceHasTouch = true;
        this.stopInertia(); // Stop any existing movement

        if (this.stateIs('multiGesture')) {
            return;
        }

        const touchCount = event.touches.length;

        if (touchCount === 2 && this.stateIs('idle', 'singleGesture')) {
            const { x, y } = this.getMidPoint(event);

            this.start.x = x;
            this.start.y = y;
            this.pinchStartCenter = { x, y }; // Capture start center
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
        this.lastMoveTime = Date.now();
        this.velocity = { x: 0, y: 0 };
    };

    private onMove = (event: TouchEvent) => {
        if (this.isPanningDisabled) return;

        if (this.stateIs('idle')) {
            return;
        }


        const touchCount = event.touches.length;

        if (this.stateIs('multiGesture') && touchCount === 2) {
            event.preventDefault();
            const scale = this.getPinchDistance(event) / this.start.distance;
            const { x, y } = this.getMidPoint(event);

            this.renderer.zoomPan({ 
                scale, 
                x: this.pinchStartCenter.x, 
                y: this.pinchStartCenter.y, 
                deltaX: x - this.start.x, 
                deltaY: y - this.start.y 
            });

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
        
        // Calculate velocity
        const now = Date.now();
        const dt = now - this.lastMoveTime;
        if (dt > 0) {
            // Simple moving average or just last frame? Last frame is jumpy.
            // But for simple fling, last frame often works if dt is small.
            // Let's just track it for onEnd.
            // We want pixels per frame effectively for the loop.
            // If dt is large (lag), deltaX might be large.
            // Normalizing to "pixels per 16ms" might be good?
            // Or just store raw delta if we assume frequent updates.
            // Let's store raw delta but decay it if no move happens?
            this.velocity = { x: deltaX, y: deltaY };
        }
        this.lastMoveTime = now;

        this.renderer.panBy({ originX: deltaX, originY: deltaY });

        this.start.x = touch.pageX;
        this.start.y = touch.pageY;
    };

    private onEndTouch = (event: TouchEvent) => {
        if (this.isPanningDisabled) return;

        if (this.stateIs('idle') || event.touches.length !== 0) {
            return;
        }

        const currentTime = new Date().getTime();
        const tapLength = currentTime - this.lastTapTime;
        const timeSinceLastMove = currentTime - this.lastMoveTime;

        if (tapLength < DOUBLE_TAP_TIME && tapLength > 0) {
            event.preventDefault();
            const [touch] = event.changedTouches;
            if (!touch) {
                return;
            }

            this.setCurrentScale(this.onDoubleTap({ x: touch.clientX, y: touch.clientY }));
        } else {
             // If we stopped moving before lifting finger (held still), velocity should be 0.
            if (timeSinceLastMove > 50) {
                this.velocity = { x: 0, y: 0 };
            }
            if (Math.abs(this.velocity.x) > MIN_VELOCITY || Math.abs(this.velocity.y) > MIN_VELOCITY) {
                this.startInertia();
            }
        }

        this.lastTapTime = currentTime;
        this.setCurrentScale(this.renderer.getScale());
        this.state = 'idle';
    };

    private onWheel = (event: WheelEvent) => {
        if (this.deviceHasTouch) {
            return;
        }
        
        this.stopInertia();

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

    private isMouseDown = false;

    private onMouseDown = (event: MouseEvent) => {
        if (this.isPanningDisabled) return;

        if (this.deviceHasTouch) return;
        this.isMouseDown = true;
        this.stopInertia();
        this.velocity = { x: 0, y: 0 };
        this.lastMoveTime = Date.now();
    };

    private onMouseMove = (event: MouseEvent) => {
        if (this.isPanningDisabled) return;

        if (this.deviceHasTouch) {
            return;
        }

        if (!this.isMouseDown || event.buttons !== 1 || this.scaleValue === this.minScale) {
            return;
        }
        event.preventDefault();

        if (event.movementX === 0 && event.movementY === 0) {
            return;
        }
        
        this.velocity = { x: event.movementX, y: event.movementY };
        this.lastMoveTime = Date.now();

        this.state = 'mouse';
        this.renderer.panBy({ originX: event.movementX, originY: event.movementY });
    };

    private onMouseEnd = () => {
        if (this.deviceHasTouch) {
            return;
        }

        this.state = 'idle';
        this.isMouseDown = false;
        
        const timeSinceLastMove = Date.now() - this.lastMoveTime;
        if (timeSinceLastMove > 50) {
             this.velocity = { x: 0, y: 0 };
        }
        
        if (Math.abs(this.velocity.x) > MIN_VELOCITY || Math.abs(this.velocity.y) > MIN_VELOCITY) {
            this.startInertia();
        }

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

        // Only trigger tap logic if we started the interaction on this container
        if (this.isMouseDown && !this.stateIs('mouse')) {
            const currentTime = new Date().getTime();
            const clickLength = currentTime - this.lastTapTime;

            if (clickLength < DOUBLE_TAP_TIME && clickLength > 0) {
                this.setCurrentScale(this.onDoubleTap({ x: event.pageX, y: event.pageY }));
            }
            this.lastTapTime = currentTime;
        }

        this.onMouseEnd();
    };

    private attachEventListeners = () => {
        this.container.addEventListener('touchstart', this.onStart, { passive: false });
        this.container.addEventListener('touchmove', this.onMove, { passive: false });
        this.container.addEventListener('touchend', this.onEndTouch, { passive: false });
        this.container.addEventListener('touchcancel', this.onEndTouch, { passive: false });

        this.container.addEventListener('mousedown', this.onMouseDown, { passive: false });
        this.container.addEventListener('mousemove', this.onMouseMove, { passive: false });
        this.container.addEventListener('mouseup', this.onMouseUp, { passive: false });
        this.container.addEventListener('mouseleave', this.onMouseEnd, { passive: false });
        this.container.addEventListener('mouseout', this.onMouseEnd, { passive: false });
        this.container.addEventListener('wheel', this.onWheel, { passive: false });
    };

    public getZoomPercentage = (): number => {
        return this.renderer.getScalePercentage();
    };

    public getTransform = () => {
        return this.renderer.getTransform();
    };

    public setZoomPercentage = (percentage: number) => {
        this.renderer.setScalePercentage(percentage);
        this.setCurrentScale(this.renderer.getScale());
    };

    public onZoom = (callback: (percentage: number) => void) => {
        this.onZoomChange = callback;
    };

    public onTransformChange = (callback: (transform: any) => void) => {
        this.onTransformChangeCallback = callback;
    };

    public reset = () => {
        this.state = 'idle';
        this.stopInertia();
        this.setCurrentScale(1);
        this.lastTapTime = 0;
        this.start = { x: 0, y: 0, distance: 0, touches: [] };
        this.renderer.reset();
    };

    public destroy = () => {
        this.stopInertia();
        this.container.removeEventListener('touchstart', this.onStart);
        this.container.removeEventListener('touchmove', this.onMove);
        this.container.removeEventListener('touchend', this.onEndTouch);
        this.container.removeEventListener('touchcancel', this.onEndTouch);

        this.container.removeEventListener('mousedown', this.onMouseDown);
        this.container.removeEventListener('mousemove', this.onMouseMove);
        this.container.removeEventListener('mouseup', this.onMouseUp);
        this.container.removeEventListener('mouseleave', this.onMouseEnd);
        this.container.removeEventListener('mouseout', this.onMouseEnd);
        this.container.removeEventListener('wheel', this.onWheel);
    };
}