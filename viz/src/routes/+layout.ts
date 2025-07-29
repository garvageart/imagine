export const ssr = false;
export const csr = true;

import { login } from "$lib/states/index.svelte.js";
import { redirect } from '@sveltejs/kit';

export function load({ url }) {
    if (!login.state && !url.pathname.startsWith("/auth")) {
        redirect(303, `/auth/register?continue=${url.pathname}`);
    }

    const queryParams = new URLSearchParams(url.search);
    const redirectURL = queryParams.get("continue")?.trim();

    if (login.state && redirectURL) {
        redirect(303, redirectURL);
    }

    if (login.state && url.pathname.startsWith("/auth")) {
        redirect(303, "/");
    }
}