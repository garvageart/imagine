import type { Collection } from "$lib/types/images";
import { generateRandomString, sleep } from "$lib/utils";
import type { PageLoad } from "./$types";
import { DateTime } from "luxon";
import { faker } from "@faker-js/faker";
import type { UserRole } from "$lib/types/users";
import CollectionData from "$lib/entities/collection";

export const load: PageLoad = ({ fetch }) => {
    // const collectionData: Collection[] = fetch(`${createServerURL(MEDIA_SERVER)}/collections/`, {
    //     method: "GET",
    //     headers: {
    //         "Content-Type": "application/json"
    //     }
    // });

    let allCollections: Collection[] = [];
    const randomCollectionCount = Math.floor(Math.random() * 30) + 10; // Random number between 5 and 25
    for (let i = 0; i < randomCollectionCount; i++) {
        const collectionId = generateRandomString(16);
        const user = {
            id: generateRandomString(8),
            first_name: faker.person.firstName(),
            last_name: faker.person.lastName(),
            username: faker.internet.username(),
            email: faker.internet.email(),
            created_on: faker.date.past({ years: 2 }),
            role: "user" as UserRole,
            updated_on: faker.date.recent({ days: 60 })
        };

        const testData: Collection = new CollectionData({
            id: collectionId,
            name: `${faker.word.adjective()} ${faker.word.noun()} Photos`.split(" ").map((word) => word.charAt(0).toUpperCase() + word.slice(1)).join(" "),
            description: faker.lorem.sentence(),
            created_on: DateTime.now().toJSDate(),
            updated_on: DateTime.now().toJSDate(),
            images: [],
            private: faker.datatype.boolean(),
            created_by: {
                id: generateRandomString(8),
                first_name: faker.person.firstName(),
                last_name: faker.person.lastName(),
                username: faker.internet.username(),
                email: faker.internet.email(),
                created_on: faker.date.past({ years: 2 }),
                role: "user",
                updated_on: faker.date.recent({ days: 60 })
            },
            image_count: Math.floor(Math.random() * 100),
            owner: user,
            thumbnail: {
                collection_id: collectionId,
                id: generateRandomString(16),
                name: `${faker.word.adjective()} ${faker.word.noun()}`,
                image_data: {
                    file_name: `${faker.word.noun()}.jpg`,
                    file_size: Math.floor(Math.random() * 1000000) + 100000, // Random size between 100KB and 1MB
                    original_file_name: `${faker.word.noun()}_original.jpg`,
                    file_type: "jpg",
                    keywords: faker.lorem.words(Math.floor(Math.random() * 15)).split(" "),
                    width: Math.floor(Math.random() * 1920) + 800, // Random width between 800 and 1920
                    height: Math.floor(Math.random() * 1080) + 600 // Random height between 600 and 1080
                },
                updated_on: faker.date.recent({ days: 30 }),
                uploaded_on: faker.date.past({ years: 1 }),
                uploaded_by: user
            }
        });

        allCollections.push(testData);
    };

    return {
        response: allCollections
    };
};
