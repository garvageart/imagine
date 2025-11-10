<script lang="ts">
	import { upload } from "$lib/states/index.svelte";
	import { fade, scale } from "svelte/transition";
	import { UploadState } from "$lib/upload/asset.svelte";
	import Button from "./Button.svelte";
	import MaterialIcon from "./MaterialIcon.svelte";

	let minimized = $state(false);
</script>

{#if minimized}
	<div id="viz-upload-panel-minimized" in:scale={{ duration: 250 }} out:fade={{ duration: 250 }}>
		<Button
			id="viz-upload-panel-minimized-button"
			onclick={() => {
				minimized = false;
			}}
			title="Show Upload Panel"
			style="background-color: var(--imag-secondary);"
			hoverColor="var(--imag-primary)"
		>
			<MaterialIcon iconName="upload" style="font-size: 1.5rem;" />
			<span>{upload.files.length} uploading file{upload.files.length === 1 ? "" : "s"}</span>
		</Button>
	</div>
{:else}
	<div in:scale={{ duration: 250 }} out:scale={{ delay: minimized ? 0 : 3000, duration: 250 }} id="viz-upload-panel">
		<div id="viz-upload-panel-header">
			<div id="upload-panel-header-info">
				<Button
					style="background-color: transparent; padding: 0em;"
					hoverColor="var(--imag-80)"
					title="Minimize Upload Panel"
					onclick={() => {
						minimized = true;
					}}
				>
					<MaterialIcon iconName="arrow_downward_alt" />
				</Button>
				<p>Uploading {upload.files.length} file{upload.files.length === 1 ? "" : "s"}</p>
			</div>
			<div class="concurrency-control">
				<label for="concurrency-input" title="Maximum simultaneous uploads">
					Concurrent:
					<input
						id="concurrency-input"
						type="number"
						min="1"
						max="10"
						bind:value={upload.concurrency}
						style="width: 3em; margin-left: 0.25em;"
					/>
				</label>
			</div>
		</div>
		<div id="viz-upload-panel-list">
			{#each upload.files as file}
				<div class="panel-file-info" data-checksum={file.data.checksum}>
					{#if file.state === UploadState.STARTED}
						<Button
							style="background-color: transparent; padding: 0em;"
							hoverColor="var(--imag-80)"
							title="Cancel Upload"
							onclick={() => {
								file.cancelRequest();
							}}
						>
							<MaterialIcon iconName="close" />
						</Button>
					{/if}
					<div class="panel-file-info-data_container">
						<div class="panel-file-info-metadata">
							<span class="viz-upload-file-name">{file.data.filename}</span>
							<span class="viz-upload-progress-text">{Math.round(file.progress)}%</span>
						</div>
						<div class="panel-file-info-progress-container">
							<span
								class="panel-file-info-progress"
								class:complete={file.state === UploadState.DONE}
								class:error={file.state === UploadState.ERROR || file.state === UploadState.CANCELED}
								style="width: {file.progress}%;"
							>
							</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
{/if}

<style lang="scss">
	#viz-upload-panel {
		min-width: 25%;
		max-width: 30%;
		display: flex;
		flex-direction: column;
		position: absolute;
		bottom: 2em;
		left: 2em;
		background-color: var(--imag-100);
		z-index: 2;
		border: 1.5px solid var(--imag-60);
		border-radius: 0.5em;
		max-height: 60vh;
		overflow: hidden;
	}

	#viz-upload-panel-minimized {
		position: absolute;
		bottom: 2em;
		left: 2em;
		z-index: 2;
	}

	#viz-upload-panel-header {
		height: 3rem;
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 1rem;
		font-size: 0.9rem;
		font-weight: 600;
		border-bottom: 1px solid var(--imag-60);
		background-color: var(--imag-90);
		box-sizing: border-box;
		gap: 1rem;
	}

	#upload-panel-header-info {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
	}

	.concurrency-control {
		display: flex;
		align-items: center;
		font-size: 0.75rem;
		font-weight: 500;

		label {
			display: flex;
			align-items: center;
			cursor: pointer;
		}

		input[type="number"] {
			background-color: var(--imag-80);
			border: 1px solid var(--imag-60);
			border-radius: 0.25em;
			color: var(--imag-text-color);
			padding: 0.25em;
			text-align: center;
			font-family: var(--imag-code-font);

			&:focus {
				outline: none;
				border-color: var(--imag-40);
			}
		}
	}

	#viz-upload-panel-list {
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		overflow-y: auto;
		font-family: var(--imag-code-font);
	}

	.panel-file-info {
		display: flex;
		flex-direction: row;
		position: relative;
		width: 100%;
		padding: 0.5rem 1rem;
		border-bottom: 1px solid var(--imag-70);
		background-color: var(--imag-100);
		box-sizing: border-box;

		&:hover {
			background-color: var(--imag-90);
		}
	}

	.panel-file-info-data_container {
		display: flex;
		flex-direction: column;
		margin-left: 0.5rem;
		width: 100%;
	}

	.panel-file-info:last-child {
		border-bottom: none;
		margin-bottom: 0;
	}

	.panel-file-info-metadata {
		font-size: 0.8rem;
		margin-bottom: 0.5rem;
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.viz-upload-file-name {
		flex: 1;
		width: 95%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-weight: 500;
	}

	.viz-upload-progress-text {
		font-weight: 600;
		font-size: 0.75rem;
		color: var(--imag-20);
		min-width: 40px;
		text-align: right;
	}

	.panel-file-info-progress-container {
		width: 100%;
		height: 4px;
		background-color: var(--imag-70);
		border-radius: 2px;
		overflow: hidden;
		position: relative;
	}

	.panel-file-info-progress {
		height: 100%;
		background: linear-gradient(90deg, var(--imag-30), var(--imag-20));
		border-radius: 2px;
		transition: width 0.3s ease;
		display: block;
	}

	.panel-file-info-progress.complete {
		background: linear-gradient(90deg, #10b981, #059669);
	}

	.panel-file-info-progress.error {
		background: linear-gradient(90deg, #ef4444, #dc2626);
	}
</style>
