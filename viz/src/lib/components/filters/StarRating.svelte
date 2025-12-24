<script lang="ts">
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import IconButton from "../IconButton.svelte";

	interface Props {
		value: number | null;
		onChange: (rating: number | null) => void;
	}

	let { value, onChange }: Props = $props();

	function handleClick(rating: number) {
		if (value === rating) {
			onChange(null);
		} else {
			onChange(rating);
		}
	}
</script>

<div class="star-rating">
	{#each Array(5) as _, i}
		{@const rating = i + 1}
		<IconButton
			iconName="star"
			fill={value !== null && rating <= value}
			weight={400}
			class="star-icon {value !== null && rating <= value ? 'active' : ''}"
			onclick={() => handleClick(rating)}
			aria-label="Rate {rating} stars"
		></IconButton>
	{/each}
</div>

<style>
	.star-rating {
		display: flex;
		gap: 2px;
	}

	.star-rating :global(.star-icon:hover),
	.star-rating :global(.star-icon.active) {
		color: var(--imag-60);
	}
</style>
