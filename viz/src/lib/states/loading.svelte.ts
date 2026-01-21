/**
 * loading.svelte.ts
 * 
 * Centralized state for tracking navigation progress.
 * Only tracks network requests that occur during an active navigation.
 */
import { untrack } from "svelte";

class LoadingState {
    private activeRequests = $state(0);
    private totalRequestsInSession = $state(0);
    private completedRequestsInSession = $state(0);

    // Whether we are currently in a "navigation" phase
    isNavigating = $state(false);

    /** Calculated progress (0-100) based on requests in the current navigation */
    progress = $state(100);

    private updateProgress() {
        if (!this.isNavigating) {
            this.progress = 100;
            return;
        }

        if (this.totalRequestsInSession === 0) {
            this.progress = 15; // Initial jump
            return;
        }

        // Calculate progress based on completed vs total
        const baseProgress = (this.completedRequestsInSession / this.totalRequestsInSession) * 100;

        // Scale to 15-95 range so it doesn't hit 100 until we say so
        const weightedProgress = 15 + (baseProgress * 0.8);

        this.progress = Math.min(weightedProgress, 95);
    }

    startRequest() {
        untrack(() => {
            // Only track requests that happen during navigation
            if (!this.isNavigating) return;

            this.activeRequests++;
            this.totalRequestsInSession++;
            this.updateProgress();
        });
    }

    endRequest() {
        untrack(() => {
            if (!this.isNavigating) return;

            this.activeRequests = Math.max(0, this.activeRequests - 1);
            this.completedRequestsInSession++;
            this.updateProgress();
        });
    }

    startNavigation() {
        this.totalRequestsInSession = 0;
        this.completedRequestsInSession = 0;
        this.activeRequests = 0;
        this.isNavigating = true;
        this.updateProgress();
    }

    endNavigation() {
        this.isNavigating = false;
        this.activeRequests = 0;
        this.updateProgress();
    }
}

export const loadingState = new LoadingState();
