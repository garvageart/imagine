<script lang="ts">
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import IconButton from "../IconButton.svelte";

	interface Props {
		value: number | null;
		onChange: (rating: number | null) => void;
		updatingRating?: boolean;
	}

	let { value, updatingRating = $bindable(false), onChange }: Props = $props();

	// Rating UI state: previewRating for hover preview, rating is the set value
	let previewRating = $state<number | null>(null);
	let rating = $derived<number | null>(value);

	let starValues = $state<number[]>([1, 2, 3, 4, 5]);

	// Prevent concurrent rating updates
	function handleClick(rating: number) {
		if (value === rating) {
			onChange(null);
		} else {
			onChange(rating);
		}
	}
</script>

<div class="rating-container">
	<div
		class="rating-stars"
		role="group"
		onmouseleave={() => (previewRating = null)}
	>
		{#each starValues as i}
			<button
				class="rating-button"
				title={`Set Rating: ${i}`}
				aria-label={`Set Rating: ${i}`}
				onmouseenter={() => (previewRating = i)}
				onmouseleave={() => (previewRating = null)}
				onclick={() => handleClick(i)}
				disabled={updatingRating}
			>
				<MaterialIcon
					fill={i <= (previewRating ?? rating ?? 0)}
					iconName="star"
					iconStyle={"sharp"}
				/>
			</button>
		{/each}
		{#if rating !== null && rating !== 0}
			<IconButton
				iconName="close"
				weight={600}
				class="rating-clear"
				aria-label="Clear rating"
				onclick={() => handleClick(0)}
				disabled={updatingRating}
			/>
		{/if}
	</div>
</div>

<style lang="scss">
	.rating-container {
		display: flex;
		align-items: center;
		gap: 0.5em;
	}

	.rating-stars {
		display: flex;
		align-items: center;
	}

	.rating-button,
	:global(.rating-clear) {
		border: none;
		background: transparent;
		cursor: pointer;
		color: var(--imag-10) !important;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	:global(.rating-clear) {
		margin: 0em 0.5em;
	}
</style>
