<script lang="ts">
	import {
		Workspace,
		type SerializedWorkspace
	} from "$lib/layouts/model.svelte";
	import LayoutNode from "./LayoutNode.svelte";
	import { VizLocalStorage } from "$lib/utils/misc";
	import { onMount } from "svelte";
	import { debugMode } from "$lib/states/index.svelte";
	import { workspaceState } from "$lib/states/workspace.svelte";
	import { createDefaultLayout } from "./layouts/registry";
	import { dev } from "$app/environment";
	import { views } from "$lib/layouts/views";
	import hotkeys from "hotkeys-js";

	interface Props {
		id: string;
	}

	let { id }: Props = $props();

	const storage = new VizLocalStorage<SerializedWorkspace>("workspaceLayout");
	let initialized = $state(false);

	onMount(() => {
		const stored = storage.get();
		workspaceState.workspace = new Workspace(undefined, views);
		if (stored) {
			try {
				workspaceState.workspace.load(stored);
				if (debugMode) {
					console.log("[Workspace] Hydrated from storage");
				}
			} catch (e) {
				console.error("[Workspace] Failed to hydrate layout", e);
				workspaceState.workspace = createDefaultLayout();
			}
		} else {
			workspaceState.workspace = createDefaultLayout();
		}

		initialized = true;
	});

	if (dev) {
		$effect(() => {
			console.log(
				"[Workspace] Workspace state:",
				workspaceState.workspace?.toJSON()
			);
		});
	}

	$effect(() => {
		if (initialized && workspaceState.workspace) {
			const serialized = workspaceState.workspace.toJSON();
			storage.set(serialized);
			if (debugMode) {
				console.log("[Workspace] Layout saved");
			}
		}
	});

	$effect(() => {
		if (!initialized) return;

		hotkeys("`", (event, handler) => {
			// Prevent default behavior (e.g. typing ` in an input, if filter logic fails)
			event.preventDefault();
			const ws = workspaceState.workspace;
			if (ws && ws.activeGroupId) {
				ws.toggleMaximize(ws.activeGroupId);
			}
		});

		return () => {
			hotkeys.unbind("`");
		};
	});
</script>

<div {id} class="viz-workspace">
	{#if initialized && workspaceState.workspace}
		<LayoutNode node={workspaceState.workspace.root} />
	{:else}
		<div class="loading">Initializing Workspace...</div>
	{/if}
</div>

<style lang="scss">
	.viz-workspace {
		height: 100%;
		width: 100%;
		overflow: hidden;
		background-color: var(--imag-100);
	}

	.loading {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--imag-60);
	}
</style>
