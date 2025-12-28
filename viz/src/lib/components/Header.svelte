<script lang="ts">
	import { dev } from "$app/environment";
	import { page } from "$app/state";
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
	import { performSearch } from "$lib/search/execute";
	import {
		debugState,
		getTheme,
		search,
		toggleTheme,
		user,
		themeState,
		isLayoutPage
	} from "$lib/states/index.svelte";
	import { historyState } from "$lib/states/history.svelte";
	import {
		SUPPORTED_IMAGE_TYPES,
		SUPPORTED_RAW_FILES,
		type SupportedImageTypes
	} from "$lib/types/images";
	import UploadManager from "$lib/upload/manager.svelte";
	import hotkeys from "hotkeys-js";
	import type { SvelteHTMLElements } from "svelte/elements";
	import OpenAccountPanel from "./AccountPanel.svelte";
	import AppMenu from "./AppMenu.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";
	import SearchInput from "./SearchInput.svelte";
	import IconButton from "./IconButton.svelte";
	import { goto } from "$app/navigation";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";
	import ContextMenu from "$lib/context-menu/ContextMenu.svelte";
	import type { MenuItem } from "$lib/context-menu/types";
	import { views } from "$lib/layouts/views";
	import { findSubPanel, getAllSubPanels } from "$lib/utils/layout";
	import {
		layoutState,
		layoutTree
	} from "$lib/third-party/svelte-splitpanes/state.svelte";
	import { duplicateView } from "$lib/layouts/panel-operations";

	let { ...props }: SvelteHTMLElements["header"] = $props();

	let searchElement = $state<HTMLInputElement | undefined>();
	let searchInputHasFocus = $derived(
		searchElement && document.activeElement === searchElement
	);

	$effect(() => {
		if (page.url.pathname === "/search") {
			const q = page.url.searchParams.get("q");
			// Only update search.value if 'q' is present and different from current search.value
			// This prevents unnecessary updates and potential infinite loops if search.value also affects the URL
			if (q && q !== search.value) {
				search.value = q;
			}
		}
	});

	// Ctrl/Cmd+I toggles dev mode
	hotkeys("ctrl+i, command+i", (e) => {
		e.preventDefault();
		debugState.toggle();
	});

	// Ctrl/Cmd+K toggles focus on the search input.
	hotkeys("ctrl+k, command+k", (e) => {
		e.preventDefault();
		if (!searchInputHasFocus) {
			searchElement?.focus();
		} else {
			searchElement?.blur();
		}
	});

	async function handleUpload(e: MouseEvent) {
		e.preventDefault();
		// allowed image types will come from the config but for now just hardcode
		const manager = new UploadManager([
			...SUPPORTED_RAW_FILES,
			...SUPPORTED_IMAGE_TYPES
		] as SupportedImageTypes[]);

		const uploadedImages = await manager.openPickerAndUpload();

		if (uploadedImages.length === 0) {
			return;
		}

		if (page.url.pathname !== "/") {
			toastState.addToast({
				title: "Upload Success",
				type: "success",
				message: `${uploadedImages.length} image(s) sucessfully uploaded`,
				actions: [
					{
						label: "Go to Photos",
						onClick: () => goto("/")
					}
				]
			});
		}
	}

	let openAccPanel = $state(false);
	let openAppMenu = $state(false);
	let appMenuButton: HTMLButtonElement | undefined = $state();

	const activeViewNames = $derived.by(() => {
		const names = new Set<string>();
		const subPanels = getAllSubPanels();

		for (const sub of subPanels) {
			if (sub.views) {
				for (const view of sub.views) {
					if (view && view.name) {
						names.add(view.name);
					}
				}
			}
		}
		return names;
	});

	// Context Menu for Theme
	let ctxShowMenu = $state(false);
	let ctxAnchor = $state<{ x: number; y: number } | null>(null);
	let ctxItems = $state<MenuItem[]>([]);

	function handleThemeContext(e: MouseEvent) {
		e.preventDefault();
		ctxAnchor = { x: e.clientX, y: e.clientY };
		ctxItems = [
			{
				id: "theme-default-system",
				label: "System",
				icon: "settings_brightness",
				action: () => themeState.setPreferredTheme("system"),
				disabled: themeState.preferredTheme === "system"
			},
			{
				id: "theme-default-light",
				label: "Light",
				icon: "light_mode",
				action: () => themeState.setPreferredTheme("light"),
				disabled: themeState.preferredTheme === "light"
			},
			{
				id: "theme-default-dark",
				label: "Dark",
				icon: "dark_mode",
				action: () => themeState.setPreferredTheme("dark"),
				disabled: themeState.preferredTheme === "dark"
			}
		];
		ctxShowMenu = true;
	}

	function handleViewMenu(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		const dynamicRouteRegex = /\[.*\].*$/;
		ctxAnchor = { x: e.clientX, y: e.clientY };
		ctxItems = views
			.filter((view) => !view.path || !dynamicRouteRegex.test(view.path))
			.map((view) => ({
				id: view.name,
				label: view.name,
				action: () => {
					let activeId = layoutTree.activeContentId;

					// Fallback to the first available content group if no panel is active
					if (!activeId) {
						const allSubPanels = getAllSubPanels();
						if (allSubPanels.length > 0) {
							activeId = allSubPanels[0].paneKeyId;
							layoutTree.activeContentId = activeId;
						} else {
							toastState.addToast({
								title: "No Panels Available",
								type: "error",
								message: "There are no panels to add the view to."
							});
							return;
						}
					}

					const result = findSubPanel("paneKeyId", activeId as any);
					if (result && result.subPanel) {
						const { parentIndex, childIndex } = result;

						const targetContent =
							layoutState.tree[parentIndex].childs.content[childIndex];
						const existingView = targetContent.views.find(
							(v) => v.name === view.name
						);

						targetContent.views.forEach((v) => (v.isActive = false));

						if (existingView) {
							existingView.isActive = true;
						} else {
							const newView = duplicateView(view);
							newView.parent = activeId;
							newView.isActive = true;
							targetContent.views.push(newView);
						}

						layoutState.tree[parentIndex].views = layoutState.tree[
							parentIndex
						].childs.content.flatMap((c) => c.views);
					}
				},
				disabled: activeViewNames.has(view.name)
			}));
		ctxShowMenu = true;
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if (e.key !== "Escape") {
			return;
		}

		if (searchInputHasFocus) {
			console.log("Escape key pressed, blurring search input");
			searchElement?.blur();
			return;
		}

		if (openAccPanel) {
			openAccPanel = false;
		}

		if (openAppMenu) {
			openAppMenu = false;
		}
	}}
	onclick={(e) => {
		if (
			openAccPanel &&
			!(e.target as HTMLElement).closest("#account-container")
		) {
			openAccPanel = false;
		}
	}}
