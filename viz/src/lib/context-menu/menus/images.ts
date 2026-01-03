import { invalidateAll } from "$app/navigation";
import { deleteCollectionImages, deleteImagesBulk, getFullImagePath, updateCollection, updateImage, type CollectionDetailResponse, type Image } from "$lib/api";
import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
import { copyToClipboard } from "$lib/utils/misc";
import type { MaterialSymbol } from "material-symbols";
import type { MenuItem } from "../types";
import { performImageDownloads } from "$lib/utils/http";
import type { SelectionScope } from "$lib/states/selection.svelte";
import { invalidateViz } from "$lib/views/views.svelte";

interface CollectionImageMenuOptions {
    downloadImages?: (images: Image[]) => void;
}

export function createCollectionImageMenu(asset: Image, collection: CollectionDetailResponse, opts?: CollectionImageMenuOptions) {
    let ctxItems: MenuItem[] = [
        {
            id: `download-${asset.uid}`,
            label: "Download",
            icon: "download",
            action: async () => {
                try {
                    opts?.downloadImages?.([asset]);
                } catch (err) {
                    console.error("Context menu download error", err);
                    toastState.addToast({
                        type: "error",
                        message: `Download failed: ${err}`
                    });
                }
            }
        },
        {
            id: "act-favourite",
            label: asset.favourited ? "Unfavourite" : "Favourite",
            icon: "favorite",
            action: async () => {
                const res = await updateImage(asset.uid, {
                    favourited: asset.favourited ? false : true
                });

                if (res.status === 200) {
                    toastState.addToast({
                        type: "success",
                        message: `Image ${asset.favourited ? "un" : ""}favourited`
                    });
                    asset = res.data;
                    await invalidateViz({ delay: 200 });
                } else {
                    toastState.addToast({
                        type: "error",
                        message: res.data.error ?? `Failed to ${asset.favourited ? "un" : ""}favourite`
                    });
                }
            }
        },
        {
            id: `collection-thumbnail-${asset.uid}`,
            label: "Make Collection Thumbnail",
            icon: "gallery_thumbnail",
            action: async () => {
                try {
                    const res = await updateCollection(collection.uid, {
                        thumbnailUID: asset.uid
                    });

                    if (res.status === 200) {
                        toastState.addToast({
                            type: "success",
                            message: `Collection thumbnail updated: **${res.data.thumbnail!.name}**`
                        });
                        await invalidateViz({ delay: 200 });
                    } else {
                        toastState.addToast({
                            type: "error",
                            message: res.data?.error ?? "Failed to update thumbnail"
                        });
                    }
                } catch (err) {
                    console.error("Context menu download error", err);
                    toastState.addToast({
                        type: "error",
                        message: `Download failed: ${err}`
                    });
                }
            }
        },
        {
            id: `remove-${asset.uid}`,
            label: "Remove from collection",
            icon: "remove_circle" as MaterialSymbol,
            action: async () => {
                if (
                    !confirm(
                        `Remove "${asset.name || asset.uid}" from collection "${collection.name}"?`
                    )
                ) {
                    return;
                }
                try {
                    const r = await deleteCollectionImages(collection.uid, {
                        uids: [asset.uid]
                    });
                    if (r.status === 200) {
                        toastState.addToast({
                            type: "success",
                            message: `Removed from collection`
                        });
                        await invalidateViz({ delay: 200 });
                    } else {
                        toastState.addToast({
                            type: "error",
                            message: r.data?.error ?? "Failed to remove"
                        });
                    }
                } catch (err) {
                    console.error("remove from collection error", err);
                    toastState.addToast({
                        type: "error",
                        message: `Failed to remove: ${err}`
                    });
                }
            }
        },
        {
            id: `copy-${asset.uid}`,
            label: "Copy link",
            icon: "link",
            action: async () => {
                try {
                    const url = getFullImagePath(asset.image_paths?.original) ?? "";
                    if (url) {
                        copyToClipboard(url);
                        toastState.addToast({
                            type: "success",
                            message: "Link copied to clipboard"
                        });
                    } else {
                        toastState.addToast({
                            type: "error",
                            message: "No URL available"
                        });
                    }
                } catch (err) {
                    console.error("copy link error", err);
                    toastState.addToast({
                        type: "error",
                        message: "Failed to copy link"
                    });
                }
            }
        },
        {
            id: `share-${asset.uid}`,
            label: "Share",
            icon: "share",
            action: () => {
                // Placeholder - open share dialog or implement later
                toastState.addToast({
                    type: "info",
                    message: "Share not implemented"
                });
            }
        }
    ];

    return ctxItems;
}

