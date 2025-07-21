import CollectionData from "$lib/entities/collection";
import UserData from "$lib/entities/user";
import { ImageObjectData } from "$lib/entities/image";
import { generateRandomString } from "$lib/utils/misc";
import { faker } from "@faker-js/faker";

/**
 * Creates a random test user for testing purposes.
 *
 * @returns A UserData object with random data.
 */
export function createTestUser() {
    return new UserData({
        id: generateRandomString(8),
        first_name: faker.person.firstName(),
        last_name: faker.person.lastName(),
        username: faker.internet.username(),
        email: faker.internet.email(),
        created_on: faker.date.past({ years: 2 }),
        role: "user",
        updated_on: faker.date.recent({ days: 60 })
    });
}

/**
 * Creates a random test image object for testing purposes.
 * 
 * @returns An ImageObjectData object with random data.
 */
export function createTestImageObject() {
    const randomImageNumber = Math.floor(Math.random() * 100);
    const name = `${faker.word.adjective()} ${faker.word.noun()}`;
    return new ImageObjectData({
        id: generateRandomString(16),
        name,
        collection_id: generateRandomString(16),
        uploaded_on: faker.date.past({ years: 2 }),
        updated_on: faker.date.recent({ days: 30 }),
        uploaded_by: createTestUser(),
        dupes: [],
        private: false,
        thumbhash: generateRandomString(16),
        urls: {
            original: `https://picsum.photos/1920/1080?random=${randomImageNumber}`,
            thumbnail: `https://picsum.photos/800/600?random=${randomImageNumber}`,
            preview: `https://picsum.photos/800/600?random=${randomImageNumber}`,
            raw: `https://picsum.photos/1920/1080?random=${randomImageNumber}`
        },
        image_data: {
            file_name: `${name.replace(/\s/g, "_")}.jpg`,
            file_size: Math.floor(Math.random() * 1000000) + 100000, // Random size between 100KB and 1MB
            original_file_name: `${faker.word.noun()}_original.jpg`,
            file_type: "jpg",
            keywords: faker.lorem.words(Math.floor(Math.random() * 15)).split(" "),
            width: Math.floor(Math.random() * 1920) + 800, // Random width between 800 and 1920
            height: Math.floor(Math.random() * 1080) + 600, // Random height between 600 and 1080
            colorSpace: "sRGB"
        }
    });
}

/**
 * Creates a new test collection with some random data.
 *
 * @returns A CollectionData object with random data.
 */
export function createTestCollection() {
    const testUser = createTestUser();

    return new CollectionData({
        id: generateRandomString(16),
        name: `${faker.word.adjective()} ${faker.word.noun()} Photos`
            .split(" ")
            .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
            .join(" "),
        description: faker.lorem.sentence(),
        created_on: faker.date.past({ years: 1 }),
        updated_on: faker.date.recent({ days: 30 }),
        images: [],
        private: faker.datatype.boolean(),
        created_by: testUser,
        image_count: Math.floor(Math.random() * 400) + 50,
        owner: testUser,
        thumbnail: createTestImageObject()
    });
}