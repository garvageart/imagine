import ZoomPan from "./preview";

export interface CropRect {
    x: number;
    y: number;
    width: number;
    height: number;
}

export type DragAction = "move" | "nw" | "ne" | "sw" | "se" | "n" | "e" | "s" | "w" | null;
export const DragActionName: Record<Exclude<DragAction, null>, string> & { null: string; } = {
    move: "Move",
    nw: "NW",
    ne: "NE",
    sw: "SW",
    se: "SE",
    n: "N",
    e: "E",
    s: "S",
    w: "W",
    null: "None"
};

export class ZoomPanCrop extends ZoomPan {
    private crop: CropRect = { x: 0, y: 0, width: 0, height: 0 };
    private onCropChangeCallback: ((crop: CropRect) => void) | undefined;

    // Crop state
    private isDraggingCrop = false;
    private dragAction: DragAction = null;
    private startX = 0;
    private startY = 0;
    private startCrop: CropRect = { x: 0, y: 0, width: 0, height: 0 };

    private aspectRatio: number | null = null;
    private minWidth = 50;
    private minHeight = 50;

    private originalWidth = 0;
    private originalHeight = 0;

    constructor(container: HTMLElement, image: HTMLElement, minScale = 1, maxScale = 4) {
        super(container, image, minScale, maxScale);

        // Initialize dimensions
        // We need the unscaled dimensions or we treat the current rendered dimensions as base.
        // Usually crop is relative to the image's intrinsic or rendered size.
        // In CropOverlay.svelte, it took 'width' and 'height' as props (rendered size).
        // Here we can get it from the element, but we must account for current scale if it's already zoomed?
        // Actually, ZoomPan initializes with scale 1 usually.
        this.originalWidth = image.offsetWidth; // clientWidth?
        this.originalHeight = image.offsetHeight;
    }

    public setCropDimensions(width: number, height: number) {
        this.originalWidth = width;
        this.originalHeight = height;
    }

    public setCrop(crop: CropRect) {
        this.crop = { ...crop };
        this.emitCropChange();
    }

    public getCrop(): CropRect {
        return { ...this.crop };
    }

    public setAspectRatio(ratio: number | null) {
        this.aspectRatio = ratio;
        if (ratio) {
            this.applyAspectRatio(ratio);
        }
    }

    public onCropChange(callback: (crop: CropRect) => void) {
        this.onCropChangeCallback = callback;
    }

    private emitCropChange() {
        if (this.onCropChangeCallback) {
            this.onCropChangeCallback(this.crop);
        }
    }

    // Initialize crop
    public initCrop(initialCrop?: CropRect | null) {
        // Update dimensions in case image loaded/changed
        // Note: this.renderer.element is the image/wrapper
        // However, ZoomPan transforms the wrapper.
        // We want the dimensions of the content being cropped.
        // If the wrapper is scaled, offsetWidth is the unscaled CSS width (layout width).
        const el = this.renderer.getElement();
        this.originalWidth = el.offsetWidth || el.clientWidth;
        this.originalHeight = el.offsetHeight || el.clientHeight;

        if (initialCrop) {
            this.crop = { ...initialCrop };
        } else {
            this.crop = {
                x: 0,
                y: 0,
                width: this.originalWidth,
                height: this.originalHeight
            };
        }
        this.emitCropChange();
    }

    private applyAspectRatio(ratio: number) {
        if (this.crop.width === 0 || this.crop.height === 0) return;

        let newH = this.crop.width / ratio;
        let newW = this.crop.width;

        const width = this.originalWidth;
        const height = this.originalHeight;

        if (this.crop.y + newH > height) {
            newH = height - this.crop.y;
            if (newH < this.minHeight) {
                newH = this.minHeight;
            }

            newW = newH * ratio;

            if (this.crop.x + newW > width) {
                if (width / height > ratio) {
                    newH = height;
                    newW = height * ratio;
                } else {
                    newW = width;
                    newH = width / ratio;
                }

                this.crop = {
                    x: (width - newW) / 2,
                    y: (height - newH) / 2,
                    width: newW,
                    height: newH
                };

                this.emitCropChange();
                return;
            }
        } else if (newH < this.minHeight) {
            newH = this.minHeight;
            newW = newH * ratio;
        }

        this.crop.height = newH;
        this.crop.width = newW;
        this.emitCropChange();
    }

