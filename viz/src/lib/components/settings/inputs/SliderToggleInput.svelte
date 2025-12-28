<script lang="ts">
	import SliderToggle from "$lib/components/SliderToggle.svelte";

	interface Props {
		label: string;
		value?: "on" | "off";
		description?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	}

	let { label, value = $bindable("off"), description = "", disabled = false, onchange }: Props = $props();

	const toggleId = $derived(`toggle-${label.replace(/\s+/g, '-').toLowerCase()}`);

	let initialRun = true;
	$effect(() => {
		if (initialRun) {
			initialRun = false;
			return;
		}

		if (onchange) {
			onchange(value);
		}
	});
</script>

<div class="input-container" class:disabled>
	<div class="label-group">
		<label class="label" for={toggleId}>{label}</label>
		{#if description}
			<span class="description">{description}</span>
		{/if}
	</div>
	<div class="toggle-wrapper" class:disabled>
		<!-- 
			We pass the label to SliderToggle because it's required, 
			but we hide it via CSS to use our own label/description layout.
			We treat the 'disabled' state by disabling pointer events if needed,
			though SliderToggle doesn't support 'disabled' prop in original version.
		-->
		<div class:pointer-events-none={disabled} class:opacity-50={disabled}>
			<SliderToggle id={toggleId} {label} bind:value labelPos="side" />
		</div>
	</div>
</div>

<style lang="scss">
	.input-container {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 0;
		border-bottom: 1px solid var(--imag-80);
		width: 100%;

		&.disabled {
			opacity: 0.5;
		}
	}

	.label-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.label {
		font-weight: 500;
		color: var(--imag-text-color);
	}

	.description {
		font-size: 0.875rem;
		color: var(--imag-40);
	}

	.toggle-wrapper {
		:global(.toggle-slider label) {
			display: none;
		}

		:global(.toggle-slider button) {
			margin-left: 0 !important;
		}
	}

	.pointer-events-none {
		pointer-events: none;
	}

	.opacity-50 {
		opacity: 0.5;
	}
</style>
