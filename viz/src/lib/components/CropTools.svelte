<script lang="ts">
	import IconButton from "./IconButton.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";

	interface Props {
		onApply: () => void;
		onCancel: () => void;
		onReset: () => void;
		onAspectRatioChange: (ratio: number | null) => void;
		x?: number;
		y?: number;
		variant?: "floating" | "placed";
	}

	let {
		onApply,
		onCancel,
		onReset,
		onAspectRatioChange,
		x = 0,
		y = 0,
		variant = "floating"
	}: Props = $props();

	let selectedRatio = $state<number | null>(null);

	const ratios = [
		{ label: "Free", value: null },
		{ label: "1:1", value: 1 },
		{ label: "4:5", value: 4 / 5 },
		{ label: "16:9", value: 16 / 9 },
		{ label: "5:4", value: 5 / 4 },
		{ label: "2:3", value: 2 / 3 }
	];

	function selectRatio(value: number | null) {
		selectedRatio = value;
		onAspectRatioChange(value);
	}

	let style = $derived(
		variant === "floating" ? `left: ${x}px; top: ${y}px;` : ""
	);
</script>

<div
	class="crop-tools-menu {variant}"
	{style}
	onclick={(e) => e.stopPropagation()}
	onkeydown={(e) => e.stopPropagation()}
	role="menu"
	tabindex="0"
>
	<div class="menu-header">Crop Tools</div>

	<div class="crop-presets">
		{#each ratios as ratio}
			<button
				class="preset-btn {selectedRatio === ratio.value ? 'active' : ''}"
				onclick={() => selectRatio(ratio.value)}
			>
				{ratio.label}
			</button>
		{/each}
	</div>

	<div class="crop-actions">
		<IconButton
			iconName="restart_alt"
			class="action-btn reset"
			onclick={onReset}
			title="Reset Crop"
		>
			<span>Reset</span>
		</IconButton>
		<IconButton
			iconName="close"
			class="action-btn cancel"
			onclick={onCancel}
			title="Cancel Crop"
		>
			<span>Cancel</span>
		</IconButton>
		<IconButton
			iconName="check"
			class="action-btn apply"
			onclick={onApply}
			title="Apply Crop"
		>
			<span>Apply</span>
		</IconButton>
	</div>
</div>

<style lang="scss">
	.crop-tools-menu {
		display: flex;
		flex-direction: column;
		gap: 0.8em;
		background: var(--imag-bg-color);
		padding: 0.8em;
		border-radius: 0.5em;
		z-index: 9999;
		pointer-events: auto;
		min-width: 200px;
		box-sizing: border-box;

		&.floating {
			position: fixed;
			box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
			border: 1px solid var(--imag-10);
			/* Basic collision avoidance: translate up/left if near edges? */
			transform: translate(0, 0);
		}

		&.placed {
			/* Sidebar styling to match metadata editor look */
			height: auto;
			width: 100%;
			background: var(--imag-bg-color);
			/* If parent doesn't constrain width, we might want a max-width, but parent likely handles it */
		}
	}

	.menu-header {
		font-size: 0.85em;
		font-weight: 600;
		color: var(--imag-30);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.2em;
	}

	.crop-presets {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr;
		gap: 0.4em;
	}

	.preset-btn {
		background: var(--imag-90);
		border: 1px solid transparent;
		color: var(--imag-text-color);
		padding: 0.4em;
		border-radius: 0.3em;
		cursor: pointer;
		font-size: 0.85em;
		text-align: center;
		transition: all 0.2s;

		&:hover {
			background: var(--imag-80);
		}

		&.active {
			background: var(--imag-primary);
			color: white;
			border-color: var(--imag-primary);
		}
	}

	.crop-actions {
		display: flex;
		justify-content: space-between;
		gap: 0.5em;
		border-top: 1px solid var(--imag-10);
		padding-top: 0.8em;
		margin-top: 0.2em;

		:global(.action-btn) {
			display: flex;
			align-items: center;
			justify-content: center;
			font-weight: 500;
			// padding: 0.5em 0.8em;
			font-size: 0.9em;
			flex: 1;
			transition: background-color 0.2s;
		}

		:global(.action-btn.apply) {
			color: #4caf50;
			background-color: rgba(76, 175, 80, 0.1);
		}
		:global(.action-btn.apply:hover) {
			background-color: rgba(76, 175, 80, 0.2);
		}

		:global(.action-btn.cancel) {
			color: #f44336;
			background-color: rgba(244, 67, 54, 0.1);
		}
		:global(.action-btn.cancel:hover) {
			background-color: rgba(244, 67, 54, 0.2);
		}

		:global(.action-btn.reset) {
			color: var(--imag-text-color);
			background-color: var(--imag-80);
		}

		:global(.action-btn.reset:hover) {
			background-color: var(--imag-70);
		}
	}
</style>
