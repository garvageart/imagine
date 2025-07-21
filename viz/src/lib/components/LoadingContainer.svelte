<script lang="ts">
	import LoadingSpinner from "./LoadingSpinner.svelte";
	let { color }: { color?: string } = $props();
</script>

<div style="height: inherit; width: 100%; display: flex; justify-content: center; align-items: center;">
	<div style="height: 2em; width: 2em">
		<LoadingSpinner {color} />
	</div>
</div>
