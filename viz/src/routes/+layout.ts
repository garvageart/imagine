import { login, user, continuePath } from "$lib/states/index.svelte.js";
import { redirect, error } from '@sveltejs/kit';
import { initApi } from "$lib/api";
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

    const isConnectionError = !!user.connectionError;

    if (isConnectionError) {
        error(503, {
            message: "Could not connect to the Imagine server. It might be down for maintenance or restarting."
        });
    }

    const isAuthed = !!user.data || !!login.state;
    if (!isAuthed && !url.pathname.startsWith("/auth")) {
        redirect(303, `/auth/register?continue=${decodeURIComponent(url.pathname)}`);
    }

    const queryParams = new URLSearchParams(url.search);
    let continueQuery = queryParams.get("continue");
    let continueState = continuePath;
    continueQuery = decodeURIComponent(continueQuery || "").trim() || null;

    if (continueQuery) {
        continueState = continueQuery;
    }

    if (continueState && url.pathname === continueState) {
        continueState = null;
    }

    const redirectURL = continueState ?? continueQuery;
    if (isAuthed && redirectURL) {
        continueState = null;
        redirect(303, redirectURL);
    }

    if (isAuthed && url.pathname.startsWith("/auth")) {
        continueState = null;
        redirect(303, "/");
    }
}