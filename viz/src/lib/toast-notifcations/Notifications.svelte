<script lang="ts">
	import { fly } from "svelte/transition";
	import { toastState } from "./notif-state.svelte";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";

	function parseNotificationText(text: string) {
		if (!text) {
			return "";
		}

		let safeText = text
			.replace(/&/g, "&amp;")
			.replace(/</g, "&lt;")
			.replace(/>/g, "&gt;")
			.replace(/"/g, "&quot;")
			.replace(/'/g, "&#039;");

		// 1. Bold: **text**
		safeText = safeText.replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>");

		// 2. Italic: *text*
		safeText = safeText.replace(/\*(.*?)\*/g, "<em>$1</em>");

		// 3. Named Links: [Link Text](url)
		safeText = safeText.replace(
			/\[([^\]]+)\]\(((?:https?:\/\/|\/)[^\s)]+)\)/g,
			'<a href="$2" target="_blank" rel="noopener noreferrer" class="viz-toast-link">$1</a>'
		);

		// 4. Raw URLs (that weren't captured by named links): https://example.com
		const urlRegex = /(?<!href=")(https?:\/\/[^\s<]+)/g;
		safeText = safeText.replace(urlRegex, (url) => {
			return `<a href="${url}" target="_blank" rel="noopener noreferrer" class="viz-toast-link">${url}</a>`;
		});

		return safeText;
	}
</script>

<section id="viz-toast-section">
	{#each toastState.toasts as toast (toast.id)}
		<article
			data-toast-id={toast.id}
			class="viz-toast viz-toast-{toast.type || 'info'}"
			role="alert"
			in:fly={{ duration: 250, x: 500, opacity: 0 }}
			out:fly={{ duration: 250, x: 500, opacity: 0 }}
		>
			<div class="viz-toast-content-wrapper">
				{#if toast.title}
					<div class="viz-toast-title">{toast.title}</div>
				{/if}

				<div class="viz-toast-message">
					{@html parseNotificationText(toast.message)}
				</div>

				{#if toast.actions && toast.actions.length > 0}
					<div class="viz-toast-actions">
						{#each toast.actions as action}
							<button class="viz-toast-action-btn" onclick={action.onClick}>
								{action.label}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			{#if toast.dismissible}
				<div class="viz-toast-close-wrapper">
					<Button
						class="viz-toast-close"
						title="Dismiss"
						aria-label="Dismiss notification"
						style="padding: 0.1em;"
						hoverColor="var(--imag-40-light)"
						onclick={() => toastState.dismissToast(toast.id)}
					>
						<MaterialIcon iconName="close" style="font-size: 1.2em;" />
					</Button>
				</div>
			{/if}
		</article>
	{/each}
</section>

<style lang="scss">
	@use "sass:color";
	@use "../styles/scss/variables" as v;

	$error: v.$imag-error-color;
	$success: v.$imag-success-color;
	$warning: v.$imag-warning-color;

	@mixin toast-btn-variant($color) {
		background: color.mix($color, #fff, 50%);
		border-color: color.mix($color, #fff, 40%);
		
		&:hover {
			background: color.mix($color, #fff, 60%);
		}
	}

	@mixin toast-variant($color) {
		background-color: color.mix($color, #fff, 30%);
		border-color: color.mix($color, #fff, 90%);
		color: var(--imag-10-light);
		--toast-border: color.mix($color, #fff, 90%);
	}

	#viz-toast-section {
		position: fixed;
		right: 2em;
		bottom: 2em;
		width: 350px;
		max-width: 90vw;
		display: flex;
		justify-content: flex-end;
		flex-direction: column;
		align-items: flex-end;
		z-index: 99999;
		pointer-events: none;
	}

	.viz-toast {
		border-radius: 0.5em;
		min-height: 3.5em;
		width: 100%;
		padding: 0.75em 1em;
		margin-top: 0.5em;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		box-shadow:
			0 4px 6px rgba(0, 0, 0, 0.1),
			0 2px 4px rgba(0, 0, 0, 0.06);
		pointer-events: auto;
		border: 1px solid transparent;

		/* Default / Info */
		background-color: var(--imag-90);
		color: var(--imag-text-color);
		border-color: var(--imag-80);
		--toast-border: var(--imag-80);

		&.viz-toast-info {
			background-color: var(--imag-90);
			color: var(--imag-text-color);
			border-color: var(--imag-80);
			--toast-border: var(--imag-80);
		}

		&.viz-toast-success {
			@include toast-variant($success);
		}

		&.viz-toast-warning {
			@include toast-variant($warning);
		}

		&.viz-toast-error {
			@include toast-variant($error);
		}
	}

	.viz-toast-content-wrapper {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.3em;
		margin-right: 0.5em;
		overflow-wrap: break-word;
		word-break: break-word;
	}

	.viz-toast-title {
		font-weight: 700;
		font-size: 0.95em;
		margin-bottom: 0.1em;
		line-height: 1.2;
	}

	.viz-toast-message {
		font-size: 0.9em;
		line-height: 1.4;
	}

	/* Deep selector for links injected via @html */
	.viz-toast-message :global(a.viz-toast-link) {
		color: inherit;
		text-decoration: underline;
		font-weight: 500;
	}

	.viz-toast-message :global(strong) {
		font-weight: 700;
	}

	.viz-toast-actions {
		display: flex;
		gap: 0.5em;
		margin-top: 0.5em;
		flex-wrap: wrap;
	}

	.viz-toast-action-btn {
		background: rgba(255, 255, 255, 0.2);
		border: 1px solid rgba(255, 255, 255, 0.3);
		color: inherit;
		padding: 0.25em 0.6em;
		border-radius: 0.25rem;
		font-size: 0.85rem;
		cursor: pointer;
		font-weight: 500;
		transition: background-color 0.2s;

		.viz-toast-success & {
			@include toast-btn-variant($success);
		}

		.viz-toast-warning & {
			@include toast-btn-variant($warning);
		}

		.viz-toast-error & {
			@include toast-btn-variant($error);
		}
	}

	.viz-toast-action-btn:hover {
		background: rgba(255, 255, 255, 0.3);
	}

	:global(.viz-toast-close) {
		background-color: var(--imag-30-light);
		outline: var(--imag-50-light) solid 1px;
	}

	.viz-toast-close-wrapper {
		margin-left: auto;
		margin-top: -0.25em;
		margin-right: -0.5em;
	}
</style>
