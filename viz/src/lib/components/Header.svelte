<script lang="ts">
	import { dev } from "$app/environment";
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
	import { search } from "$lib/states/index.svelte";
	import { VizLocalStorage } from "$lib/utils/misc";
	import { onMount } from "svelte";
	import { performSearch } from "$lib/search/execute";
	import MaterialIcon from "./MaterialIcon.svelte";
	import { page } from "$app/state";
	import type { SvelteHTMLElements } from "svelte/elements";

	let { ...props }: SvelteHTMLElements["header"] = $props();

	// eventually this will move to a different page with a different way of enabling, this is just temporary
	const storeDebug = new VizLocalStorage<boolean>("debugMode");
	let devEnabled = $state(storeDebug.get() === null ? false : storeDebug.get());

	// i'd probably forget to remove this in prod setting so just check lmao
	if (dev || !CLIENT_IS_PRODUCTION) {
		$effect(() => {
			console.log("Debug mode is", devEnabled ? "enabled" : "disabled");
			storeDebug.set(devEnabled!);
			if (window.debug !== devEnabled) {
				window.debug === devEnabled;
			}
		});
	}

	onMount(() => {
		if (page.url.pathname !== "/search") {
			return;
		}

		// If URL has search params, perform search automatically
		const urlParams = new URLSearchParams(window.location.search);
		const q = urlParams.get("q");
		if (q) {
			search.value = q;
		}
	});

	let searchInputHasFocus = $state(false);
</script>

<header {...props} class="{props.class} no-select">
	<a id="viz-title" href="/">viz</a>
	<div class="search-container{searchInputHasFocus ? ' has-focus' : ''}">
		<button
			id="search-button"
			type="button"
			aria-label="Search"
			aria-disabled={search.loading}
			aria-pressed={search.loading}
			title="Search"
			onclick={performSearch}
			onkeydown={(e) => e.key === "Enter" && performSearch()}
			disabled={search.loading}
		>
			<MaterialIcon iconName="search" />
		</button>
		<input
			type="search"
			class="search-input"
			placeholder="Search"
			aria-label="Search"
			aria-disabled={search.loading}
			disabled={search.loading}
			onkeydown={(e) => e.key === "Enter" && performSearch()}
			onfocus={() => (searchInputHasFocus = true)}
			onblur={() => (searchInputHasFocus = false)}
			bind:value={search.value}
		/>
		{#if search.value}
			<button
				id="clear-search-button"
				type="button"
				aria-label="Clear Search"
				title="Clear Search"
				aria-disabled={search.loading}
				aria-pressed={search.loading}
				disabled={search.loading}
				onclick={() => (search.value = "")}
			>
				<MaterialIcon iconName="close" />
			</button>
		{/if}
	</div>
	<div class="header-button-container">
		{#if dev || !CLIENT_IS_PRODUCTION}
			<button type="button" class="header-button" aria-label="Toggle Debug Mode" onclick={() => (devEnabled = !devEnabled)}>
				{#if devEnabled}
					<span class="debug-mode-text">ON</span>
				{:else}
					<span class="debug-mode-text">OFF</span>
				{/if}
				<MaterialIcon iconName="bug_report" showHoverBG />
			</button>
		{/if}
	</div>
</header>

<style lang="scss">
	header {
		background-color: var(--imag-bg-color);
		max-height: 2.5em;
		padding: 0.7em 1em;
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--imag-60);
		position: relative;
		justify-content: center;
		flex-direction: row;
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 1.2em;
		position: absolute;
		left: 1em;
	}

	#search-button {
		background-color: var(--imag-100);
		border: none;
		outline: none;
		height: 100%;
		padding: 0.2em 0.5em;
		font-size: 1.2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-text-color);
		cursor: pointer;

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	.search-container {
		display: flex;
		align-items: center;
		width: 30%;
		height: 2em;
		border: 1px solid;
		border-color: var(--imag-90);
		border-radius: 2em;
		overflow: hidden;
		&:focus {
			border-color: var(--imag-80);
			height: 1.3em;
		}
	}

	.search-input {
		font-size: 0.9em;
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
		outline: none;
		border: none;
		width: 100%;
		height: 100%;
		border-radius: 2em;
		border-top-left-radius: 0%;
		border-bottom-left-radius: 0%;
		padding: 0 0.7em;
		font-family: var(--imag-font-family);

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-text-color);
		}
	}

	.search-container.has-focus {
		border: 1.5px solid var(--imag-primary);
	}

	#clear-search-button {
		border: none;
		outline: none;
		height: 100%;
		padding: 0.2em 0.5em;
		font-size: 1.2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-80);
		cursor: pointer;

		&:hover {
			color: var(--imag-70);
		}

		&:active {
			color: var(--imag-60);
		}
	}

	.header-button-container {
		position: absolute;
		right: 1em;
		display: flex;
		align-items: center;
	}

	:global(.header-button) {
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--imag-100);
		border: none;
		border-radius: 4em;
		outline: none;
		padding: 0.2em 0.5em;
		font-size: 1rem;
		color: var(--imag-text-color);
		margin-right: 1em;
		cursor: pointer;

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}

		.debug-mode-text {
			margin-right: 0.5em;
			font-family: var(--imag-code-font);
			font-weight: bold;
			font-size: 0.9em;
			color: var(--imag-text-color);
		}
	}
</style>
