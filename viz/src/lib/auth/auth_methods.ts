import { goto } from "$app/navigation";
import { sleep } from "$lib/utils/misc";
import { cookieMethods } from "$lib/utils/cookie";
import { defaults, getCurrentUser, logout, type User } from "$lib/api";
import { user } from "$lib/states/index.svelte";

interface OAuthResponseUserData {
    id?: string;
    email: string;
    name: string;
    picture: string;
}

export async function fetchCurrentUser(): Promise<User | null> {
    user.loading = true;
    try {
        const result = await getCurrentUser();

        if (result.status === 200) {
            user.data = result.data;
            user.error = null;
            user.isAdmin = result.data.role.includes('admin');
            return result.data;
        } else {
            user.data = null;
            user.error = null;
            return null;
        }
    } catch (err: any) {
        user.data = null;
        user.error = err?.message ?? 'Failed to fetch current user';
        user.connectionError = true;
        return null;
    } finally {
        user.loading = false;
        user.fetched = true;
    }
}

export function clearUser() {
    user.data = null;
    user.error = null;
    user.fetched = true;
}

export function logoutUser() {
    clearUser();
    logout();
    cookieMethods.delete("imag-state");
    goto('/login');
}


export const authServerURL = defaults.baseUrl;

export async function sendOAuthParams(provider: string | null): Promise<boolean> {
    const queryParams = Object.fromEntries(new URLSearchParams(location.search).entries());

    if (!queryParams.code) {
        return false;
    }

    if (!provider) {
        await sleep(3000);
        goto("/");
        return false;
    }

    const fetchURL = new URL(`${authServerURL}/auth/oauth/${provider}`);
    for (const [key, value] of Object.entries(queryParams)) {
        fetchURL.searchParams.set(key, value);
    }

    try {
        const response = await fetch(fetchURL, {
            method: "POST",
            mode: "cors",
            credentials: "include"
        });

        if (response.status !== 200) {
            console.error(response.statusText);
            goto("/?error=40");
            return false;
        }

        const authData: OAuthResponseUserData = await response.json();

        if (authData.email) {
            goto("/auth/register", {
                state: {
                    email: authData.email,
                    name: authData.name,
                    picture: authData.picture,
                    provider: provider
                }
            });
            return true;
        } else {
            cookieMethods.delete("imag-state");
            return false;
        }
    } catch (err) {
        console.error("OAuth error:", err);
        goto("/?error=40");
        return false;
    }
}
