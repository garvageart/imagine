import { openDB } from "idb";

export async function initDB() {
    return await openDB("imagine", 1, {
        upgrade(db) {
            db.createObjectStore("preferences", {
                keyPath: "id",
                autoIncrement: true
            });
            
        }
    });
} 