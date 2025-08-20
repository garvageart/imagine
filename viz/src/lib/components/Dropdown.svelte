<script lang="ts" module>
	export type DropdownOption<T> = {
		title: string;
		icon?: MaterialSymbol;
		disabled?: boolean;
	};
</script>

<script lang="ts" generics="T">
	import { isEqual } from "lodash-es";
	import { fly } from "svelte/transition";
	import MaterialIcon from "./MaterialIcon.svelte";
	import type { MaterialSymbol } from "material-symbols";

	interface Props {
		class?: string;
		options: DropdownOption<T>[];
		selectedOption?: DropdownOption<T>;
		showMenu?: boolean;
		controlable?: boolean;
		title: string;
		icon?: MaterialSymbol;
		onSelect?: (option: DropdownOption<T>) => void;
	}

	let { options, selectedOption = $bindable(options[0]), showMenu = $bindable(false), title, icon, onSelect }: Props = $props();

	function handleOptionSelect(option: DropdownOption<T>) {
		onSelect?.(option);
		selectedOption = option;
		showMenu = false;
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if (e.key === "Escape") {
			showMenu = false;
		}
	}}
	onclick={(e) => {
		if (!document.querySelector(".viz-dropdown-container")?.contains(e.target as Node)) {
			showMenu = false;
		}
	}}
/>

<div class="viz-dropdown-container">
	<button class="viz-dropdown-button" onclick={() => (showMenu = !showMenu)} {title}>
		{#if selectedOption}
			<span class="viz-dropdown-icon">
				{#if selectedOption.icon}
					<MaterialIcon iconName={selectedOption.icon} />
				{:else if icon}
					<MaterialIcon iconName={icon} />
				{/if}
			</span>
			<p class="viz-dropdown-title">
				{selectedOption.title}
			</p>
		{:else}
			<p class="viz-dropdown-title">
				{title}
			</p>
		{/if}
	</button>

	{#if showMenu}
		<div transition:fly={{ y: -30, duration: 250 }} class="viz-dropdown">
			{#each options as option}
				<button
					class="viz-dropdown-option"
					disabled={option.disabled}
					class:selected-option={isEqual(selectedOption, option)}
					title={option.title}
					onclick={() => !option.disabled && handleOptionSelect(option)}
				>
					{#if isEqual(selectedOption, option)}
						<span class="viz-dropdown-icon">
							<MaterialIcon iconName={option.icon ?? "check"} />
						</span>
						<p class="viz-dropdown-title">
							{option.title}
						</p>
					{:else}
						<span class="viz-dropdown-icon">
							{#if option.icon}
								<MaterialIcon iconName={option.icon} />
							{/if}
						</span>
						<p class="viz-dropdown-title">
							{option.title}
						</p>
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>

<style lang="scss">
	.viz-dropdown-container {
		position: relative;
		display: inline-block;
	}

	.selected-option {
		background-color: var(--imag-90);

		&:hover {
			background-color: var(--imag-80);
		}
	}

	.viz-dropdown-title {
		margin: 0em 0.2rem;
		text-wrap: nowrap;
	}

	.viz-dropdown-button {
		display: flex;
		align-items: center;
		padding: 0.5em 0.5em;
		border-radius: 10em;
		text-wrap: nowrap;
		cursor: pointer;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
			background-color: var(--imag-80);
		}

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}

	.viz-dropdown {
		position: absolute;
		z-index: 1;
		top: 100%;
		left: 0;
		border-radius: 0.5rem;
		background-color: var(--imag-100);
		box-shadow: 0 0.5em 1em rgba(0, 0, 0, 0.1);
		min-width: 10em;
		max-width: 30em;
		max-height: 20em;
		overflow: auto;
	}

	.viz-dropdown button {
		display: flex;
		align-items: center;
		padding: 0.5em 1em;
		width: 100%;
		flex-direction: row;
		justify-content: flex-start;

		&:focus {
			box-shadow: 0px 0px 0px 1.5px inset var(--imag-primary);
			outline: none;
			background-color: var(--imag-80);
		}

		&:hover {
			background-color: var(--imag-90);
		}

		&:active {
			background-color: var(--imag-80);
		}
	}
</style>
