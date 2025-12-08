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
 * dynamic queue processor that respects global concurrency.
 * Can be called repeatedly to fill available slots.
 */
export function processGlobalQueue() {
    // Find pending tasks
    const pendingTasks = upload.files.filter(t => t.state === UploadState.PENDING);

    if (pendingTasks.length === 0 && activeCount === 0) {
        return;
    }

    // Fill available slots
    while (activeCount < upload.concurrency && pendingTasks.length > 0) {
        const task = pendingTasks.shift();
        if (!task) break;

        activeCount++;

        // We don't await here so the loop continues to fill slots
        task.upload()
            .then((result) => {
                if (result) {
                    upload.stats.success++;
                }
            })
            .catch((err) => {
                console.error(`Upload failed for ${task.data.file_name}:`, err);
                upload.stats.errors++;
            })
            .finally(() => {
                activeCount--;

                // Clean up completed tasks from the list if needed
                if (task.state === UploadState.DONE || task.state === UploadState.CANCELED || task.state === UploadState.ERROR) {
                    const idx = upload.files.indexOf(task);
                    if (idx !== -1) {
                        upload.files.splice(idx, 1);
                    }
                } else if (task.state === UploadState.DUPLICATE) {
                    upload.stats.duplicates++;
                    // We keep duplicates visible
                }

                // Trigger next processing cycle
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
            if (!this.allowedTypes.includes(fileType as any)) {
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
        // We no longer return a promise of all results because the queue is dynamic.
        // The UI updates reactively based on task state.
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
     */
    async openPickerAndUpload(): Promise<void> {
        const files = await this.openPicker();
        if (files.length === 0) return;

        this.addFiles(files);
        await this.start();
    }
}