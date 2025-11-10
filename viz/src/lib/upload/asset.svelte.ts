import { uploadImageWithProgress, type UploadImageResult } from "$lib/api";
import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
import type { ImageUploadFileData, ImageUploadSuccess } from "./manager.svelte";

export enum UploadState {
    PENDING,
    STARTED,
    DONE,
    ERROR,
    CANCELED,
    INVALID
}

export interface UploadImageStats {
    progress: number;
    state: UploadState;
    startTime?: Date;
    endTime?: Date;
}

export class UploadImage implements UploadImageStats {
    progress: number = $state(0);
    state: UploadState = $state(UploadState.PENDING);
    startTime?: Date = $state(new Date());
    checksum?: string;
    data: ImageUploadFileData;
    request: XMLHttpRequest | undefined = $state(undefined);

    constructor(data: ImageUploadFileData) {
        this.checksum = data.checksum;
        this.data = data;
    }

    reset() {
        this.progress = 0;
        this.state = UploadState.PENDING;
    }

    cancelRequest() {
        this.state = UploadState.CANCELED;
        if (this.request) {
            this.request.abort();
        }
    }

    private updateProgress = (event: ProgressEvent<XMLHttpRequestEventTarget>) => {
        // Some browsers don't provide total (lengthComputable=false). Fallback to file size when possible.
        if (event.lengthComputable && event.total > 0) {
            this.progress = Math.min(100, (event.loaded / event.total) * 100);
        } else if ((this.data as any)?.data?.size) {
            const total = (this.data as any).data.size as number;
            this.progress = Math.min(100, (event.loaded / total) * 100);
        } else {
            // As a last resort, show indeterminate progress by nudging a bit until completion
            this.progress = Math.min(95, this.progress + 1);
        }
    };

    async upload(): Promise<UploadImageResult | undefined> {
        try {
            this.state = UploadState.STARTED;
            const responseData = await uploadImageWithProgress({
                data: this.data,
                onUploadProgress: this.updateProgress,
                request: this.request
            });

            this.state = (responseData.status === 200) || (responseData.status === 201) ? UploadState.DONE : UploadState.INVALID;
            if (responseData.status !== 200 && responseData.status !== 201) {
                throw new Error(`Upload failed with status ${responseData.status}`);
            }

            toastState.addToast({
                message: this.state === UploadState.DONE ? `Uploaded ${this.data.filename} successfully.` : `Failed to upload ${this.data.filename}.`,
                type: this.state === UploadState.DONE ? 'success' : 'error',
                dismissible: true,
                timeout: 5000
            });

            return responseData.data;
        } catch (error) {
            this.state = UploadState.ERROR;
            toastState.addToast({
                message: `Error uploading ${this.data.filename}.`,
                type: 'error',
                dismissible: true,
                timeout: 2000
            });
            return undefined;
        }
    }
}