/**
 * Recursively traverse file system entries to collect all files,
 * including those in nested folders.
 */
export async function traverseFileTree(item: FileSystemEntry): Promise<File[]> {
    const files: File[] = [];

    if (item.isFile) {
        const fileEntry = item as FileSystemFileEntry;
        const file = await new Promise<File>((resolve, reject) => {
            fileEntry.file(resolve, reject);
        });
        files.push(file);
    } else if (item.isDirectory) {
        const dirEntry = item as FileSystemDirectoryEntry;
        const reader = dirEntry.createReader();

        const entries = await new Promise<FileSystemEntry[]>(
            (resolve, reject) => {
                reader.readEntries(resolve, reject);
            }
        );

        for (const entry of entries) {
            const nestedFiles = await traverseFileTree(entry);
            files.push(...nestedFiles);
        }
    }

    return files;
}