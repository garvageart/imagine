<script lang="ts">
	import { fly } from "svelte/transition";
	import { toastState } from "./notif-state.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";

	const vizToastColours = {
		info: "var(--imag-90)",
		success: "#38a169",
		warning: "#dd6b20",
		error: "#e53e3e"
	};

	function convertTextURLsToHref(text: string) {
		const urlRegex = /https?:\/\/[^\s]+/g;

		return text.replaceAll(urlRegex, (url) => {
			return `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`;
		});
	}
</script>

<section id="viz-toast-section">
	{#each toastState.toasts as toast (toast.id)}
		<article
			data-toast-id={toast.id}
			class="viz-toast"
			style="background-color: {toast.type ? vizToastColours[toast.type] : vizToastColours.info};"
			role="alert"
			in:fly={{ duration: 250, x: 500, opacity: 0 }}
			out:fly={{ duration: 250, x: 500, opacity: 0 }}
		>
			<div class="viz-toast-message">
				{@html convertTextURLsToHref(toast.message)}
			</div>

			{#if toast.dismissible}
				<Button
					class="viz-toast-close"
					title="Dismiss"
					aria-label="Dismiss notification"
					style="background-color: transparent;"
					hoverColor={toast.type ? vizToastColours[toast.type] : vizToastColours.info}
					onclick={() => toastState.dismissToast(toast.id)}
				>
					<MaterialIcon iconName="close" />
				</Button>
			{/if}
		</article>
	{/each}
</section>

<style>
	#viz-toast-section {
		position: fixed;
		right: 2em;
		bottom: 1em;
		width: 20%;
		display: flex;
		justify-content: center;
		flex-direction: column;
		align-items: center;
		z-index: 99999;
	}

	.viz-toast {
		background-color: var(--imag-90);
		color: var(--imag-text-color);
		border-radius: 0.5em;
		height: 4em;
		width: 100%;
		padding: 0.5em 1em;
		margin: 0.2em;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.viz-toast-message {
		font-size: 0.9em;
	}
</style>