export function createImageMenu(images: Image[], selectionScope: SelectionScope<Image>) {
    let items = Array.from(selectionScope.selected);
    let firstItem = items[0];
    let actionMenuItems: MenuItem[] = [
        {
            id: "act-download",
            label: "Download",
            icon: "download",
            action: () => {
                try {
                    performImageDownloads(items);
                } catch (err) {
                    console.error("Download error", err);
                    toastState.addToast({
                        type: "error",
                        message: `Download failed: ${err}`,
                        timeout: 5000
                    });
                }
            }
        },
        {
            id: "act-share",
            label: "Share",
            icon: "share",
            action: () => {
                // TODO: Open share dialog
                toastState.addToast({
                    type: "info",
                    message: `Share ${items.length} image(s) - Not yet implemented`,
                    timeout: 3000
                });
            }
        },
        {
            id: "act-copy-link",
            label: "Copy Link",
            icon: "link",
            action: () => {
                if (items.length === 1) {
                    const url = getFullImagePath(firstItem.image_paths?.original);
                    copyToClipboard(url);
                    toastState.addToast({
                        type: "success",
                        message: "Link copied to clipboard",
                        timeout: 3000
                    });
                } else {
                    toastState.addToast({
                        type: "warning",
                        message: "Can only copy link for a single image",
                        timeout: 3000
                    });
                }
            }
        },
        {
            id: "act-favourite",
            label: items[0].favourited ? "Unfavourite" : "Favourite",
            icon: "favorite",
            action: async () => {
                const res = await updateImage(firstItem.uid, {
                    favourited: firstItem.favourited ? false : true
                });

                if (res.status === 200) {
                    toastState.addToast({
                        type: "success",
                        message: `Image ${firstItem.favourited ? "un" : ""}favourited`
                    });
                    firstItem = res.data;
                    await invalidateViz({ delay: 200 });
                } else {
                    toastState.addToast({
                        type: "error",
                        message: res.data.error ?? `Failed to ${firstItem.favourited ? "un" : ""}favourite`
                    });
                }
            }
        },
        {
            id: "act-edit-metadata",
            label: "Edit Metadata",
            icon: "edit",
            action: () => {
                // TODO: Open metadata editor
                toastState.addToast({
                    type: "info",
                    message: `Edit metadata for ${items.length} image(s) - Not yet implemented`,
                    timeout: 3000
                });
            }
        },
        {
            id: "act-move-to-trash",
            label: "Move to Trash",
            icon: "delete",
            action: async () => {
                const okTrash = confirm(
                    `Move ${items.length} selected image(s) to trash?`
                );

                if (!okTrash) {
                    return;
                }

                try {
                    const res = await deleteImagesBulk({
                        uids: items.map((i) => i.uid),
                        force: false
                    });

                    if (res.status === 200 || res.status === 207) {
                        const deletedUIDs = (res.data.results ?? [])
                            .filter((r) => r.deleted)
                            .map((r) => r.uid);
                        images = images.filter((img) => !deletedUIDs.includes(img.uid));
                        selectionScope.clear();
                    } else {
                        toastState.addToast({
                            type: "error",
                            message: res.data?.error ?? "Failed to delete images",
                            timeout: 4000
                        });
                    }
                } catch (err) {
                    toastState.addToast({
                        type: "error",
                        message: `Delete failed: ${err}`,
                        timeout: 5000
                    });
                }
            }
        },
        {
            id: "act-force-delete",
            label: "Force Delete",
            icon: "delete_forever",
            action: async () => {
                const okForce = confirm(
                    `Permanently delete ${items.length} image(s)? This action cannot be undone!`
                );

                if (!okForce) {
                    return;
                }

                try {
                    const res = await deleteImagesBulk({
                        uids: items.map((i) => i.uid),
                        force: true
                    });

                    if (res.status === 200 || res.status === 207) {
                        const deletedUIDs = (res.data.results ?? [])
                            .filter((r) => r.deleted)
                            .map((r) => r.uid);
                        images = images.filter((img) => !deletedUIDs.includes(img.uid));
                        selectionScope.clear();
                    } else {
                        toastState.addToast({
                            type: "error",
                            message: (res as any).data?.error ?? "Failed to delete images",
                            timeout: 4000
                        });
                    }
                } catch (err) {
                    toastState.addToast({
                        type: "error",
                        message: `Delete failed: ${err}`,
                        timeout: 5000
                    });
                }
            }
        }
    ];

    return actionMenuItems;
}