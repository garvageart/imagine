import { user, continuePath, system } from "$lib/states/index.svelte.js";
import { redirect, error } from '@sveltejs/kit';
import { initApi, getSystemStatus } from "$lib/api";
import { fetchCurrentUser } from "$lib/auth/auth_methods";
import { browser } from '$app/environment';

export const ssr = false;
export const csr = true;

export async function load({ url, fetch }) {
    // Initialize API client with SvelteKit's enhanced fetch
    initApi(fetch);

    // Fetch system status if not already fetched
    if (!system.fetched) {
        system.loading = true;

        if (browser && window.__VIZ_CONFIG__) {
            system.data = window.__VIZ_CONFIG__.system;
            system.fetched = true;
            // Optional: clear it so we don't rely on stale data later if we ever re-fetch?
            window.__VIZ_CONFIG__ = undefined;
        } else {
            const res = await getSystemStatus();
            if (res.status === 200) {
                system.data = res.data;
                system.fetched = true;
            } else {
                system.error = "Failed to load system status";
            }
        }
        system.loading = false;
    }

    const needsSuperadmin = system.data?.needs_superadmin;
    const userOnboardingRequired = system.data?.user_onboarding_required;
    const isOnboardingPage = url.pathname.startsWith('/onboarding');
    const isAuthPage = url.pathname.startsWith('/auth');

    // 1. Superadmin setup takes priority over everything
    if (needsSuperadmin) {
        if (!isOnboardingPage) {
            redirect(303, '/onboarding');
        }
        return;
    }

    // 2. Ensure we have the user state
    if (!user.fetched) {
        await fetchCurrentUser();
    }

    const isConnectionError = !!user.connectionError;

    if (isConnectionError) {
        error(503, {
            message: "Could not connect to the Viz server. It might be down for maintenance or restarting."
        });
    }

    const isAuthed = !!user.data;

    // 3. Handle Auth Pages (Login/Register)
    if (isAuthPage) {
        if (isAuthed) {
            // Already logged in? Go home or to continue path
            const queryParams = new URLSearchParams(url.search);
            const continueQuery = queryParams.get("continue");
            redirect(303, continueQuery ? decodeURIComponent(continueQuery) : '/');
        }
        // Not logged in, stay on auth page
        return;
    }

    // 4. Handle Protected Routes (everything else)
    if (!isAuthed) {
        const continueUrl = url.pathname === '/' ? '' : `?continue=${encodeURIComponent(url.pathname)}`;
        redirect(303, `/auth/login${continueUrl}`);
    }

    // 5. User Onboarding (only if authenticated)
    if (userOnboardingRequired) {
        if (!isOnboardingPage) {
            redirect(303, '/onboarding');
        }
        return;
    }

    // 6. Prevent access to /onboarding if not needed
    if (isOnboardingPage && !userOnboardingRequired) {
        redirect(303, '/');
    }

    // 7. Handle "continue" param if present on a non-auth page (rare, but possible if redirected manually)
    const queryParams = new URLSearchParams(url.search);
    const continueQuery = queryParams.get("continue");

    if (continueQuery) {
        const decoded = decodeURIComponent(continueQuery).trim();
        // Avoid infinite redirect loop
        if (decoded && decoded !== url.pathname) {
            redirect(303, decoded);
        }
    }
}