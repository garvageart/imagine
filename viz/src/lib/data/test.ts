import CollectionData from "$lib/entities/collection";
import { UserData } from "$lib/types/users";
import { generateRandomString } from "$lib/utils";
import { faker } from "@faker-js/faker";
import { DateTime } from "luxon";

export const testUser = new UserData({
    id: generateRandomString(8),
    first_name: faker.person.firstName(),
    last_name: faker.person.lastName(),
    username: faker.internet.username(),
    email: faker.internet.email(),
    created_on: faker.date.past({ years: 2 }),
    role: "user",
    updated_on: faker.date.recent({ days: 60 })
});

export const testCollection = new CollectionData({
    id: generateRandomString(16),
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
    owner: testUser,
    thumbnail: {
        collection_id: generateRandomString(16),
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
        uploaded_by: testUser,
    }
});