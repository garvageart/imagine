import { Workspace } from "$lib/layouts/model.svelte";

export const workspaceState = $state({
	workspace: null as Workspace | null
});
