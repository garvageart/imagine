import { upload } from "$lib/states/index.svelte";
import type { SupportedImageTypes, SupportedRAWFiles } from "$lib/types/images";
import { UploadImage, UploadState } from "./asset.svelte";

export interface ImageUploadFileData {
    file_name: string;
    data: File;
    checksum?: string;
}

export interface ImageUploadSuccess {
    uid: string;
    metadata?: any;
}

// Module-level state to be shared across all UploadManager instances
let activeCount = $state(0);

/**
 * Waits for a list of upload tasks to complete (success, error, cancel, or duplicate).
 */
export function waitForUploadCompletion(tasks: UploadImage[]): Promise<void> {
    return new Promise((resolve) => {
        const check = () => {
            const allDone = tasks.every(
                (t) =>
                    t.state === UploadState.DONE ||
                    t.state === UploadState.ERROR ||
                    t.state === UploadState.CANCELED ||
                    t.state === UploadState.DUPLICATE
            );

            if (allDone) {
                resolve();
            } else {
                setTimeout(check, 200);
            }
        };
        check();
    });
}

/**
 * dynamic queue processor that respects global concurrency.
 * Can be called repeatedly to fill available slots.
 */
export function processGlobalQueue() {
    if (activeCount >= upload.concurrency) {
        return;
    }

    const pendingTasks = upload.files.filter(t => t.state === UploadState.PENDING);

    if (pendingTasks.length === 0) {
        return;
    }

    const slotsAvailable = upload.concurrency - activeCount;
    const tasksToStart = pendingTasks.slice(0, slotsAvailable);

    for (const task of tasksToStart) {
        activeCount++;

        task.upload().finally(() => {
            activeCount--;
            processGlobalQueue();
        });
    }
}

/**
 * Complete rewrite: Clean upload manager for drag-and-drop and file picker.
 * Files are immediately added to global upload state so the panel shows right away.
 */
export default class UploadManager {
    allowedTypes: string[];

    constructor(allowedTypes: (SupportedImageTypes | SupportedRAWFiles)[]) {
        this.allowedTypes = allowedTypes;
    }

    /**
     * Add files programmatically (e.g., from drag-and-drop).
     * Files are immediately added to the global upload.files array so the panel appears.
     * Returns array of created UploadImage tasks.
     */
    addFiles(files: File[]): UploadImage[] {
        const tasks: UploadImage[] = [];

        for (const file of files) {
            // Validate file type
            const fileType = file.type.split("/")[1];
            if (!this.allowedTypes.includes(fileType)) {
                console.warn(`Skipping unsupported file type: ${file.type}`);
                continue;
            }

            // Create upload task
            const task = new UploadImage({
                file_name: file.name,
                data: file
            });

            tasks.push(task);
        }

        // Immediately add to global state (panel shows when upload.files.length > 0)
        if (tasks.length > 0) {
            upload.files.push(...tasks);
            upload.stats.total += tasks.length;
        }

        return tasks;
    }

    /**
     * Start uploading tasks with concurrency control.
     * If no tasks provided, uploads all pending tasks in the global store.
     */
    async start(tasks?: UploadImage[]): Promise<void> {
        processGlobalQueue();
    }

    /**
     * dynamic queue processor that respects global concurrency.
     * Can be called repeatedly to fill available slots.
     */
    processQueue() {
        processGlobalQueue();
    }

    /**
     * Open file picker dialog.
     * Creates a hidden input, triggers click, and returns selected files.
     */
    openPicker(): Promise<File[]> {
        return new Promise((resolve) => {
            const input = document.createElement("input");
            input.type = "file";
            input.multiple = true;
            input.accept = this.allowedTypes.map(t => `image/${t}`).join(",");

            input.onchange = () => {
                const files = Array.from(input.files || []);
                input.remove();
                resolve(files);
            };

            input.click();
        });
    }

    /**
     * Open picker, add files, and start upload in one call.
     * Convenience method for backward compatibility.
     * Waits for all selected files to be uploaded (or failed) before returning.
     */
    async openPickerAndUpload(): Promise<ImageUploadSuccess[]> {
        const files = await this.openPicker();
        if (files.length === 0) return [];

        const tasks = this.addFiles(files);
        await this.start();

        await waitForUploadCompletion(tasks);

        const success = tasks
            .filter(t => (t.state === UploadState.DONE || t.state === UploadState.DUPLICATE) && t.imageData)
            .map(t => ({
                uid: t.imageData!.uid,
                metadata: t.imageData
            }));
        return success;
    }
} 