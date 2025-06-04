import { goto } from "$app/navigation";
import { AUTH_SERVER } from "$lib/constants";
import { cookieMethods, createServerURL, getURLParams, sleep } from "$lib/utils";

interface AuthorizationCodeFlowResponse {
    code: string;
    state: string;
}

interface AuthorizationCodeGrantResponse {
    access_token: string;
    expires_in: number;
    refresh_token: string;
    scope: string;
    token_type: string;
}

interface OAuthResponseUserData {
    id: string;
    email: string;
    verified_email: boolean;
    name: string;
    picture: string;
    hd: string;
}

export const authServerURL = createServerURL(AUTH_SERVER);

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

    const fetchURL = new URL(`${authServerURL}/oauth/${provider}`);
    for (const [key, value] of Object.entries(queryParams)) {
        fetchURL.searchParams.set(key, value);
    }

    const authData: OAuthResponseUserData = await fetch(fetchURL, {
        method: "POST",
        mode: "cors",
        credentials: "include"
    }).then(async (res) => {
        return await res.json();
    }).catch((err) => {
        console.error(err);
        return null;
    });


    if (authData.email) {
        goto("/signup");
        return true;
    } else {
        cookieMethods.delete("imag-state");
        return false;
    }
}
