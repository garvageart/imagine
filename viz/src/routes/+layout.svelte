<script>
	import Header from "$lib/components/Header.svelte";
	import "$lib/styles/scss/main.scss";

	let { children } = $props();
</script>

<Header></Header>

{@render children()}