/>

<header {...props} class="{props.class} no-select">
	<div id="left-menu-container">
		<button
			bind:this={appMenuButton}
			id="viz-title"
			onclick={() => (openAppMenu = !openAppMenu)}
			aria-label="App Menu"
			title="App Menu"
		>
			viz
			<MaterialIcon
				iconName="arrow_drop_down"
				weight={300}
				style="font-size: 1.2em; margin-left: 0.15em;"
			/>
		</button>
		{#if isLayoutPage()}
			<div class="menu-seperator"></div>
			<IconButton
				class="header-button"
				iconName="grid_view"
				title="Views"
				onclick={handleViewMenu}
			/>
		{/if}
		<AppMenu bind:isOpen={openAppMenu} bind:anchor={appMenuButton} />
		<div class="menu-seperator"></div>
		<div class="icon-group-container">
			<a class="page-nav-btn" href="/photos" title="Go to Photos">
				<IconButton class="header-button" iconName="gallery_thumbnail" />
			</a>
			<a class="page-nav-btn" href="/collections" title="Go to Collections">
				<IconButton class="header-button" iconName="photo_library" />
			</a>
		</div>
	</div>
	<div class="center-container">
		<IconButton
			class="header-button"
			iconName="arrow_back"
			disabled={!historyState.canGoBack}
			onclick={() => history.back()}
		/>
		<IconButton
			class="header-button"
			iconName="arrow_forward"
			disabled={!historyState.canGoForward}
			onclick={() => history.forward()}
		/>
		<div class="menu-seperator"></div>
		<SearchInput
			inputId="header-search"
			placeholder="Search (Ctrl/Cmd + K)"
			bind:searchInputHasFocus
			bind:loading={search.loading}
			bind:value={search.value}
			bind:element={searchElement}
			{performSearch}
			style="width: 100%; border-color: var(--imag-80); height: 1.5em; font-size: 0.9em;"
		/>
	</div>
	<div class="header-button-container">
		<IconButton
			weight={300}
			iconName={getTheme() === "dark" ? "dark_mode" : "light_mode"}
			id="theme-toggle"
			class="header-button theme-toggle"
			title="Toggle theme (Right-click to set default)"
			aria-label="Toggle Theme"
			onclick={() => toggleTheme()}
			oncontextmenu={handleThemeContext}
		></IconButton>
		<IconButton
			iconName="upload"
			iconStyle="sharp"
			id="header-upload-button"
			class="header-button"
			aria-label="Upload"
			onclick={handleUpload}
		>
			<span style="font-size: 0.75rem; font-weight: 500;"> Upload </span>
		</IconButton>
		{#if dev || !CLIENT_IS_PRODUCTION}
			<IconButton
				iconName="bug_report"
				id="debug-button"
				class="header-button"
				aria-label="Toggle Debug Mode"
				onclick={() => debugState.toggle()}
				title="Toggle Debug Mode"
			>
				<span class="debug-mode-text">{debugState.value ? "ON" : "OFF"}</span>
			</IconButton>
		{/if}
		<div id="account-container">
			<button
				id="account-button"
				class="header-button"
				aria-label="Account"
				onclick={() => (openAccPanel = !openAccPanel)}
				title={user.data?.username
					? `${user.data.username} (${user.data.email})`
					: "Account"}
			>
				<figure
					style="height: 100%; display: flex; align-items: center; justify-content: center;"
				>
					<span style="font-weight: 700; font-size: 0.8em;"
						>{user.data ? user.data.username[0] : "?"}</span
					>
				</figure>
			</button>
			{#if openAccPanel}
				<OpenAccountPanel bind:openAccPanel />
			{/if}
		</div>
	</div>
</header>

<ContextMenu
	bind:showMenu={ctxShowMenu}
	bind:items={ctxItems}
	anchor={ctxAnchor}
/>

<style lang="scss">
	header {
		background-color: var(--imag-bg-color);
		max-height: 2em;
		padding: 0.1em 0.8em;
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--imag-60);
		position: relative;
		justify-content: space-between;
		flex-direction: row;
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 1em;
		display: flex;
		align-items: center;
		gap: 0.1em;
		background: transparent;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0em 0.5em;
		border-radius: 0.5rem;
		transition: background-color 0.1s ease;

		&:hover {
			background-color: var(--imag-95);
		}

		&:active {
			background-color: var(--imag-90);
		}
	}

	#left-menu-container {
		border-radius: 0.25rem;
		z-index: 300;
		gap: 0.5rem;
		height: 100%;
		display: flex;
		flex-direction: row;
		align-items: center;
	}

	.menu-seperator {
		background-color: var(--imag-60);
		height: 60%;
		width: 1px;
	}

	.icon-group-container {
		gap: 0.5rem;
		display: flex;
		flex-direction: row;
	}

	.page-nav-btn {
		display: flex;
		justify-content: center;
		align-items: center;
		text-decoration: none;
	}

	.center-container {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
		width: 30%;
		height: 100%;
	}

	#account-container {
		position: relative;
	}

	#account-button {
		height: 1.4rem;
		width: 1.4rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10em;
		outline: 1px solid var(--imag-60);
		background-color: var(--imag-80);
	}

	figure {
		display: block;
		margin-block-start: 0em;
		margin-block-end: 0em;
		margin-inline-start: 0px;
		margin-inline-end: 0px;
		unicode-bidi: isolate;
	}

	:global(#header-upload-button) {
		margin: auto 1rem;
		color: var(--imag-text-color);
		font-size: 0.8rem;
		padding: 0.25em 0.5em;
	}

	.header-button-container {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	:global(#theme-toggle) {
		margin: auto 0.5rem;
	}

	:global(.header-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 10em;
		padding: 0.15em 0.4em;
		font-size: 0.8rem;
		color: var(--imag-10);
		cursor: pointer;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
			background-color: var(--imag-90);
			border-radius: 4em;
		}

		&:hover {
			background-color: var(--imag-90);
			// outline: 1px solid var(--imag-60);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	.debug-mode-text {
		margin-right: 0.4em;
		font-family: var(--imag-code-font);
		font-weight: 500;
		font-size: 1em;
	}
</style>
