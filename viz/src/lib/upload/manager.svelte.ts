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
    async start(tasks?: UploadImage[]): Promise<ImageUploadSuccess[]> {
        const pending = tasks || upload.files.filter(t => t.state === UploadState.PENDING);

        if (pending.length === 0) {
            return [];
        }

        const results = await this.uploadWithConcurrency(pending);

        // Update stats based on task states
        const successfulCount = pending.filter(t => t.state === UploadState.DONE).length;
        const duplicateCount = pending.filter(t => t.state === UploadState.DUPLICATE).length;
        const errorCount = pending.filter(t => t.state === UploadState.ERROR || t.state === UploadState.INVALID).length;

        upload.stats.success += successfulCount;
        upload.stats.duplicates += duplicateCount;
        upload.stats.errors += errorCount;

        const successful = results.filter(r => r !== undefined) as ImageUploadSuccess[];

        // Remove only fully completed/terminal tasks from panel.
        // Keep DUPLICATE tasks visible so the UI can show duplicate state to the user.
        for (const task of pending) {
            const idx = upload.files.indexOf(task);
            if (idx === -1) {
                continue;
            }

            // Remove if DONE, CANCELED or ERROR. Keep DUPLICATE and INVALID for visibility.
            if (task.state === UploadState.DONE || task.state === UploadState.CANCELED || task.state === UploadState.ERROR) {
                upload.files.splice(idx, 1);
            }
        }

        return successful;
    }

    /**
     * Upload tasks with concurrency limit using a proper queue.
     */
    private async uploadWithConcurrency(
        tasks: UploadImage[]
    ): Promise<(ImageUploadSuccess | undefined)[]> {
        const results: (ImageUploadSuccess | undefined)[] = new Array(tasks.length);
        let activeCount = 0;
        let nextIndex = 0;

        return new Promise((resolve) => {
            const startNext = () => {
                // If all tasks started and none active, we're done
                if (nextIndex >= tasks.length && activeCount === 0) {
                    resolve(results);
                    return;
                }

                // Start new tasks while under concurrency limit
                while (activeCount < upload.concurrency && nextIndex < tasks.length) {
                    const index = nextIndex++;
                    const task = tasks[index];

                    activeCount++;

                    task.upload()
                        .then((result) => {
                            results[index] = result;
                        })
                        .catch((err) => {
                            console.error(`Upload failed for ${task.data.file_name}:`, err);
                            results[index] = undefined;
                        })
                        .finally(() => {
                            activeCount--;
                            startNext();
                        });
                }
            };

            startNext();
        });
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
    async openPickerAndUpload(): Promise<ImageUploadSuccess[]> {
        const files = await this.openPicker();
        if (files.length === 0) return [];

        const tasks = this.addFiles(files);
        return await this.start(tasks);
    }
}