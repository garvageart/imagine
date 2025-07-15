import type { User, UserRole } from "$lib/types/users";

class UserData implements User {
    id: string;
    first_name: string;
    last_name: string;
    username: string;
    email: string;
    created_on: Date;
    updated_on: Date;
    role: UserRole;

    constructor(data: Partial<UserData>) {
        this.id = data.id || '';
        this.first_name = data.first_name || '';
        this.last_name = data.last_name || '';
        this.username = data.username || '';
        this.email = data.email || '';
        this.created_on = data.created_on || new Date();
        this.updated_on = data.updated_on || new Date();
        this.role = data.role || 'user';

        for (const [key, value] of Object.entries(data)) {
            if (value === undefined || value === null || value === '') {
                console.warn(`UserData: Missing value for ${key}`);
            }
        }
    }
}

export default UserData;