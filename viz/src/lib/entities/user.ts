import type { UserRole } from "$lib/types/users";

class UserData {
    uid: string;
    first_name: string;
    last_name: string;
    username: string;
    email: string;
    role: UserRole;
    created_at: Date;
    updated_at: Date;

    constructor(data: Partial<UserData> & Pick<UserData, 'uid' | 'username' | 'email' | 'created_at' | 'updated_at'>) {
        this.uid = data.uid;
        this.first_name = data.first_name ?? '';
        this.last_name = data.last_name ?? '';
        this.username = data.username;
        this.email = data.email;
        this.role = data.role || 'user';
        this.created_at = data.created_at;
        this.updated_at = data.updated_at;

        for (const [key, value] of Object.entries(data)) {
            if (value === undefined || value === null || value === '') {
                console.warn(`UserData: Missing value for ${key}`);
            }
        }
    }
}

export default UserData;