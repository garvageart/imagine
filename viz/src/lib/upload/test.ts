import { upload } from "$lib/states/index.svelte";
import { onMount } from "svelte";
import { UploadImage, UploadState } from "./asset.svelte";

export function setupDummyData() {
    onMount(() => {
        upload.files = [
            new UploadImage({
                filename: "mountain_sunset_4k.jpg",
                checksum: "abc123def456",
                data: new File([], "mountain_sunset_4k.jpg")
            }),
            new UploadImage({
                filename: "beach_vacation_2024.png",
                checksum: "ghi789jkl012",
                data: new File([], "beach_vacation_2024.png")
            }),
            new UploadImage({
                filename: "family_portrait_hires.jpg",
                checksum: "mno345pqr678",
                data: new File([], "family_portrait_hires.jpg")
            }),
            new UploadImage({
                filename: "city_lights_night_photography.jpg",
                checksum: "stu901vwx234",
                data: new File([], "city_lights_night_photography.jpg")
            })
        ];

        upload.files[0].progress = 75;
        upload.files[0].state = UploadState.STARTED;

        upload.files[1].progress = 100;
        upload.files[1].state = UploadState.DONE;

        upload.files[2].progress = 45;
        upload.files[2].state = UploadState.STARTED;

        upload.files[3].progress = 10;
        upload.files[3].state = UploadState.STARTED;
    });
}