    private getClientPos(e: MouseEvent | TouchEvent) {
        if ("touches" in e) {
            return { x: e.touches[0].clientX, y: e.touches[0].clientY };
        }

        return { x: (e as MouseEvent).clientX, y: (e as MouseEvent).clientY };
    }

    public startCropDrag(action: DragAction, e: MouseEvent | TouchEvent) {
        if (e instanceof MouseEvent && e.button !== 0) {
            return;
        }

        e.preventDefault();
        e.stopPropagation(); // Stop ZoomPan from seeing this

        this.isDraggingCrop = true;
        this.dragAction = action;

        const pos = this.getClientPos(e);
        this.startX = pos.x;
        this.startY = pos.y;
        this.startCrop = { ...this.crop };

        window.addEventListener("mousemove", this.handleCropMove);
        window.addEventListener("touchmove", this.handleCropMove, { passive: false });
        window.addEventListener("mouseup", this.handleCropEnd);
        window.addEventListener("touchend", this.handleCropEnd);
    }

    private handleCropMove = (e: MouseEvent | TouchEvent) => {
        if (!this.isDraggingCrop || !this.dragAction) return;

        if (e instanceof MouseEvent && e.buttons === 0) {
            this.handleCropEnd(e);
            return;
        }

        e.preventDefault();
        e.stopPropagation();

        const pos = this.getClientPos(e);

        // Calculate Scale
        const scale = this.renderer.getScale();

        const dx = (pos.x - this.startX) / scale;
        const dy = (pos.y - this.startY) / scale;

        const width = this.originalWidth;
        const height = this.originalHeight;

        let newCrop = { ...this.startCrop };

        if (this.dragAction === "move") {
            newCrop.x = Math.max(0, Math.min(width - newCrop.width, this.startCrop.x + dx));
            newCrop.y = Math.max(0, Math.min(height - newCrop.height, this.startCrop.y + dy));
        } else {
            const isWest = this.dragAction.includes("w");
            const isEast = this.dragAction.includes("e");
            const isNorth = this.dragAction.includes("n");
            const isSouth = this.dragAction.includes("s");

            const isAlt = e.altKey;
            const startCx = this.startCrop.x + this.startCrop.width / 2;
            const startCy = this.startCrop.y + this.startCrop.height / 2;

            // 1. Calculate Unconstrained New Dimensions
            let candidateX = newCrop.x;
            let candidateY = newCrop.y;
            let candidateW = newCrop.width;
            let candidateH = newCrop.height;

            if (isAlt) {
                // Resize from Center
                if (isWest) candidateW = this.startCrop.width - dx * 2;
                else if (isEast) candidateW = this.startCrop.width + dx * 2;

                if (isNorth) candidateH = this.startCrop.height - dy * 2;
                else if (isSouth) candidateH = this.startCrop.height + dy * 2;

                candidateX = startCx - candidateW / 2;
                candidateY = startCy - candidateH / 2;
            } else {
                // Resize from Corner/Edge
                if (isWest) {
                    candidateW = this.startCrop.width - dx;
                    candidateX = this.startCrop.x + dx;
                }
                if (isEast) {
                    candidateW = this.startCrop.width + dx;
                }
                if (isNorth) {
                    candidateH = this.startCrop.height - dy;
                    candidateY = this.startCrop.y + dy;
                }
                if (isSouth) {
                    candidateH = this.startCrop.height + dy;
                }
            }

            // 2. Enforce Minimums (before aspect ratio)
            if (candidateW < this.minWidth) {
                const diff = this.minWidth - candidateW;
                candidateW = this.minWidth;
                if (isAlt) candidateX = startCx - candidateW / 2;
                else if (isWest) candidateX -= diff;
            }
            if (candidateH < this.minHeight) {
                const diff = this.minHeight - candidateH;
                candidateH = this.minHeight;
                if (isAlt) candidateY = startCy - candidateH / 2;
                else if (isNorth) candidateY -= diff;
            }

            // 3. Enforce Aspect Ratio
            // Check for explicit aspect ratio OR Shift key override
            const effectiveAspectRatio = this.aspectRatio ?? (e.shiftKey ? this.startCrop.width / this.startCrop.height : null);

            if (effectiveAspectRatio) {
                // Simplified approach: width drives height unless N/S drag only.
                const widthDriven = isEast || isWest || (!isNorth && !isSouth);

                if (widthDriven) {
                    let targetH = candidateW / effectiveAspectRatio;

                    if (isAlt) {
                        candidateH = targetH;
                        candidateY = startCy - candidateH / 2;
                    } else {
                        // If strict side drag (E/W), center Y. Else (Corner), anchor opposite corner.
                        if (!isNorth && !isSouth) {
                            candidateY = this.startCrop.y + (this.startCrop.height - targetH) / 2;
                            candidateH = targetH;
                        } else {
                            candidateH = targetH;
                            if (isNorth) {
                                // Anchor South (bottom)
                                const oldBottom = this.startCrop.y + this.startCrop.height;
                                candidateY = oldBottom - candidateH;
                            }
                        }
                    }
                } else {
                    // Height driven (North/South drag only)
                    let targetW = candidateH * effectiveAspectRatio;

                    if (isAlt) {
                        candidateW = targetW;
                        candidateX = startCx - candidateW / 2;
                    } else {
                        // Center X expansion for N/S drag
                        candidateX = this.startCrop.x + (this.startCrop.width - targetW) / 2;
                        candidateW = targetW;
                    }
                }
            }

            // 4. Enforce Bounds
            if (isAlt) {
                // Symmetric Bounds: Limit W/H to 2x distance to nearest edge from center
                const maxW = Math.min(startCx, width - startCx) * 2;
                const maxH = Math.min(startCy, height - startCy) * 2;

                if (candidateW > maxW) {
                    candidateW = maxW;
                    candidateX = startCx - candidateW / 2;
                }
                if (candidateH > maxH) {
                    candidateH = maxH;
                    candidateY = startCy - candidateH / 2;
                }

                // If Aspect Ratio exists, clipping one dimension might break ratio.
                if (effectiveAspectRatio) {
                    if (candidateW / candidateH > effectiveAspectRatio + 0.001) {
                        // Width too big relative to Height (because Height was clamped)
                        candidateW = candidateH * effectiveAspectRatio;
                        candidateX = startCx - candidateW / 2;
                    } else if (candidateW / candidateH < effectiveAspectRatio - 0.001) {
                        // Height too big relative to Width
                        candidateH = candidateW / effectiveAspectRatio;
                        candidateY = startCy - candidateH / 2;
                    }
                }

            } else {
                // Standard Bounds (Iterative correction)
                // Helper to apply bounds to X/W
                const clampX = () => {
                    if (candidateX < 0) {
                        candidateW += candidateX; // shrinking width
                        candidateX = 0;
                    }
                    if (candidateX + candidateW > width) {
                        candidateW = width - candidateX;
                    }
                };

                // Helper to apply bounds to Y/H
                const clampY = () => {
                    if (candidateY < 0) {
                        candidateH += candidateY;
                        candidateY = 0;
                    }
                    if (candidateY + candidateH > height) {
                        candidateH = height - candidateY;
                    }
                };

                clampX();
                clampY();

                // If aspect ratio, clamping one might break ratio. Recalculate and Re-clamp.
                if (effectiveAspectRatio) {
                    // Try fitting width first
                    if (candidateW / candidateH > effectiveAspectRatio) {
                        // Too wide: shrink width
                        candidateW = candidateH * effectiveAspectRatio;
                    } else if (candidateW / candidateH < effectiveAspectRatio) {
                        // Too tall: shrink height
                        candidateH = candidateW / effectiveAspectRatio;
                    }

                    // Re-check bounds
                    if (candidateX < 0) candidateX = 0;
                    if (candidateY < 0) candidateY = 0;
                    if (candidateW > width) candidateW = width;
                    if (candidateH > height) candidateH = height;
                }
            }

            newCrop = {
                x: candidateX,
                y: candidateY,
                width: candidateW,
                height: candidateH
            };
        }

        this.crop = newCrop;
        this.emitCropChange();
    };

    private handleCropEnd = (e: MouseEvent | TouchEvent) => {
        if (this.isDraggingCrop) {
            e.stopPropagation();
            if (e.cancelable) {
                e.preventDefault();
            }
        }

        this.isDraggingCrop = false;
        this.dragAction = null;
        window.removeEventListener("mousemove", this.handleCropMove);
        window.removeEventListener("touchmove", this.handleCropMove);
        window.removeEventListener("mouseup", this.handleCropEnd);
        window.removeEventListener("touchend", this.handleCropEnd);
    };
}
