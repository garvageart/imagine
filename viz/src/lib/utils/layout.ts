import { generateRandomString } from "./misc";

// TODO: Move utility functions to purpose made files
// instead of shoving them down all one utility file
export function generateKeyId(length = 10): string {
    return "sp-" + generateRandomString(length);
}
