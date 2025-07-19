/**
 * Various methods for storing, retrieving and deleting cookies from the browser
 */
export const cookieMethods = {
    set: (key: string, value: string, expiresDate?: Date | string) => {
        document.cookie = `${key}=${value}; expires=${expiresDate}; Secure; path =/`;
    },
    get: (key: string): string | undefined => {
        const allCookies = document?.cookie;
        const cookieValue = allCookies.split("; ").find(cookie => cookie.startsWith(`${key}`))?.split("=")[1];

        return cookieValue;
    },
    delete: (key: string) => {
        document.cookie = `${key}=; max-age=0; path =/`;
    }
};