<script lang="ts">
	import { modal } from "$lib/states/index.svelte";
	import ModalContainer from "./ModalContainer.svelte";
	import FilterPanel from "../panels/workspace/FilterPanel.svelte";
</script>

{#if modal.show}
	<ModalContainer>
		<div class="filter-modal-content">
			<FilterPanel />
		</div>
	</ModalContainer>
{/if}

<style lang="scss">
	.filter-modal-content {
		width: 100%;
		height: 80vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	:global(#viz-modal) {
		max-height: 90% !important;
		padding: 0 !important;
	}
</style>
