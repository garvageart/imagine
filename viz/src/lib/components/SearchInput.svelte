<script lang="ts">
	import type { SvelteHTMLElements } from "svelte/elements";
	import MaterialIcon from "./MaterialIcon.svelte";

	interface Props {
		inputId: string;
		loading?: boolean;
		value: string;
		element?: HTMLInputElement;
		placeholder?: string;
		searchInputHasFocus?: boolean;
		performSearch?: (e: KeyboardEvent | MouseEvent) => void;
	}

	let {
		loading = $bindable(false),
		value = $bindable(),
		element = $bindable(),
		placeholder = "Search",
		searchInputHasFocus = $bindable(),
		performSearch,
		inputId,
		...props
	}: Props & SvelteHTMLElements["div"] = $props();

	function handleSearch(e: KeyboardEvent) {
		e.key === "Enter" && performSearch?.(e);
	}
</script>

<div class="search-input" class:has-focus={searchInputHasFocus} {...props}>
	<button
		class="search-button"
		aria-label="Search"
		aria-disabled={loading}
		aria-pressed={loading}
		title="Search"
		onclick={performSearch}
		onkeydown={handleSearch}
		disabled={loading}
	>
		<MaterialIcon iconName="search" style="margin: 0 0.2em; color: var(--imag-20);" />
	</button>
	<input
		id={inputId}
		type="search"
		class="search-input__field"
		{placeholder}
		aria-label="Search"
		aria-disabled={loading}
		disabled={loading}
		onkeydown={handleSearch}
		onfocus={() => (searchInputHasFocus = true)}
		onblur={() => (searchInputHasFocus = false)}
		bind:value
		bind:this={element}
	/>
	{#if value}
		<button
			class="clear-search-button"
			type="button"
			aria-label="Clear Search"
			title="Clear Search"
			aria-disabled={loading}
			aria-pressed={loading}
			disabled={loading}
			onclick={() => (value = "")}
		>
			<MaterialIcon iconName="close" />
		</button>
	{/if}
</div>

<style lang="scss">
	.search-input {
		display: flex;
		align-items: center;
		width: 20%;
		border: 1px solid;
		border-color: var(--imag-80);
		border-radius: 2em;
		background-color: var(--imag-bg-color);
		overflow: hidden;
		&:focus {
			border-color: var(--imag-80);
		}
	}

	.search-button {
		background-color: var(--imag-90);
		border: none;
		border-radius: 2em;
		border-top-right-radius: 0%;
		border-bottom-right-radius: 0%;
		height: 100%;
		padding: 0.15em 0.4em;
		font-size: 0.85rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--imag-text-color);
		cursor: pointer;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-60);
			outline: none;
			background-color: var(--imag-80);
		}

		&:hover {
			background-color: var(--imag-80);
		}

		&:active {
			background-color: var(--imag-90);
		}
	}

	.search-input__field {
		font-size: 0.7rem;
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
		outline: none;
		border: none;
		width: 100%;
		height: 100%;
		border-top-left-radius: 0%;
		border-bottom-left-radius: 0%;
		padding: 0 0.5em;
		font-family: var(--imag-font-family);

		&::placeholder {
			color: var(--imag-40);
			font-family: var(--imag-font-family);
		}

		&:focus::placeholder {
			color: var(--imag-text-color);
		}
	}

	.has-focus {
		outline: 1.5px solid var(--imag-primary);
	}

	.clear-search-button {
		border: none;
		outline: none;
		height: 100%;
		padding: 0.15em 0.4em;
		font-size: 1rem;
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
</style>
