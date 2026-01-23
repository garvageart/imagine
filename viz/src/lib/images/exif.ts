import { updateImage, type ImageAsset, type ImageUpdate } from "$lib/api";

export async function setRating(image: ImageAsset, prevRating: number | null, newRating: number | null) {
    if (prevRating === newRating) {
        return newRating;
    }

    const uid = image.uid;
    const dateToUpdate: ImageUpdate = {
        image_metadata: { rating: newRating }
    };

    const res = await updateImage(uid, dateToUpdate);
    if (res.status === 200) {
        image = res.data;
        return image.image_metadata?.rating ?? null;
    } else {
        throw new Error(`Failed to update rating: ${res.data.error}`);
    }
}

