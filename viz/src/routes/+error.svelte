<script lang="ts">
	import { page } from "$app/state";
	import Button from "$lib/components/Button.svelte";
	import MaterialIcon from "$lib/components/MaterialIcon.svelte";
	import { goto } from "$app/navigation";

	let statusCode = $derived(page.status);
	let errorMessage = $derived(page.error?.message ?? "An unexpected error occurred.");
	let stackTrace = $derived((page.error as any)?.stack);
</script>

<div class="error-container">
	<div class="error-card">
		<div class="icon-wrapper">
			{#if statusCode === 404}
				<MaterialIcon iconName="search_off" style="font-size: 4rem; color: var(--imag-20);" />
			{:else if statusCode === 403 || statusCode === 401}
				<MaterialIcon iconName="lock" style="font-size: 4rem; color: var(--imag-20);" />
			{:else}
				<MaterialIcon iconName="error_med" style="font-size: 4rem; color: var(--imag-20);" />
			{/if}
		</div>

		<h1 class="status-code">{statusCode}</h1>

		<div class="message-container">
			<h2 class="error-title">
				{#if statusCode === 404}
					Not Found
				{:else if statusCode === 403}
					Access Denied
				{:else if statusCode === 401}
					Unauthorized
				{:else if statusCode === 500}
					Server Error
				{:else}
					Something Went Wrong
				{/if}
			</h2>
			<p class="error-message">{errorMessage}</p>
		</div>

		{#if stackTrace}
			<div class="trace-container">
				<details>
					<summary>Error Details</summary>
					<pre class="trace-content">{stackTrace}</pre>
				</details>
			</div>
		{/if}

		<div class="actions">
			<Button class="actions-button" onclick={() => goto("/")}>
				<MaterialIcon iconName="home" />
				Go Home
			</Button>

			<Button class="actions-button" onclick={() => history.back()}>
				<MaterialIcon iconName="arrow_back" />
				Go Back
			</Button>
		</div>
	</div>
</div>

<style lang="scss">
	.error-container {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100vh;
		width: 100vw;
		background-color: var(--imag-bg-color);
		color: var(--imag-text-color);
	}

	.error-card {
		background-color: var(--imag-100);
		padding: 3rem;
		border-radius: 12px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
		text-align: center;
		max-width: 500px;
		width: 90%;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		border: 1px solid var(--imag-60);
	}

	.icon-wrapper {
		display: flex;
		justify-content: center;
		margin-bottom: 1rem;
	}

	.status-code {
		font-size: 4rem;
		font-weight: 700;
		color: var(--imag-text-color);
		line-height: 1;
		font-family: var(--imag-code-font);
	}

	.message-container {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.error-title {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
	}

	.error-message {
		color: var(--imag-20);
		font-size: 1rem;
		margin: 0;
		line-height: 1.5;
	}

	.trace-container {
		text-align: left;
		width: 100%;
		margin-top: 1rem;

		details {
			background-color: var(--imag-bg-color);
			border-radius: 0.5em;
			padding: 0.5rem;
			border: 1px solid var(--imag-60);

			summary {
				cursor: pointer;
				font-weight: 500;
				color: var(--imag-20);
				padding: 0.25rem;
				user-select: none;

				&:hover {
					color: var(--imag-text-color);
				}
			}
		}

		.trace-content {
			margin-top: 0.5rem;
			padding: 0.5rem;
			overflow-x: auto;
			font-size: 0.8rem;
			color: var(--imag-text-color);
			background-color: var(--imag-100);
			border-radius: 4px;
			font-family: var(--imag-code-font);
			white-space: pre-wrap;
			word-break: break-word;
			max-height: 12rem;
			overflow-y: auto;
		}
	}

	.actions {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-top: 1rem;
		flex-wrap: wrap;
	}

	:global(.actions-button) {
		background-color: var(--imag-100);
		color: var(--imag-text-color);
		padding: 0.75em 1.5em;
		display: flex;
		align-items: center;
		gap: 0.5em;

		&:hover {
			background-color: var(--imag-90);
		}
	}
</style>
