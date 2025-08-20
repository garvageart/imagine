import { createApi } from 'unsplash-js';
import dotenv from "dotenv";
import path from "path";

async function main() {
    const dotEnvPath = path.resolve(__dirname, "..", "..", ".env");

    dotenv.config({
        path: dotEnvPath
    });

    const accessKey = process.env.UNSPLASH_ACCESS_KEY;
    if (!accessKey) {
        console.error("UNSPLASH_ACCESS_KEY is not defined in the environment variables.");
        process.exit(1);
    }

    const unsplash = createApi({
        accessKey: accessKey
    });

    console.log("Fetching random images from Unsplash...");
    const result = await unsplash.photos.getRandom({ count: 30 });

    if (result.type === 'error') {
        throw new Error(`Failed to fetch random images from Unsplash: ${result.errors.join(', ')}`);
    }

    const randomImgs = Array.isArray(result.response) ? result.response : [result.response];
    const randomURLs = randomImgs.map(img => {
        return img.urls.full;
    });

    console.log("Ingesting random images from Unsplash...");
    const ingestResponse = await fetch("http://localhost:7770/images/urls", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer tes`
        },
        body: JSON.stringify({ urls: randomURLs })
    });

    if (!ingestResponse.ok || ingestResponse.status !== 201) {
        const errorText = await ingestResponse.text();
        throw new Error(`Failed to ingest images. Status: ${ingestResponse.status}. Body: ${errorText}`);
    }

    console.log("All images have been ingested");
    console.log(await ingestResponse.json());

    // Track a photo download
    // https://help.unsplash.com/api-guidelines/guideline-triggering-a-download
    console.log("Tracking downloads...");
    const downloadPromises = randomImgs.map(img => {
        return unsplash.photos.trackDownload({
            downloadLocation: img.links.download_location,
        }).then(trackResult => {
            if (trackResult.type === 'error') {
                console.error(`Failed to track download for ${img.id}:`, trackResult.errors.join(', '));
            } else {
                console.log(`Download tracked for ${img.id}`);
            }
        });
    });

    await Promise.all(downloadPromises);

    console.log("Done");
}

main().catch(error => {
    console.error("An error occurred during the script execution:");
    if (error instanceof Error) {
        console.error(error.message);
    } else {
        console.error(error);
    }
    process.exit(1);
});

