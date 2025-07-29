import { createApi } from 'unsplash-js';
import { Random } from "unsplash-js/dist/methods/photos/types";
import dotenv from "dotenv";
import path from "path";

const dotEnvPath = path.resolve("../../") + "\\.env";

dotenv.config({
    path: dotEnvPath
});


const unsplash = createApi({
    accessKey: process.env.UNSPLASH_ACCESS_KEY!
});

try {
    console.log("Fetching random images from Unsplash...");
    const result = await fetch(`https://api.unsplash.com/photos/random?count=30&client_id=${process.env.UNSPLASH_ACCESS_KEY}`, {
        headers: {
            "Accept-Version": "v1"
        }
    });

    if (!result.ok) {
        console.log(result.status, result.statusText);
        throw new Error("Failed to fetch random images from Unsplash");
    }

    const resultJson = await result.json();

    const randomImgs = resultJson as Random[];
    const randomURLs = randomImgs.map(img => {
        return img.urls.full;
    });

    console.log("Ingesting random images from Unsplash...");
    fetch("http://localhost:7770/images/urls", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer tes`
        },
        body: JSON.stringify({ urls: randomURLs })
    }).then(async res => {
        if (res.status !== 201) {
            throw new Error("Failed to ingest images from Unsplash");
        }

        console.log("All images have been ingested");
        console.log(await res.json());
        // Track a photo download
        // https://help.unsplash.com/api-guidelines/guideline-triggering-a-download
        for (const img of randomImgs) {
            await new Promise(resolve => setTimeout(resolve, 1000));
            fetch(`https://api.unsplash.com/photos/${img.id}/download`).then(async (res) => {
                res.json().then((res) => {
                    console.log(res);
                    console.log("Download tracked for", img.id, "\n");
                });

            });
        }

        console.log("Done");
    }).catch(err => {
        console.error(err);
    });
} catch (error) {
    console.error("Failed to fetch random images from Unsplash");
    // @ts-ignore
    console.error(error?.message);
}