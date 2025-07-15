export type UserRole = "user" | "admin" | "superadmin" | "guest";

export interface User {
    id: string;
    first_name: string;
    last_name: string;
    username: string;
    email: string;
    created_on: Date;
    updated_on: Date;
    role: UserRole;
}