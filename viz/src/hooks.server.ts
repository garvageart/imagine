import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
    event.locals.user = (await event.fetch("https://localhost:7777/user")).json();

    const response = await resolve(event);

    return response;
};