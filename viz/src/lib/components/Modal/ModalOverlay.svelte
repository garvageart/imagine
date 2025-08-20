<script lang="ts">
	import { lightbox, modal } from "$lib/states/index.svelte";
	import Lightbox from "../Lightbox.svelte";
	import Modal from "./Modal.svelte";

	let { children } = $props();

	if (window.debug) {
		$effect(() => {
			if (modal.show) {
				console.log("modal is showing");
			} else {
				console.log("modal is not showing");
			}
		});
	}

	$effect(() => {
		lightbox.show = modal.show;
	});
</script>

{#if modal.show}
	<Lightbox onclick={() => (modal.show = false)}>
		<Modal>
			{@render children()}
		</Modal>
	</Lightbox>
{/if}
