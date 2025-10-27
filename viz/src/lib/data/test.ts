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
        uid: generateRandomString(8),
        first_name: faker.person.firstName(),
        last_name: faker.person.lastName(),
        username: faker.internet.username(),
        email: faker.internet.email(),
        role: "user",
        created_at: faker.date.past({ years: 2 }),
        updated_at: faker.date.recent({ days: 60 })
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
    const testUser = createTestUser();

    return new ImageObjectData({
        uid: generateRandomString(16),
        name,
        collection_id: generateRandomString(16),
        uploaded_by: testUser.uid,
        private: false,
        width: Math.floor(Math.random() * 1920) + 800,
        height: Math.floor(Math.random() * 1080) + 600,
        processed: true,
        thumbhash: generateRandomString(16),
        created_at: faker.date.past({ years: 2 }),
        updated_at: faker.date.recent({ days: 30 }),
        added_at: faker.date.past({ years: 2 }),
        added_by: testUser.uid,
        image_metadata: {
            file_name: `${name.replace(/\s/g, "_")}.jpg`,
            file_size: Math.floor(Math.random() * 1000000) + 100000,
            original_file_name: `${faker.word.noun()}_original.jpg`,
            file_type: "jpg",
            keywords: faker.lorem.words(Math.floor(Math.random() * 15)).split(" "),
            color_space: "sRGB",
            file_modified_at: faker.date.recent({ days: 30 }).toISOString(),
            file_created_at: faker.date.past({ years: 2 }).toISOString(),
            thumbhash: generateRandomString(16),
            checksum: generateRandomString(32)
        },
        image_paths: {
            original_path: `https://picsum.photos/1920/1080?random=${randomImageNumber}`,
            thumbnail_path: `https://picsum.photos/800/600?random=${randomImageNumber}`,
            preview_path: `https://picsum.photos/800/600?random=${randomImageNumber}`,
            raw_path: `https://picsum.photos/1920/1080?random=${randomImageNumber}`
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
        uid: generateRandomString(16),
        name: `${faker.word.adjective()} ${faker.word.noun()} Photos`
            .split(" ")
            .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
            .join(" "),
        description: faker.lorem.sentence(),
        images: [],
        private: faker.datatype.boolean(),
        created_by: testUser.uid,
        created_at: faker.date.past({ years: 1 }),
        updated_at: faker.date.recent({ days: 30 }),
        image_count: Math.floor(Math.random() * 400) + 50,
        thumbnail: createTestImageObject()
    });
}