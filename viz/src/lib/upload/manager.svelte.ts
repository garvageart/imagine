import { upload } from "$lib/states/index.svelte";
import type { SupportedImageTypes, SupportedRAWFiles } from "$lib/types/images";
import { UploadImage } from "./asset.svelte";

export interface ImageUploadFileData {
    filename: string;
    data: File;
    checksum?: string;
}

export interface ImageUploadSuccess {
    uid: string;
    metadata?: any;
}

export default class UploadManager {
    allowedTypes: string[];
    fileHolder: HTMLInputElement;

    constructor(allowedTypes: SupportedImageTypes[] | SupportedRAWFiles[]) {
        this.allowedTypes = allowedTypes;
        this.fileHolder = document.createElement("input");

        this.createFileHolder();
    }

    private async readFile(fileList: FileList) {
        const allFiles = [...fileList];

        const allFileData = allFiles.map(async (file) => {
            const rawData = await new Promise<string | ArrayBuffer>((resolve) => {
                const reader = new FileReader();

                reader.onloadend = (e) => {
                    if (e.target && e.target.result) {
                        resolve(e.target.result);
                    }
                };

                reader.readAsDataURL(file);
            });

            return {
                file,
                rawData
            };
        });

        return Promise.all(allFileData);
    }

    private createFileHolder() {
        const allowedMimeTypesString = this.allowedTypes.map((mimeType) => "image/" + mimeType).join(", ");

        this.fileHolder.setAttribute("type", "file");
        this.fileHolder.setAttribute("accept", allowedMimeTypesString);
        this.fileHolder.multiple = true;
    }

    /**
     * Upload files with controlled concurrency
     * @param files Array of UploadImage instances to upload
     * @param concurrency Maximum number of simultaneous uploads
     * @returns Promise resolving to array of upload results
     */
    private async uploadWithConcurrency(files: UploadImage[], concurrency: number): Promise<(ImageUploadSuccess | undefined)[]> {
        const results: (ImageUploadSuccess | undefined)[] = [];
        const executing: Promise<void>[] = [];

        for (let i = 0; i < files.length; i++) {
            const file = files[i];

            // Create a promise for this upload
            const uploadPromise = (async () => {
                const result = await file.upload();
                results[i] = result;
            })();

            executing.push(uploadPromise);

            // When concurrency limit is reached, wait for one to finish
            if (executing.length >= concurrency) {
                await Promise.race(executing);
                // Remove completed promises
                executing.splice(0, executing.findIndex(p => p === uploadPromise) + 1);
            }
        }

        // Wait for remaining uploads to complete
        await Promise.all(executing);

        return results;
    }

    openFileHolder() {
        this.fileHolder.click();
    }

    async uploadImage() {
        return await new Promise<ImageUploadSuccess[]>((resolve, reject) => {
            this.fileHolder.addEventListener("change", async () => {
                try {
                    if (!this.fileHolder.files) {
                        resolve([]);
                        return;
                    }

                    const allFileData = await this.readFile(this.fileHolder.files);
                    this.fileHolder.remove();

                    const uploadFiles = [];

                    for (const fileData of allFileData) {
                        if (!this.allowedTypes.includes(fileData.file.type.split("/")[1])) {
                            resolve([]);
                            return;
                        }

                        const fileInformation: ImageUploadFileData = {
                            filename: fileData.file.name,
                            data: fileData.file
                        };

                        // Checking for duplicates should be optional maybe but just do it now anyways
                        if (crypto?.subtle?.digest) {
                            const hashBuffer = await crypto.subtle.digest("SHA-1", await fileData.file.arrayBuffer());
                            const hashArray = Array.from(new Uint8Array(hashBuffer));
                            fileInformation.checksum = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
                        }

                        uploadFiles.push(new UploadImage(fileInformation));
                        upload.stats.total += 1;
                    }

                    upload.files.push(...uploadFiles);

                    // Upload files with concurrency control
                    const results = await this.uploadWithConcurrency(uploadFiles, upload.concurrency);

                    const successfulUploads = results.filter((r) => r !== undefined);
                    upload.stats.success = successfulUploads.length;
                    upload.stats.errors = upload.stats.total - successfulUploads.length;

                    for (const file of uploadFiles) {
                        const index = upload.files.indexOf(file);
                        if (index > -1) {
                            upload.files.splice(index, 1);
                        }
                    }

                    resolve(successfulUploads);
                } catch (error) {
                    reject(error);
                }
            });
        });
    }

}