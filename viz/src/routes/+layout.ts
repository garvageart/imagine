import { login, user } from "$lib/states/index.svelte.js";
import { redirect } from '@sveltejs/kit';
import { initApi } from "$lib/api/client";
import { fetchCurrentUser } from "$lib/auth/auth_methods";

export const ssr = false;
export const csr = true;

export async function load({ url, fetch }) {
    // Initialize API client with SvelteKit's enhanced fetch
    initApi(fetch);

    // Prefer API-backed auth over cookie when available
    if (!user.fetched && !url.pathname.startsWith('/auth')) {
        await fetchCurrentUser();
    }

    const isAuthed = !!user.data || !!login.state;

    if (!isAuthed && !url.pathname.startsWith("/auth")) {
        redirect(303, `/auth/register?continue=${url.pathname}`);
    }

    const queryParams = new URLSearchParams(url.search);
    const redirectURL = queryParams.get("continue")?.trim();

    if (isAuthed && redirectURL) {
        redirect(303, redirectURL);
    }

    if (isAuthed && url.pathname.startsWith("/auth")) {
        redirect(303, "/");
    }
}