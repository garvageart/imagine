<script lang="ts">
	import Button from "$lib/components/Button.svelte";
	import ModalContainer from "./ModalContainer.svelte";
	import type { Snippet } from "svelte";

	interface Props {
		title: string;
		children?: Snippet;
		actions?: Snippet;
		confirmText?: string;
		cancelText?: string;
		onConfirm?: () => void;
		onCancel?: () => void;
	}

	let { title, children, actions, confirmText = "Confirm", cancelText = "Cancel", onConfirm, onCancel }: Props = $props();
</script>

<ModalContainer>
	<div class="confirmation-modal">
		<h2>{title}</h2>
		<div class="confirmation-content">
			{#if children}
				{@render children()}
			{/if}
		</div>

		<div class="confirm-actions">
			{#if actions}
				{@render actions()}
			{:else}
				<Button onclick={onCancel}>{cancelText}</Button>
				<Button
					onclick={onConfirm}
					style="background-color: var(--imag-primary); color: var(--imag-10);"
					hoverColor="var(--imag-100)"
				>
					{confirmText}
				</Button>
			{/if}
		</div>
	</div>
</ModalContainer>

<style lang="scss">
	.confirmation-modal {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		width: 100%;
		color: var(--imag-text-color);

		h2 {
			margin: 0;
			font-size: 1.5rem;
		}

		.confirmation-content {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;
		}

		.confirm-actions {
			display: flex;
			gap: 1rem;
			justify-content: flex-end;
			margin-top: 0.5rem;
		}
	}
</style>
