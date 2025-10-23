<script lang="ts">
	import { upload } from "$lib/states/index.svelte";
	import { fade } from "svelte/transition";

	for (const file of upload.files) {
		file.upload().then(() => {
			upload.files.splice(upload.files.indexOf(file), 1);
		});
	}
</script>

<div in:fade={{ duration: 250 }} out:fade={{ duration: 250 }} id="viz-upload-panel">
	<div id="viz-upload-panel-header">
		<p>Uploading {upload.files.length} file{upload.files.length === 1 ? "" : "s"}</p>
	</div>
	<div id="viz-upload-panel-list">
		{#each upload.files as file}
			<div class="viz-upload-panel-file-info" data-checksum={file.data.checksum}>
				<div class="viz-upload-panel-file-info-metadata">
					<span>{file.data.filename}</span>
					<span>{file.progress}%</span>
				</div>
				<span class="viz-upload-panel-file-info-progress" style="width: {file.progress}%;"></span>
			</div>
		{/each}
	</div>
</div>

<style>
	#viz-upload-panel {
		width: 20%;
		display: flex;
		flex-direction: column;
		position: absolute;
		bottom: calc(2em);
		right: calc(2em);
		background-color: var(--imag-80);
		z-index: 2;
		border: 1.5px solid var(--imag-60);
		border-radius: 0.5em;
		margin: auto;
	}

	#viz-upload-panel-header {
		height: 3rem;
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.9rem;
	}

	.viz-upload-panel-file-info {
		width: 100%;
		padding: 0.5rem 1rem;
		position: relative;
		border-bottom: 1px solid var(--imag-60);
		overflow: hidden;
	}
</style>
